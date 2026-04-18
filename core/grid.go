package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
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
	return &Grid{
		Axis:      axis,
		Color:     render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}, // light gray
		LineWidth: 0.5,
		Alpha:     0, // use Color.A
		Major:     true,
		Minor:     false,
		z:         -1000, // behind everything else
	}
}

// Draw renders grid lines at tick positions.
func (g *Grid) Draw(r render.Renderer, ctx *DrawContext) {
	if !g.Major && !g.Minor {
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

// Z returns the z-order for sorting.
func (g *Grid) Z() float64 {
	return g.z
}

// Bounds returns an empty rect for now.
func (g *Grid) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}
