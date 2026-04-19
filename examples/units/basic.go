package main

import (
	"fmt"
	"time"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
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
	_, _ = dateAx.PlotUnits(
		[]time.Time{
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
		[]float64{12, 18, 9, 21},
	)
	dateAx.AutoScale(0.05)

	categoryAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.38, Y: 0.18}, Max: geom.Pt{X: 0.64, Y: 0.86}})
	categoryAx.SetTitle("Categories")
	categoryAx.SetYLabel("Count")
	_, _ = categoryAx.BarUnits([]string{"alpha", "beta", "gamma", "delta"}, []float64{4, 9, 6, 7})
	categoryAx.AutoScale(0.1)

	unitAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.70, Y: 0.18}, Max: geom.Pt{X: 0.96, Y: 0.86}})
	unitAx.SetTitle("Custom Units")
	unitAx.SetXLabel("Distance")
	unitAx.SetYLabel("Pace")
	_, _ = unitAx.ScatterUnits(
		[]distanceKM{5, 10, 21.1, 42.2},
		[]float64{6.4, 5.8, 5.2, 5.5},
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
