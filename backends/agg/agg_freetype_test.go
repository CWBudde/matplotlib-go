//go:build freetype

package agg

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
)

func TestUsesDejaVuSansWithoutFallback(t *testing.T) {
	r := mustNew(t, 200, 100)
	if r.fontPath == "" {
		t.Fatal("expected DejaVu Sans font path to be configured")
	}

	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	r.DrawText("Hello", geom.Pt{X: 20, Y: 40}, 12, white)
	r.DrawTextRotated("Hello", geom.Pt{X: 120, Y: 50}, 12, math.Pi/2, white)

	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if r.fallback {
		t.Fatal("expected DejaVu Sans path to be used without falling back to GSV")
	}
}
