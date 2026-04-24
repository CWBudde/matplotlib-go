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

// TickLabelStyle captures axis-owned label placement and orientation.
type TickLabelStyle struct {
	Rotation  float64
	Pad       float64
	HAlign    TextAlign
	VAlign    TextVerticalAlign
	AutoAlign bool
}

// TickLevel adds an optional additional tick/label row to an axis.
type TickLevel struct {
	Locator    Locator
	Formatter  Formatter
	Size       float64
	ShowTicks  bool
	ShowLabels bool
	LabelStyle TickLabelStyle
}

// Axis renders axis spines, ticks, and labels for a single dimension.
type Axis struct {
	Side            AxisSide     // which side of the plot
	Locator         Locator      // major tick position calculator
	MinorLocator    Locator      // minor tick position calculator (nil = no minor ticks)
	Formatter       Formatter    // major tick label formatter
	MinorFormatter  Formatter    // optional minor tick label formatter
	Color           render.Color // axis line and tick color
	LineWidth       float64      // width of axis line and ticks
	TickSize        float64      // length of major tick marks (in pixels)
	MinorTickSize   float64      // length of minor tick marks (in pixels); 0 uses TickSize*0.6
	MajorTickCount  int          // target major tick count for automatic locators
	MinorTickCount  int          // target minor tick count for automatic locators
	ShowSpine       bool         // whether to draw the axis line
	ShowTicks       bool         // whether to draw major/minor tick marks
	ShowLabels      bool         // whether to draw major tick labels
	ShowMinorLabels bool         // whether to draw minor tick labels
	MajorLabelStyle TickLabelStyle
	MinorLabelStyle TickLabelStyle
	ExtraTickLevels []TickLevel
	z               float64 // z-order
}

// NewXAxis creates an axis for the bottom (x-axis).
func NewXAxis() *Axis {
	return &Axis{
		Side:            AxisBottom,
		Locator:         LinearLocator{},
		Formatter:       ScalarFormatter{Prec: 3},
		Color:           render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:       defaultAxisLineWidth,
		TickSize:        5.0,
		MajorTickCount:  6,
		MinorTickCount:  30,
		ShowSpine:       true,
		ShowTicks:       true,
		ShowLabels:      true,
		MajorLabelStyle: defaultTickLabelStyle(),
		MinorLabelStyle: defaultTickLabelStyle(),
	}
}

// NewYAxis creates an axis for the left (y-axis).
func NewYAxis() *Axis {
	return &Axis{
		Side:            AxisLeft,
		Locator:         LinearLocator{},
		Formatter:       ScalarFormatter{Prec: 3},
		Color:           render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:       defaultAxisLineWidth,
		TickSize:        5.0,
		MajorTickCount:  6,
		MinorTickCount:  30,
		ShowSpine:       true,
		ShowTicks:       true,
		ShowLabels:      true,
		MajorLabelStyle: defaultTickLabelStyle(),
		MinorLabelStyle: defaultTickLabelStyle(),
	}
}

// Draw renders the axis spine on the axes edge.
func (a *Axis) Draw(r render.Renderer, ctx *DrawContext) {
	if isPolarProjection(ctx.Projection) {
		a.drawPolarSpine(r, ctx)
		return
	}
	if a.ShowSpine {
		a.drawSpine(r, ctx)
	}
}

