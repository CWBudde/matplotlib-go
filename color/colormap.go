package color

import (
	"math"
	"strings"

	"matplotlib-go/render"
)

// ColorStop defines a point in a piecewise-linear colormap.
type ColorStop struct {
	Pos float64
	Color render.Color
}

// Colormap maps normalized values in [0,1] to colors.
type Colormap struct {
	name  string
	stops []ColorStop
}

// Name returns the user-visible colormap name.
func (c Colormap) Name() string {
	return c.name
}

// At returns a color for normalized input t in [0,1].
func (c Colormap) At(t float64) render.Color {
	if len(c.stops) == 0 {
		return render.Color{R: 0, G: 0, B: 0, A: 1}
	}

	v := clamp01(t)

	if v <= c.stops[0].Pos {
		return c.stops[0].Color
	}
	if v >= c.stops[len(c.stops)-1].Pos {
		return c.stops[len(c.stops)-1].Color
	}

	for i := 1; i < len(c.stops); i++ {
		left := c.stops[i-1]
		right := c.stops[i]
		if v < right.Pos {
			if right.Pos-left.Pos <= 0 {
				return right.Color
			}
			local := (v - left.Pos) / (right.Pos - left.Pos)
			return mixColor(left.Color, right.Color, local)
		}
	}

	return c.stops[len(c.stops)-1].Color
}

// NewColormap creates a new linear colormap from color stops.
// Stops are sorted by Pos and clamped to [0,1].
func NewColormap(name string, stops []ColorStop) Colormap {
	norm := normalizeColormapName(name)
	if len(stops) == 0 {
		return Colormap{name: norm, stops: []ColorStop{
			{Pos: 0, Color: render.Color{R: 0, G: 0, B: 0, A: 1}},
			{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
		}}
	}

	normalized := make([]ColorStop, len(stops))
	for i, stop := range stops {
		normalized[i] = ColorStop{
			Pos:   clamp01(stop.Pos),
			Color: stop.Color,
		}
	}

	return Colormap{name: norm, stops: normalized}
}

var defaultColormapName = "viridis"

var colormaps = map[string]Colormap{
	"viridis": {
		name: "viridis",
		stops: []ColorStop{
			{0.00, render.Color{R: 0.267, G: 0.004, B: 0.329, A: 1}},
			{0.13, render.Color{R: 0.283, G: 0.141, B: 0.458, A: 1}},
			{0.26, render.Color{R: 0.254, G: 0.265, B: 0.531, A: 1}},
			{0.38, render.Color{R: 0.168, G: 0.467, B: 0.557, A: 1}},
			{0.51, render.Color{R: 0.128, G: 0.566, B: 0.551, A: 1}},
			{0.64, render.Color{R: 0.278, G: 0.640, B: 0.494, A: 1}},
			{0.76, render.Color{R: 0.572, G: 0.750, B: 0.300, A: 1}},
			{0.89, render.Color{R: 0.838, G: 0.873, B: 0.132, A: 1}},
			{1.00, render.Color{R: 0.993, G: 0.906, B: 0.145, A: 1}},
		},
	},
	"gray": {
		name: "gray",
		stops: []ColorStop{
			{0, render.Color{R: 0, G: 0, B: 0, A: 1}},
			{1, render.Color{R: 1, G: 1, B: 1, A: 1}},
		},
	},
	"inferno": {
		name: "inferno",
		stops: []ColorStop{
			{0.00, render.Color{R: 0.001, G: 0.000, B: 0.014, A: 1}},
			{0.20, render.Color{R: 0.200, G: 0.065, B: 0.497, A: 1}},
			{0.40, render.Color{R: 0.867, G: 0.221, B: 0.320, A: 1}},
			{0.60, render.Color{R: 0.988, G: 0.683, B: 0.250, A: 1}},
			{0.80, render.Color{R: 0.992, G: 0.988, B: 0.643, A: 1}},
			{1.00, render.Color{R: 1.000, G: 1.000, B: 1.000, A: 1}},
		},
	},
}

// RegisterColormap adds a named colormap to the runtime registry.
func RegisterColormap(name string, cmap Colormap) {
	key := normalizeColormapName(name)
	if key == "" {
		return
	}
	cmap.name = key
	colormaps[key] = cmap
}

// GetColormap returns a colormap by name.
// Unknown names fall back to the default `viridis` colormap.
func GetColormap(name string) Colormap {
	key := normalizeColormapName(name)
	if cmap, ok := colormaps[key]; ok {
		return cmap
	}
	return colormaps[defaultColormapName]
}

// DefaultColormap returns the configured default colormap.
func DefaultColormap() Colormap {
	return GetColormap(defaultColormapName)
}

func mixColor(a, b render.Color, t float64) render.Color {
	return render.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}

func normalizeColormapName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	if math.IsNaN(v) {
		return 0
	}
	return v
}
