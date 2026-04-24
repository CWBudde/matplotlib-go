package core

import (
	"image"

	matcolor "matplotlib-go/color"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// ColorbarOptions configures figure-level colorbar placement.
type ColorbarOptions struct {
	Width    float64
	Padding  float64
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

// AddColorbar creates a dedicated axes to the right of a plot and populates it
// with a colorbar derived from a scalar-mappable artist.
func (f *Figure) AddColorbar(parent *Axes, mappable ScalarMappable, opts ...ColorbarOptions) *Axes {
	if f == nil || parent == nil || mappable == nil {
		return nil
	}

	cfg := ColorbarOptions{
		Width:   0.035,
		Padding: 0.02,
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Width <= 0 {
			cfg.Width = 0.035
		}
		if cfg.Padding <= 0 {
			cfg.Padding = 0.02
		}
	}

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

	rect := geom.Rect{
		Min: geom.Pt{
			X: parent.RectFraction.Max.X + cfg.Padding,
			Y: parent.RectFraction.Min.Y,
		},
		Max: geom.Pt{
			X: parent.RectFraction.Max.X + cfg.Padding + cfg.Width,
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
	if cfg.Label != "" {
		ax.SetYLabel(cfg.Label)
	}

	ax.Add(&Colorbar{
		Colormap:    cmap,
		Alpha:       1,
		BorderColor: render.Color{R: 0.2, G: 0.2, B: 0.2, A: 0.9},
		BorderWidth: 1,
		z:           -10,
	})

	return ax
}

// Draw renders a vertical gradient across the colorbar axes.
func (c *Colorbar) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}

	const (
		gradientWidth  = 16
		gradientHeight = 256
	)

	cmap := matcolor.GetColormap(c.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, gradientWidth, gradientHeight))
	for y := 0; y < gradientHeight; y++ {
		t := 1.0
		if gradientHeight > 1 {
			t = 1 - float64(y)/float64(gradientHeight-1)
		}
		col := cmap.At(t)
		col.A *= alpha
		rgba := toRGBAColor(col)
		for x := 0; x < gradientWidth; x++ {
			img.Set(x, y, rgba)
		}
	}

	r.Image(render.NewImageData(img), ctx.Clip)
	r.Path(pixelRectPath(ctx.Clip), &render.Paint{
		Stroke:    c.BorderColor,
		LineWidth: c.BorderWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

// Bounds returns an empty rect so colorbars do not affect autoscaling.
func (c *Colorbar) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the colorbar z-order.
func (c *Colorbar) Z() float64 { return c.z }