// DrawTicks renders tick marks pointing outward from the plot area.
// Called outside the clip region so ticks are visible.
func (a *Axis) DrawTicks(r render.Renderer, ctx *DrawContext) {
	if isPolarProjection(ctx.Projection) {
		a.drawPolarTicks(r, ctx)
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

	if a.ShowTicks {
		// Minor ticks first
		if a.MinorLocator != nil {
			minorTicks := visibleTicks(a.MinorLocator.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
			if len(minorTicks) > 0 {
				a.drawMinorTicks(r, ctx, minorTicks, isXAxis)
			}
		}

		// Major ticks
		ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		if len(ticks) > 0 {
			a.drawTicks(r, ctx, ticks, isXAxis)
		}
	}

	for _, level := range a.ExtraTickLevels {
		if !level.ShowTicks || level.Locator == nil {
			continue
		}
		ticks := visibleTicks(level.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		if len(ticks) == 0 {
			continue
		}
		size := tickLevelSize(level, a.TickSize)
		for _, tickValue := range ticks {
			a.drawSingleTick(r, ctx, tickValue, size, isXAxis)
		}
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
	if isPolarProjection(ctx.Projection) {
		a.drawPolarTickLabels(r, ctx)
		return
	}
	if !a.ShowLabels && !a.ShowMinorLabels && len(a.ExtraTickLevels) == 0 {
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
	if a.ShowLabels && a.Locator != nil && a.Formatter != nil {
		ticks := visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		a.drawTickLabels(r, ctx, ticks, a.Formatter, a.MajorLabelStyle, a.TickSize, isXAxis)
	}
	if a.ShowMinorLabels && a.MinorLocator != nil && a.MinorFormatter != nil {
		ticks := visibleTicks(a.MinorLocator.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		a.drawTickLabels(r, ctx, ticks, a.MinorFormatter, a.MinorLabelStyle, a.minorTickSize(), isXAxis)
	}
	for _, level := range a.ExtraTickLevels {
		if !level.ShowLabels || level.Locator == nil || level.Formatter == nil {
			continue
		}
		ticks := visibleTicks(level.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax)
		a.drawTickLabels(r, ctx, ticks, level.Formatter, normalizeTickLabelStyle(level.LabelStyle), tickLevelSize(level, a.TickSize), isXAxis)
	}
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

func (a *Axis) majorTickTargetCountForContext(ctx *DrawContext, isXAxis bool) int {
	target := a.majorTickTargetCount()
	if ctx == nil || target <= 1 {
		return target
	}

	length := ctx.Clip.H()
	minSpacing := 42.0
	if isXAxis {
		length = ctx.Clip.W()
		minSpacing = 72.0
	}
	if length <= 0 {
		return target
	}

	capacity := int(math.Floor(length / minSpacing))
	if capacity < 1 {
		capacity = 1
	}
	if capacity < target {
		return capacity
	}
	return target
}

func (a *Axis) minorTickTargetCountForContext(ctx *DrawContext, isXAxis bool) int {
	target := a.minorTickTargetCount()
	major := a.majorTickTargetCountForContext(ctx, isXAxis)
	limit := major * 5
	if limit < 1 {
		limit = 1
	}
	if target > limit {
		return limit
	}
	return target
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

// drawTickLabels draws text labels for a single tick level if the renderer supports text.
func (a *Axis) drawTickLabels(r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64, isXAxis bool) {
	textRen, ok := r.(render.TextDrawer)
	if !ok || formatter == nil {
		return
	}

	fontSize := tickLabelFontSize(a, ctx)
	style = normalizeTickLabelStyle(style)
	labelPadPx := tickLabelPadForSize(tickSize, style, ctx)

	step := 0.0
	if len(ticks) >= 2 {
		step = ticks[1] - ticks[0]
	}

	var rotRen render.RotatedTextDrawer
	if drawer, ok := r.(render.RotatedTextDrawer); ok {
		rotRen = drawer
	}

	for i, tickValue := range ticks {
		label := formatTickLabel(formatter, tickValue, i, ticks)
		if scalarFormatter, ok := formatter.(ScalarFormatter); ok && step > 0 {
			label = formatScalarTickLabel(scalarFormatter, tickValue, step)
		}
		if label == "" {
			continue
		}

		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)

		labelPos, ok := tickLabelOrigin(a, ctx, tickValue, layout, labelPadPx, style, isXAxis)
		if !ok {
			continue
		}

		if style.Rotation != 0 && rotRen != nil {
			drawDisplayTextRotated(rotRen, label, tickLabelRotationAnchor(labelPos, layout), fontSize, style.Rotation*math.Pi/180.0, a.Color)
			continue
		}

		drawDisplayText(textRen, label, labelPos, fontSize, a.Color)
	}
}

func tickLabelFontSize(a *Axis, ctx *DrawContext) float64 {
	if ctx == nil {
		return 8
	}

	switch {
	case a != nil && (a.Side == AxisLeft || a.Side == AxisRight):
		return ctx.RC.TickLabelSize("y")
	default:
		return ctx.RC.TickLabelSize("x")
	}
}

func tickLabelPadPx(a *Axis, ctx *DrawContext) float64 {
	return tickLabelPadForSize(a.TickSize, a.MajorLabelStyle, ctx)
}

func tickLabelPadForSize(tickSize float64, style TickLabelStyle, ctx *DrawContext) float64 {
	const majorLabelPadPt = 3.5

	padPx := majorLabelPadPt * 96.0 / 72.0
	if ctx != nil && ctx.RC.DPI > 0 {
		padPx = majorLabelPadPt * ctx.RC.DPI / 72.0
	}
	if style.Pad > 0 {
		padPx = style.Pad
	}
	return tickSize + padPx
}

func tickLabelOrigin(a *Axis, ctx *DrawContext, tickValue float64, layout singleLineTextLayout, labelPadPx float64, style TickLabelStyle, isXAxis bool) (geom.Pt, bool) {
	if a == nil || ctx == nil {
		return geom.Pt{}, false
	}
	style = normalizeTickLabelStyle(style)

	if isXAxis {
		spineY := getSpinePosition(a.Side, ctx)
		tickPos := ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineY})
		hAlign, vAlign := resolvedTickLabelLayoutAlignments(a.Side, style, true)

		switch a.Side {
		case AxisBottom:
			anchor := geom.Pt{X: tickPos.X, Y: tickPos.Y + labelPadPx}
			return geom.Pt{
				X: anchor.X - tickLabelLeftOffset(hAlign, layout),
				Y: anchor.Y + textBaselineOffset(layout, vAlign),
			}, true
		case AxisTop:
			anchor := geom.Pt{X: tickPos.X, Y: tickPos.Y - labelPadPx}
			return geom.Pt{
				X: anchor.X - tickLabelLeftOffset(hAlign, layout),
				Y: anchor.Y + textBaselineOffset(layout, vAlign),
			}, true
		default:
			return geom.Pt{}, false
		}
	}

	spineX := getSpinePosition(a.Side, ctx)
	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: spineX, Y: tickValue})
	hAlign, vAlign := resolvedTickLabelLayoutAlignments(a.Side, style, false)

	switch a.Side {
	case AxisLeft:
		anchor := geom.Pt{X: tickPos.X - labelPadPx, Y: tickPos.Y}
		return geom.Pt{
			X: anchor.X - tickLabelLeftOffsetForLeftAxis(hAlign, layout),
			Y: anchor.Y + textBaselineOffset(layout, vAlign),
		}, true
	case AxisRight:
		anchor := geom.Pt{X: tickPos.X + labelPadPx, Y: tickPos.Y}
		return geom.Pt{
			X: anchor.X - tickLabelLeftOffsetForRightAxis(hAlign, layout),
			Y: anchor.Y + textBaselineOffset(layout, vAlign),
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

func textInkRect(origin geom.Pt, layout singleLineTextLayout) (geom.Rect, bool) {
	if layout.HaveInkBounds && layout.InkBounds.W > 0 && layout.InkBounds.H > 0 {
		return geom.Rect{
			Min: geom.Pt{X: origin.X + layout.InkBounds.X, Y: origin.Y + layout.InkBounds.Y},
			Max: geom.Pt{X: origin.X + layout.InkBounds.X + layout.InkBounds.W, Y: origin.Y + layout.InkBounds.Y + layout.InkBounds.H},
		}, true
	}
	if layout.Width <= 0 || layout.Height <= 0 {
		return geom.Rect{}, false
	}
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - layout.Ascent},
		Max: geom.Pt{X: origin.X + layout.Width, Y: origin.Y - layout.Ascent + layout.Height},
	}, true
}

func axisTickLabelBounds(a *Axis, r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || ctx == nil {
		return geom.Rect{}, false
	}
	if isPolarProjection(ctx.Projection) {
		return a.polarTickLabelBounds(r, ctx)
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

	var (
		union geom.Rect
		have  bool
	)

	type labelLevel struct {
		ticks     []float64
		formatter Formatter
		style     TickLabelStyle
		tickSize  float64
	}

	levels := make([]labelLevel, 0, 2+len(a.ExtraTickLevels))
	if a.ShowLabels && a.Locator != nil && a.Formatter != nil {
		levels = append(levels, labelLevel{
			ticks:     visibleTicks(a.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax),
			formatter: a.Formatter,
			style:     a.MajorLabelStyle,
			tickSize:  a.TickSize,
		})
	}
	if a.ShowMinorLabels && a.MinorLocator != nil && a.MinorFormatter != nil {
		levels = append(levels, labelLevel{
			ticks:     visibleTicks(a.MinorLocator.Ticks(domainMin, domainMax, a.minorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax),
			formatter: a.MinorFormatter,
			style:     a.MinorLabelStyle,
			tickSize:  a.minorTickSize(),
		})
	}
	for _, level := range a.ExtraTickLevels {
		if !level.ShowLabels || level.Locator == nil || level.Formatter == nil {
			continue
		}
		levels = append(levels, labelLevel{
			ticks:     visibleTicks(level.Locator.Ticks(domainMin, domainMax, a.majorTickTargetCountForContext(ctx, isXAxis)), domainMin, domainMax),
			formatter: level.Formatter,
			style:     normalizeTickLabelStyle(level.LabelStyle),
			tickSize:  tickLevelSize(level, a.TickSize),
		})
	}

	for _, level := range levels {
		bounds, ok := tickLabelBoundsForLevel(a, r, ctx, level.ticks, level.formatter, level.style, level.tickSize, isXAxis)
		if !ok {
			continue
		}

		if !have {
			union = bounds
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{
				X: math.Min(union.Min.X, bounds.Min.X),
				Y: math.Min(union.Min.Y, bounds.Min.Y),
			},
			Max: geom.Pt{
				X: math.Max(union.Max.X, bounds.Max.X),
				Y: math.Max(union.Max.Y, bounds.Max.Y),
			},
		}
	}

	return union, have
}

func (a *Axis) drawPolarSpine(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || !a.ShowSpine {
		return
	}

	center, radius := polarCenterAndRadius(ctx.Clip)
	paint := polarTickPaint(a.Color, a.LineWidth, nil)

	switch a.Side {
	case AxisBottom, AxisTop:
		path := polarProjectionFramePath(ctx.Projection, ctx.Clip)
		r.Path(path, &paint)
	case AxisLeft, AxisRight:
		labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)
		path := geom.Path{}
		path.MoveTo(center)
		path.LineTo(polarPixelPoint(center, radius, labelAngle))
		r.Path(path, &paint)
	}
}

func (a *Axis) drawPolarTicks(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || !a.ShowTicks {
		return
	}

	switch a.Side {
	case AxisBottom, AxisTop:
		a.drawPolarThetaTicks(r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, true), a.minorTickSize())
		a.drawPolarThetaTicks(r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, false), a.TickSize)
	case AxisLeft, AxisRight:
		a.drawPolarRadialTicks(r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, true), a.minorTickSize())
		a.drawPolarRadialTicks(r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, false), a.TickSize)
	}
}

func (a *Axis) drawPolarThetaTicks(r render.Renderer, ctx *DrawContext, ticks []float64, tickSize float64) {
	if len(ticks) == 0 || tickSize <= 0 {
		return
	}
	center, radius := polarCenterAndRadius(ctx.Clip)
	paint := polarTickPaint(a.Color, a.LineWidth, nil)

	for _, tick := range ticks {
		angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
		path := geom.Path{}
		path.MoveTo(polarPixelPoint(center, radius, angle))
		path.LineTo(polarPixelPoint(center, radius+tickSize, angle))
		r.Path(path, &paint)
	}
}

func (a *Axis) drawPolarRadialTicks(r render.Renderer, ctx *DrawContext, ticks []float64, tickSize float64) {
	if len(ticks) == 0 || tickSize <= 0 {
		return
	}
	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	paint := polarTickPaint(a.Color, a.LineWidth, nil)
	labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)

	for _, tick := range ticks {
		u := ctx.DataToPixel.YScale.Fwd(tick)
		radius := outerRadius * u
		if radius <= 0 {
			continue
		}
		span := tickSize / math.Max(radius, 1)
		segments := int(math.Max(float64(polarArcMinSegments), math.Ceil(span*24)))
		path := polarArcPath(center, radius, labelAngle-span/2, labelAngle+span/2, segments, false)
		r.Path(path, &paint)
	}
}

