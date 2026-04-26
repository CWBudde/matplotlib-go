package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

// Grid renders grid lines at tick positions.
type Grid struct {
	Axis           AxisSide     // which axis to use for tick positions
	Color          render.Color // major grid line color
	LineWidth      float64      // width of major grid lines
	Dashes         []float64    // dash pattern for major grid (nil = solid)
	Alpha          float64      // alpha override (0-1), if 0 uses Color.A
	Major          bool         // draw grid at major ticks
	Minor          bool         // draw grid at minor ticks
	MinorColor     render.Color // minor grid line color (zero value uses Color with lower alpha)
	MinorLineWidth float64      // width of minor grid lines (0 uses LineWidth*0.5)
	MinorDashes    []float64    // dash pattern for minor grid (nil = solid)
	Locator        Locator      // major tick locator (nil = LinearLocator)
	MinorLocator   Locator      // minor tick locator (nil = MinorLinearLocator{N:5})
	z              float64      // z-order (should be behind data)
}

// NewGrid creates a new grid for the specified axis.
func NewGrid(axis AxisSide) *Grid {
	rc := style.CurrentDefaults()
	return &Grid{
		Axis:           axis,
		Color:          rc.GridColor,
		LineWidth:      rc.GridLineWidth,
		MinorColor:     rc.MinorGridColor,
		MinorLineWidth: rc.MinorGridLineWidth,
		Alpha:          0, // use Color.A
		Major:          true,
		Minor:          false,
		z:              -1000, // behind everything else
	}
}

// Draw renders grid lines at tick positions.
func (g *Grid) Draw(r render.Renderer, ctx *DrawContext) {
	if !g.Major && !g.Minor {
		return
	}
	if isPolarProjection(ctx.Projection) {
		g.drawPolar(r, ctx)
		return
	}
	if isGeoProjection(ctx.Projection) {
		g.drawGeo(r, ctx)
		return
	}
	if projectionGridUsesSampling(ctx) {
		g.drawCurvelinear(r, ctx)
		return
	}

	var domainMin, domainMax float64
	var isXAxis bool

	switch g.Axis {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	// Draw minor grid first (behind major)
	if g.Minor {
		minorLoc := g.MinorLocator
		if minorLoc == nil {
			minorLoc = MinorLinearLocator{N: 5}
		}
		minorTicks := minorLoc.Ticks(domainMin, domainMax, 30)

		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}

		for _, v := range minorTicks {
			g.drawLine(r, ctx, v, isXAxis, minorColor, minorWidth, g.MinorDashes)
		}
	}

	// Draw major grid
	if g.Major {
		loc := g.Locator
		if loc == nil {
			loc = LinearLocator{}
		}
		ticks := loc.Ticks(domainMin, domainMax, 8)

		for _, v := range ticks {
			g.drawLine(r, ctx, v, isXAxis, majorColor, g.LineWidth, g.Dashes)
		}
	}
}

func (g *Grid) drawGeo(r render.Renderer, ctx *DrawContext) {
	var domainMin, domainMax float64
	switch g.Axis {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	if g.Minor {
		minorLoc := g.MinorLocator
		if minorLoc == nil {
			minorLoc = MinorLinearLocator{N: 2}
		}
		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}
		paint := polarTickPaint(minorColor, minorWidth, g.MinorDashes)
		for _, tick := range visibleTicks(minorLoc.Ticks(domainMin, domainMax, 20), domainMin, domainMax) {
			drawGeoGridLine(r, ctx, g.Axis, tick, paint)
		}
	}

	if g.Major {
		loc := g.Locator
		if loc == nil {
			switch g.Axis {
			case AxisBottom, AxisTop:
				loc = ctx.Axes.effectiveXAxis().Locator
			default:
				loc = ctx.Axes.effectiveYAxis().Locator
			}
		}
		paint := polarTickPaint(majorColor, g.LineWidth, g.Dashes)
		for _, tick := range visibleTicks(loc.Ticks(domainMin, domainMax, 8), domainMin, domainMax) {
			drawGeoGridLine(r, ctx, g.Axis, tick, paint)
		}
	}
}

