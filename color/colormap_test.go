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

func TestListedColormapMatchesMatplotlibViridisBytes(t *testing.T) {
	c := GetColormap("viridis")
	tests := []struct {
		t          float64
		r, g, b, a uint8
	}{
		{0.000, 68, 1, 84, 255},
		{0.125, 71, 44, 123, 255},
		{0.250, 58, 82, 139, 255},
		{0.375, 44, 114, 142, 255},
		{0.500, 32, 144, 140, 255},
		{0.625, 40, 174, 127, 255},
		{0.750, 94, 201, 97, 255},
		{0.875, 173, 220, 48, 255},
		{1.000, 253, 231, 36, 255},
	}

	for _, tt := range tests {
		got := colorBytes(c.At(tt.t))
		want := [4]uint8{tt.r, tt.g, tt.b, tt.a}
		if got != want {
			t.Fatalf("viridis.At(%v) bytes = %v, want %v", tt.t, got, want)
		}
	}
}

func TestListedColormapRepresentativeBytes(t *testing.T) {
	tests := []struct {
		name string
		t    float64
		want [4]uint8
	}{
		{name: "inferno", t: 0.00, want: [4]uint8{0, 0, 3, 255}},
		{name: "inferno", t: 0.50, want: [4]uint8{187, 55, 84, 255}},
		{name: "inferno", t: 1.00, want: [4]uint8{252, 254, 164, 255}},
		{name: "magma", t: 0.00, want: [4]uint8{0, 0, 3, 255}},
		{name: "magma", t: 0.50, want: [4]uint8{182, 54, 121, 255}},
		{name: "magma", t: 1.00, want: [4]uint8{251, 252, 191, 255}},
	}

	for _, tt := range tests {
		got := colorBytes(GetColormap(tt.name).At(tt.t))
		if got != tt.want {
			t.Fatalf("%s.At(%v) bytes = %v, want %v", tt.name, tt.t, got, tt.want)
		}
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

func colorBytes(c render.Color) [4]uint8 {
	return [4]uint8{
		uint8(c.R * 255),
		uint8(c.G * 255),
		uint8(c.B * 255),
		uint8(c.A * 255),
	}
}
