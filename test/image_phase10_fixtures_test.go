package test

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestImshowClipped_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_clipped", renderImshowClipped)
}

func TestImshowTransformed_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_transformed", renderImshowTransformed)
}

func TestImageAlpha_Golden(t *testing.T) {
	runGoldenTest(t, "image_alpha", renderImageAlpha)
}

func TestMatshowBasic_Golden(t *testing.T) {
	runGoldenTest(t, "matshow_basic", renderMatshowBasic)
}

func TestSpyMarker_Golden(t *testing.T) {
	runGoldenTest(t, "spy_marker", renderSpyMarker)
}

func TestSpyImage_Golden(t *testing.T) {
	runGoldenTest(t, "spy_image", renderSpyImage)
}

func renderImshowClipped() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Clipped Imshow")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	cmap := "viridis"
	nearest := "nearest"
	extent := [4]float64{0, 8, 0, 8}
	ax.ImShow(waveImageData(8, 8), core.ImShowOptions{
		Colormap:      &cmap,
		Extent:        &extent,
		Origin:        core.ImageOriginLower,
		Aspect:        "auto",
		Interpolation: &nearest,
	})
	ax.SetXLim(2, 6)
	ax.SetYLim(1, 7)

	return renderImageFixture(fig, 640, 360)
}

func renderImshowTransformed() image.Image {
	fig := core.NewFigure(420, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.16, Y: 0.14}, Max: geom.Pt{X: 0.9, Y: 0.88}})
	ax.SetTitle("Transformed Imshow")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-1, 5)
	ax.SetYLim(-1, 5)
	_ = ax.SetAspect("equal")

	cmap := "magma"
	vmin := 0.0
	vmax := 1.0
	xmin := 0.0
	xmax := 4.0
	ymin := 0.0
	ymax := 4.0
	angle := 28.0
	bilinear := "bilinear"
	ax.Image(waveImageData(6, 6), core.ImageOptions{
		Colormap:      &cmap,
		VMin:          &vmin,
		VMax:          &vmax,
		XMin:          &xmin,
		XMax:          &xmax,
		YMin:          &ymin,
		YMax:          &ymax,
		Origin:        core.ImageOriginLower,
		Angle:         &angle,
		Interpolation: &bilinear,
	})

	return renderImageFixture(fig, 420, 420)
}

func renderImageAlpha() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	ax.SetTitle("Image Alpha")
	ax.SetXLabel("column")
	ax.SetYLabel("row")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 6)

	lineColor := render.Color{R: 0.08, G: 0.08, B: 0.10, A: 1}
	lineWidth := 2.0
	ax.Plot([]float64{0, 6}, []float64{0, 6}, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
	})
	ax.Plot([]float64{0, 6}, []float64{6, 0}, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
	})

	cmap := "plasma"
	alpha := 0.45
	vmin := 0.0
	vmax := 1.0
	xmin := 0.0
	xmax := 6.0
	ymin := 0.0
	ymax := 6.0
	bilinear := "bilinear"
	ax.Image(waveImageData(6, 6), core.ImageOptions{
		Colormap:      &cmap,
		VMin:          &vmin,
		VMax:          &vmax,
		Alpha:         &alpha,
		XMin:          &xmin,
		XMax:          &xmax,
		YMin:          &ymin,
		YMax:          &ymax,
		Origin:        core.ImageOriginLower,
		Interpolation: &bilinear,
	})

	return renderImageFixture(fig, 640, 360)
}

func renderMatshowBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.22, Y: 0.12}, Max: geom.Pt{X: 0.78, Y: 0.9}})
	ax.SetTitle("Matshow")
	cmap := "cividis"
	ax.MatShow([][]float64{
		{0.10, 0.20, 0.35, 0.55},
		{0.18, 0.32, 0.48, 0.70},
		{0.28, 0.46, 0.66, 0.86},
		{0.40, 0.58, 0.78, 0.96},
	}, core.MatShowOptions{Colormap: &cmap})

	return renderImageFixture(fig, 640, 360)
}

func renderSpyMarker() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.22, Y: 0.14}, Max: geom.Pt{X: 0.78, Y: 0.9}})
	ax.SetTitle("Spy Marker")
	color := render.Color{R: 0.16, G: 0.38, B: 0.72, A: 1}
	marker := core.MarkerSquare
	ax.Spy(sparseFixtureData(14, 14), core.SpyOptions{
		Precision:  0.1,
		Marker:     &marker,
		MarkerSize: 8,
		Color:      &color,
	})

	return renderImageFixture(fig, 640, 360)
}

func renderSpyImage() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.22, Y: 0.14}, Max: geom.Pt{X: 0.78, Y: 0.9}})
	ax.SetTitle("Spy Image")
	useImage := true
	ax.Spy(sparseFixtureData(14, 14), core.SpyOptions{
		Precision: 0.1,
		UseImage:  &useImage,
	})

	return renderImageFixture(fig, 640, 360)
}

func renderImageFixture(fig *core.Figure, width, height int) image.Image {
	r, err := agg.New(width, height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func waveImageData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		yy := float64(row) / float64(max(1, rows-1))
		for col := range data[row] {
			xx := float64(col) / float64(max(1, cols-1))
			data[row][col] = 0.5 + 0.25*math.Sin((xx*2.4+0.15)*math.Pi) + 0.22*math.Cos((yy*2.7-0.2)*math.Pi)
		}
	}
	return data
}

func sparseFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			if row == col || row+col == cols-1 || (col+2*row)%5 == 0 || (2*col+row)%9 == 0 {
				data[row][col] = 1
			}
		}
	}
	return data
}
