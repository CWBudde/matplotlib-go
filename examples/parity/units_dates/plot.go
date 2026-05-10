package units_dates

import (
	"image"
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/parity/internal/common"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(720, 380)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Date Units")
	ax.SetXLabel("Date")
	ax.SetYLabel("Requests")
	common.AddReferenceYGrid(ax)

	dates := []time.Time{
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC),
	}
	x := make([]float64, len(dates))
	for i, d := range dates {
		x[i] = common.ReferenceDateNumber(d)
	}
	lower := []float64{6, 7, 5, 8, 7}
	upper := []float64{10, 15, 13, 18, 16}
	fillColor := render.Color{R: 0.85, G: 0.91, B: 0.96, A: 1}
	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	lineWidth := 2.0
	ax.FillBetween(x, lower, upper, core.FillOptions{Color: &fillColor})
	ax.Plot(x, []float64{8, 12, 9, 15, 13}, core.PlotOptions{Color: &lineColor, LineWidth: &lineWidth})
	ax.XAxis.Locator = core.FixedLocator{TicksList: []float64{
		common.ReferenceDateNumber(time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC)),
		common.ReferenceDateNumber(time.Date(2024, 2, 12, 0, 0, 0, 0, time.UTC)),
		common.ReferenceDateNumber(time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC)),
	}}
	ax.XAxis.Formatter = core.DateFormatter{Layout: "02 Jan", Location: time.UTC}
	ax.AutoScale(0.06)
	return common.RenderFixtureFigure(fig, 720, 380)
}
