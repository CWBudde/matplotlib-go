package main

import (
	"fmt"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/render"
	"matplotlib-go/style"
)

func main() {
	fig := core.NewFigure(960, 640)
	applyMatplotlibGridSpecStyle(fig)

	outer := fig.GridSpec(
		2,
		2,
		core.WithGridSpecPadding(0.08, 0.96, 0.10, 0.92),
		core.WithGridSpecSpacing(0.06/(2+0.06), 0.28/(2+0.28)),
		core.WithGridSpecWidthRatios(2, 1),
	)

	mainAx := outer.Span(0, 0, 2, 1).AddAxes()
	configureAxes(mainAx, "Main Span", []float64{0, 1, 2, 3, 4}, []float64{1.2, 2.8, 2.1, 3.6, 3.1}, render.Color{R: 0.15, G: 0.35, B: 0.72, A: 1})
	configureTicks(mainAx, []float64{0, 1, 2, 3, 4}, []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5}, "%.1f")

	nested := outer.Cell(0, 1).GridSpec(2, 1, core.WithGridSpecSpacing(0, 0.75/(2+0.75)))
	topRight := nested.Cell(0, 0).AddAxes()
	configureAxes(topRight, "Nested Top", []float64{0, 1, 2, 3}, []float64{3.4, 2.6, 2.9, 1.8}, render.Color{R: 0.72, G: 0.32, B: 0.18, A: 1})
	configureTicks(topRight, []float64{0, 1, 2, 3}, []float64{2, 3}, "%.0f")

	bottomRight := nested.Cell(1, 0).AddAxes(core.WithSharedX(topRight))
	configureAxes(bottomRight, "Nested Bottom", []float64{0, 1, 2, 3}, []float64{1.0, 1.6, 1.3, 2.2}, render.Color{R: 0.18, G: 0.55, B: 0.34, A: 1})
	configureTicks(bottomRight, []float64{0, 1, 2, 3}, []float64{1, 2}, "%.0f")

	sub := outer.Cell(1, 1).SubFigure()
	inset := sub.AddSubplot(1, 1, 1)
	configureAxes(inset, "SubFigure", []float64{0, 1, 2, 3}, []float64{2.0, 2.4, 1.9, 2.7}, render.Color{R: 0.55, G: 0.22, B: 0.50, A: 1})
	configureTicks(inset, []float64{0, 1, 2, 3}, []float64{2.0, 2.2, 2.4, 2.6}, "%.1f")

	r, err := agg.New(960, 640, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	if err := core.SavePNG(fig, r, "gridspec.png"); err != nil {
		panic(err)
	}
	fmt.Println("saved gridspec.png")
}

func applyMatplotlibGridSpecStyle(fig *core.Figure) {
	fig.RC = style.Apply(fig.RC, style.WithFont("DejaVu Sans", 10))
	fig.RC.TitleFontSize = 12 * fig.RC.DPI / 72
	fig.RC.AxisLabelFontSize = 10 * fig.RC.DPI / 72
	fig.RC.XTickLabelFontSize = 10 * fig.RC.DPI / 72
	fig.RC.YTickLabelFontSize = 10 * fig.RC.DPI / 72
}

func configureAxes(ax *core.Axes, title string, x, y []float64, c render.Color) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{
		Color:     &c,
		LineWidth: &width,
		Label:     title,
	})
	ax.AutoScale(0.10)
}

func configureTicks(ax *core.Axes, xTicks, yTicks []float64, yFormat string) {
	ax.XAxis.Locator = core.FixedLocator{TicksList: xTicks}
	ax.YAxis.Locator = core.FixedLocator{TicksList: yTicks}
	ax.YAxis.Formatter = core.FormatStrFormatter{Pattern: yFormat}
}
