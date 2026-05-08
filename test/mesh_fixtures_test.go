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

func TestPColorFlat_Golden(t *testing.T) {
	runGoldenTest(t, "pcolor_flat", renderPColorFlat)
}

func TestPColorMeshNearest_Golden(t *testing.T) {
	runGoldenTest(t, "pcolormesh_nearest", renderPColorMeshNearest)
}

func TestPColorMeshGouraud_Golden(t *testing.T) {
	runGoldenTest(t, "pcolormesh_gouraud", renderPColorMeshGouraud)
}

func TestPColorMeshMasked_Golden(t *testing.T) {
	runGoldenTest(t, "pcolormesh_masked", renderPColorMeshMasked)
}

func TestHist2DWeightedDensity_Golden(t *testing.T) {
	runGoldenTest(t, "hist2d_weighted_density", renderHist2DWeightedDensity)
}

func renderPColorFlat() image.Image {
	fig, ax := meshFixtureFigure("pcolor flat")
	cmap := "plasma"
	edge := render.Color{R: 0.96, G: 0.96, B: 0.96, A: 1}
	width := 0.75
	ax.PColor(meshFixtureData(4, 5), core.MeshOptions{
		XEdges:    []float64{0.0, 0.9, 2.0, 3.4, 4.1, 5.2},
		YEdges:    []float64{-0.2, 0.8, 1.6, 2.9, 4.0},
		Shading:   core.MeshShadingFlat,
		Colormap:  &cmap,
		EdgeColor: &edge,
		EdgeWidth: &width,
	})
	ax.SetXLim(0, 5.2)
	ax.SetYLim(-0.2, 4.0)
	return renderFixtureFigure(fig, 640, 360)
}

func renderPColorMeshNearest() image.Image {
	fig, ax := meshFixtureFigure("nearest centers")
	cmap := "viridis"
	vmin, vmax := -0.75, 0.95
	ax.PColorMesh(meshFixtureData(4, 5), core.MeshOptions{
		XEdges:   []float64{0.0, 0.8, 2.1, 3.5, 5.0},
		YEdges:   []float64{-1.0, 0.2, 1.4, 2.7},
		Shading:  core.MeshShadingNearest,
		Colormap: &cmap,
		VMin:     &vmin,
		VMax:     &vmax,
	})
	ax.SetXLim(-0.4, 5.7)
	ax.SetYLim(-1.6, 3.35)
	return renderFixtureFigure(fig, 640, 360)
}

func renderPColorMeshGouraud() image.Image {
	fig, ax := meshFixtureFigure("Gouraud mesh")
	cmap := "viridis"
	vmin, vmax := -0.85, 0.85
	mesh := ax.PColorMesh(meshFixtureData(5, 6), core.MeshOptions{
		XEdges:   []float64{-2.5, -1.5, -0.4, 0.8, 1.7, 2.8},
		YEdges:   []float64{-1.8, -0.8, 0.1, 1.1, 2.0},
		Shading:  core.MeshShadingGouraud,
		Colormap: &cmap,
		VMin:     &vmin,
		VMax:     &vmax,
	})
	if mesh != nil {
		fig.AddColorbar(ax, mesh, core.ColorbarOptions{Label: "value"})
	}
	ax.SetXLim(-2.5, 2.8)
	ax.SetYLim(-1.8, 2.0)
	return renderFixtureFigure(fig, 640, 360)
}

func renderPColorMeshMasked() image.Image {
	bad := render.Color{R: 0.62, G: 0.62, B: 0.62, A: 0.78}
	cmapName := "mesh fixture mask"
	matcolor.RegisterColormap(cmapName, matcolor.GetColormap("viridis").WithBad(bad))

	fig, ax := meshFixtureFigure("masked mesh")
	edge := render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1}
	width := 0.7
	ax.PColorMesh(meshFixtureData(4, 5), core.MeshOptions{
		XEdges:    []float64{0, 1, 2, 3, 4, 5},
		YEdges:    []float64{0, 1, 2, 3, 4},
		Colormap:  &cmapName,
		EdgeColor: &edge,
		EdgeWidth: &width,
		Mask: [][]bool{
			{false, true, false, false, false},
			{false, false, false, true, false},
			{true, false, false, false, false},
			{false, false, true, false, false},
		},
	})
	ax.SetXLim(0, 5)
	ax.SetYLim(0, 4)
	return renderFixtureFigure(fig, 640, 360)
}

func renderHist2DWeightedDensity() image.Image {
	fig, ax := meshFixtureFigure("hist2d weighted density")
	x, y, weights := hist2DWeightedData()
	cmap := "magma"
	result := ax.Hist2D(x, y, core.Hist2DOptions{
		XBinEdges: []float64{-2, -1, 0, 1, 2, 3},
		YBinEdges: []float64{-1.5, -0.5, 0.5, 1.5, 2.5},
		Weights:   weights,
		Norm:      core.HistNormDensity,
		Colormap:  &cmap,
	})
	if result != nil {
		fig.AddColorbar(ax, result.Mesh, core.ColorbarOptions{Label: "density"})
	}
	ax.SetXLim(-2, 3)
	ax.SetYLim(-1.5, 2.5)
	return renderFixtureFigure(fig, 640, 360)
}

func meshFixtureFigure(title string) (*core.Figure, *core.Axes) {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.88}})
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	return fig, ax
}

func meshFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for yi := range data {
		data[yi] = make([]float64, cols)
		y := float64(yi) / float64(max(1, rows-1))
		for xi := range data[yi] {
			x := float64(xi) / float64(max(1, cols-1))
			data[yi][xi] = 0.65*math.Sin((x*2.3+0.15)*math.Pi) + 0.28*math.Cos((y*2.1-0.25)*math.Pi)
		}
	}
	return data
}

func hist2DWeightedData() (x, y, weights []float64) {
	x = []float64{-1.8, -1.4, -0.8, -0.3, 0.2, 0.7, 1.1, 1.6, 2.1, 2.5, -0.6, 0.4, 1.3, 2.7}
	y = []float64{-1.1, -0.4, -0.8, 0.1, 0.5, 0.9, 1.2, 1.7, 2.0, 2.2, 0.7, -0.2, 0.4, 1.3}
	weights = []float64{0.8, 1.3, 0.7, 1.1, 1.6, 0.9, 1.4, 1.2, 1.8, 0.6, 1.5, 0.9, 1.1, 1.7}
	return x, y, weights
}