func (a *Axis) drawPolarTickLabels(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	switch a.Side {
	case AxisBottom, AxisTop:
		if a.ShowLabels {
			a.drawPolarThetaTickLabels(textRen, r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, false), a.Formatter, a.MajorLabelStyle, a.TickSize)
		}
		if a.ShowMinorLabels {
			a.drawPolarThetaTickLabels(textRen, r, ctx, polarThetaTicks(a, ctx.DataToPixel.XScale, true), a.MinorFormatter, a.MinorLabelStyle, a.minorTickSize())
		}
	case AxisLeft, AxisRight:
		if a.ShowLabels {
			a.drawPolarRadialTickLabels(textRen, r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, false), a.Formatter, a.MajorLabelStyle, a.TickSize)
		}
		if a.ShowMinorLabels {
			a.drawPolarRadialTickLabels(textRen, r, ctx, polarRadialTicks(a, ctx.DataToPixel.YScale, true), a.MinorFormatter, a.MinorLabelStyle, a.minorTickSize())
		}
	}
}

func (a *Axis) drawPolarThetaTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) {
	if formatter == nil || len(ticks) == 0 {
		return
	}

	center, radius := polarCenterAndRadius(ctx.Clip)
	fontSize := tickLabelFontSize(a, ctx)
	labelPadPx := tickLabelPadForSize(tickSize, normalizeTickLabelStyle(style), ctx)

	for i, tick := range ticks {
		label := formatTickLabel(formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
		angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
		anchor := polarPixelPoint(center, radius+labelPadPx, angle)
		hAlign, vAlign := polarTickLabelAlignments(angle)
		drawDisplayText(textRen, label, alignedSingleLineOrigin(anchor, layout, hAlign, vAlign), fontSize, a.Color)
	}
}

func (a *Axis) drawPolarRadialTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) {
	if formatter == nil || len(ticks) == 0 {
		return
	}

	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	fontSize := tickLabelFontSize(a, ctx)
	labelPadPx := tickLabelPadForSize(tickSize, normalizeTickLabelStyle(style), ctx)
	labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)

	for i, tick := range ticks {
		label := formatTickLabel(formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
		radius := outerRadius * ctx.DataToPixel.YScale.Fwd(tick)
		anchor := polarPixelPoint(center, radius+labelPadPx, labelAngle)
		hAlign, vAlign := polarTickLabelAlignments(labelAngle)
		drawDisplayText(textRen, label, alignedSingleLineOrigin(anchor, layout, hAlign, vAlign), fontSize, a.Color)
	}
}

