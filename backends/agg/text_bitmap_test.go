package agg

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestBlendAlphaMaskAppliesTextAlphaAndClips(t *testing.T) {
	r := mustNew(t, 3, 3)
	mask := image.NewAlpha(image.Rect(0, 0, 3, 3))
	mask.SetAlpha(0, 0, color.Alpha{A: 255})
	mask.SetAlpha(1, 1, color.Alpha{A: 128})
	mask.SetAlpha(2, 2, color.Alpha{A: 255})

	if !r.blendAlphaMask(mask, -1, -1, render.Color{R: 1, G: 0, B: 0, A: 0.5}) {
		t.Fatal("blendAlphaMask returned false")
	}

	img := r.GetImage()
	if got := img.RGBAAt(0, 0); got.R != 255 || got.G < 188 || got.G > 193 || got.B < 188 || got.B > 193 || got.A != 255 {
		t.Fatalf("half-covered clipped text pixel = %+v, want red over white with alpha-scaled coverage", got)
	}
	if got := img.RGBAAt(1, 1); got.R != 255 || got.G < 126 || got.G > 129 || got.B < 126 || got.B > 129 || got.A != 255 {
		t.Fatalf("full-covered clipped text pixel = %+v, want red over white with text alpha", got)
	}
	if got := img.RGBAAt(2, 2); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("out-of-mask destination pixel changed: %+v", got)
	}
}
