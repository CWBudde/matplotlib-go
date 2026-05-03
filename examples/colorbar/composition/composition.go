package composition

import (
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1000
	Height = 700
	DPI    = 100
)

// ColorbarComposition builds the same plot as
// test/matplotlib_ref/plots/colorbar_composition.py.
func ColorbarComposition() *core.Figure {
	fig := core.NewFigure(Width, Height)
	fig.ConstrainedLayout()
	ax := fig.AddSubplot(1, 1, 1)

	rows, cols := 80, 120
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			x := (float64(col)/float64(cols-1))*4 - 2
			y := (float64(row)/float64(rows-1))*4 - 2
			r := math.Hypot(x, y)
			data[row][col] = math.Sin(3*r) * math.Exp(-0.6*r)
		}
	}

	cmap := "inferno"
	xMin := 0.0
	xMax := float64(cols)
	yMin := 0.0
	yMax := float64(rows)
	im := ax.Image(data, core.ImageOptions{
		Colormap: &cmap,
		XMin:     &xMin,
		XMax:     &xMax,
		YMin:     &yMin,
		YMax:     &yMax,
		Origin:   core.ImageOriginLower,
	})

	ax.SetTitle("Heatmap with Colorbar")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, float64(cols))
	ax.SetYLim(0, float64(rows))
	yTicks := make([]float64, 0, rows/20+1)
	for tick := 0; tick <= rows; tick += 20 {
		yTicks = append(yTicks, float64(tick))
	}
	ax.YAxis.Locator = core.FixedLocator{TicksList: yTicks}

	gridColor := render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = gridColor
		grid.LineWidth = 0.5
	}

	cbar := fig.AddColorbar(ax, im)
	cbar.SetYLabel("Intensity")

	return fig
}
