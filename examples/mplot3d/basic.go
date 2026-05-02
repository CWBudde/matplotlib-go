package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(760, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		fmt.Printf("Error creating 3D axes: %v\n", err)
		return
	}

	ax.SetTitle("3D Toolkit Scaffold")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(30, -60)

	// Small shared data set mirrors the Python meshgrid example and touches each
	// scaffolded 3D artist without making the rendering hard to inspect.
	x := []float64{0, 1}
	y := []float64{0, 1}
	zGrid := [][]float64{
		{0, 1},
		{1, 2},
	}
	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	_ = line
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	ax.Wireframe(x, y, zGrid)
	ax.Surface(x, y, zGrid)
	ax.Contour(x, y, zGrid)
	ax.Bar3D([]float64{0.2}, []float64{0.3}, []float64{0.4}, []float64{0.2}, []float64{0.2}, []float64{0.3})
	ax.Text3D(0.2, 0.8, 0.6, "demo point")

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      760,
		Height:     560,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "mplot3d_basic.png"); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Println("Created mplot3d_basic.png")
}
