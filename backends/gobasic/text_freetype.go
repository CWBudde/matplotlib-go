//go:build freetype

package gobasic

import (
	"errors"
	"image"
	"os"
	"strings"
	"sync"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"codeberg.org/go-fonts/dejavu/dejavusansmono"
	"codeberg.org/go-fonts/dejavu/dejavuserif"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"matplotlib-go/render"
)

var (
	freetypeFontCacheMu sync.RWMutex
	freetypeFontCache   = map[string]*opentype.Font{}
)

func measureText(text string, size float64, fontKey string) render.TextMetrics {
	face, err := openTypeFace(fontKey, size)
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

func renderTextBitmap(text string, size float64, textColor render.Color, fontKey string) *image.RGBA {
	face, err := openTypeFace(fontKey, size)
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

func openTypeFace(fontKey string, size float64) (font.Face, error) {
	t, err := parseOpentypeFont(fontKey)
	if err != nil {
		return nil, err
	}

	return opentype.NewFace(t, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func parseOpentypeFont(fontKey string) (*opentype.Font, error) {
	key := resolveFontFamily(fontKey)
	path := discoveredFontPath(fontKey, key)
	if path != "" {
		key = "file:" + path
	}
	freetypeFontCacheMu.RLock()
	if cached, ok := freetypeFontCache[key]; ok {
		freetypeFontCacheMu.RUnlock()
		return cached, nil
	}
	freetypeFontCacheMu.RUnlock()

	fontTTF := fontSourceForFamily(key)
	if path != "" {
		var err error
		fontTTF, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	if len(fontTTF) == 0 {
		return nil, errors.New("unsupported font family")
	}

	parsed, err := opentype.Parse(fontTTF)
	if err != nil {
		return nil, err
	}

	freetypeFontCacheMu.Lock()
	freetypeFontCache[key] = parsed
	freetypeFontCacheMu.Unlock()
	return parsed, nil
}

func discoveredFontPath(fontKey, resolvedFamily string) string {
	if usesEmbeddedFont(fontKey, resolvedFamily) {
		return ""
	}
	return render.DefaultFontManager().FindFontPath(fontKey)
}

func usesEmbeddedFont(fontKey, resolvedFamily string) bool {
	normalized := normalizeFontFamily(fontKey)
	return normalized == "" ||
		normalized == "default" ||
		normalized == resolvedFamily ||
		normalized == "dejavusans" ||
		normalized == "dejavuserif" ||
		normalized == "dejavusansmono" ||
		normalized == "sansserif" ||
		normalized == "serif" ||
		normalized == "mono" ||
		normalized == "monospace"
}

func resolveFontFamily(fontKey string) string {
	normalized := normalizeFontFamily(fontKey)
	switch {
	case strings.Contains(normalized, "serif"):
		return "serif"
	case strings.Contains(normalized, "mono") || strings.Contains(normalized, "monospace"):
		return "mono"
	default:
		return "sans"
	}
}

func normalizeFontFamily(fontKey string) string {
	normalized := strings.ToLower(strings.TrimSpace(fontKey))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	if normalized == "" || normalized == "default" {
		return "sans"
	}
	return normalized
}

func fontSourceForFamily(family string) []byte {
	switch family {
	case "serif":
		return dejavuserif.TTF
	case "mono":
		return dejavusansmono.TTF
	default:
		return dejavusans.TTF
	}
}
