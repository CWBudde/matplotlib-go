package scatter_marker_types

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 640
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Scatter Marker Types")
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 8)

	markerTypes := []core.MarkerType{
		core.MarkerCircle, core.MarkerSquare, core.MarkerTriangle,
		core.MarkerDiamond, core.MarkerPlus, core.MarkerCross,
	}
	colors := []render.Color{
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 1, G: 1, B: 0, A: 1},
		{R: 1, G: 0, B: 1, A: 1},
		{R: 0, G: 1, B: 1, A: 1},
	}

	for i, markerType := range markerTypes {
		lineWidth := 0.0
		if markerType == core.MarkerPlus || markerType == core.MarkerCross {
			lineWidth = 2.0
		}
		scatter := &core.Scatter2D{
			XY:        []geom.Pt{{X: float64(1 + i), Y: 4}},
			Size:      core.ScatterAreaFromRadius(12.0, style.Default.DPI),
			Color:     colors[i],
			EdgeColor: colors[i],
			EdgeWidth: lineWidth,
			Marker:    markerType,
			Alpha:     1.0,
		}
		ax.Add(scatter)
	}
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
