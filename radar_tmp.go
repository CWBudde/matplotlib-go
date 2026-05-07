package main

import (
	"fmt"
	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func main() {
	cases := []string{"default", "ccw", "cw"}
	for _, v := range cases {
		fig := core.NewFigure(720, 720)
		labels := []string{"Speed", "Power", "Range", "Handling", "Comfort"}
		ax, err := fig.AddRadarAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.10}, Max: geom.Pt{X: 0.88, Y: 0.88}}, labels)
		if err != nil {
			panic(err)
		}
		if v == "cw" {
			_ = ax.SetThetaDirection("clockwise")
		}
		if v == "ccw" {
			_ = ax.SetThetaDirection("counterclockwise")
		}
		ax.SetTitle("Radar Projection")
		ax.YScale = transform.NewLinear(0, 1)
		ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0.25, 0.5, 0.75, 1.0}}
		ax.YAxis.MinorLocator = nil
		ax.YAxis.Formatter = core.PercentFormatter{Decimals: 0}

		angles := core.RadarAngles(len(labels))
		values := []float64{0.72, 0.88, 0.64, 0.79, 0.58}
		closedAngles := append(append([]float64(nil), angles...), angles[0])
		closedValues := append(append([]float64(nil), values...), values[0])
		lineColor := render.Color{R: 0.15, G: 0.35, B: 0.70, A: 1}
		fillColor := render.Color{R: 0.18, G: 0.50, B: 0.82, A: 0.22}
		lineWidth := 2.2
		ax.FillToBaseline(closedAngles, closedValues, core.FillOptions{Color: &fillColor})
		ax.Plot(closedAngles, closedValues, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth, Label: "model A"})

		r, _, err := backends.NewRendererFromEnv(backends.Config{Width: 720, Height: 720, Background: render.Color{R: 1, G: 1, B: 1, A: 1}, DPI: 100}, backends.TextCapabilities)
		if err != nil {
			panic(err)
		}
		if err := core.SavePNG(fig, r, "/tmp/radar_"+v+".png"); err != nil {
			panic(err)
		}
		fmt.Println("wrote", v)
	}
}
