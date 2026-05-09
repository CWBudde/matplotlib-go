package test

import (
	"image"
	"math"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestBoundaryNormPColorMesh_Golden(t *testing.T) {
	runGoldenTest(t, "boundarynorm_pcolormesh", renderBoundaryNormPColorMesh)
}

func TestLogNormImshow_Golden(t *testing.T) {
	runGoldenTest(t, "lognorm_imshow", renderLogNormImshow)
}

func TestTwoSlopeNormImage_Golden(t *testing.T) {
	runGoldenTest(t, "twoslope_norm_image", renderTwoSlopeNormImage)
}

func TestColorbarExtensions_Golden(t *testing.T) {
	runGoldenTest(t, "colorbar_extensions", renderColorbarExtensions)
}

func renderBoundaryNormPColorMesh() image.Image {
	fig, ax := colorNormFixtureFigure("BoundaryNorm PColorMesh")
	cmap := "viridis"
	mesh := ax.PColorMesh([][]float64{
		{0.2, 0.8, 1.2, 1.8},
		{2.2, 2.8, 3.2, 3.8},
		{0.5, 1.5, 2.5, 3.5},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3, 4},
		YEdges:   []float64{0, 1, 2, 3},
		Colormap: &cmap,
		Norm: core.BoundaryNorm{
			Boundaries: []float64{0, 1, 2, 3, 4},
			NColors:    256,
		},
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "band"})
	}
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3)
	return renderFixtureFigure(fig, 640, 360)
}

func renderLogNormImshow() image.Image {
	fig, ax := colorNormFixtureFigure("LogNorm Imshow")
	cmap := "magma"
	xmin, xmax := 0.0, 6.0
	ymin, ymax := 0.0, 5.0
	img := ax.Image(logNormFixtureData(5, 6), core.ImageOptions{
		Colormap: &cmap,
		Norm:     core.LogNorm{VMin: 1, VMax: 1000},
		XMin:     &xmin,
		XMax:     &xmax,
		YMin:     &ymin,
		YMax:     &ymax,
		Origin:   core.ImageOriginLower,
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "log value"})
	}
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5)
	return renderFixtureFigure(fig, 640, 360)
}

func renderTwoSlopeNormImage() image.Image {
	fig, ax := colorNormFixtureFigure("TwoSlopeNorm Diverging")
	cmap := "diverging fixture"
	matcolor.RegisterColormap(cmap, matcolor.NewColormap(cmap, []matcolor.ColorStop{
		{Pos: 0.0, Color: render.Color{R: 0.23, G: 0.30, B: 0.75, A: 1}},
		{Pos: 0.5, Color: render.Color{R: 0.86, G: 0.86, B: 0.86, A: 1}},
		{Pos: 1.0, Color: render.Color{R: 0.71, G: 0.02, B: 0.15, A: 1}},
	}))
	xmin, xmax := 0.0, 7.0
	ymin, ymax := 0.0, 5.0
	img := ax.Image(twoSlopeFixtureData(5, 7), core.ImageOptions{
		Colormap: &cmap,
		Norm:     core.TwoSlopeNorm{VMin: -3, VCenter: 0, VMax: 6},
		XMin:     &xmin,
		XMax:     &xmax,
		YMin:     &ymin,
		YMax:     &ymax,
		Origin:   core.ImageOriginLower,
	})
	if img != nil {
		fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "anomaly"})
	}
	ax.SetXLim(0, 7)
	ax.SetYLim(0, 5)
	return renderFixtureFigure(fig, 640, 360)
}

func renderColorbarExtensions() image.Image {
	under := render.Color{R: 0.08, G: 0.16, B: 0.72, A: 1}
	over := render.Color{R: 0.78, G: 0.12, B: 0.08, A: 1}
	cmapName := "extension fixture"
	matcolor.RegisterColormap(cmapName, matcolor.GetColormap("viridis").Copy(cmapName).WithUnder(under).WithOver(over))

	fig, ax := colorNormFixtureFigure("Colorbar Extensions")
	vmin, vmax := 0.0, 1.0
	mesh := ax.PColorMesh([][]float64{
		{-0.35, 0.15, 0.35},
		{0.55, 0.85, 1.35},
	}, core.MeshOptions{
		XEdges:   []float64{0, 1, 2, 3},
		YEdges:   []float64{0, 1, 2},
		Colormap: &cmapName,
		VMin:     &vmin,
		VMax:     &vmax,
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "extended", Extend: "both"})
	}
	ax.SetXLim(0, 3)
	ax.SetYLim(0, 2)
	return renderFixtureFigure(fig, 640, 360)
}

func colorNormFixtureFigure(title string) (*core.Figure, *core.Axes) {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.88}})
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	return fig, ax
}

func logNormFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			t := float64(row*cols+col) / float64(rows*cols-1)
			data[row][col] = math.Pow(10, 3*t)
		}
	}
	return data
}

func twoSlopeFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		y := float64(row) / float64(max(1, rows-1))
		for col := range data[row] {
			x := float64(col) / float64(max(1, cols-1))
			data[row][col] = 6*x - 3 + 1.5*math.Sin((y-0.5)*math.Pi)
		}
	}
	return data
}
