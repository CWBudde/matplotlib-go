package core

import (
	"sort"

	"matplotlib-go/color"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

// Artist is anything that can draw itself with a z-order and optional bounds.
type Artist interface {
	Draw(r render.Renderer, ctx *DrawContext)
	Z() float64
	Bounds(ctx *DrawContext) geom.Rect
}

// ArtistFunc adapts a function to an Artist.
type ArtistFunc func(r render.Renderer, ctx *DrawContext)

func (f ArtistFunc) Draw(r render.Renderer, ctx *DrawContext) { f(r, ctx) }
func (f ArtistFunc) Z() float64                               { return 0 }
func (f ArtistFunc) Bounds(_ *DrawContext) geom.Rect          { return geom.Rect{} }

// DrawContext carries per-draw state like transforms and style.
type DrawContext struct {
	// DataToPixel maps data coordinates to pixels.
	DataToPixel Transform2D
	// Styling configuration in effect.
	RC style.RC
	// Clip is the axes pixel rectangle.
	Clip geom.Rect
}

// Transform2D wires x/y scales with an axes->pixel affine transform.
type Transform2D struct {
	XScale      transform.Scale
	YScale      transform.Scale
	AxesToPixel transform.AffineT
}

// Apply transforms a data-space point to pixel coordinates.
func (t *Transform2D) Apply(p geom.Pt) geom.Pt {
	u := t.XScale.Fwd(p.X)
	v := t.YScale.Fwd(p.Y)
	return t.AxesToPixel.Apply(geom.Pt{X: u, Y: v})
}

// Figure is the root of the Artist tree. It contains Axes children.
type Figure struct {
	SizePx   geom.Pt
	RC       style.RC
	Children []*Axes
}

// NewFigure creates a new figure with pixel dimensions and optional style overrides.
func NewFigure(w, h int, opts ...style.Option) *Figure {
	rc := style.Apply(style.Default, opts...)
	return &Figure{
		SizePx:   geom.Pt{X: float64(w), Y: float64(h)},
		RC:       rc,
		Children: nil,
	}
}

// Axes represents an axes region inside a figure.
type Axes struct {
	RectFraction geom.Rect // [0..1] fraction in figure coords
	RC           *style.RC // nil => inherit figure RC
	XScale       transform.Scale
	YScale       transform.Scale
	Artists      []Artist
	zsorted      bool

	// Axis control
	XAxis     *Axis // bottom x-axis
	YAxis     *Axis // left y-axis
	ShowFrame bool  // draw top and right border lines (default true)

	// Text labels
	Title  string // title above the plot
	XLabel string // x-axis label below ticks
	YLabel string // y-axis label left of ticks

	// Color cycling for multiple series
	ColorCycle *color.ColorCycle
}

// AddAxes appends an Axes to the Figure. If opts are provided, the Axes gets its
// own RC copy; otherwise it inherits from the Figure.
func (f *Figure) AddAxes(r geom.Rect, opts ...style.Option) *Axes {
	var rc *style.RC
	if len(opts) > 0 {
		v := style.Apply(f.RC, opts...)
		rc = &v
	}
	ax := &Axes{
		RectFraction: r,
		RC:           rc,
		XScale:       transform.NewLinear(0, 1),
		YScale:       transform.NewLinear(0, 1),
		XAxis:        NewXAxis(),
		YAxis:        NewYAxis(),
		ShowFrame:    true,
		ColorCycle:   color.NewDefaultColorCycle(),
	}
	f.Children = append(f.Children, ax)
	return ax
}

// Add registers an Artist with the Axes.
func (a *Axes) Add(art Artist) { a.Artists = append(a.Artists, art); a.zsorted = false }

// SetXLim sets the x-axis limits.
func (a *Axes) SetXLim(min, max float64) {
	a.XScale = transform.NewLinear(min, max)
}

// SetYLim sets the y-axis limits.
func (a *Axes) SetYLim(min, max float64) {
	a.YScale = transform.NewLinear(min, max)
}

// SetXLimLog sets the x-axis to logarithmic scale with given limits.
func (a *Axes) SetXLimLog(min, max, base float64) {
	a.XScale = transform.NewLog(min, max, base)
	if a.XAxis != nil {
		a.XAxis.Locator = LogLocator{Base: base, Minor: false}
		a.XAxis.Formatter = LogFormatter{Base: base}
	}
}

// SetYLimLog sets the y-axis to logarithmic scale with given limits.
func (a *Axes) SetYLimLog(min, max, base float64) {
	a.YScale = transform.NewLog(min, max, base)
	if a.YAxis != nil {
		a.YAxis.Locator = LogLocator{Base: base, Minor: false}
		a.YAxis.Formatter = LogFormatter{Base: base}
	}
}

// AutoScale computes axis limits from the data bounds of all artists,
// adding the specified margin fraction on each side (e.g. 0.05 = 5%).
// A margin of 0 fits exactly to the data. If no artists have non-zero bounds,
// limits remain unchanged.
func (a *Axes) AutoScale(margin float64) {
	var xMin, xMax, yMin, yMax float64
	first := true

	for _, art := range a.Artists {
		b := art.Bounds(nil)
		if b.W() == 0 && b.H() == 0 && b.Min.X == 0 && b.Min.Y == 0 {
			continue // skip zero-bounds artists (grids, etc.)
		}
		if first {
			xMin, xMax = b.Min.X, b.Max.X
			yMin, yMax = b.Min.Y, b.Max.Y
			first = false
		} else {
			if b.Min.X < xMin {
				xMin = b.Min.X
			}
			if b.Max.X > xMax {
				xMax = b.Max.X
			}
			if b.Min.Y < yMin {
				yMin = b.Min.Y
			}
			if b.Max.Y > yMax {
				yMax = b.Max.Y
			}
		}
	}
	if first {
		return // no data artists
	}

	// Apply margin
	xSpan := xMax - xMin
	ySpan := yMax - yMin
	if xSpan == 0 {
		xSpan = 1 // avoid zero-span
	}
	if ySpan == 0 {
		ySpan = 1
	}
	xMin -= xSpan * margin
	xMax += xSpan * margin
	yMin -= ySpan * margin
	yMax += ySpan * margin

	a.XScale = transform.NewLinear(xMin, xMax)
	a.YScale = transform.NewLinear(yMin, yMax)
}

// AddGrid adds grid lines for the specified axis.
func (a *Axes) AddGrid(axis AxisSide) *Grid {
	grid := NewGrid(axis)
	a.Add(grid)
	return grid
}

// AddXGrid adds vertical grid lines based on x-axis ticks.
func (a *Axes) AddXGrid() *Grid {
	return a.AddGrid(AxisBottom)
}

// AddYGrid adds horizontal grid lines based on y-axis ticks.
func (a *Axes) AddYGrid() *Grid {
	return a.AddGrid(AxisLeft)
}

// SetTitle sets the title displayed above the plot.
func (a *Axes) SetTitle(title string) { a.Title = title }

// SetXLabel sets the label displayed below the x-axis.
func (a *Axes) SetXLabel(label string) { a.XLabel = label }

// SetYLabel sets the label displayed left of the y-axis.
func (a *Axes) SetYLabel(label string) { a.YLabel = label }

// NextColor returns the next color in the axes color cycle.
func (a *Axes) NextColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewDefaultColorCycle()
	}
	return a.ColorCycle.Next()
}

