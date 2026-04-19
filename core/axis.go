package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// AxisSide specifies which side of the plot area an axis is on.
type AxisSide uint8

const (
	AxisBottom AxisSide = iota // x-axis at bottom
	AxisTop                    // x-axis at top
	AxisLeft                   // y-axis at left
	AxisRight                  // y-axis at right
)

// Matplotlib defaults axes.linewidth to 0.8 pt. The reference images are
// rasterized at 100 DPI, so the equivalent stroke is about 1.11 px.
const defaultAxisLineWidth = 0.8 * 100.0 / 72.0

// Axis renders axis spines, ticks, and labels for a single dimension.
type Axis struct {
	Side          AxisSide     // which side of the plot
	Locator       Locator      // major tick position calculator
	MinorLocator  Locator      // minor tick position calculator (nil = no minor ticks)
	Formatter     Formatter    // tick label formatter
	Color         render.Color // axis line and tick color
	LineWidth     float64      // width of axis line and ticks
	TickSize      float64      // length of major tick marks (in pixels)
	MinorTickSize float64      // length of minor tick marks (in pixels); 0 uses TickSize*0.6
	MajorTickCount int         // target major tick count for automatic locators
	MinorTickCount int         // target minor tick count for automatic locators
	ShowSpine     bool         // whether to draw the axis line
	ShowTicks     bool         // whether to draw tick marks
	ShowLabels    bool         // whether to draw tick labels
	z             float64      // z-order
}

// NewXAxis creates an axis for the bottom (x-axis).
func NewXAxis() *Axis {
	return &Axis{
		Side:       AxisBottom,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 3},
		Color:      render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:  defaultAxisLineWidth,
		TickSize:   5.0,
		MajorTickCount: 6,
		MinorTickCount: 30,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: true,
	}
}

// NewYAxis creates an axis for the left (y-axis).
func NewYAxis() *Axis {
	return &Axis{
		Side:       AxisLeft,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 3},
		Color:      render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:  defaultAxisLineWidth,
		TickSize:   5.0,
		MajorTickCount: 6,
		MinorTickCount: 30,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: true,
	}
}

// Draw renders the axis spine on the axes edge.
func (a *Axis) Draw(r render.Renderer, ctx *DrawContext) {
	if a.ShowSpine {
		a.drawSpine(r, ctx)
	}
}

// DrawTicks renders tick marks pointing outward from the plot area.
// Called outside the clip region so ticks are visible.
func (a *Axis) DrawTicks(r render.Renderer, ctx *DrawContext) {
	if !a.ShowTicks {
		return
	}

	var domainMin, domainMax float64
	var isXAxis bool

	switch a.Side {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	// Minor ticks first
	if a.MinorLocator != nil {
		minorTicks := visibleTicks(a.MinorLocator.Ticks(domainMin, domainMax, a.minorTickTargetCount()), domainMin, domainMax)
		if len(minorTicks) > 0 {
			a.drawMinorTicks(r, ctx, minorTicks, isXAxis)
		}
	}

	// Major ticks
	ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCount()), domainMin, domainMax)
	if len(ticks) > 0 {
		a.drawTicks(r, ctx, ticks, isXAxis)
	}
}

// drawSpine draws the main axis line directly in pixel space, centered on the
// axes edge so the stroke can extend on both sides like Matplotlib's spines.
func (a *Axis) drawSpine(r render.Renderer, ctx *DrawContext) {
	px := ctx.Clip
	lw := a.LineWidth
	p1, p2 := spinePixelEndpoints(a.Side, px)

	paint := render.Paint{
		LineWidth: lw,
		Stroke:    a.Color,
		LineCap:   render.CapSquare,
		LineJoin:  render.JoinMiter,
	}
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}
	r.Path(path, &paint)
}