func (a *Axis) polarTickLabelBounds(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || ctx == nil {
		return geom.Rect{}, false
	}

	type polarLabel struct {
		ticks    []float64
		format   Formatter
		style    TickLabelStyle
		tickSize float64
	}

	levels := []polarLabel{}
	if a.ShowLabels {
		levels = append(levels, polarLabel{
			ticks: func() []float64 {
				if a.Side == AxisBottom || a.Side == AxisTop {
					return polarThetaTicks(a, ctx.DataToPixel.XScale, false)
				}
				return polarRadialTicks(a, ctx.DataToPixel.YScale, false)
			}(),
			format:   a.Formatter,
			style:    a.MajorLabelStyle,
			tickSize: a.TickSize,
		})
	}
	if a.ShowMinorLabels {
		levels = append(levels, polarLabel{
			ticks: func() []float64 {
				if a.Side == AxisBottom || a.Side == AxisTop {
					return polarThetaTicks(a, ctx.DataToPixel.XScale, true)
				}
				return polarRadialTicks(a, ctx.DataToPixel.YScale, true)
			}(),
			format:   a.MinorFormatter,
			style:    a.MinorLabelStyle,
			tickSize: a.minorTickSize(),
		})
	}

	var (
		union geom.Rect
		have  bool
	)

	for _, level := range levels {
		if level.format == nil || len(level.ticks) == 0 {
			continue
		}
		bounds, ok := a.polarTickLabelBoundsForLevel(r, ctx, level.ticks, level.format, level.style, level.tickSize)
		if !ok {
			continue
		}
		if !have {
			union = bounds
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{X: math.Min(union.Min.X, bounds.Min.X), Y: math.Min(union.Min.Y, bounds.Min.Y)},
			Max: geom.Pt{X: math.Max(union.Max.X, bounds.Max.X), Y: math.Max(union.Max.Y, bounds.Max.Y)},
		}
	}

	return union, have
}

