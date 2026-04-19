package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// ErrorBar renders symmetric horizontal and/or vertical error bars for points.
type ErrorBar struct {
	XY        []geom.Pt    // data-space points
	XErr      []float64    // symmetric x errors (same length as XY or broadcast scalar)
	YErr      []float64    // symmetric y errors (same length as XY or broadcast scalar)
	Color     render.Color // stroke color
	LineWidth float64      // stroke width in pixels
	CapSize   float64      // cap size in pixels (full width/height)
	Alpha     float64      // alpha transparency (0-1), if 0 uses 1.0
	Label     string       // series label for legend
	z         float64      // z-order
}

// Draw renders each error bar from XY to XY with symmetric offsets.
func (e *ErrorBar) Draw(r render.Renderer, ctx *DrawContext) {
	if len(e.XY) == 0 {
		return
	}

	lineWidth := e.LineWidth
	if lineWidth <= 0 {
		lineWidth = 1.0
	}

	capSize := e.CapSize
	if capSize < 0 {
		capSize = 0
	}
	capHalf := capSize * 0.5

	alpha := e.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}
	if alpha > 1 {
		alpha = 1.0
	}

	color := e.Color
	color.A *= alpha
	if color.A <= 0 {
		return
	}

	paint := render.Paint{
		Stroke:    color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	}

	for i, pt := range e.XY {
		xErr := resolveError(e.XErr, i)
		yErr := resolveError(e.YErr, i)

		if xErr > 0 {
			left := geom.Pt{X: pt.X - xErr, Y: pt.Y}
			right := geom.Pt{X: pt.X + xErr, Y: pt.Y}
			r.Path(linePath(ctx, left, right), &paint)

			if capHalf > 0 {
				leftTop := addPixelOffset(ctx, left, 0, -capHalf)
				leftBottom := addPixelOffset(ctx, left, 0, capHalf)
				rightTop := addPixelOffset(ctx, right, 0, -capHalf)
				rightBottom := addPixelOffset(ctx, right, 0, capHalf)
				r.Path(linePath(ctx, leftTop, leftBottom), &paint)
				r.Path(linePath(ctx, rightTop, rightBottom), &paint)
			}
		}

		if yErr > 0 {
			lower := geom.Pt{X: pt.X, Y: pt.Y - yErr}
			upper := geom.Pt{X: pt.X, Y: pt.Y + yErr}
			r.Path(linePath(ctx, lower, upper), &paint)

			if capHalf > 0 {
				lowerLeft := addPixelOffset(ctx, lower, -capHalf, 0)
				lowerRight := addPixelOffset(ctx, lower, capHalf, 0)
				upperLeft := addPixelOffset(ctx, upper, -capHalf, 0)
				upperRight := addPixelOffset(ctx, upper, capHalf, 0)
				r.Path(linePath(ctx, lowerLeft, lowerRight), &paint)
				r.Path(linePath(ctx, upperLeft, upperRight), &paint)
			}
		}
	}
}

// Z returns the z-order for sorting.
func (e *ErrorBar) Z() float64 {
	return e.z
}

// Bounds returns the data-space bounding box for bars and error extents.
func (e *ErrorBar) Bounds(*DrawContext) geom.Rect {
	if len(e.XY) == 0 {
		return geom.Rect{}
	}

	bounds := geom.Rect{Min: e.XY[0], Max: e.XY[0]}
	for i, pt := range e.XY {
		if math.IsNaN(pt.X) || math.IsNaN(pt.Y) || math.IsInf(pt.X, 0) || math.IsInf(pt.Y, 0) {
			continue
		}

		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}

		xErr := resolveError(e.XErr, i)
		yErr := resolveError(e.YErr, i)
		if xErr > 0 {
			left := pt.X - xErr
			right := pt.X + xErr
			if left < bounds.Min.X {
				bounds.Min.X = left
			}
			if right > bounds.Max.X {
				bounds.Max.X = right
			}
		}
		if yErr > 0 {
			lower := pt.Y - yErr
			upper := pt.Y + yErr
			if lower < bounds.Min.Y {
				bounds.Min.Y = lower
			}
			if upper > bounds.Max.Y {
				bounds.Max.Y = upper
			}
		}
	}

	return bounds
}

// resolveError returns scalar or broadcasted symmetric error magnitude at index i.
func resolveError(err []float64, i int) float64 {
	if len(err) == 0 {
		return 0
	}
	if len(err) == 1 {
		return math.Abs(err[0])
	}
	if i < len(err) {
		return math.Abs(err[i])
	}
	return 0
}

func addPixelOffset(ctx *DrawContext, dataPt geom.Pt, dxPx, dyPx float64) geom.Pt {
	basePx := ctx.DataToPixel.Apply(dataPt)
	targetPx := geom.Pt{
		X: basePx.X + dxPx,
		Y: basePx.Y + dyPx,
	}

	dataSpace, ok := invertToData(ctx, targetPx)
	if !ok {
		return dataPt
	}
	return dataSpace
}

func invertToData(ctx *DrawContext, px geom.Pt) (geom.Pt, bool) {
	if ctx == nil {
		return geom.Pt{}, false
	}
	return ctx.DataToPixel.Invert(px)
}
