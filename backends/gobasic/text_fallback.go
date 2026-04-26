//go:build !freetype

package gobasic

import (
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"matplotlib-go/render"
)

func measureText(text string, size float64, _ string, dpi uint) render.TextMetrics {
	size = scalePointSize(size, dpi)
	scale := size / defaultFontHeight
	if size <= 0 || scale <= 0 {
		return render.TextMetrics{}
	}

	face := basicfont.Face7x13
	width := font.MeasureString(face, text).Ceil()
	height := float64(face.Metrics().Height.Ceil())
	ascent := float64(face.Metrics().Ascent.Ceil())
	descent := float64(face.Metrics().Descent.Ceil())

	return render.TextMetrics{
		W:       quantize(float64(width) * scale),
		H:       quantize(height * scale),
		Ascent:  quantize(ascent * scale),
		Descent: quantize(descent * scale),
	}
}

func renderTextBitmap(text string, size float64, textColor render.Color, _ string, dpi uint) *image.RGBA {
	size = scalePointSize(size, dpi)
	face := basicfont.Face7x13
	baseFontMetrics := face.Metrics()
	baseW := int(font.MeasureString(face, text).Ceil())
	baseH := baseFontMetrics.Height.Ceil()
	if baseW <= 0 || baseH <= 0 {
		return nil
	}

	src := image.NewRGBA(image.Rect(0, 0, baseW, baseH))
	d := font.Drawer{
		Dst:  src,
		Src:  image.NewUniform(renderColorToRGBA(textColor)),
		Face: face,
		Dot:  fixed.P(0, baseFontMetrics.Ascent.Ceil()),
	}
	d.DrawString(text)

	scale := size / defaultFontHeight
	if scale == 1 {
		return src
	}
	if scale <= 0 {
		return nil
	}

	return scaleImageNearest(src, scale, scale)
}