// spinePixelEndpoints returns the two pixel-space endpoints for a spine on the
// given side of px. AGG strokes align crisply when centered on pixel centers,
// so boundaries are shifted by 0.5 px instead of landing on pixel edges.
func spinePixelEndpoints(side AxisSide, px geom.Rect) (geom.Pt, geom.Pt) {
	x1 := math.Round(px.Min.X) + 0.5
	y1 := math.Round(px.Min.Y) + 0.5
	x2 := math.Round(px.Max.X) + 0.5
	y2 := math.Round(px.Max.Y) + 0.5

	switch side {
	case AxisBottom:
		return geom.Pt{X: x1, Y: y2}, geom.Pt{X: x2, Y: y2}
	case AxisTop:
		return geom.Pt{X: x1, Y: y1}, geom.Pt{X: x2, Y: y1}
	case AxisLeft:
		return geom.Pt{X: x1, Y: y1}, geom.Pt{X: x1, Y: y2}
	case AxisRight:
		return geom.Pt{X: x2, Y: y1}, geom.Pt{X: x2, Y: y2}
	}
	return geom.Pt{}, geom.Pt{}
}

// drawTicks draws tick marks at the specified positions.
func (a *Axis) drawTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, a.TickSize, isXAxis)
	}
}

// drawMinorTicks draws smaller tick marks at the specified positions.
func (a *Axis) drawMinorTicks(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	sz := a.MinorTickSize
	if sz <= 0 {
		sz = a.TickSize * 0.6
	}
	for _, tickValue := range ticks {
		a.drawSingleTick(r, ctx, tickValue, sz, isXAxis)
	}
}

// drawSingleTick draws a single tick mark pointing outward from the plot area.
func (a *Axis) drawSingleTick(r render.Renderer, ctx *DrawContext, tickValue, tickSize float64, isXAxis bool) {
	var p1, p2 geom.Pt

	if isXAxis {
		spineY := getSpinePosition(a.Side, ctx)
		spinePixel := ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineY})
		spinePixel.X = math.Round(spinePixel.X) + 0.5
		spinePixel.Y = math.Round(spinePixel.Y) + 0.5

		switch a.Side {
		case AxisBottom:
			// Bottom spine: ticks point downward (positive Y in pixel space = outward)
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X, Y: spinePixel.Y + tickSize}
		case AxisTop:
			// Top spine: ticks point upward (negative Y = outward)
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X, Y: spinePixel.Y - tickSize}
		}
	} else {
		spineX := getSpinePosition(a.Side, ctx)
		spinePixel := ctx.DataToPixel.Apply(geom.Pt{X: spineX, Y: tickValue})
		spinePixel.X = math.Round(spinePixel.X) + 0.5
		spinePixel.Y = math.Round(spinePixel.Y) + 0.5

		switch a.Side {
		case AxisLeft:
			// Left spine: ticks point leftward (negative X = outward)
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X - tickSize, Y: spinePixel.Y}
		case AxisRight:
			// Right spine: ticks point rightward (positive X = outward)
			p1 = spinePixel
			p2 = geom.Pt{X: spinePixel.X + tickSize, Y: spinePixel.Y}
		}
	}

	// Create tick path
	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	// Draw the tick
	paint := render.Paint{
		LineWidth: a.LineWidth,
		Stroke:    a.Color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}
	r.Path(path, &paint)
}

// DrawFrame draws fallback top/right border lines when those sides are not
// represented by explicit axes.
func DrawFrame(r render.Renderer, ctx *DrawContext, ref *Axis, drawTop, drawRight bool) {
	if ref == nil || !ref.ShowSpine {
		return
	}
	if !drawTop && !drawRight {
		return
	}
	paint := render.Paint{
		LineWidth: ref.LineWidth,
		Stroke:    ref.Color,
		LineCap:   render.CapSquare,
		LineJoin:  render.JoinMiter,
	}
	drawLine := func(p1, p2 geom.Pt) {
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{p1, p2},
		}
		r.Path(path, &paint)
	}

	if drawTop {
		p1, p2 := spinePixelEndpoints(AxisTop, ctx.Clip)
		drawLine(p1, p2)
	}
	if drawRight {
		p1, p2 := spinePixelEndpoints(AxisRight, ctx.Clip)
		drawLine(p1, p2)
	}
}

// getSpinePosition returns the data coordinate where the spine should be drawn.
func getSpinePosition(side AxisSide, ctx *DrawContext) float64 {
	switch side {
	case AxisBottom, AxisTop:
		// For x-axis, spine is at y coordinate
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		if side == AxisBottom {
			return yMin // bottom of plot
		}
		return yMax // top of plot
	case AxisLeft, AxisRight:
		// For y-axis, spine is at x coordinate
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		if side == AxisLeft {
			return xMin // left of plot
		}
		return xMax // right of plot
	}
	return 0
}

