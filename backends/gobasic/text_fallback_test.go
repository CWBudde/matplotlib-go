//go:build !freetype

package gobasic

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestFallbackTextUsesAntialiasedEmbeddedFont(t *testing.T) {
	if _, err := fallbackOpenTypeFont(""); err != nil {
		t.Fatalf("fallbackOpenTypeFont returned error: %v", err)
	}
	img := renderTextBitmap("Ag", 24, render.Color{R: 0, G: 0, B: 0, A: 1}, "", 72)
	if img == nil {
		t.Fatal("renderTextBitmap returned nil")
	}

	var partialAlpha int
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] > 0 && img.Pix[i] < 255 {
			partialAlpha++
		}
	}
	if partialAlpha == 0 {
		t.Fatal("renderTextBitmap produced no antialiased pixels")
	}
}