// PeekColor returns the current color without advancing the cycle.
func (a *Axes) PeekColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewDefaultColorCycle()
	}
	return a.ColorCycle.Peek()
}

// ResetColorCycle resets the color cycle to the first color.
func (a *Axes) ResetColorCycle() {
	if a.ColorCycle != nil {
		a.ColorCycle.Reset()
	}
}

// layout computes the pixel rectangle for this Axes inside the Figure.
func (a *Axes) layout(f *Figure) (pixelRect geom.Rect) {
	// Map fraction [0..1] to pixel coordinates
	min := geom.Pt{X: f.SizePx.X * a.RectFraction.Min.X, Y: f.SizePx.Y * a.RectFraction.Min.Y}
	max := geom.Pt{X: f.SizePx.X * a.RectFraction.Max.X, Y: f.SizePx.Y * a.RectFraction.Max.Y}
	return geom.Rect{Min: min, Max: max}
}

// effectiveRC resolves the RC for this axes, inheriting from the Figure if needed.
func (a *Axes) effectiveRC(f *Figure) style.RC {
	if a.RC != nil {
		return *a.RC
	}
	return f.RC
}

// DrawFigure performs a traversal and draws the figure into the renderer.
func DrawFigure(fig *Figure, r render.Renderer) {
	vp := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: fig.SizePx.X, Y: fig.SizePx.Y}}
	_ = r.Begin(vp)
	defer r.End()

	for _, ax := range fig.Children {
		px := ax.layout(fig)

		// Build DrawContext with composed transform
		ctx := &DrawContext{
			DataToPixel: Transform2D{
				XScale:      ax.XScale,
				YScale:      ax.YScale,
				AxesToPixel: transform.NewAffine(axesToPixel(px)),
			},
			RC:   ax.effectiveRC(fig),
			Clip: px,
		}

		// Draw clipped content (data, grids, spines, tick marks)
		r.Save()
		r.ClipRect(px)

		if !ax.zsorted {
			sort.SliceStable(ax.Artists, func(i, j int) bool {
				zi, zj := ax.Artists[i].Z(), ax.Artists[j].Z()
				if zi == zj {
					return i < j
				}
				return zi < zj
			})
			ax.zsorted = true
		}
		for _, art := range ax.Artists {
			art.Draw(r, ctx)
		}

		if ax.XAxis != nil {
			ax.XAxis.Draw(r, ctx)
		}
		if ax.YAxis != nil {
			ax.YAxis.Draw(r, ctx)
		}
		if ax.ShowFrame {
			ref := ax.XAxis
			if ref == nil {
				ref = ax.YAxis
			}
			DrawFrame(r, ctx, ref)
		}
		r.Restore()

		// Draw tick labels and text labels outside the clip rect (in the margins)
		if ax.XAxis != nil {
			ax.XAxis.DrawTickLabels(r, ctx)
		}
		if ax.YAxis != nil {
			ax.YAxis.DrawTickLabels(r, ctx)
		}
		drawAxesLabels(ax, r, ctx, px)
	}
}