// Z returns the z-order for sorting.
func (a *Axis) Z() float64 {
	return a.z
}

// Bounds returns an empty rect for now.
func (a *Axis) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

// DrawTickLabels draws tick labels outside the clip region (call after r.Restore()).
func (a *Axis) DrawTickLabels(r render.Renderer, ctx *DrawContext) {
	if !a.ShowLabels {
		return
	}
	var domainMin, domainMax float64
	var isXAxis bool
	switch a.Side {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
		isXAxis = false
	}
	ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCount()), domainMin, domainMax)
	a.drawTickLabels(r, ctx, ticks, isXAxis)
}

func (a *Axis) majorTickTargetCount() int {
	if a == nil || a.MajorTickCount <= 0 {
		return 6
	}
	return a.MajorTickCount
}

func (a *Axis) minorTickTargetCount() int {
	if a == nil || a.MinorTickCount <= 0 {
		return 30
	}
	return a.MinorTickCount
}

func visibleTicks(ticks []float64, minVal, maxVal float64) []float64 {
	if len(ticks) == 0 {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	span := maxVal - minVal
	tol := 1e-9 * math.Max(1, span)

	out := make([]float64, 0, len(ticks))
	for _, tick := range ticks {
		if tick < minVal-tol || tick > maxVal+tol {
			continue
		}
		if approx(tick, minVal, tol) {
			tick = minVal
		} else if approx(tick, maxVal, tol) {
			tick = maxVal
		}
		out = append(out, tick)
	}
	return out
}

// drawTickLabels draws text labels for the ticks if the renderer supports text.
func (a *Axis) drawTickLabels(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	fontSize := tickLabelFontSize(ctx)
	labelPadPx := tickLabelPadPx(a, ctx)

	step := 0.0
	if len(ticks) >= 2 {
		step = ticks[1] - ticks[0]
	}

	for _, tickValue := range ticks {
		label := a.Formatter.Format(tickValue)
		if scalarFormatter, ok := a.Formatter.(ScalarFormatter); ok && step > 0 {
			label = formatScalarTickLabel(scalarFormatter, tickValue, step)
		}
		if label == "" {
			continue
		}

		// Measure text for centering
		metrics := r.MeasureText(label, fontSize, ctx.RC.FontKey)
		bounds, haveBounds := measureTextBounds(r, label, fontSize, ctx.RC.FontKey)

		labelPos, ok := tickLabelOrigin(a, ctx, tickValue, metrics, bounds, haveBounds, labelPadPx, isXAxis)
		if !ok {
			continue
		}

		textRen.DrawText(label, labelPos, fontSize, a.Color)
	}
}

func tickLabelFontSize(ctx *DrawContext) float64 {
	const mediumOverLarge = 10.0 / 12.0

	fontSize := 12.0
	if ctx != nil && ctx.RC.FontSize > 0 {
		fontSize = ctx.RC.FontSize * mediumOverLarge
	}
	if fontSize < 8 {
		return 8
	}
	return fontSize
}

func tickLabelPadPx(a *Axis, ctx *DrawContext) float64 {
	const majorLabelPadPt = 3.5

	padPx := majorLabelPadPt * 96.0 / 72.0
	if ctx != nil && ctx.RC.DPI > 0 {
		padPx = majorLabelPadPt * ctx.RC.DPI / 72.0
	}
	return a.TickSize + padPx
}

func tickLabelOrigin(a *Axis, ctx *DrawContext, tickValue float64, metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool, labelPadPx float64, isXAxis bool) (geom.Pt, bool) {
	if a == nil || ctx == nil {
		return geom.Pt{}, false
	}

	if isXAxis {
		spineY := getSpinePosition(a.Side, ctx)
		tickPos := ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineY})

		switch a.Side {
		case AxisBottom:
			return geom.Pt{
				X: tickPos.X - tickLabelCenterOffsetX(metrics, bounds, haveBounds),
				Y: tickPos.Y + labelPadPx + metrics.Ascent,
			}, true
		case AxisTop:
			return geom.Pt{
				X: tickPos.X - tickLabelCenterOffsetX(metrics, bounds, haveBounds),
				Y: tickPos.Y - labelPadPx - metrics.Descent,
			}, true
		default:
			return geom.Pt{}, false
		}
	}

	spineX := getSpinePosition(a.Side, ctx)
	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: spineX, Y: tickValue})

	switch a.Side {
	case AxisLeft:
		anchorX := tickPos.X - labelPadPx
		return geom.Pt{
			X: tickLabelBaselineForRight(anchorX, metrics, bounds, haveBounds),
			Y: tickPos.Y + (metrics.Ascent-metrics.Descent)/2,
		}, true
	case AxisRight:
		anchorX := tickPos.X + labelPadPx
		return geom.Pt{
			X: tickLabelBaselineForLeft(anchorX, bounds, haveBounds),
			Y: tickPos.Y + (metrics.Ascent-metrics.Descent)/2,
		}, true
	default:
		return geom.Pt{}, false
	}
}

