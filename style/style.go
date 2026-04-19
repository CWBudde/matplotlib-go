package style

import (
	"matplotlib-go/color"
	"matplotlib-go/render"
)

// RC holds global rendering defaults (rc-like configuration).
// Fields are simple value types to keep configuration immutable-ish by copy.
type RC struct {
	DPI        float64
	FontKey    string
	FontSize   float64
	LineWidth  float64
	TextColor  [4]float64
	LineColor  [4]float64
	Background [4]float64
	TickCountX int
	TickCountY int

	AxesBackground     render.Color
	AxesEdgeColor      render.Color
	AxisLineWidth      float64
	GridColor          render.Color
	MinorGridColor     render.Color
	GridLineWidth      float64
	MinorGridLineWidth float64
	LegendBackground   render.Color
	LegendBorderColor  render.Color
	LegendTextColor    render.Color
	ColorCycle         color.Palette
}

// Default contains the library defaults. Copy and apply options to customize.
var Default = RC{
	DPI:                100,
	FontKey:            "DejaVu Sans",
	FontSize:           12,
	LineWidth:          1.25,
	TextColor:          [4]float64{0, 0, 0, 1},
	LineColor:          [4]float64{0, 0, 0, 1},
	Background:         [4]float64{1, 1, 1, 1},
	TickCountX:         5,
	TickCountY:         5,
	AxesBackground:     render.Color{R: 1, G: 1, B: 1, A: 1},
	AxesEdgeColor:      render.Color{R: 0, G: 0, B: 0, A: 1},
	AxisLineWidth:      0.8 * 100.0 / 72.0,
	GridColor:          render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
	MinorGridColor:     render.Color{R: 0.8, G: 0.8, B: 0.8, A: 0.4},
	GridLineWidth:      0.5,
	MinorGridLineWidth: 0.25,
	LegendBackground:   render.Color{R: 1, G: 1, B: 1, A: 0.9},
	LegendBorderColor:  render.Color{R: 0.2, G: 0.2, B: 0.2, A: 0.7},
	LegendTextColor:    render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1},
	ColorCycle:         color.Tab10,
}

// Option mutates an RC. Options should be applied on a copy derived from Default.
type Option func(*RC)

// Apply copies base and applies the given options in order, returning the result.
//
//nolint:gocritic // RC is intentionally passed by value to preserve copy-on-apply semantics.
func Apply(base RC, opts ...Option) RC {
	rc := base
	rc.ColorCycle = clonePalette(rc.ColorCycle)
	for _, opt := range opts {
		if opt != nil {
			opt(&rc)
		}
	}
	return rc
}

// WithDPI sets the DPI.
func WithDPI(d float64) Option { return func(rc *RC) { rc.DPI = d } }

// WithFont sets the font key and size.
func WithFont(key string, size float64) Option {
	return func(rc *RC) { rc.FontKey, rc.FontSize = key, size }
}

// WithLineWidth sets the default line width.
func WithLineWidth(w float64) Option { return func(rc *RC) { rc.LineWidth = w } }

// WithTextColor sets the default text color as normalized sRGBA (0..1).
func WithTextColor(r, g, b, a float64) Option {
	return func(rc *RC) { rc.TextColor = [4]float64{r, g, b, a} }
}

// WithLineColor sets the default stroke color as normalized sRGBA (0..1).
func WithLineColor(r, g, b, a float64) Option {
	return func(rc *RC) { rc.LineColor = [4]float64{r, g, b, a} }
}

// WithBackground sets the default background color as normalized sRGBA (0..1).
func WithBackground(r, g, b, a float64) Option {
	return func(rc *RC) { rc.Background = [4]float64{r, g, b, a} }
}

// WithTickCounts sets the target tick counts for X and Y.
func WithTickCounts(nx, ny int) Option { return func(rc *RC) { rc.TickCountX, rc.TickCountY = nx, ny } }

// WithAxesBackground sets the axes face color.
func WithAxesBackground(c render.Color) Option {
	return func(rc *RC) { rc.AxesBackground = c }
}

// WithAxesEdgeColor sets the axes spine and tick color.
func WithAxesEdgeColor(c render.Color) Option {
	return func(rc *RC) { rc.AxesEdgeColor = c }
}

// WithAxisLineWidth sets the default axes spine and tick width.
func WithAxisLineWidth(w float64) Option {
	return func(rc *RC) { rc.AxisLineWidth = w }
}

// WithGridColors sets the default major and minor grid colors.
func WithGridColors(major, minor render.Color) Option {
	return func(rc *RC) {
		rc.GridColor = major
		rc.MinorGridColor = minor
	}
}

// WithGridLineWidths sets the default major and minor grid widths.
func WithGridLineWidths(major, minor float64) Option {
	return func(rc *RC) {
		rc.GridLineWidth = major
		rc.MinorGridLineWidth = minor
	}
}

// WithLegendColors sets the legend box and text colors.
func WithLegendColors(background, border, text render.Color) Option {
	return func(rc *RC) {
		rc.LegendBackground = background
		rc.LegendBorderColor = border
		rc.LegendTextColor = text
	}
}

// WithColorCycle sets the automatic series color palette.
func WithColorCycle(palette color.Palette) Option {
	return func(rc *RC) { rc.ColorCycle = clonePalette(palette) }
}

// WithTheme replaces the current RC with the named theme preset.
func WithTheme(theme Theme) Option {
	return func(rc *RC) { *rc = Apply(theme.RC) }
}

// FigureBackground returns the figure face color as a renderer color.
func (rc RC) FigureBackground() render.Color {
	return render.Color{
		R: rc.Background[0],
		G: rc.Background[1],
		B: rc.Background[2],
		A: rc.Background[3],
	}
}

// DefaultTextColor returns the default text color as a renderer color.
func (rc RC) DefaultTextColor() render.Color {
	return render.Color{
		R: rc.TextColor[0],
		G: rc.TextColor[1],
		B: rc.TextColor[2],
		A: rc.TextColor[3],
	}
}

// DefaultLineColor returns the default line color as a renderer color.
func (rc RC) DefaultLineColor() render.Color {
	return render.Color{
		R: rc.LineColor[0],
		G: rc.LineColor[1],
		B: rc.LineColor[2],
		A: rc.LineColor[3],
	}
}

// Palette returns a copy of the configured automatic color cycle.
func (rc RC) Palette() color.Palette {
	return clonePalette(rc.ColorCycle)
}

func clonePalette(palette color.Palette) color.Palette {
	if len(palette) == 0 {
		palette = color.Tab10
	}
	cloned := make(color.Palette, len(palette))
	copy(cloned, palette)
	return cloned
}
