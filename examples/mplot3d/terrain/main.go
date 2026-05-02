package main

import (
	"log"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(900, 640)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.08},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	})
	if err != nil {
		log.Fatalf("add 3D axes: %v", err)
	}

	ax.SetTitle("3D Surface + Filled Contours")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(35, -60)

	// Use the same deterministic terrain formula as the Python counterpart so
	// surface, contour, and contourf behavior can be compared directly.
	x, y, z := sinusoidalTerrain(90, 70)
	ax.PlotSurfaceGrid(x, y, z)

	// Additional primitives exercise mixed 3D artist ordering on top of the
	// surface: a floor outline, sample points, a triangular patch, and text.
	ax.Plot3D([]float64{0, 0.9, 0.9, 0, 0}, []float64{0, 0, 0.9, 0.9, 0}, []float64{-0.2, -0.2, -0.2, -0.2, -0.2})
	ax.Scatter3D([]float64{0.2, 0.5, 0.8}, []float64{0.2, 0.5, 0.8}, []float64{0.3, 0.35, 0.2})
	ax.Contour(x, y, z)
	ax.Contourf(x, y, z)

	tri := core.Triangulation{
		X:         []float64{0, 0.5, 1},
		Y:         []float64{0, 0, 0.4},
		Triangles: [][3]int{{0, 1, 2}},
	}
	triZ := []float64{0.1, 0.4, 0.9}
	ax.Trisurf(tri, triZ)
	ax.Text3D(0.9, 0.1, 0.65, "3D demo")

	r, err := agg.New(900, 640, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		log.Fatal(err)
	}
	core.DrawFigure(fig, r)
	if err := r.SavePNG("mplot3d_terrain.png"); err != nil {
		log.Fatalf("save PNG: %v", err)
	}
}

func sinusoidalTerrain(xCount, yCount int) ([]float64, []float64, [][]float64) {
	if xCount < 2 {
		xCount = 2
	}
	if yCount < 2 {
		yCount = 2
	}
	x := make([]float64, xCount)
	y := make([]float64, yCount)
	z := make([][]float64, yCount)

	for yi := 0; yi < yCount; yi++ {
		y[yi] = -math.Pi + 2*math.Pi*float64(yi)/float64(yCount-1)
	}
	for xi := 0; xi < xCount; xi++ {
		x[xi] = -math.Pi + 2*math.Pi*float64(xi)/float64(xCount-1)
	}
	for yi := 0; yi < yCount; yi++ {
		row := make([]float64, xCount)
		for xi := 0; xi < xCount; xi++ {
			angleX := x[xi]
			angleY := y[yi]
			row[xi] = 0.5*math.Sin(angleX)*math.Cos(angleY) +
				0.35*math.Sin(2*angleX+0.6)*math.Cos(angleY/2) +
				0.15*math.Cos(3*angleY-angleX)
		}
		z[yi] = row
	}
	return x, y, z
}
