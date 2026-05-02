package render

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

// GraphicsContext centralizes draw-state that Matplotlib keeps separate from
// renderer drawing verbs.
//
// Renderers are not required to consume this type directly. It is the port's
// shared state model for backends and higher-level artists that need explicit
// ownership of opacity, clipping, transforms, and path paint state.
type GraphicsContext struct {
	Alpha             float64
	Transform         transform.T
	Paint             Paint
	ClipRect          geom.Rect
	HasClipRect       bool
	ClipPath          geom.Path
	HasClipPath       bool
	ClipPathTransform geom.Affine
	HasClipPathTrans  bool
	Antialias         AntialiasMode
	Snap              SnapMode
	Hatch             string
	HatchColor        Color
	HatchLineWidth    float64
	HatchSpacing      float64
	Sketch            SketchParams
	ForceAlpha        bool
	ForcedAlpha       float64
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

// WithClipPathTransform stores the affine transform for the current clip path.
func (gc GraphicsContext) WithClipPathTransform(affine geom.Affine) GraphicsContext {
	gc.ClipPathTransform = affine
	gc.HasClipPathTrans = true
	return gc
}

// WithAntialias returns a context with the given antialiasing mode.
func (gc GraphicsContext) WithAntialias(mode AntialiasMode) GraphicsContext {
	gc.Antialias = mode
	return gc
}

// WithSnap returns a context with the given snap mode.
func (gc GraphicsContext) WithSnap(mode SnapMode) GraphicsContext {
	gc.Snap = mode
	return gc
}

// WithHatch returns a context with hatch metadata.
func (gc GraphicsContext) WithHatch(pattern string, color Color, linewidth float64) GraphicsContext {
	gc.Hatch = pattern
	gc.HatchColor = color
	gc.HatchLineWidth = linewidth
	return gc
}

// WithHatchSpacing returns a context with explicit hatch spacing.
func (gc GraphicsContext) WithHatchSpacing(spacing float64) GraphicsContext {
	gc.HatchSpacing = spacing
	return gc
}

// WithSketch returns a context with sketch/jitter parameters.
func (gc GraphicsContext) WithSketch(params SketchParams) GraphicsContext {
	gc.Sketch = params
	return gc
}

// WithForcedAlpha returns a context that overrides paint alpha before applying
// the context alpha multiplier.
func (gc GraphicsContext) WithForcedAlpha(alpha float64) GraphicsContext {
	gc.ForceAlpha = true
	gc.ForcedAlpha = alpha
	return gc
}

// EffectivePaint returns the path paint after applying the context alpha.
func (gc GraphicsContext) EffectivePaint() Paint {
	paint := gc.Paint
	paint.Dashes = cloneFloat64s(paint.Dashes)
	if gc.Antialias != AntialiasDefault {
		paint.Antialias = gc.Antialias
	}
	if gc.Snap != SnapOff {
		paint.Snap = gc.Snap
	}
	if gc.Hatch != "" {
		paint.Hatch = gc.Hatch
		paint.HatchColor = gc.HatchColor
		paint.HatchLineWidth = gc.HatchLineWidth
		paint.HatchSpacing = gc.HatchSpacing
	}
	if gc.Sketch != (SketchParams{}) {
		paint.Sketch = gc.Sketch
	}
	if gc.HasClipPathTrans {
		paint.ClipPathTransform = gc.ClipPathTransform
		paint.HasClipPathTrans = true
	}
	if gc.ForceAlpha {
		paint.ForceAlpha = true
		paint.Alpha = gc.ForcedAlpha
		if paint.Stroke.A > 0 {
			paint.Stroke.A = gc.ForcedAlpha
		}
		if paint.Fill.A > 0 {
			paint.Fill.A = gc.ForcedAlpha
		}
		if paint.HatchColor.A > 0 {
			paint.HatchColor.A = gc.ForcedAlpha
		}
	}
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
