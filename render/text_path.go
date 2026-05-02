package render

import (
	"errors"
	"math"
	"os"
	"sync"
	"unicode/utf8"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var (
	textPathFontCacheMu sync.RWMutex
	textPathFontCache   = map[string]*sfnt.Font{}
)

// TextPath converts text to filled glyph outline paths at a baseline origin.
func TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if text == "" || size <= 0 {
		return geom.Path{}, false
	}

	manager := DefaultFontManager()
	runs, ok := manager.ResolveTextRuns(text, fontKey)
	if !ok {
		return geom.Path{}, false
	}

	path, ok := textRunsPath(runs, origin, size)
	if !ok || !path.Validate() || len(path.C) == 0 {
		return geom.Path{}, false
	}
	return path, true
}

func loadTextPathFont(path string) (*sfnt.Font, error) {
	textPathFontCacheMu.RLock()
	if cached, ok := textPathFontCache[path]; ok {
		textPathFontCacheMu.RUnlock()
		return cached, nil
	}
	textPathFontCacheMu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsed, err := sfnt.Parse(data)
	if err != nil {
		return nil, err
	}

	textPathFontCacheMu.Lock()
	textPathFontCache[path] = parsed
	textPathFontCacheMu.Unlock()
	return parsed, nil
}

func textPathFromFont(f *sfnt.Font, text string, origin geom.Pt, size float64) (geom.Path, error) {
	if f == nil || text == "" || size <= 0 {
		return geom.Path{}, errors.New("invalid text path input")
	}

	ppem := fixed.Int26_6(math.Round(size * 64))
	var buf sfnt.Buffer
	var out geom.Path
	penX := origin.X
	var previous sfnt.GlyphIndex
	havePrevious := false

	for _, r := range text {
		glyphIndex, err := f.GlyphIndex(&buf, r)
		if err != nil {
			return geom.Path{}, err
		}
		if glyphIndex == 0 {
			havePrevious = false
			continue
		}

		if havePrevious {
			kern, err := f.Kern(&buf, previous, glyphIndex, ppem, font.HintingNone)
			if err == nil {
				penX += fixedToFloat(kern)
			}
		}

		segments, err := f.LoadGlyph(&buf, glyphIndex, ppem, nil)
		if err != nil {
			return geom.Path{}, err
		}
		appendGlyphSegments(&out, segments, geom.Pt{X: penX, Y: origin.Y})

		advance, err := f.GlyphAdvance(&buf, glyphIndex, ppem, font.HintingNone)
		if err != nil {
			return geom.Path{}, err
		}
		penX += fixedToFloat(advance)
		previous = glyphIndex
		havePrevious = true
	}

	return out, nil
}

func textRunsPath(runs []FontRun, origin geom.Pt, size float64) (geom.Path, bool) {
	var out geom.Path
	ok := false
	layout, haveLayout := LayoutTextGlyphRuns(runs, origin, size)
	if !haveLayout {
		return geom.Path{}, false
	}
	ppem := fixed.Int26_6(math.Round(size * 64))
	var buf sfnt.Buffer
	for _, glyph := range layout.Glyphs {
		fontData, err := loadTextPathFont(glyph.Face.Path)
		if err != nil {
			return geom.Path{}, false
		}
		segments, err := fontData.LoadGlyph(&buf, glyph.GlyphIndex, ppem, nil)
		if err != nil {
			return geom.Path{}, false
		}
		before := len(out.C)
		appendGlyphSegments(&out, segments, glyph.Origin)
		ok = ok || len(out.C) > before
	}
	return out, ok
}

func textPathAndAdvanceFromFont(f *sfnt.Font, text string, origin geom.Pt, size float64) (geom.Path, float64, error) {
	if f == nil || text == "" || size <= 0 {
		return geom.Path{}, 0, errors.New("invalid text path input")
	}

	ppem := fixed.Int26_6(math.Round(size * 64))
	var buf sfnt.Buffer
	var out geom.Path
	penX := origin.X
	startX := origin.X
	var previous sfnt.GlyphIndex
	havePrevious := false

	for _, r := range text {
		glyphIndex, err := f.GlyphIndex(&buf, r)
		if err != nil {
			return geom.Path{}, 0, err
		}
		if glyphIndex == 0 {
			havePrevious = false
			continue
		}

		if havePrevious {
			kern, err := f.Kern(&buf, previous, glyphIndex, ppem, font.HintingNone)
			if err == nil {
				penX += fixedToFloat(kern)
			}
		}

		segments, err := f.LoadGlyph(&buf, glyphIndex, ppem, nil)
		if err != nil {
			return geom.Path{}, 0, err
		}
		appendGlyphSegments(&out, segments, geom.Pt{X: penX, Y: origin.Y})

		advance, err := f.GlyphAdvance(&buf, glyphIndex, ppem, font.HintingNone)
		if err != nil {
			return geom.Path{}, 0, err
		}
		penX += fixedToFloat(advance)
		previous = glyphIndex
		havePrevious = true
	}

	return out, penX - startX, nil
}

func fontFaceSupportsRune(face FontFace, r rune) bool {
	if face.Path == "" || r == utf8.RuneError {
		return false
	}
	fontData, err := loadTextPathFont(face.Path)
	if err != nil {
		return false
	}
	var buf sfnt.Buffer
	glyphIndex, err := fontData.GlyphIndex(&buf, r)
	return err == nil && glyphIndex != 0
}

func appendGlyphSegments(path *geom.Path, segments sfnt.Segments, origin geom.Pt) {
	for _, segment := range segments {
		switch segment.Op {
		case sfnt.SegmentOpMoveTo:
			path.MoveTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpLineTo:
			path.LineTo(sfntPoint(segment.Args[0], origin))
		case sfnt.SegmentOpQuadTo:
			path.QuadTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
			)
		case sfnt.SegmentOpCubeTo:
			path.CubicTo(
				sfntPoint(segment.Args[0], origin),
				sfntPoint(segment.Args[1], origin),
				sfntPoint(segment.Args[2], origin),
			)
		}
	}
}

func sfntPoint(p fixed.Point26_6, origin geom.Pt) geom.Pt {
	return geom.Pt{
		X: origin.X + fixedToFloat(p.X),
		Y: origin.Y + fixedToFloat(p.Y),
	}
}

func fixedToFloat(v fixed.Int26_6) float64 {
	return float64(v) / 64
}
