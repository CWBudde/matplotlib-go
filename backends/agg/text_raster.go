package agg

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sync"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	rasterFontCacheMu sync.RWMutex
	rasterFontCache   = map[string]*opentype.Font{}
)

func loadRasterFont(face render.FontFace) (*opentype.Font, error) {
	key := rasterFontCacheKey(face)
	if key == "" {
		return nil, errors.New("agg: raster font face has no path or data")
	}
	rasterFontCacheMu.RLock()
	if cached, ok := rasterFontCache[key]; ok {
		rasterFontCacheMu.RUnlock()
		return cached, nil
	}
	rasterFontCacheMu.RUnlock()

	fontData := face.Data
	if face.Path != "" {
		var err error
		fontData, err = os.ReadFile(face.Path)
		if err != nil {
			return nil, err
		}
	}

	parsed, err := opentype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	rasterFontCacheMu.Lock()
	rasterFontCache[key] = parsed
	rasterFontCacheMu.Unlock()
	return parsed, nil
}

func (r *Renderer) openRasterFace(face render.FontFace, size float64) (font.Face, error) {
	parsed, err := loadRasterFont(face)
	if err != nil {
		return nil, err
	}

	dpi := float64(r.resolution)
	if dpi <= 0 {
		dpi = 72
	}

	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

func (r *Renderer) measureRasterText(text string, face render.FontFace, size float64) (render.TextMetrics, bool) {
	fontKey := fontReference(face)
	if text == "" || fontKey == "" || size <= 0 {
		return render.TextMetrics{}, false
	}

	fontFace, err := r.openRasterFace(face, size)
	if err != nil {
		return render.TextMetrics{}, false
	}
	defer func() { _ = fontFace.Close() }()

	layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, r.fontPixelSize(size), fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}

	metrics := fontFace.Metrics()
	ascent := quantize(float64(metrics.Ascent.Ceil()))
	descent := quantize(float64(metrics.Descent.Ceil()))
	height := quantize(float64(metrics.Height.Ceil()))
	if fontHeights, ok := rasterFontHeightMetrics(face, size, r.resolution); ok {
		ascent = math.Max(fontHeights.ascent, math.Max(0, -layout.Bounds.Y))
		descent = math.Max(fontHeights.descent, math.Max(0, layout.Bounds.Y+layout.Bounds.H))
		height = ascent + descent
	}
	return render.TextMetrics{
		W:       quantize(layout.Advance),
		H:       height,
		Ascent:  ascent,
		Descent: descent,
	}, true
}

func (r *Renderer) drawRasterText(text string, face render.FontFace, origin geom.Pt, size float64, textColor render.Color) bool {
	fontKey := fontReference(face)
	if text == "" || fontKey == "" || size <= 0 {
		return false
	}

	primaryFace, err := r.openRasterFace(face, size)
	if err != nil {
		return false
	}
	defer func() { _ = primaryFace.Close() }()

	layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, r.fontPixelSize(size), fontKey)
	if !ok {
		return false
	}

	metrics := primaryFace.Metrics()
	faces := map[string]font.Face{rasterFontCacheKey(face): primaryFace}
	defer closeRasterFaces(faces, rasterFontCacheKey(face))

	minX := 0
	maxX := int(math.Ceil(layout.Advance))
	for _, glyph := range layout.Glyphs {
		glyphFace := glyph.Face
		if rasterFontCacheKey(glyphFace) == "" {
			glyphFace = face
		}
		face, err := r.rasterFaceForFace(faces, glyphFace, size)
		if err != nil {
			return false
		}
		bounds, _, ok := face.GlyphBounds(glyph.Rune)
		if !ok {
			continue
		}
		glyphMinX := int(math.Floor(glyph.Origin.X + float64(bounds.Min.X)/64.0))
		glyphMaxX := int(math.Ceil(glyph.Origin.X + float64(bounds.Max.X)/64.0))
		minX = min(minX, glyphMinX)
		maxX = max(maxX, glyphMaxX)
	}

	width := maxX - minX
	height := metrics.Height.Ceil()
	if width <= 0 || height <= 0 {
		return false
	}

	rawLeft := origin.X + float64(minX)
	rawTop := origin.Y - float64(metrics.Ascent.Ceil())
	topLeft := geom.Pt{
		X: math.Floor(rawLeft),
		Y: math.Floor(rawTop),
	}
	width = int(math.Ceil(origin.X+float64(maxX))) - int(topLeft.X)
	height = int(math.Ceil(rawTop+float64(height))) - int(topLeft.Y)
	if width <= 0 || height <= 0 {
		return false
	}

	src := image.NewRGBA(image.Rect(0, 0, width, height))
	uniform := image.NewUniform(renderColorToRGBA(textColor))
	for _, glyph := range layout.Glyphs {
		glyphFace := glyph.Face
		if rasterFontCacheKey(glyphFace) == "" {
			glyphFace = face
		}
		face, err := r.rasterFaceForFace(faces, glyphFace, size)
		if err != nil {
			return false
		}
		dot := fixed.Point26_6{
			X: fixed.Int26_6(math.Round((origin.X + glyph.Origin.X - topLeft.X) * 64.0)),
			Y: fixed.Int26_6(math.Round((origin.Y - topLeft.Y) * 64.0)),
		}
		dr, mask, maskp, _, ok := face.Glyph(dot, glyph.Rune)
		if !ok || dr.Empty() {
			continue
		}
		draw.DrawMask(src, dr, uniform, image.Point{}, mask, maskp, draw.Over)
	}

	img, err := agglib.NewImageFromStandardImage(src)
	if err != nil {
		return false
	}

	return r.ctx.DrawImageScaled(img, topLeft.X, topLeft.Y, float64(width), float64(height)) == nil
}

func (r *Renderer) rasterFaceForFace(faces map[string]font.Face, fontFace render.FontFace, size float64) (font.Face, error) {
	key := rasterFontCacheKey(fontFace)
	if face := faces[key]; face != nil {
		return face, nil
	}
	face, err := r.openRasterFace(fontFace, size)
	if err != nil {
		return nil, err
	}
	faces[key] = face
	return face, nil
}

func closeRasterFaces(faces map[string]font.Face, keepPath string) {
	for fontPath, face := range faces {
		if fontPath != keepPath && face != nil {
			_ = face.Close()
		}
	}
}

func rasterFontCacheKey(face render.FontFace) string {
	if face.Path != "" {
		return "path:" + filepath.Clean(face.Path)
	}
	if face.Family != "" && len(face.Data) > 0 {
		return "embedded:" + face.Family
	}
	return ""
}

func renderColorToRGBA(c render.Color) color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(clamp01(c.R) * 255)),
		G: uint8(math.Round(clamp01(c.G) * 255)),
		B: uint8(math.Round(clamp01(c.B) * 255)),
		A: uint8(math.Round(clamp01(c.A) * 255)),
	}
}