// drawAxesLabels renders title, xlabel, and ylabel outside the clipped axes area.
func drawAxesLabels(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) {
	type textRenderer interface {
		DrawText(text string, origin geom.Pt, size float64, textColor render.Color)
	}
	type verticalTextRenderer interface {
		textRenderer
		DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color)
	}

	textRen, ok := r.(textRenderer)
	if !ok {
		return
	}

	labelColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	titleSize := ctx.RC.FontSize + 2
	labelSize := ctx.RC.FontSize

	// Title: centered above the plot
	if ax.Title != "" {
		metrics := r.MeasureText(ax.Title, titleSize, ctx.RC.FontKey)
		x := px.Min.X + (px.W()-metrics.W)/2
		y := px.Min.Y - 10
		textRen.DrawText(ax.Title, geom.Pt{X: x, Y: y}, titleSize, labelColor)
	}

	// XLabel: centered below the x-axis tick labels
	if ax.XLabel != "" {
		metrics := r.MeasureText(ax.XLabel, labelSize, ctx.RC.FontKey)
		x := px.Min.X + (px.W()-metrics.W)/2
		y := px.Max.Y + 35
		textRen.DrawText(ax.XLabel, geom.Pt{X: x, Y: y}, labelSize, labelColor)
	}

	// YLabel: vertical text if supported, else horizontal fallback
	if ax.YLabel != "" {
		if vertRen, ok := r.(verticalTextRenderer); ok {
			x := px.Min.X - 55
			y := px.Min.Y + px.H()/2
			vertRen.DrawTextVertical(ax.YLabel, geom.Pt{X: x, Y: y}, labelSize, labelColor)
		} else {
			metrics := r.MeasureText(ax.YLabel, labelSize, ctx.RC.FontKey)
			x := px.Min.X - 60 - metrics.W/2
			y := px.Min.Y + px.H()/2 + metrics.H/2
			textRen.DrawText(ax.YLabel, geom.Pt{X: x, Y: y}, labelSize, labelColor)
		}
	}
}

// axesToPixel returns an affine mapping [0..1]^2 (axes space) -> pixel rect.
// This maps axes coordinates to pixel coordinates with Y-flip for mathematical orientation:
// - axes (0,0) -> pixel (px.Min.X, px.Max.Y) [bottom-left]
// - axes (1,1) -> pixel (px.Max.X, px.Min.Y) [top-right]
func axesToPixel(px geom.Rect) geom.Affine {
	sx := px.W()
	sy := -px.H() // Negative to flip Y-axis
	tx := px.Min.X
	ty := px.Max.Y // Start from bottom of pixel rect
	return geom.Affine{A: sx, D: sy, E: tx, F: ty}
}