func (g *Grid) drawCurvelinear(r render.Renderer, ctx *DrawContext) {
	var domainMin, domainMax float64
	switch g.Axis {
	case AxisBottom, AxisTop:
		domainMin, domainMax = ctx.DataToPixel.XScale.Domain()
	case AxisLeft, AxisRight:
		domainMin, domainMax = ctx.DataToPixel.YScale.Domain()
	}

	axis := g.axisForContext(ctx)
	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	if g.Minor {
		minorLoc := g.curvelinearLocator(axis, true)
		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}
		paint := polarTickPaint(minorColor, minorWidth, g.MinorDashes)
		for _, tick := range visibleTicks(minorLoc.Ticks(domainMin, domainMax, g.curvelinearTargetTickCount(axis, true)), domainMin, domainMax) {
			drawSampledGridLine(r, ctx, g.Axis, tick, geoGridSegments, paint)
		}
	}

	if g.Major {
		loc := g.curvelinearLocator(axis, false)
		paint := polarTickPaint(majorColor, g.LineWidth, g.Dashes)
		for _, tick := range visibleTicks(loc.Ticks(domainMin, domainMax, g.curvelinearTargetTickCount(axis, false)), domainMin, domainMax) {
			drawSampledGridLine(r, ctx, g.Axis, tick, geoGridSegments, paint)
		}
	}
}

func (g *Grid) drawPolar(r render.Renderer, ctx *DrawContext) {
	majorColor := g.Color
	if g.Alpha > 0 && g.Alpha <= 1 {
		majorColor.A = g.Alpha
	}

	if g.Minor {
		minorColor := g.MinorColor
		if minorColor == (render.Color{}) {
			minorColor = majorColor
			minorColor.A = majorColor.A * 0.4
		}
		minorWidth := g.MinorLineWidth
		if minorWidth <= 0 {
			minorWidth = g.LineWidth * 0.5
		}
		g.drawPolarLines(r, ctx, true, minorColor, minorWidth, g.MinorDashes)
	}

	if g.Major {
		g.drawPolarLines(r, ctx, false, majorColor, g.LineWidth, g.Dashes)
	}
}

func (g *Grid) drawPolarLines(r render.Renderer, ctx *DrawContext, minor bool, color render.Color, width float64, dashes []float64) {
	center, outerRadius := polarCenterAndRadius(ctx.Clip)
	paint := polarTickPaint(color, width, dashes)

	switch g.Axis {
	case AxisBottom, AxisTop:
		axis := axisForPolarGrid(ctx, true)
		ticks := polarThetaTicks(axis, ctx.DataToPixel.XScale, minor)
		if locator := g.polarLocator(axis, minor); locator != nil && ctx.DataToPixel.XScale != nil {
			minVal, maxVal := ctx.DataToPixel.XScale.Domain()
			ticks = visibleTicks(locator.Ticks(minVal, maxVal, polarTargetTickCount(axis, minor)), minVal, maxVal)
		}
		for _, tick := range ticks {
			angle := polarAngleForTheta(ctx.Projection, ctx.DataToPixel.XScale, tick)
			path := geom.Path{}
			path.MoveTo(center)
			path.LineTo(polarPixelPoint(center, outerRadius, angle))
			r.Path(path, &paint)
		}
	case AxisLeft, AxisRight:
		axis := axisForPolarGrid(ctx, false)
		ticks := polarRadialTicks(axis, ctx.DataToPixel.YScale, minor)
		if locator := g.polarLocator(axis, minor); locator != nil && ctx.DataToPixel.YScale != nil {
			minVal, maxVal := ctx.DataToPixel.YScale.Domain()
			ticks = visibleTicks(locator.Ticks(minVal, maxVal, polarTargetTickCount(axis, minor)), minVal, maxVal)
		}
		for _, tick := range ticks {
			radius := outerRadius * ctx.DataToPixel.YScale.Fwd(tick)
			if radius <= 0 {
				continue
			}
			path := polarCirclePath(center, radius)
			if sides := radarFrameSidesForProjection(ctx.Projection); sides >= 3 {
				path = polarPolygonFramePath(ctx.Projection, center, radius, sides)
			}
			r.Path(path, &paint)
		}
	}
}

func axisForPolarGrid(ctx *DrawContext, isTheta bool) *Axis {
	if ctx == nil || ctx.Axes == nil {
		return nil
	}
	if isTheta {
		return ctx.Axes.effectiveXAxis()
	}
	return ctx.Axes.effectiveYAxis()
}

