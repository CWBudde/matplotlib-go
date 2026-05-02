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
	fig := core.NewFigure(720, 420)
	ax, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.92, Y: 0.86},
	}, "hammer")
	if err != nil {
		log.Fatal(err)
	}
	ax.SetTitle("Hammer Projection")
	ax.SetXLabel("longitude")
	ax.SetYLabel("latitude")

	gridColor := render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	lonGrid := ax.AddGrid(core.AxisBottom)
	lonGrid.Color = gridColor
	lonGrid.LineWidth = 0.8
	latGrid := ax.AddGrid(core.AxisLeft)
	latGrid.Color = gridColor
	latGrid.LineWidth = 0.8

	lon, lat := denseGeoPath()
	lineColor := render.Color{R: 0.14, G: 0.34, B: 0.70, A: 1}
	lineWidth := 2.0
	ax.Plot(lon, lat, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
	})

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		log.Fatal(err)
	}
	core.DrawFigure(fig, r)
	if err := r.SavePNG("hammer.png"); err != nil {
		log.Fatal(err)
	}
}

func denseGeoPath() ([]float64, []float64) {
	const n = 361
	lon := make([]float64, n)
	lat := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		lon[i] = -math.Pi + 2*math.Pi*t
		lat[i] = 0.35 * math.Sin(3*lon[i])
	}
	return lon, lat
}
