package core

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Line2D is a minimal polyline artist (stroke only).
type Line2D struct {
	XY     []geom.Pt    // data space points
	W      float64      // stroke width (px for now)
	Col    render.Color // stroke color
	Dashes []float64    // dash pattern (on/off pairs)
	Label  string       // series label for legend
	z      float64      // z-order
}

// Draw renders the line by transforming points to pixel space and drawing a path.
func (l *Line2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(l.XY) == 0 {
		return // nothing to draw
	}

	p := geom.Path{}
	for i, v := range l.XY {
		q := (&ctx.DataToPixel).Apply(v)
		if i == 0 {
			p.C = append(p.C, geom.MoveTo)
		} else {
			p.C = append(p.C, geom.LineTo)
		}
		p.V = append(p.V, q)
	}

	paint := render.Paint{
		LineWidth:  l.W,
		LineJoin:   render.JoinRound, // Default to round joins
		LineCap:    render.CapRound,  // Default to round caps
		MiterLimit: 10.0,             // Standard miter limit
		Stroke:     l.Col,
		Dashes:     l.Dashes, // Use dash pattern if provided
	}
	r.Path(p, &paint)
}

// Z returns the z-order for sorting.
func (l *Line2D) Z() float64 {
	return l.z
}

// Bounds returns the bounding box of all points in data space.
func (l *Line2D) Bounds(*DrawContext) geom.Rect {
	if len(l.XY) == 0 {
		return geom.Rect{}
	}
	r := geom.Rect{Min: l.XY[0], Max: l.XY[0]}
	for _, pt := range l.XY[1:] {
		if pt.X < r.Min.X {
			r.Min.X = pt.X
		}
		if pt.Y < r.Min.Y {
			r.Min.Y = pt.Y
		}
		if pt.X > r.Max.X {
			r.Max.X = pt.X
		}
		if pt.Y > r.Max.Y {
			r.Max.Y = pt.Y
		}
	}
	return r
}
