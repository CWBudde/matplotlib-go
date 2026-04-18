package color

import (
	"math"
	"testing"

	"matplotlib-go/render"
)

func TestGetColormap_UnknownFallsBackToViridis(t *testing.T) {
	c := GetColormap("does-not-exist")
	if c.Name() != "viridis" {
		t.Fatalf("expected fallback colormap viridis, got %q", c.Name())
	}
}

func TestRegisterColormap_NormalizesNameAndClampsStops(t *testing.T) {
	name := "Custom Test"
	RegisterColormap(name, NewColormap(name, []ColorStop{
		{Pos: -0.5, Color: render.Color{R: 1, G: 0, B: 0, A: 1}},
		{Pos: 1.4, Color: render.Color{R: 0, G: 0, B: 1, A: 1}},
	}))

	c := GetColormap("  CuStOm tEsT  ")
	if c.Name() != "custom test" {
		t.Fatalf("expected normalized colormap name %q, got %q", "custom test", c.Name())
	}

	mid := c.At(0.5)
	if math.Abs(mid.R-0.5) > 1e-9 || math.Abs(mid.G-0.5) > 1e-9 || math.Abs(mid.B-0.5) > 1e-9 {
		t.Fatalf("unexpected midpoint color: %#v", mid)
	}
}

func TestRegisterColormap_IgnoreEmptyName(t *testing.T) {
	// Preserve the fallback behavior when name normalization would become empty.
	defaultBefore := DefaultColormap()
	RegisterColormap("   ", NewColormap("ignored", []ColorStop{}))
	got := GetColormap("ignored")
	if got.Name() != defaultBefore.Name() {
		t.Fatalf("empty name registration should be ignored, expected %q got %q", defaultBefore.Name(), got.Name())
	}
}
