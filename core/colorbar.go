package core

import (
	"math"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// ColorbarOptions configures figure-level colorbar placement.
type ColorbarOptions struct {
	Width    float64
	Padding  float64
	Aspect   float64
	Label    string
	Colormap *string
	VMin     *float64
	VMax     *float64
}

// Colorbar renders a vertical gradient keyed to a scalar colormap.
type Colorbar struct {
	Colormap    string
	Alpha       float64
	BorderColor render.Color
	BorderWidth float64
	z           float64
}

const (
	defaultColorbarFraction = 0.15
	defaultColorbarPadding  = 0.05
	defaultColorbarAspect   = 20.0
)

// AddColorbar creates a dedicated axes to the right of a plot and populates it
// with a colorbar derived from a scalar-mappable artist.
func (f *Figure) AddColorbar(parent *Axes, mappable ScalarMappable, opts ...ColorbarOptions) *Axes {
	if f == nil || parent == nil || mappable == nil {
		return nil
	}

	cfg := ColorbarOptions{}
	if len(opts) > 0 {
		cfg = opts[0]
	}
	cfg.Aspect = resolvedColorbarAspect(cfg.Aspect)

	mapping := mappable.ScalarMap().Resolved()
	cmap := mapping.Colormap
	if cfg.Colormap != nil && *cfg.Colormap != "" {
		cmap = *cfg.Colormap
	}
	vmin := mapping.VMin
	if cfg.VMin != nil {
		vmin = *cfg.VMin
	}
	vmax := mapping.VMax
	if cfg.VMax != nil {
		vmax = *cfg.VMax
	}
	if vmin == vmax {
		vmax = vmin + 1
	}

	base := parent.RectFraction
	width := resolvedColorbarWidth(f, base, cfg.Width, cfg.Aspect)
	padding := resolvedColorbarLayoutPadding(f, base, cfg.Padding)
	parent.RectFraction = colorbarParentRect(base, width, padding)
	rect := geom.Rect{
		Min: geom.Pt{
			X: parent.RectFraction.Max.X + padding,
			Y: parent.RectFraction.Min.Y,
		},
		Max: geom.Pt{
			X: parent.RectFraction.Max.X + padding + width,
			Y: parent.RectFraction.Max.Y,
		},
	}
	if rect.Min.X >= rect.Max.X {
		return nil
	}

	ax := f.AddAxes(rect)
	ax.colorbarParent = parent
	ax.colorbarWidth = cfg.Width
	ax.colorbarPadding = cfg.Padding
	ax.colorbarAspect = cfg.Aspect
	ax.colorbarBase = base
	ax.ShowFrame = false
	ax.SetXLim(0, 1)
	ax.SetYLim(vmin, vmax)

	if ax.XAxis != nil {
		ax.XAxis.ShowSpine = false
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
	}
	if ax.YAxis != nil {
		ax.YAxis.ShowSpine = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		ax.YAxis.MinorLocator = nil
	}
	if right := ax.RightAxis(); right != nil {
		right.MinorLocator = nil
	}
	_ = ax.SetYTickLabelPosition("right")
	_ = ax.SetYLabelPosition("right")
	if cfg.Label != "" {
		ax.SetYLabel(cfg.Label)
	}

	ax.Add(&Colorbar{
		Colormap:    cmap,
		Alpha:       1,
		BorderColor: f.RC.AxesEdgeColor,
		BorderWidth: f.RC.AxisLineWidth,
		z:           -10,
	})

	return ax
}

func resolvedColorbarPadding(base geom.Rect, padding float64) float64 {
	if padding > 0 {
		return padding
	}
	return defaultColorbarPadding
}

func resolvedColorbarLayoutPadding(fig *Figure, base geom.Rect, padding float64) float64 {
	resolved := resolvedColorbarPadding(base, padding)
	if padding > 0 || fig == nil || fig.layoutEngine != LayoutEngineConstrained || fig.SizePx.X <= 0 {
		return resolved
	}
	return resolved + layoutPadPx(fig, LayoutEngineConstrained)/fig.SizePx.X
}

func resolvedColorbarAspect(aspect float64) float64 {
	if aspect > 0 {
		return aspect
	}
	return defaultColorbarAspect
}

func resolvedColorbarWidth(fig *Figure, base geom.Rect, width, aspect float64) float64 {
	if width > 0 {
		return width
	}
	fractionWidth := base.W() * defaultColorbarFraction
	if fig == nil || fig.SizePx.X <= 0 || fig.SizePx.Y <= 0 || aspect <= 0 {
		return fractionWidth
	}
	aspectWidth := base.H() * fig.SizePx.Y / (aspect * fig.SizePx.X)
	if aspectWidth <= 0 {
		return fractionWidth
	}
	return math.Min(fractionWidth, aspectWidth)
}

func colorbarParentRect(base geom.Rect, width, padding float64) geom.Rect {
	reserved := width + padding
	if reserved <= 0 {
		return base
	}
	shrunk := base
	if base.Max.X-reserved <= base.Min.X {
		return shrunk
	}
	shrunk.Max.X -= reserved
	return shrunk
}

// Draw renders a vertical gradient across the colorbar axes.
func (c *Colorbar) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}

	const gradientHeight = 256

	cmap := matcolor.GetColormap(c.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	outlinePath := pixelRectPath(ctx.Clip)
	if snapped := snappedStrokeRectPath(ctx.Clip); len(snapped.C) > 0 {
		outlinePath = snapped
	}

	for i := 0; i < gradientHeight; i++ {
		t := (float64(i) + 0.5) / float64(gradientHeight)
		col := cmap.At(t)
		col.A *= alpha

		path := snappedFillRectPath(colorbarCellRect(ctx.Clip, i, gradientHeight))
		if len(path.C) == 0 {
			continue
		}
		r.Path(path, &render.Paint{
			Fill:      col,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Antialias: render.AntialiasDefault,
		})
	}

	r.Path(outlinePath, &render.Paint{
		Stroke:    c.BorderColor,
		LineWidth: c.BorderWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func colorbarCellRect(clip geom.Rect, index, count int) geom.Rect {
	if count <= 0 {
		return geom.Rect{}
	}
	y0 := clip.Max.Y - clip.H()*float64(index+1)/float64(count)
	y1 := clip.Max.Y - clip.H()*float64(index)/float64(count)
	return geom.Rect{
		Min: geom.Pt{X: clip.Min.X, Y: y0},
		Max: geom.Pt{X: clip.Max.X, Y: y1},
	}
}

// Bounds returns an empty rect so colorbars do not affect autoscaling.
func (c *Colorbar) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the colorbar z-order.
func (c *Colorbar) Z() float64 { return c.z }
