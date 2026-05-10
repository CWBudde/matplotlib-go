package render

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// TextDirection describes the intended logical flow of shaped text.
type TextDirection string

const (
	TextDirectionLTR TextDirection = "ltr"
	TextDirectionRTL TextDirection = "rtl"
)

// TextFeature describes one OpenType feature toggle requested for shaping.
// The current fallback shaper records the request surface but does not yet
// apply GSUB/GPOS features.
type TextFeature struct {
	Tag   string
	Value int
}

// TextShapingOptions carries font and script context for text shaping.
type TextShapingOptions struct {
	FontKey   string
	Direction TextDirection
	Script    string
	Language  string
	Features  []TextFeature
}

// ShapedGlyph is one glyph positioned by the shared text shaping layer.
type ShapedGlyph struct {
	Rune       rune
	Cluster    int
	Face       FontFace
	GlyphIndex sfnt.GlyphIndex
	Origin     geom.Pt
	Offset     geom.Pt
	Advance    geom.Pt
	Kern       float64
}

// ShapedRun is a contiguous shaped glyph run using one resolved font face.
type ShapedRun struct {
	Face      FontFace
	Direction TextDirection
	Script    string
	Language  string
	Features  []TextFeature
	Glyphs    []ShapedGlyph
	Advance   geom.Pt
	Bounds    TextBounds
}

// ShapedText is a renderer-independent shaped single-line text layout. Size is
// interpreted as pixels-per-em by this low-level helper.
type ShapedText struct {
	Runs    []ShapedRun
	Glyphs  []ShapedGlyph
	Advance geom.Pt
	Bounds  TextBounds
}

// ShapeText resolves text to font runs and shapes it into positioned glyphs.
func ShapeText(text string, origin geom.Pt, size float64, opts TextShapingOptions) (ShapedText, bool) {
	if text == "" || size <= 0 {
		return ShapedText{}, false
	}
	runs, ok := DefaultFontManager().ResolveTextRuns(text, opts.FontKey)
	if !ok {
		return ShapedText{}, false
	}
	return ShapeTextRuns(runs, origin, size, opts)
}

// ShapeTextRuns shapes already-resolved font runs.
func ShapeTextRuns(runs []FontRun, origin geom.Pt, size float64, opts TextShapingOptions) (ShapedText, bool) {
	if len(runs) == 0 || size <= 0 {
		return ShapedText{}, false
	}
	if opts.Direction == "" {
		opts.Direction = TextDirectionLTR
	}

	ppem := fixed.Int26_6(math.Round(size * 64))
	if ppem <= 0 {
		return ShapedText{}, false
	}

	var (
		shaped        ShapedText
		penX          = origin.X
		kernEnabled   = openTypeFeatureEnabled(opts, "kern", true)
		haveBounds    bool
		minX, minY    float64
		maxX, maxY    float64
		previous      sfnt.GlyphIndex
		previousFace  string
		havePrevious  bool
		clusterOffset int
		laidOutGlyphs bool
	)

	for _, inputRun := range runs {
		faceKey := fontFaceCacheKey(inputRun.Face)
		if inputRun.Text == "" || faceKey == "" {
			clusterOffset += len(inputRun.Text)
			continue
		}
		fontData, err := loadTextPathFontFace(inputRun.Face)
		if err != nil {
			return ShapedText{}, false
		}

		run := ShapedRun{
			Face:      inputRun.Face,
			Direction: opts.Direction,
			Script:    opts.Script,
			Language:  opts.Language,
			Features:  append([]TextFeature(nil), opts.Features...),
		}
		runStartX := penX
		runHaveBounds := false
		var runMinX, runMinY, runMaxX, runMaxY float64

		var buf sfnt.Buffer
		var inputGlyphs []shapingGlyph
		for localCluster, r := range inputRun.Text {
			glyphIndex, err := fontData.GlyphIndex(&buf, r)
			if err != nil {
				return ShapedText{}, false
			}
			if glyphIndex == 0 {
				havePrevious = false
				previousFace = ""
				continue
			}
			inputGlyphs = append(inputGlyphs, shapingGlyph{
				Rune:       r,
				Cluster:    clusterOffset + localCluster,
				GlyphIndex: glyphIndex,
			})
		}
		inputGlyphs = applyGSUBLigatures(inputRun.Face, inputGlyphs, opts)

		for _, inputGlyph := range inputGlyphs {
			kern := 0.0
			if kernEnabled && havePrevious && previousFace == faceKey {
				if k, err := fontData.Kern(&buf, previous, inputGlyph.GlyphIndex, ppem, font.HintingNone); err == nil {
					kern = fixedToFloat(k)
					penX += kern
				}
			}

			originPt := geom.Pt{X: penX, Y: origin.Y}
			segments, err := fontData.LoadGlyph(&buf, inputGlyph.GlyphIndex, ppem, nil)
			if err != nil {
				return ShapedText{}, false
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
				if !runHaveBounds {
					runMinX, runMinY, runMaxX, runMaxY = x0, y0, x1, y1
					runHaveBounds = true
				} else {
					runMinX = math.Min(runMinX, x0)
					runMinY = math.Min(runMinY, y0)
					runMaxX = math.Max(runMaxX, x1)
					runMaxY = math.Max(runMaxY, y1)
				}
			}

			advance, err := fontData.GlyphAdvance(&buf, inputGlyph.GlyphIndex, ppem, font.HintingNone)
			if err != nil {
				return ShapedText{}, false
			}
			advancePt := geom.Pt{X: fixedToFloat(advance)}
			glyph := ShapedGlyph{
				Rune:       inputGlyph.Rune,
				Cluster:    inputGlyph.Cluster,
				Face:       inputRun.Face,
				GlyphIndex: inputGlyph.GlyphIndex,
				Origin:     originPt,
				Advance:    advancePt,
				Kern:       kern,
			}
			run.Glyphs = append(run.Glyphs, glyph)
			shaped.Glyphs = append(shaped.Glyphs, glyph)
			penX += advancePt.X
			previous = inputGlyph.GlyphIndex
			previousFace = faceKey
			havePrevious = true
			laidOutGlyphs = true
		}

		run.Advance = geom.Pt{X: penX - runStartX}
		if runHaveBounds {
			run.Bounds = TextBounds{X: runMinX - origin.X, Y: runMinY - origin.Y, W: runMaxX - runMinX, H: runMaxY - runMinY}
		}
		if len(run.Glyphs) > 0 {
			shaped.Runs = append(shaped.Runs, run)
		}
		clusterOffset += len(inputRun.Text)
	}

	shaped.Advance = geom.Pt{X: penX - origin.X}
	if haveBounds {
		shaped.Bounds = TextBounds{X: minX - origin.X, Y: minY - origin.Y, W: maxX - minX, H: maxY - minY}
	}
	return shaped, laidOutGlyphs || shaped.Advance.X > 0 || shaped.Advance.Y > 0
}
