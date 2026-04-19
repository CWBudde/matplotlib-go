package agg

import (
	"image"
	"image/color"
	"math"
	"os"
	"sync"

	agglib "github.com/cwbudde/agg_go"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

var (
	rasterFontCacheMu sync.RWMutex
	rasterFontCache   = map[string]*opentype.Font{}
)

func loadRasterFont(fontPath string) (*opentype.Font, error) {
	rasterFontCacheMu.RLock()
	if cached, ok := rasterFontCache[fontPath]; ok {
		rasterFontCacheMu.RUnlock()
		return cached, nil
	}
	rasterFontCacheMu.RUnlock()

	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	parsed, err := opentype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	rasterFontCacheMu.Lock()
	rasterFontCache[fontPath] = parsed
	rasterFontCacheMu.Unlock()
	return parsed, nil
}

func (r *Renderer) openRasterFace(fontPath string, size float64) (font.Face, error) {
	parsed, err := loadRasterFont(fontPath)
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

func (r *Renderer) measureRasterText(text, fontPath string, size float64) (render.TextMetrics, bool) {
	if text == "" || fontPath == "" || size <= 0 {
		return render.TextMetrics{}, false
	}

	face, err := r.openRasterFace(fontPath, size)
	if err != nil {
		return render.TextMetrics{}, false
	}
	defer func() { _ = face.Close() }()

	metrics := face.Metrics()
	return render.TextMetrics{
		W:       quantize(float64(font.MeasureString(face, text).Ceil())),
		H:       quantize(float64(metrics.Height.Ceil())),
		Ascent:  quantize(float64(metrics.Ascent.Ceil())),
		Descent: quantize(float64(metrics.Descent.Ceil())),
	}, true
}

func (r *Renderer) drawRasterText(text, fontPath string, origin geom.Pt, size float64, textColor render.Color) bool {
	if text == "" || fontPath == "" || size <= 0 {
		return false
	}

	face, err := r.openRasterFace(fontPath, size)
	if err != nil {
		return false
	}
	defer func() { _ = face.Close() }()

	metrics := face.Metrics()
	width := font.MeasureString(face, text).Ceil()
	height := metrics.Height.Ceil()
	if width <= 0 || height <= 0 {
		return false
	}

	src := image.NewRGBA(image.Rect(0, 0, width, height))
	drawer := font.Drawer{
		Dst:  src,
		Src:  image.NewUniform(renderColorToRGBA(textColor)),
		Face: face,
		Dot:  fixed.P(0, metrics.Ascent.Ceil()),
	}
	drawer.DrawString(text)

	img, err := agglib.NewImageFromStandardImage(src)
	if err != nil {
		return false
	}

	topLeft := geom.Pt{
		X: float64(int(origin.X)),
		Y: float64(int(origin.Y - float64(metrics.Ascent.Ceil()))),
	}
	return r.ctx.DrawImageScaled(img, topLeft.X, topLeft.Y, float64(width), float64(height)) == nil
}

func renderColorToRGBA(c render.Color) color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(clamp01(c.R) * 255)),
		G: uint8(math.Round(clamp01(c.G) * 255)),
		B: uint8(math.Round(clamp01(c.B) * 255)),
		A: uint8(math.Round(clamp01(c.A) * 255)),
	}
}
