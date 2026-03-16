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

// Axis renders axis spines, ticks, and labels for a single dimension.
type Axis struct {
	Side         AxisSide     // which side of the plot
	Locator      Locator      // major tick position calculator
	MinorLocator Locator      // minor tick position calculator (nil = no minor ticks)
	Formatter    Formatter    // tick label formatter
	Color        render.Color // axis line and tick color
	LineWidth    float64      // width of axis line and ticks
	TickSize     float64      // length of major tick marks (in pixels)
	MinorTickSize float64     // length of minor tick marks (in pixels); 0 uses TickSize*0.6
	ShowSpine    bool         // whether to draw the axis line
	ShowTicks    bool         // whether to draw tick marks
	ShowLabels   bool         // whether to draw tick labels
	z            float64      // z-order
}

// NewXAxis creates an axis for the bottom (x-axis).
func NewXAxis() *Axis {
	return &Axis{
		Side:       AxisBottom,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 3},
		Color:      render.Color{R: 0, G: 0, B: 0, A: 1}, // black
		LineWidth:  1.0,
		TickSize:   5.0,
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
		LineWidth:  1.0,
		TickSize:   5.0,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: true,
	}
}

// Draw renders the axis spine (called inside clip region).
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

	var min, max float64
	var isXAxis bool

	switch a.Side {
	case AxisBottom, AxisTop:
		min, max = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		min, max = ctx.DataToPixel.YScale.Domain()
	}

	// Minor ticks first
	if a.MinorLocator != nil {
		minorTicks := a.MinorLocator.Ticks(min, max, 30)
		if len(minorTicks) > 0 {
			a.drawMinorTicks(r, ctx, minorTicks, isXAxis)
		}
	}

	// Major ticks
	ticks := a.Locator.Ticks(min, max, 6)
	if len(ticks) > 0 {
		a.drawTicks(r, ctx, ticks, isXAxis)
	}
}

// drawSpine draws the main axis line directly in pixel space.
// Lines are snapped inward by half their width so the stroke falls entirely
// within the clip rectangle, avoiding sub-pixel anti-aliasing artifacts.
func (a *Axis) drawSpine(r render.Renderer, ctx *DrawContext) {
	px := ctx.Clip
	lw := a.LineWidth
	half := lw / 2
	p1, p2 := spinePixelEndpoints(a.Side, px, half)

	paint := render.Paint{
		LineWidth: lw,
		Stroke:    a.Color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}
	r.Path(path, &paint)
}

// spinePixelEndpoints returns the two pixel-space endpoints for a spine on the
// given side of px, snapped inward by `inset` pixels.
func spinePixelEndpoints(side AxisSide, px geom.Rect, inset float64) (geom.Pt, geom.Pt) {
	x1 := math.Round(px.Min.X)
	y1 := math.Round(px.Min.Y)
	x2 := math.Round(px.Max.X)
	y2 := math.Round(px.Max.Y)

	switch side {
	case AxisBottom:
		y := y2 - inset
		return geom.Pt{X: x1, Y: y}, geom.Pt{X: x2, Y: y}
	case AxisTop:
		y := y1 + inset
		return geom.Pt{X: x1, Y: y}, geom.Pt{X: x2, Y: y}
	case AxisLeft:
		x := x1 + inset
		return geom.Pt{X: x, Y: y1}, geom.Pt{X: x, Y: y2}
	case AxisRight:
		x := x2 - inset
		return geom.Pt{X: x, Y: y1}, geom.Pt{X: x, Y: y2}
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

// DrawFrame draws the top and right border lines of the axes box directly in
// pixel space, snapped to crisp integer-aligned positions.
func DrawFrame(r render.Renderer, ctx *DrawContext, ref *Axis) {
	if ref == nil || !ref.ShowSpine {
		return
	}
	paint := render.Paint{
		LineWidth: ref.LineWidth,
		Stroke:    ref.Color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
	}
	half := ref.LineWidth / 2

	drawLine := func(p1, p2 geom.Pt) {
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{p1, p2},
		}
		r.Path(path, &paint)
	}

	p1, p2 := spinePixelEndpoints(AxisTop, ctx.Clip, half)
	drawLine(p1, p2)
	p1, p2 = spinePixelEndpoints(AxisRight, ctx.Clip, half)
	drawLine(p1, p2)
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
	var min, max float64
	var isXAxis bool
	switch a.Side {
	case AxisBottom, AxisTop:
		min, max = ctx.DataToPixel.XScale.Domain()
		isXAxis = true
	case AxisLeft, AxisRight:
		min, max = ctx.DataToPixel.YScale.Domain()
		isXAxis = false
	}
	ticks := a.Locator.Ticks(min, max, 6)
	a.drawTickLabels(r, ctx, ticks, isXAxis)
}

// drawTickLabels draws text labels for the ticks if the renderer supports text.
func (a *Axis) drawTickLabels(r render.Renderer, ctx *DrawContext, ticks []float64, isXAxis bool) {
	type textRenderer interface {
		DrawText(text string, origin geom.Pt, size float64, textColor render.Color)
	}

	textRen, ok := r.(textRenderer)
	if !ok {
		return
	}

	fontSize := ctx.RC.FontSize
	if fontSize <= 0 {
		fontSize = 12.0
	}

	for _, tickValue := range ticks {
		label := a.Formatter.Format(tickValue)
		if label == "" {
			continue
		}

		// Measure text for centering
		metrics := r.MeasureText(label, fontSize, ctx.RC.FontKey)

		var labelPos geom.Pt

		if isXAxis {
			spineY := getSpinePosition(a.Side, ctx)
			tickPos := ctx.DataToPixel.Apply(geom.Pt{X: tickValue, Y: spineY})

			switch a.Side {
			case AxisBottom:
				// Center horizontally, below tick end + gap for ascent
				labelPos = geom.Pt{
					X: tickPos.X - metrics.W/2,
					Y: tickPos.Y + a.TickSize + metrics.Ascent + 2,
				}
			case AxisTop:
				labelPos = geom.Pt{
					X: tickPos.X - metrics.W/2,
					Y: tickPos.Y - a.TickSize - metrics.Descent - 2,
				}
			}
		} else {
			spineX := getSpinePosition(a.Side, ctx)
			tickPos := ctx.DataToPixel.Apply(geom.Pt{X: spineX, Y: tickValue})

			switch a.Side {
			case AxisLeft:
				// Right-align to left of tick end
				labelPos = geom.Pt{
					X: tickPos.X - a.TickSize - metrics.W - 3,
					Y: tickPos.Y + metrics.Ascent/2,
				}
			case AxisRight:
				labelPos = geom.Pt{
					X: tickPos.X + a.TickSize + 3,
					Y: tickPos.Y + metrics.Ascent/2,
				}
			}
		}

		textRen.DrawText(label, labelPos, fontSize, a.Color)
	}
}