func (g *Grid) polarLocator(axis *Axis, minor bool) Locator {
	if minor {
		if g.MinorLocator != nil {
			return g.MinorLocator
		}
		if axis != nil {
			return axis.MinorLocator
		}
		return nil
	}
	if g.Locator != nil {
		return g.Locator
	}
	if axis != nil {
		return axis.Locator
	}
	return nil
}

func polarTargetTickCount(axis *Axis, minor bool) int {
	if axis == nil {
		if minor {
			return 30
		}
		return 6
	}
	if minor {
		return axis.minorTickTargetCount()
	}
	return axis.majorTickTargetCount()
}

func projectionGridUsesSampling(ctx *DrawContext) bool {
	if ctx == nil || ctx.Projection == nil || isPolarProjection(ctx.Projection) || isGeoProjection(ctx.Projection) {
		return false
	}
	if _, ok := ctx.TransProjection().(transform.Separable); ok {
		return false
	}
	return true
}

func (g *Grid) axisForContext(ctx *DrawContext) *Axis {
	if g == nil || ctx == nil || ctx.Axes == nil {
		return nil
	}
	switch g.Axis {
	case AxisTop:
		if axis := ctx.Axes.effectiveTopAxis(); axis != nil {
			return axis
		}
		return ctx.Axes.effectiveXAxis()
	case AxisRight:
		if axis := ctx.Axes.effectiveRightAxis(); axis != nil {
			return axis
		}
		return ctx.Axes.effectiveYAxis()
	case AxisLeft:
		return ctx.Axes.effectiveYAxis()
	default:
		return ctx.Axes.effectiveXAxis()
	}
}

func (g *Grid) curvelinearLocator(axis *Axis, minor bool) Locator {
	if minor {
		if g.MinorLocator != nil {
			return g.MinorLocator
		}
		if axis != nil && axis.MinorLocator != nil {
			return axis.MinorLocator
		}
		return MinorLinearLocator{N: 5}
	}
	if g.Locator != nil {
		return g.Locator
	}
	if axis != nil && axis.Locator != nil {
		return axis.Locator
	}
	return LinearLocator{}
}

func (g *Grid) curvelinearTargetTickCount(axis *Axis, minor bool) int {
	if axis == nil {
		if minor {
			return 30
		}
		return 8
	}
	if minor {
		return axis.minorTickTargetCount()
	}
	return axis.majorTickTargetCount()
}

// drawLine draws a single grid line.
func (g *Grid) drawLine(r render.Renderer, ctx *DrawContext, tickValue float64, isXAxis bool, color render.Color, width float64, dashes []float64) {
	var p1, p2 geom.Pt

	if isXAxis {
		yMin, yMax := ctx.DataToPixel.YScale.Domain()
		p1 = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: yMin})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: yMax})
	} else {
		xMin, xMax := ctx.DataToPixel.XScale.Domain()
		p1 = ctx.DataToPixel.Apply(geom.Pt{X: xMin, Y: tickValue})
		p2 = ctx.DataToPixel.Apply(geom.Pt{X: xMax, Y: tickValue})
	}

	path := geom.Path{}
	path.C = append(path.C, geom.MoveTo)
	path.V = append(path.V, p1)
	path.C = append(path.C, geom.LineTo)
	path.V = append(path.V, p2)

	paint := render.Paint{
		LineWidth: width,
		Stroke:    color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
		Dashes:    dashes,
	}
	r.Path(path, &paint)
}

func drawSampledGridLine(r render.Renderer, ctx *DrawContext, axis AxisSide, tick float64, segments int, paint render.Paint) {
	if ctx == nil || segments <= 0 {
		return
	}
	xMin, xMax := ctx.DataToPixel.XScale.Domain()
	yMin, yMax := ctx.DataToPixel.YScale.Domain()
	path := geom.Path{}

	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		var data geom.Pt
		switch axis {
		case AxisBottom, AxisTop:
			data = geom.Pt{X: tick, Y: yMin + (yMax-yMin)*t}
		default:
			data = geom.Pt{X: xMin + (xMax-xMin)*t, Y: tick}
		}
		pt := ctx.DataToPixel.Apply(data)
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	r.Path(path, &paint)
}

// Z returns the z-order for sorting.
func (g *Grid) Z() float64 {
	return g.z
}

// Bounds returns an empty rect for now.
func (g *Grid) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}
