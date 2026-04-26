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

func TestGetColormap_PlasmaRegistered(t *testing.T) {
	c := GetColormap("plasma")
	if c.Name() != "plasma" {
		t.Fatalf("expected plasma colormap, got %q", c.Name())
	}
	if got := c.At(0); got.B < got.R || got.B < got.G {
		t.Fatalf("expected plasma low end to be purple, got %#v", got)
	}
	if got := c.At(1); got.R < 0.9 || got.G < 0.9 {
		t.Fatalf("expected plasma high end to be yellow, got %#v", got)
	}
}

func TestGetColormap_ChannelMapsRegistered(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "red channel", want: "red channel"},
		{name: "green channel", want: "green channel"},
		{name: "blue channel", want: "blue channel"},
	}
	for _, tt := range tests {
		c := GetColormap(tt.name)
		if c.Name() != tt.want {
			t.Fatalf("GetColormap(%q).Name() = %q, want %q", tt.name, c.Name(), tt.want)
		}
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
	if math.Abs(mid.R-0.5) > 1e-9 || math.Abs(mid.G-0) > 1e-9 || math.Abs(mid.B-0.5) > 1e-9 {
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