func (a *Axis) polarTickLabelBoundsForLevel(r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64) (geom.Rect, bool) {
	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	fontSize := tickLabelFontSize(a, ctx)
	labelPadPx := tickLabelPadForSize(tickSize, normalizeTickLabelStyle(style), ctx)
	labelAngle := polarRadialLabelAngleForProjection(ctx.Projection)

	var (
		union geom.Rect
		have  bool
	)

	for i, tick := range ticks {
		label := formatTickLabel(formatter, tick, i, ticks)
		if label == "" {
			continue
		}
		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)

		var (
			anchor geom.Pt
			hAlign TextAlign
			vAlign textLayoutVerticalAlign
		)

		if a.Side == AxisBottom || a.Side == AxisTop {
			angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
			anchor = polarPixelPoint(center, outerRadius+labelPadPx, angle)
			hAlign, vAlign = polarTickLabelAlignments(angle)
		} else {
			radius := outerRadius * ctx.DataToPixel.YScale.Fwd(tick)
			anchor = polarPixelPoint(center, radius+labelPadPx, labelAngle)
			hAlign, vAlign = polarTickLabelAlignments(labelAngle)
		}

		origin := alignedSingleLineOrigin(anchor, layout, hAlign, vAlign)
		inkRect, ok := textInkRect(origin, layout)
		if !ok {
			continue
		}
		if !have {
			union = inkRect
			have = true
			continue
		}
		union = geom.Rect{
			Min: geom.Pt{X: math.Min(union.Min.X, inkRect.Min.X), Y: math.Min(union.Min.Y, inkRect.Min.Y)},
			Max: geom.Pt{X: math.Max(union.Max.X, inkRect.Max.X), Y: math.Max(union.Max.Y, inkRect.Max.Y)},
		}
	}

	return union, have
}

