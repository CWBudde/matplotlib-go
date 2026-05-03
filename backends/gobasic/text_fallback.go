//go:build !freetype

package gobasic

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"sync"

	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	fallbackFontCacheMu sync.RWMutex
	fallbackFontCache   = map[string]*opentype.Font{}
)

func measureText(text string, size float64, fontKey string, dpi uint) render.TextMetrics {
	face, err := fallbackOpenTypeFace(fontKey, size, dpi)
	if err != nil {
		return render.TextMetrics{}
	}
	defer func() { _ = face.Close() }()

	textWidth := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()

	return render.TextMetrics{
		W:       quantize(float64(textWidth)),
		H:       quantize(float64(metrics.Height.Ceil())),
		Ascent:  quantize(float64(metrics.Ascent.Ceil())),
		Descent: quantize(float64(metrics.Descent.Ceil())),
	}
}

func renderTextBitmap(text string, size float64, textColor render.Color, fontKey string, dpi uint) *image.RGBA {
	face, err := fallbackOpenTypeFace(fontKey, size, dpi)
	if err != nil {
		return nil
	}
	defer func() { _ = face.Close() }()

	metrics := face.Metrics()
	baseW := font.MeasureString(face, text).Ceil()
	baseH := metrics.Height.Ceil()
	if baseW <= 0 || baseH <= 0 {
		return nil
	}

	src := image.NewRGBA(image.Rect(0, 0, baseW, baseH))
	d := font.Drawer{
		Dst:  src,
		Src:  image.NewUniform(renderColorToRGBA(textColor)),
		Face: face,
		Dot:  fixed.P(0, metrics.Ascent.Ceil()),
	}
	d.DrawString(text)
	return src
}

func fallbackOpenTypeFace(fontKey string, size float64, dpi uint) (font.Face, error) {
	parsed, err := fallbackOpenTypeFont(fontKey)
	if err != nil {
		return nil, err
	}
	if dpi == 0 {
		dpi = 72
	}

	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     float64(dpi),
		Hinting: font.HintingFull,
	})
}

func fallbackOpenTypeFont(fontKey string) (*opentype.Font, error) {
	face, ok := render.DefaultFontManager().FindFont(render.ParseFontProperties(fontKey))
	if !ok && fontKey == "" {
		face, ok = render.DefaultFontManager().FindFont(render.ParseFontProperties("DejaVu Sans"))
	}
	if !ok {
		return nil, errors.New("gobasic: no fallback font face")
	}

	key := fallbackFontCacheKey(face)
	if key == "" {
		return nil, errors.New("gobasic: fallback font face has no data")
	}
	fallbackFontCacheMu.RLock()
	if cached, ok := fallbackFontCache[key]; ok {
		fallbackFontCacheMu.RUnlock()
		return cached, nil
	}
	fallbackFontCacheMu.RUnlock()

	data := face.Data
	if face.Path != "" {
		var err error
		data, err = os.ReadFile(face.Path)
		if err != nil {
			return nil, err
		}
	}
	parsed, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}

	fallbackFontCacheMu.Lock()
	fallbackFontCache[key] = parsed
	fallbackFontCacheMu.Unlock()
	return parsed, nil
}

func fallbackFontCacheKey(face render.FontFace) string {
	if face.Path != "" {
		return "path:" + filepath.Clean(face.Path)
	}
	if face.Family != "" && len(face.Data) > 0 {
		return "embedded:" + face.Family
	}
	return ""
}
