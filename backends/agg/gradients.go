package agg

import (
	"github.com/cwbudde/matplotlib-go/render"
)

// SupportsGradientFill reports that the AGG backend renders Paint.FillGradient
// natively via Agg2D's linear and radial gradient span generators.
//
// The current AGG bridge supports two-stop linear and radial gradients. For
// gradients with more than two stops, the first and last stops drive the
// linear gradient endpoints; the radial path additionally uses the middle stop
// when exactly three stops are supplied via the multi-stop variant.
func (r *Renderer) SupportsGradientFill() bool { return true }

// SupportsPatternFill reports that the AGG backend does NOT yet consume
// Paint.FillPattern natively. Callers that need pattern fills should fall back
// to expanding the pattern into a tile of paths or to a vector backend.
func (r *Renderer) SupportsPatternFill() bool { return false }

// applyGradientFill configures the AGG fill state for the gradient described
// by paint.FillGradient and returns true. The caller is responsible for
// resetting the fill color to a solid value afterwards if it issues further
// fill operations with a different source.
func (r *Renderer) applyGradientFill(paint *render.Paint) bool {
	g := &paint.FillGradient
	if g.Kind == render.GradientNone || len(g.Stops) == 0 {
		return false
	}

	stops := g.Stops
	first := stops[0].Color
	last := stops[len(stops)-1].Color
	first = colorWithForcedAlpha(first, paint)
	last = colorWithForcedAlpha(last, paint)

	switch g.Kind {
	case render.LinearGradient:
		r.ctx.SetFillLinearGradient(
			g.Start.X, g.Start.Y,
			g.End.X, g.End.Y,
			renderColorToAGG(first), renderColorToAGG(last),
			1.0,
		)
		return true
	case render.RadialGradient:
		if len(stops) >= 3 {
			mid := colorWithForcedAlpha(stops[len(stops)/2].Color, paint)
			r.ctx.SetFillRadialGradientMultiStop(
				g.Center.X, g.Center.Y, g.Radius,
				renderColorToAGG(first),
				renderColorToAGG(mid),
				renderColorToAGG(last),
			)
			return true
		}
		r.ctx.SetFillRadialGradient(
			g.Center.X, g.Center.Y, g.Radius,
			renderColorToAGG(first), renderColorToAGG(last),
			1.0,
		)
		return true
	}
	return false
}