func tickLabelBoundsForLevel(a *Axis, r render.Renderer, ctx *DrawContext, ticks []float64, formatter Formatter, style TickLabelStyle, tickSize float64, isXAxis bool) (geom.Rect, bool) {
	if len(ticks) == 0 || formatter == nil {
		return geom.Rect{}, false
	}

	fontSize := tickLabelFontSize(a, ctx)
	style = normalizeTickLabelStyle(style)
	labelPadPx := tickLabelPadForSize(tickSize, style, ctx)

	var (
		union geom.Rect
		have  bool
	)

	step := 0.0
	if len(ticks) >= 2 {
		step = ticks[1] - ticks[0]
	}

	for i, tickValue := range ticks {
		label := formatTickLabel(formatter, tickValue, i, ticks)
		if scalarFormatter, ok := formatter.(ScalarFormatter); ok && step > 0 {
			label = formatScalarTickLabel(scalarFormatter, tickValue, step)
		}
		if label == "" {
			continue
		}

		layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
		origin, ok := tickLabelOrigin(a, ctx, tickValue, layout, labelPadPx, style, isXAxis)
		if !ok {
			continue
		}
		inkRect, ok := textInkRect(origin, layout)
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

func tickLabelCenterOffsetX(layout singleLineTextLayout) float64 {
	if layout.HaveInkBounds {
		return layout.InkBounds.X + layout.InkBounds.W/2
	}
	return layout.Width / 2
}

func tickLabelLeftOffset(hAlign TextAlign, layout singleLineTextLayout) float64 {
	switch hAlign {
	case TextAlignLeft:
		if layout.HaveInkBounds {
			return -layout.InkBounds.X
		}
		return 0
	case TextAlignRight:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W
		}
		return layout.Width
	default:
		return tickLabelCenterOffsetX(layout)
	}
}

func tickLabelLeftOffsetForLeftAxis(hAlign TextAlign, layout singleLineTextLayout) float64 {
	switch hAlign {
	case TextAlignLeft:
		if layout.HaveInkBounds {
			return layout.InkBounds.X
		}
		return 0
	case TextAlignCenter:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W/2
		}
		return layout.Width / 2
	default:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W
		}
		return layout.Width
	}
}

