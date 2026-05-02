package render

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// TextGlyph is one positioned glyph from a laid-out text run.
type TextGlyph struct {
	Rune       rune
	Face       FontFace
	GlyphIndex sfnt.GlyphIndex
	Origin     geom.Pt
	Advance    float64
	Kern       float64
}

// TextGlyphLayout is a renderer-neutral, glyph-positioned single-line text
// layout. Size is interpreted as pixels-per-em by this low-level helper.
type TextGlyphLayout struct {
	Glyphs  []TextGlyph
	Advance float64
	Bounds  TextBounds
}

// LayoutTextGlyphs resolves text to font runs and applies pair kerning with
// the same ppem used for glyph advances. It intentionally avoids
// opentype.Face.Kern because that API has historically exposed adjustments in
// font-unit space rather than active-size pixel space.
func LayoutTextGlyphs(text string, origin geom.Pt, size float64, fontKey string) (TextGlyphLayout, bool) {
	if text == "" || size <= 0 {
		return TextGlyphLayout{}, false
	}

	runs, ok := DefaultFontManager().ResolveTextRuns(text, fontKey)
	if !ok {
		return TextGlyphLayout{}, false
	}
	return LayoutTextGlyphRuns(runs, origin, size)
}

// LayoutTextGlyphRuns lays out already-resolved font runs.
func LayoutTextGlyphRuns(runs []FontRun, origin geom.Pt, size float64) (TextGlyphLayout, bool) {
	if len(runs) == 0 || size <= 0 {
		return TextGlyphLayout{}, false
	}

	ppem := fixed.Int26_6(math.Round(size * 64))
	if ppem <= 0 {
		return TextGlyphLayout{}, false
	}

	var (
		layout        TextGlyphLayout
		penX          = origin.X
		haveBounds    bool
		minX, minY    float64
		maxX, maxY    float64
		previous      sfnt.GlyphIndex
		previousFace  string
		havePrevious  bool
		laidOutGlyphs bool
	)

	for _, run := range runs {
		if run.Text == "" || run.Face.Path == "" {
			continue
		}
		fontData, err := loadTextPathFont(run.Face.Path)
		if err != nil {
			return TextGlyphLayout{}, false
		}

		var buf sfnt.Buffer
		for _, r := range run.Text {
			glyphIndex, err := fontData.GlyphIndex(&buf, r)
			if err != nil {
				return TextGlyphLayout{}, false
			}
			if glyphIndex == 0 {
				havePrevious = false
				previousFace = ""
				continue
			}

			kern := 0.0
			if havePrevious && previousFace == run.Face.Path {
				if k, err := fontData.Kern(&buf, previous, glyphIndex, ppem, font.HintingNone); err == nil {
					kern = fixedToFloat(k)
					penX += kern
				}
			}

			originPt := geom.Pt{X: penX, Y: origin.Y}
			segments, err := fontData.LoadGlyph(&buf, glyphIndex, ppem, nil)
			if err != nil {
				return TextGlyphLayout{}, false
			}
			if len(segments) > 0 {
				bounds := segments.Bounds()
				x0 := originPt.X + fixedToFloat(bounds.Min.X)
				y0 := originPt.Y + fixedToFloat(bounds.Min.Y)
				x1 := originPt.X + fixedToFloat(bounds.Max.X)
				y1 := originPt.Y + fixedToFloat(bounds.Max.Y)
				if !haveBounds {
					minX, minY, maxX, maxY = x0, y0, x1, y1
					haveBounds = true
				} else {
					minX = math.Min(minX, x0)
					minY = math.Min(minY, y0)
					maxX = math.Max(maxX, x1)
					maxY = math.Max(maxY, y1)
				}
			}

			advance, err := fontData.GlyphAdvance(&buf, glyphIndex, ppem, font.HintingNone)
			if err != nil {
				return TextGlyphLayout{}, false
			}
			advanceFloat := fixedToFloat(advance)
			layout.Glyphs = append(layout.Glyphs, TextGlyph{
				Rune:       r,
				Face:       run.Face,
				GlyphIndex: glyphIndex,
				Origin:     originPt,
				Advance:    advanceFloat,
				Kern:       kern,
			})
			penX += advanceFloat
			previous = glyphIndex
			previousFace = run.Face.Path
			havePrevious = true
			laidOutGlyphs = true
		}
	}

	layout.Advance = penX - origin.X
	if haveBounds {
		layout.Bounds = TextBounds{X: minX - origin.X, Y: minY - origin.Y, W: maxX - minX, H: maxY - minY}
	}
	return layout, laidOutGlyphs || layout.Advance > 0
}
