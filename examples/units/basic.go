package main

import (
	"fmt"
	"time"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type distanceKM float64

type distanceConverter struct{}

func (distanceConverter) Convert(value any) (float64, error) {
	v, ok := value.(distanceKM)
	if !ok {
		return 0, fmt.Errorf("unexpected distance value %T", value)
	}
	return float64(v), nil
}

func (distanceConverter) AxisInfo([]float64) core.AxisInfo {
	return core.AxisInfo{
		Formatter: core.FormatStrFormatter{Pattern: "%.0f km"},
	}
}

func main() {
	core.MustRegisterUnitConverter(distanceKM(0), func() core.UnitsConverter {
		return distanceConverter{}
	})

	fig := core.NewFigure(1200, 420)

	dateAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.06, Y: 0.18}, Max: geom.Pt{X: 0.32, Y: 0.86}})
	dateAx.SetTitle("Dates")
	dateAx.SetYLabel("Requests")
	addReferenceYGrid(dateAx)
	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	lineWidth := 2.0
	// time.Time values are converted by the built-in date converter before the
	// normal line artist is created.
	_, _ = dateAx.PlotUnits(
		[]time.Time{
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
		[]float64{12, 18, 9, 21},
		core.PlotOptions{
			Color:     &lineColor,
			LineWidth: &lineWidth,
		},
	)
	dateAx.AutoScale(0.05)

	categoryAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.38, Y: 0.18}, Max: geom.Pt{X: 0.64, Y: 0.86}})
	categoryAx.SetTitle("Categories")
	categoryAx.SetYLabel("Count")
	addReferenceYGrid(categoryAx)
	barColor := render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1}
	barEdgeColor := render.Color{R: 0.60, G: 0.30, B: 0.03, A: 1}
	barEdgeWidth := 1.0
	barWidth := 0.8
	// String categories are assigned stable ordinal positions and matching tick
	// labels by the category converter.
	_, _ = categoryAx.BarUnits([]string{"alpha", "beta", "gamma", "delta"}, []float64{4, 9, 6, 7}, core.BarOptions{
		Color:     &barColor,
		EdgeColor: &barEdgeColor,
		EdgeWidth: &barEdgeWidth,
		Width:     &barWidth,
	})
	categoryAx.AutoScale(0.1)

	unitAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.70, Y: 0.18}, Max: geom.Pt{X: 0.96, Y: 0.86}})
	unitAx.SetTitle("Custom Units")
	unitAx.SetXLabel("Distance")
	unitAx.SetYLabel("Pace")
	addReferenceXYGrid(unitAx)
	scatterColor := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.92}
	scatterEdgeColor := render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1}
	scatterEdgeWidth := 1.0
	scatterSize := core.ScatterAreaFromRadius(8.0, 100.0)
	// distanceKM demonstrates a custom converter and axis formatter registered
	// above; the plotted artist receives converted float64 coordinates.
	_, _ = unitAx.ScatterUnits(
		[]distanceKM{5, 10, 21.1, 42.2},
		[]float64{6.4, 5.8, 5.2, 5.5},
		core.ScatterOptions{
			Color:     &scatterColor,
			Size:      &scatterSize,
			EdgeColor: &scatterEdgeColor,
			EdgeWidth: &scatterEdgeWidth,
		},
	)
	unitAx.AutoScale(0.08)

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:  1200,
		Height: 420,
	}, backends.TextCapabilities)
	if err != nil {
		panic(err)
	}

	if err := core.SavePNG(fig, r, "units_basic.png"); err != nil {
		panic(err)
	}
}

func addReferenceYGrid(ax *core.Axes) {
	grid := ax.AddGrid(core.AxisLeft)
	grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	grid.LineWidth = 0.5
}

func addReferenceXYGrid(ax *core.Axes) {
	xGrid := ax.AddGrid(core.AxisBottom)
	xGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	xGrid.LineWidth = 0.5
	addReferenceYGrid(ax)
}