func measureTextBounds(r render.Renderer, text string, size float64, fontKey string) (render.TextBounds, bool) {
	if bounder, ok := r.(render.TextBounder); ok {
		return bounder.MeasureTextBounds(text, size, fontKey)
	}
	return render.TextBounds{}, false
}

func textInkRect(origin geom.Pt, metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) (geom.Rect, bool) {
	if haveBounds && bounds.W > 0 && bounds.H > 0 {
		return geom.Rect{
			Min: geom.Pt{X: origin.X + bounds.X, Y: origin.Y + bounds.Y},
			Max: geom.Pt{X: origin.X + bounds.X + bounds.W, Y: origin.Y + bounds.Y + bounds.H},
		}, true
	}
	if metrics.W <= 0 || metrics.H <= 0 {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent},
		Max: geom.Pt{X: origin.X + metrics.W, Y: origin.Y - metrics.Ascent + metrics.H},
	}, true
}

func axisTickLabelBounds(a *Axis, r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || !a.ShowLabels || ctx == nil || a.Locator == nil || a.Formatter == nil {
		return geom.Rect{}, false
	}

	var (
		domainMin float64
		domainMax float64
		isXAxis   bool
	)

	switch a.Side {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	default:
		return geom.Rect{}, false
	}

	ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, 6), domainMin, domainMax)
	if len(ticks) == 0 {
		return geom.Rect{}, false
	}

	fontSize := tickLabelFontSize(ctx)
	labelPadPx := tickLabelPadPx(a, ctx)

	var (
		union geom.Rect
		have  bool
	)

	step := 0.0
	if len(ticks) >= 2 {
		step = ticks[1] - ticks[0]
	}

	for _, tickValue := range ticks {
		label := a.Formatter.Format(tickValue)
		if scalarFormatter, ok := a.Formatter.(ScalarFormatter); ok && step > 0 {
			label = formatScalarTickLabel(scalarFormatter, tickValue, step)
		}
		if label == "" {
			continue
		}

		metrics := r.MeasureText(label, fontSize, ctx.RC.FontKey)
		bounds, haveBounds := measureTextBounds(r, label, fontSize, ctx.RC.FontKey)
		origin, ok := tickLabelOrigin(a, ctx, tickValue, metrics, bounds, haveBounds, labelPadPx, isXAxis)
		if !ok {
			continue
		}
		inkRect, ok := textInkRect(origin, metrics, bounds, haveBounds)
		if !ok {
			continue
		}

		if !have {
			union = inkRect
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{
				X: math.Min(union.Min.X, inkRect.Min.X),
				Y: math.Min(union.Min.Y, inkRect.Min.Y),
			},
			Max: geom.Pt{
				X: math.Max(union.Max.X, inkRect.Max.X),
				Y: math.Max(union.Max.Y, inkRect.Max.Y),
			},
		}
	}

	return union, have
}

func tickLabelCenterOffsetX(metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) float64 {
	if haveBounds {
		return bounds.X + bounds.W/2
	}
	return metrics.W / 2
}

func tickLabelBaselineForRight(anchorX float64, metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) float64 {
	if haveBounds {
		return anchorX - (bounds.X + bounds.W)
	}
	return anchorX - metrics.W
}

func tickLabelBaselineForLeft(anchorX float64, bounds render.TextBounds, haveBounds bool) float64 {
	if haveBounds {
		return anchorX - bounds.X
	}
	return anchorX
}
