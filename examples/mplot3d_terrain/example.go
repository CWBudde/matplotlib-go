package mplot3d_terrain

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

func Render() image.Image {
	fig := core.NewFigure(900, 640)
	ax, err := fig.AddAxes3D(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.08}, Max: geom.Pt{X: 0.92, Y: 0.88}})
	if err != nil {
		panic(err)
	}
	ax.SetTitle("3D Surface + Filled Contours")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(35, -60)
	x, y, z := common.SinusoidalTerrain(90, 70)
	zeroWidth := 0.0
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	contourWidth := 0.6
	contourLevels := 8
	contourOffset := common.GridMin(z) - 0.2
	orange := render.Color{R: 1.0, G: 0.4980392156862745, B: 0.054901960784313725, A: 1}
	triAlpha := 0.7
	ax.PlotSurfaceGrid(x, y, z, core.PlotOptions{LineWidth: &zeroWidth})
	ax.Plot3D([]float64{0, 0.9, 0.9, 0, 0}, []float64{0, 0, 0.9, 0.9, 0}, []float64{-0.2, -0.2, -0.2, -0.2, -0.2}, core.PlotOptions{Color: &black})
	ax.Scatter3D([]float64{0.2, 0.5, 0.8}, []float64{0.2, 0.5, 0.8}, []float64{0.3, 0.35, 0.2})
	ax.Contour(x, y, z, core.PlotOptions{Color: &black, LineWidth: &contourWidth, LevelCount: contourLevels})
	ax.Contourf(x, y, z, core.PlotOptions{LevelCount: contourLevels, Offset: &contourOffset})
	tri := core.Triangulation{X: []float64{0, 0.5, 1}, Y: []float64{0, 0, 0.4}, Triangles: [][3]int{{0, 1, 2}}}
	ax.Trisurf(tri, []float64{0.1, 0.4, 0.9}, core.PlotOptions{Color: &orange, Alpha: &triAlpha})
	ax.Text(0.70, 0.62, "3D demo", core.TextOptions{Coords: core.Coords(core.CoordAxes)})
	return common.RenderFixtureFigure(fig, 900, 640)
}
