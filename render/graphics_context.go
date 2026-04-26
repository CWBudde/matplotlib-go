package render

import (
	"matplotlib-go/internal/geom"
	"matplotlib-go/transform"
)

// GraphicsContext centralizes draw-state that Matplotlib keeps separate from
// renderer drawing verbs.
//
// Renderers are not required to consume this type directly. It is the port's
// shared state model for backends and higher-level artists that need explicit
// ownership of opacity, clipping, transforms, and path paint state.
type GraphicsContext struct {
	Alpha       float64
	Transform   transform.T
	Paint       Paint
	ClipRect    geom.Rect
	HasClipRect bool
	ClipPath    geom.Path
	HasClipPath bool
}

// NewGraphicsContext returns a graphics context with opaque alpha.
func NewGraphicsContext() GraphicsContext {
	return GraphicsContext{Alpha: 1}
}

// Clone returns an independent copy of the graphics context.
func (gc GraphicsContext) Clone() GraphicsContext {
	gc.Paint.Dashes = cloneFloat64s(gc.Paint.Dashes)
	gc.ClipPath = clonePath(gc.ClipPath)
	return gc
}

// WithAlpha returns a context with the given alpha multiplier.
func (gc GraphicsContext) WithAlpha(alpha float64) GraphicsContext {
	gc.Alpha = alpha
	return gc
}

// WithTransform returns a context with the given current transform.
func (gc GraphicsContext) WithTransform(tr transform.T) GraphicsContext {
	gc.Transform = tr
	return gc
}

// WithClipRect returns a context with a rectangular clip.
func (gc GraphicsContext) WithClipRect(rect geom.Rect) GraphicsContext {
	gc.ClipRect = rect
	gc.HasClipRect = true
	return gc
}

// WithClipPath returns a context with a path clip.
func (gc GraphicsContext) WithClipPath(path geom.Path) GraphicsContext {
	gc.ClipPath = clonePath(path)
	gc.HasClipPath = true
	return gc
}

// EffectivePaint returns the path paint after applying the context alpha.
func (gc GraphicsContext) EffectivePaint() Paint {
	paint := gc.Paint
	paint.Dashes = cloneFloat64s(paint.Dashes)
	paint.Stroke.A *= gc.Alpha
	paint.Fill.A *= gc.Alpha
	return paint
}

func cloneFloat64s(in []float64) []float64 {
	if len(in) == 0 {
		return nil
	}
	out := make([]float64, len(in))
	copy(out, in)
	return out
}

func clonePath(in geom.Path) geom.Path {
	out := geom.Path{}
	if len(in.V) > 0 {
		out.V = make([]geom.Pt, len(in.V))
		copy(out.V, in.V)
	}
	if len(in.C) > 0 {
		out.C = make([]geom.Cmd, len(in.C))
		copy(out.C, in.C)
	}
	return out
}
