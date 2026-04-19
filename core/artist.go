package core

import (
	"math"
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

// OverlayArtist is an optional extension for artists that need to render
// outside the axes clip, such as annotations with labels offset from the plot.
type OverlayArtist interface {
	DrawOverlay(r render.Renderer, ctx *DrawContext)
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

	shareX *Axes
	shareY *Axes
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
func (a *Axes) SetXLim(minVal, maxVal float64) {
	target := a.xScaleRoot()
	target.XScale = transform.NewLinear(minVal, maxVal)
}

// SetYLim sets the y-axis limits.
func (a *Axes) SetYLim(minVal, maxVal float64) {
	target := a.yScaleRoot()
	target.YScale = transform.NewLinear(minVal, maxVal)
}

// SetXLimLog sets the x-axis to logarithmic scale with given limits.
func (a *Axes) SetXLimLog(minVal, maxVal, base float64) {
	target := a.xScaleRoot()
	target.XScale = transform.NewLog(minVal, maxVal, base)
	if target.XAxis != nil {
		target.XAxis.Locator = LogLocator{Base: base, Minor: false}
		target.XAxis.Formatter = LogFormatter{Base: base}
	}
}

// SetYLimLog sets the y-axis to logarithmic scale with given limits.
func (a *Axes) SetYLimLog(minVal, maxVal, base float64) {
	target := a.yScaleRoot()
	target.YScale = transform.NewLog(minVal, maxVal, base)
	if target.YAxis != nil {
		target.YAxis.Locator = LogLocator{Base: base, Minor: false}
		target.YAxis.Formatter = LogFormatter{Base: base}
	}
}

// AutoScale computes axis limits from the data bounds of all artists,
// adding the specified margin fraction on each side (e.g. 0.05 = 5%).
// A margin of 0 fits exactly to the data. If no artists have non-zero bounds,
// limits remain unchanged.
func (a *Axes) AutoScale(margin float64) {
	targetX := a.xScaleRoot()
	targetY := a.yScaleRoot()

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

	targetX.XScale = transform.NewLinear(xMin, xMax)
	targetY.YScale = transform.NewLinear(yMin, yMax)
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
	minPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Min.X, Y: f.SizePx.Y * a.RectFraction.Min.Y}
	maxPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Max.X, Y: f.SizePx.Y * a.RectFraction.Max.Y}
	return geom.Rect{Min: minPt, Max: maxPt}
}

func (a *Axes) xScaleRoot() *Axes {
	if a == nil {
		return nil
	}
	cur := a
	for cur.shareX != nil {
		cur = cur.shareX
		if cur == nil {
			return a
		}
	}
	return cur
}

func (a *Axes) yScaleRoot() *Axes {
	if a == nil {
		return nil
	}
	cur := a
	for cur.shareY != nil {
		cur = cur.shareY
		if cur == nil {
			return a
		}
	}
	return cur
}

func (a *Axes) effectiveXAxis() *Axis {
	if a.shareX != nil {
		return a.shareX.effectiveXAxis()
	}
	return a.XAxis
}

func (a *Axes) effectiveYAxis() *Axis {
	if a.shareY != nil {
		return a.shareY.effectiveYAxis()
	}
	return a.YAxis
}

func (a *Axes) effectiveXScale() transform.Scale {
	if a.shareX != nil {
		return a.shareX.effectiveXScale()
	}
	return a.XScale
}

func (a *Axes) effectiveYScale() transform.Scale {
	if a.shareY != nil {
		return a.shareY.effectiveYScale()
	}
	return a.YScale
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
		xAxis := ax.effectiveXAxis()
		yAxis := ax.effectiveYAxis()

		// Build DrawContext with composed transform
		ctx := &DrawContext{
			DataToPixel: Transform2D{
				XScale:      ax.effectiveXScale(),
				YScale:      ax.effectiveYScale(),
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

		if xAxis != nil {
			xAxis.Draw(r, ctx)
		}
		if yAxis != nil {
			yAxis.Draw(r, ctx)
		}
		if ax.ShowFrame {
			ref := xAxis
			if ref == nil {
				ref = yAxis
			}
			DrawFrame(r, ctx, ref)
		}
		r.Restore()

		setRendererResolution(r, ctx.RC.DPI)
		for _, art := range ax.Artists {
			if overlay, ok := art.(OverlayArtist); ok {
				overlay.DrawOverlay(r, ctx)
			}
		}

		// Draw ticks (outward), tick labels, and text labels outside the clip rect
		if xAxis != nil {
			xAxis.DrawTicks(r, ctx)
			xAxis.DrawTickLabels(r, ctx)
		}
		if yAxis != nil {
			yAxis.DrawTicks(r, ctx)
			yAxis.DrawTickLabels(r, ctx)
		}
		drawAxesLabels(ax, r, ctx, px)
	}
}

func setRendererResolution(r render.Renderer, dpi float64) {
	if dpi <= 0 {
		return
	}
	if setter, ok := r.(render.DPIAware); ok {
		setter.SetResolution(uint(math.Round(dpi)))
	}
}

// drawAxesLabels renders title, xlabel, and ylabel outside the clipped axes area.
func drawAxesLabels(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	labelColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	titleSize := ctx.RC.FontSize
	labelSize := ctx.RC.FontSize * 0.97
	if labelSize < 8 {
		labelSize = 8
	}

	// Title: centered above the plot
	if ax.Title != "" {
		metrics := r.MeasureText(ax.Title, titleSize, ctx.RC.FontKey)
		x := px.Min.X + (px.W()-metrics.W)/2
		// Matplotlib positions axes titles at y=1.0 with a baseline-aligned
		// transform offset by axes.titlepad (default 6 pt).
		titlePadPx := ctx.RC.DPI * 6.0 / 72.0
		y := px.Min.Y - titlePadPx
		textRen.DrawText(ax.Title, geom.Pt{X: x, Y: y}, titleSize, labelColor)
	}

	// XLabel: centered below the x-axis tick labels
	if ax.XLabel != "" {
		metrics := r.MeasureText(ax.XLabel, labelSize, ctx.RC.FontKey)
		x := px.Min.X + (px.W()-metrics.W)/2
		// Same baseline rule as the title: place the baseline below the axes
		// using the ascent so the glyph box sits in the intended margin.
		y := px.Max.Y + 18 + metrics.Ascent
		textRen.DrawText(ax.XLabel, geom.Pt{X: x, Y: y}, labelSize, labelColor)
	}

	// YLabel: vertical text if supported, else horizontal fallback
	if ax.YLabel != "" {
		center := geom.Pt{X: px.Min.X - 36, Y: px.Min.Y + px.H()/2}
		switch ren := r.(type) {
		case render.RotatedTextDrawer:
			ren.DrawTextRotated(ax.YLabel, center, labelSize, math.Pi/2, labelColor)
		case render.VerticalTextDrawer:
			ren.DrawTextVertical(ax.YLabel, center, labelSize, labelColor)
		default:
			metrics := r.MeasureText(ax.YLabel, labelSize, ctx.RC.FontKey)
			x := px.Min.X - 42 - metrics.W/2
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
