package main

import (
	"fmt"
	"math"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(980, 640)

	host := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.14},
		Max: geom.Pt{X: 0.56, Y: 0.88},
	})
	host.SetTitle("AxisArtist / Parasite")
	host.SetXLabel("phase")
	host.SetYLabel("signal")
	host.SetXLim(-3.5, 3.5)
	host.SetYLim(-1.3, 1.3)
	host.AddYGrid()

	x := make([]float64, 240)
	sine := make([]float64, len(x))
	cosScaled := make([]float64, len(x))
	for i := range x {
		x[i] = -3.5 + 7*float64(i)/float64(len(x)-1)
		sine[i] = math.Sin(x[i])
		cosScaled[i] = 55 + 35*math.Cos(x[i]*0.8)
	}

	hostColor := render.Color{R: 0.14, G: 0.34, B: 0.72, A: 1}
	hostWidth := 2.2
	host.Plot(x, sine, core.PlotOptions{
		Color:     &hostColor,
		LineWidth: &hostWidth,
		Label:     "sin(x)",
	})

	floatX := host.FloatingXAxis(0)
	floatX.Axis.Color = render.Color{R: 0.26, G: 0.26, B: 0.30, A: 1}
	floatX.Axis.SetLineStyle(render.CapRound, render.JoinRound, 5, 3)
	_ = floatX.SetTickDirection("inout")

	floatY := host.FloatingYAxis(0)
	floatY.Axis.Color = render.Color{R: 0.26, G: 0.26, B: 0.30, A: 1}
	floatY.Axis.SetLineStyle(render.CapRound, render.JoinRound, 5, 3)
	_ = floatY.SetTickDirection("inout")

	parasite := host.ParasiteAxes(core.WithParasiteSharedX(host))
	if parasite != nil && parasite.Axes != nil {
		overlay := parasite.Axes
		overlay.SetYLim(0, 100)
		overlay.YAxis.Color = render.Color{R: 0.74, G: 0.28, B: 0.18, A: 1}
		overlay.YAxis.ShowSpine = false
		overlay.YAxis.ShowTicks = false
		overlay.YAxis.ShowLabels = false
		overlay.XAxis.ShowSpine = false
		overlay.XAxis.ShowTicks = false
		overlay.XAxis.ShowLabels = false

		right := overlay.RightAxis()
		right.Color = render.Color{R: 0.74, G: 0.28, B: 0.18, A: 1}
		right.SetLineStyle(render.CapRound, render.JoinRound)

		overlayColor := render.Color{R: 0.74, G: 0.28, B: 0.18, A: 1}
		overlayWidth := 1.8
		overlay.Plot(x, cosScaled, core.PlotOptions{
			Color:     &overlayColor,
			LineWidth: &overlayWidth,
			Label:     "55 + 35 cos(0.8x)",
		})
	}

	host.AddAnchoredText("floating axes at x=0 / y=0\nparasite right scale", core.AnchoredTextOptions{
		Location: core.LegendUpperLeft,
	})
	host.AddLegend()

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      980,
		Height:     640,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}
	if err := core.SavePNG(fig, r, "axisartist_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved axisartist_basic.png")
}
