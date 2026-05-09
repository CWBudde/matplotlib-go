package color

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
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
		{name: "cividis", t: 0.00, want: [4]uint8{0, 34, 78, 255}},
		{name: "cividis", t: 0.50, want: [4]uint8{125, 124, 120, 255}},
		{name: "cividis", t: 1.00, want: [4]uint8{254, 232, 56, 255}},
	}

	for _, tt := range tests {
		got := colorBytes(GetColormap(tt.name).At(tt.t))
		if got != tt.want {
			t.Fatalf("%s.At(%v) bytes = %v, want %v", tt.name, tt.t, got, tt.want)
		}
	}
}

func TestBinaryColormapMatchesMatplotlibSpyDefaults(t *testing.T) {
	cmap := GetColormap("binary")

	if got := cmap.At(0); got != (render.Color{R: 1, G: 1, B: 1, A: 1}) {
		t.Fatalf("binary at 0 = %+v, want white", got)
	}
	if got := cmap.At(1); got != (render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("binary at 1 = %+v, want black", got)
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

func TestColormapAtValueUsesBadUnderAndOverColors(t *testing.T) {
	bad := render.Color{R: 0.7, G: 0.7, B: 0.7, A: 0.4}
	under := render.Color{R: 0.1, G: 0.2, B: 0.9, A: 1}
	over := render.Color{R: 0.9, G: 0.2, B: 0.1, A: 1}
	c := NewColormap("bounded", []ColorStop{
		{Pos: 0, Color: render.Color{R: 0, G: 0, B: 0, A: 1}},
		{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
	}).WithBad(bad).WithUnder(under).WithOver(over)

	if got := c.AtValue(math.NaN()); got != bad {
		t.Fatalf("bad color = %#v, want %#v", got, bad)
	}
	if got := c.AtValue(-0.01); got != under {
		t.Fatalf("under color = %#v, want %#v", got, under)
	}
	if got := c.AtValue(1.01); got != over {
		t.Fatalf("over color = %#v, want %#v", got, over)
	}
	if got := c.AtValue(0.5); got.R < 0.49 || got.R > 0.51 {
		t.Fatalf("in-range color = %#v, want midpoint", got)
	}
}

func TestColormapAtValueDefaultsBadTransparentAndUnderOverEndpoints(t *testing.T) {
	c := NewColormap("defaults", []ColorStop{
		{Pos: 0, Color: render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}},
		{Pos: 1, Color: render.Color{R: 0.8, G: 0.7, B: 0.6, A: 1}},
	})

	if got := c.AtValue(math.NaN()); got.A != 0 {
		t.Fatalf("default bad color = %#v, want transparent", got)
	}
	if got, want := c.AtValue(-1), c.At(0); got != want {
		t.Fatalf("default under = %#v, want low endpoint %#v", got, want)
	}
	if got, want := c.AtValue(2), c.At(1); got != want {
		t.Fatalf("default over = %#v, want high endpoint %#v", got, want)
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
