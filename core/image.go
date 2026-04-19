package core

import (
	"image"
	"image/color"
	"math"

	matcolor "matplotlib-go/color"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// Draw renders the rasterized image through the renderer.
func (i *Image2D) Draw(r render.Renderer, ctx *DrawContext) {
	if i == nil || r == nil {
		return
	}

	dst := i.destinationRect(ctx)
	if dst.W() <= 0 || dst.H() <= 0 {
		return
	}

	raster, ok := i.rasterize()
	if !ok {
		return
	}

	angleRad := i.AngleDeg * math.Pi / 180
	if angleRad == 0 {
		r.Image(raster, dst)
		return
	}

	if tr, ok := r.(render.ImageTransformer); ok {
		anchor := i.rotationAnchor(ctx, dst)
		transform := imageTransform(dst, raster, anchor, angleRad)
		tr.ImageTransformed(raster, dst, transform)
		return
	}

	// Fallback: ignore rotation and render axis-aligned image.
	r.Image(raster, dst)
}

func (i *Image2D) rasterize() (render.Image, bool) {
	rows := len(i.Data)
	if rows == 0 {
		return nil, false
	}

	cols := 0
	for _, row := range i.Data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return nil, false
	}

	vmin := i.VMin
	vmax := i.VMax
	if !isFinite(vmin) || !isFinite(vmax) {
		vmin, vmax = dataRange(i.Data)
	}
	if vmin == vmax {
		vmax = vmin + 1
	}

	cm := matcolor.GetColormap(i.Colormap)
	alpha := clampOneToOne(i.Alpha)
	img := image.NewRGBA(image.Rect(0, 0, cols, rows))

	span := vmax - vmin
	if span == 0 {
		span = 1
	}
	for row, values := range i.Data {
		for col := 0; col < cols; col++ {
			if col >= len(values) {
				continue
			}
			v := values[col]
			if !isFinite(v) {
				continue
			}

			n := (v - vmin) / span
			c := cm.At(n)
			c.A *= alpha

			pixelY := row
			if i.Origin == ImageOriginLower {
				pixelY = rows - 1 - row
			}

			img.Set(col, pixelY, toRGBAColor(c))
		}
	}

	return render.NewImageData(img), true
}

func (i *Image2D) destinationRect(ctx *DrawContext) geom.Rect {
	if ctx == nil {
		return geom.Rect{}
	}
	p1 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMin, Y: i.YMin})
	p2 := ctx.DataToPixel.Apply(geom.Pt{X: i.XMax, Y: i.YMax})
	minX := minF(p1.X, p2.X)
	maxX := maxF(p1.X, p2.X)
	minY := minF(p1.Y, p2.Y)
	maxY := maxF(p1.Y, p2.Y)
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func (i *Image2D) rotationAnchor(ctx *DrawContext, dst geom.Rect) geom.Pt {
	switch i.RotateAt {
	case ImageAnchorTopLeft:
		return dst.Min
	case ImageAnchorTopRight:
		return geom.Pt{X: dst.Max.X, Y: dst.Min.Y}
	case ImageAnchorBottomLeft:
		return geom.Pt{X: dst.Min.X, Y: dst.Max.Y}
	case ImageAnchorBottomRight:
		return dst.Max
	case ImageAnchorCustom:
		if ctx != nil {
			return ctx.DataToPixel.Apply(geom.Pt{X: i.RotateX, Y: i.RotateY})
		}
		fallthrough
	case ImageAnchorCenter:
		fallthrough
	default:
		return geom.Pt{
			X: (dst.Min.X + dst.Max.X) * 0.5,
			Y: (dst.Min.Y + dst.Max.Y) * 0.5,
		}
	}
}

func imageTransform(dst geom.Rect, raster render.Image, anchor geom.Pt, angle float64) geom.Affine {
	srcW, srcH := raster.Size()
	if srcW <= 0 || srcH <= 0 {
		return geom.Identity()
	}

	sx := dst.W() / float64(srcW)
	sy := dst.H() / float64(srcH)
	scale := geom.Affine{
		A: sx,
		D: sy,
		E: dst.Min.X,
		F: dst.Min.Y,
	}

	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rot := geom.Affine{
		A: cos,
		B: sin,
		C: -sin,
		D: cos,
		E: anchor.X,
		F: anchor.Y,
	}
	negAnchor := geom.Affine{
		A: 1,
		D: 1,
		E: -anchor.X,
		F: -anchor.Y,
	}

	around := rot.Mul(negAnchor)
	return around.Mul(scale)
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func toRGBAColor(c render.Color) color.Color {
	red := toByte(c.R)
	green := toByte(c.G)
	blue := toByte(c.B)
	alpha := toByte(c.A)
	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}

func toByte(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func dataRange(data [][]float64) (min float64, max float64) {
	first := true
	for _, row := range data {
		for _, v := range row {
			if !isFinite(v) {
				continue
			}
			if first {
				min = v
				max = v
				first = false
				continue
			}
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
	}
	if first {
		return 0, 1
	}
	return min, max
}