func tickLabelLeftOffsetForRightAxis(hAlign TextAlign, layout singleLineTextLayout) float64 {
	switch hAlign {
	case TextAlignRight:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W
		}
		return layout.Width
	case TextAlignCenter:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W/2
		}
		return layout.Width / 2
	default:
		if layout.HaveInkBounds {
			return layout.InkBounds.X
		}
		return 0
	}
}

func tickLabelRotationAnchor(origin geom.Pt, layout singleLineTextLayout) geom.Pt {
	left := origin.X
	if layout.HaveInkBounds {
		left += layout.InkBounds.X
	}
	width := tickLabelWidth(layout)
	top := origin.Y + layout.InkBounds.Y
	if !layout.HaveInkBounds {
		top = origin.Y - layout.Ascent
	}
	return geom.Pt{X: left + width/2, Y: top + tickLabelHeight(layout)}
}

func tickLabelWidth(layout singleLineTextLayout) float64 {
	if layout.HaveInkBounds && layout.InkBounds.W > 0 {
		return layout.InkBounds.W
	}
	return layout.Width
}

func tickLabelHeight(layout singleLineTextLayout) float64 {
	if layout.HaveInkBounds && layout.InkBounds.H > 0 {
		return layout.InkBounds.H
	}
	return layout.Height
}

func resolvedTickLabelAlignments(side AxisSide, style TickLabelStyle, isXAxis bool) (TextAlign, TextVerticalAlign) {
	if !style.AutoAlign {
		return style.HAlign, style.VAlign
	}
	if isXAxis {
		switch side {
		case AxisBottom:
			return TextAlignCenter, TextVAlignTop
		case AxisTop:
			return TextAlignCenter, TextVAlignBottom
		}
	}
	switch side {
	case AxisLeft:
		return TextAlignRight, TextVAlignMiddle
	case AxisRight:
		return TextAlignLeft, TextVAlignMiddle
	default:
		return TextAlignCenter, TextVAlignMiddle
	}
}

func resolvedTickLabelLayoutAlignments(side AxisSide, style TickLabelStyle, isXAxis bool) (TextAlign, textLayoutVerticalAlign) {
	hAlign, vAlign := resolvedTickLabelAlignments(side, style, isXAxis)
	if isXAxis {
		return hAlign, layoutVerticalAlign(vAlign, false)
	}
	return hAlign, layoutVerticalAlign(vAlign, true)
}

func defaultTickLabelStyle() TickLabelStyle {
	return TickLabelStyle{AutoAlign: true}
}

func normalizeTickLabelStyle(style TickLabelStyle) TickLabelStyle {
	if !style.AutoAlign && style.HAlign == TextAlignLeft && style.VAlign == TextVAlignBaseline && style.Pad == 0 && style.Rotation == 0 {
		style.AutoAlign = true
	}
	return style
}

func tickLevelSize(level TickLevel, fallback float64) float64 {
	if level.Size > 0 {
		return level.Size
	}
	return fallback
}

func (a *Axis) minorTickSize() float64 {
	if a.MinorTickSize > 0 {
		return a.MinorTickSize
	}
	return a.TickSize * 0.6
}

func (a *Axis) AddTickLevel(level TickLevel) {
	if a == nil {
		return
	}
	level.LabelStyle = normalizeTickLabelStyle(level.LabelStyle)
	a.ExtraTickLevels = append(a.ExtraTickLevels, level)
}

func (a *Axis) ClearTickLevels() {
	if a == nil {
		return
	}
	a.ExtraTickLevels = nil
}
