package test

import (
	"fmt"
	"image"
	"math"
	"testing"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestUnstructuredShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "unstructured_showcase", renderUnstructuredShowcase)
}

func TestArraysShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "arrays_showcase", renderArraysShowcase)
}

func TestAxisArtistShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axisartist_showcase", renderAxisArtistShowcase)
}

func TestAxesGrid1Showcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_grid1_showcase", renderAxesGrid1Showcase)
}

func renderUnstructuredShowcase() image.Image {
	fig := core.NewFigure(1320, 520)

	tri, values := parityTriangulation()

	meshAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.16},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	configureParityUnstructuredAxes(meshAx, "Triangulation")
	meshColor := render.Color{R: 0.18, G: 0.24, B: 0.34, A: 1}
	meshWidth := 1.35
	meshAx.TriPlot(tri, core.TriPlotOptions{
		Color:     &meshColor,
		LineWidth: &meshWidth,
		Label:     "triplot",
	})
	meshAx.AddAnchoredText("explicit triangular mesh", core.AnchoredTextOptions{
		Location: core.LegendLowerRight,
	})

	colorAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.37, Y: 0.16},
		Max: geom.Pt{X: 0.63, Y: 0.88},
	})
	configureParityUnstructuredAxes(colorAx, "Tripcolor + Tricontour")
	cmap := "viridis"
	edgeColor := render.Color{R: 1, G: 1, B: 1, A: 0.55}
	edgeWidth := 0.6
	colorAx.TriColor(tri, values, core.TriColorOptions{
		Colormap:  &cmap,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
		Label:     "tripcolor",
	})
	contourColor := render.Color{R: 0.08, G: 0.12, B: 0.18, A: 0.95}
	contourWidth := 1.15
	colorAx.TriContour(tri, values, core.ContourOptions{
		Color:      &contourColor,
		LineWidth:  &contourWidth,
		LevelCount: 6,
		LabelLines: true,
		LabelColor: &contourColor,
	})

	fillAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.69, Y: 0.16},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	configureParityUnstructuredAxes(fillAx, "Filled Tricontour")
	fillMap := "plasma"
	fillAx.TriContourf(tri, values, core.ContourOptions{
		Colormap:   &fillMap,
		LevelCount: 7,
		Label:      "tricontourf",
	})
	highlight := render.Color{R: 1, G: 1, B: 1, A: 0.88}
	highlightWidth := 0.95
	fillAx.TriContour(tri, values, core.ContourOptions{
		Color:      &highlight,
		LineWidth:  &highlightWidth,
		LevelCount: 7,
	})

	fig.AddAnchoredText("unstructured gallery family\ntriangulation, tripcolor, tricontour", core.AnchoredTextOptions{
		Location: core.LegendUpperRight,
	})

	r, err := agg.New(1320, 520, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderArraysShowcase() image.Image {
	fig := core.NewFigure(1240, 620)

	heatAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.14},
		Max: geom.Pt{X: 0.31, Y: 0.88},
	})
	heatAx.SetTitle("Annotated Heatmap")
	heatAx.SetXLabel("column")
	heatAx.SetYLabel("row")
	heatMap := "viridis"
	heatData := parityAnnotatedData()
	heatRows, heatCols := len(heatData), len(heatData[0])
	heatAx.Image(heatData, core.ImageOptions{
		Colormap: &heatMap,
		XMin:     parityFloatPtr(-0.5),
		XMax:     parityFloatPtr(float64(heatCols) - 0.5),
		YMin:     parityFloatPtr(-0.5),
		YMax:     parityFloatPtr(float64(heatRows) - 0.5),
		Origin:   core.ImageOriginUpper,
	})
	heatAx.SetXLim(-0.5, float64(heatCols)-0.5)
	heatAx.SetYLim(float64(heatRows)-0.5, -0.5)
	_ = heatAx.SetAspect("equal")
	paritySetMatrixTicks(heatAx, heatRows, heatCols)
	parityUseMatrixTopAxis(heatAx)
	heatThreshold := 0.5 * (0.12 + 0.97)
	for row := range heatRows {
		for col := range heatCols {
			textColor := render.Color{R: 0.12, G: 0.12, B: 0.14, A: 1}
			if heatData[row][col] >= heatThreshold {
				textColor = render.Color{R: 1, G: 1, B: 1, A: 1}
			}
			heatAx.Text(float64(col), float64(row), fmt.Sprintf("%.2f", heatData[row][col]), core.TextOptions{
				HAlign:   core.TextAlignCenter,
				VAlign:   core.TextVAlignMiddle,
				FontSize: 10,
				Color:    textColor,
			})
		}
	}

	meshAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.37, Y: 0.14},
		Max: geom.Pt{X: 0.63, Y: 0.88},
	})
	meshAx.SetTitle("PColorMesh + Contour")
	meshAx.SetXLabel("x bin")
	meshAx.SetYLabel("y bin")
	meshAx.SetXLim(0, 10)
	meshAx.SetYLim(0, 8)
	meshMap := "plasma"
	meshEdges := render.Color{R: 1, G: 1, B: 1, A: 0.48}
	meshEdgeWidth := 0.65
	meshData := parityWaveGrid(8, 10, 0.35)
	meshAx.PColorMesh(meshData, core.MeshOptions{
		Colormap:  &meshMap,
		EdgeColor: &meshEdges,
		EdgeWidth: &meshEdgeWidth,
		Label:     "pcolormesh",
	})
	contourColor := render.Color{R: 0.14, G: 0.10, B: 0.16, A: 0.95}
	contourWidth := 1.1
	meshAx.Contour(meshData, core.ContourOptions{
		Color:      &contourColor,
		LineWidth:  &contourWidth,
		LevelCount: 6,
		LabelLines: true,
		LabelColor: &contourColor,
	})

	spyAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.69, Y: 0.14},
		Max: geom.Pt{X: 0.95, Y: 0.88},
	})
	spyAx.SetTitle("Spy")
	spyAx.SetXLabel("column")
	spyAx.SetYLabel("row")
	spyColor := render.Color{R: 0.16, G: 0.38, B: 0.72, A: 1}
	spyMarker := core.MarkerSquare
	spyAx.Spy(paritySparsePattern(18, 18), core.SpyOptions{
		Precision:  0.1,
		Marker:     &spyMarker,
		MarkerSize: 10,
		Color:      &spyColor,
		Label:      "spy",
	})
	spyAx.AddAnchoredText("sparse structure view", core.AnchoredTextOptions{
		Location: core.LegendLowerRight,
	})

	fig.AddAnchoredText("arrays gallery family\nheatmap, quad mesh, sparse matrix", core.AnchoredTextOptions{
		Location: core.LegendUpperRight,
	})

	r, err := agg.New(1240, 620, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(100)
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderAxisArtistShowcase() image.Image {
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
	floatX.Axis.ShowTicks = false
	floatX.Axis.ShowLabels = false
	_ = floatX.SetTickDirection("inout")

	floatY := host.FloatingYAxis(0)
	floatY.Axis.Color = render.Color{R: 0.26, G: 0.26, B: 0.30, A: 1}
	floatY.Axis.SetLineStyle(render.CapRound, render.JoinRound, 5, 3)
	floatY.Axis.ShowTicks = false
	floatY.Axis.ShowLabels = false
	_ = floatY.SetTickDirection("inout")

	overlay := host.TwinX()
	if overlay != nil {
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
	legend := host.AddLegend()
	legend.SetLocator(core.RelativeAnchoredBoxLocator{
		X:       0.5,
		Y:       0,
		OffsetY: 10,
		HAlign:  core.BoxAlignCenter,
		VAlign:  core.BoxAlignTop,
	})

	r, err := agg.New(980, 640, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderAxesGrid1Showcase() image.Image {
	fig := core.NewFigure(1100, 720)

	grid := fig.NewImageGrid(
		2,
		2,
		geom.Rect{
			Min: geom.Pt{X: 0.06, Y: 0.12},
			Max: geom.Pt{X: 0.60, Y: 0.88},
		},
		core.WithAxesDividerHorizontalSpace(0.18/11.0),
		core.WithAxesDividerVerticalSpace(0.20/7.2),
	)
	if grid == nil {
		panic("image grid creation failed")
	}

	for row := range 2 {
		for col := range 2 {
			ax := grid.At(row, col)
			ax.SetTitle("Tile " + string(rune('1'+row)) + "," + string(rune('1'+col)))
			ax.ImShow(parityGridSurface(24, 24, float64(row*2+col)))
			ax.AddAnchoredText("image grid", core.AnchoredTextOptions{
				Location:        core.LegendLowerRight,
				Locator:         core.NewAnchoredOffsetLocator(core.LegendLowerRight, 3, 1, 3),
				Padding:         2,
				CornerRadius:    4,
				BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 1},
				BorderColor:     render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
				TextColor:       render.Color{R: 0, G: 0, B: 0, A: 1},
				FontSize:        10,
			})
		}
	}

	channels := []struct {
		ax    *core.Axes
		title string
		cmap  string
	}{
		{
			ax: fig.AddAxes(geom.Rect{
				Min: geom.Pt{X: 0.66, Y: 0.34},
				Max: geom.Pt{X: 0.75, Y: 0.56},
			}),
			title: "Red",
			cmap:  "red channel",
		},
		{
			ax: fig.AddAxes(geom.Rect{
				Min: geom.Pt{X: 0.775, Y: 0.34},
				Max: geom.Pt{X: 0.865, Y: 0.56},
			}),
			title: "Green",
			cmap:  "green channel",
		},
		{
			ax: fig.AddAxes(geom.Rect{
				Min: geom.Pt{X: 0.89, Y: 0.34},
				Max: geom.Pt{X: 0.98, Y: 0.56},
			}),
			title: "Blue",
			cmap:  "blue channel",
		},
	}
	for idx, channel := range channels {
		channel.ax.SetTitle(channel.title)
		channel.ax.XAxis.Locator = core.FixedLocator{TicksList: []float64{0, 10, 20}}
		channel.ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 10, 20}}
		cmap := channel.cmap
		channel.ax.ImShow(parityChannelSurface(28, 28, idx), core.ImShowOptions{
			Colormap: &cmap,
		})
	}

	notePad := 0.35 * 11 * fig.RC.DPI / 72
	noteInset := 0.5 * 11 * fig.RC.DPI / 72
	fig.AddAnchoredText("axes_grid1-style layout\nImageGrid + RGB channel views", core.AnchoredTextOptions{
		Location:        core.LegendUpperRight,
		Padding:         notePad,
		Inset:           noteInset,
		BoxPadding:      notePad,
		CornerRadius:    notePad,
		BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		BorderColor:     render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
		FontSize:        11,
	})

	r, err := agg.New(1100, 720, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func configureParityUnstructuredAxes(ax *core.Axes, title string) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-0.1, 3.1)
	ax.SetYLim(-0.15, 2.65)
	_ = ax.SetAspect("equal")
}

func parityTriangulation() (core.Triangulation, []float64) {
	tri := core.Triangulation{
		X: []float64{0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55},
		Y: []float64{0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05},
		Triangles: [][3]int{
			{0, 1, 4},
			{1, 5, 4},
			{1, 2, 5},
			{2, 6, 5},
			{2, 3, 6},
			{4, 5, 7},
			{5, 8, 7},
			{5, 6, 8},
			{6, 9, 8},
		},
	}

	values := make([]float64, len(tri.X))
	for i := range values {
		values[i] = math.Sin(tri.X[i]*1.4) + 0.7*math.Cos((tri.Y[i]+0.15)*2.1)
	}
	return tri, values
}

func parityAnnotatedData() [][]float64 {
	return [][]float64{
		{0.12, 0.28, 0.46, 0.64, 0.82},
		{0.18, 0.34, 0.58, 0.74, 0.88},
		{0.24, 0.42, 0.63, 0.79, 0.91},
		{0.16, 0.38, 0.61, 0.83, 0.97},
	}
}

func parityWaveGrid(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.55 + 0.25*math.Sin((xx*2.3+phase)*math.Pi) + 0.20*math.Cos((yy*2.8-phase*0.4)*math.Pi)
		}
	}
	return values
}

func parityGridSurface(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.5 + 0.25*math.Sin((xx+phase)*2*math.Pi) + 0.25*math.Cos((yy+phase*0.3)*3*math.Pi)
		}
	}
	return values
}

func parityChannelSurface(rows, cols, channel int) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			switch channel {
			case 1:
				values[y][x] = 0.5 + 0.32*math.Sin(yy*4*math.Pi) + 0.18*math.Cos(xx*2*math.Pi)
			case 2:
				dx := xx - 0.5
				dy := yy - 0.5
				values[y][x] = 0.58 + 0.36*math.Sin((xx+yy)*3*math.Pi) - 0.18*math.Hypot(dx, dy)
			default:
				dx := xx - 0.35
				dy := yy - 0.42
				values[y][x] = 0.35 + 0.65*math.Exp(-7*(dx*dx+dy*dy))
			}
		}
	}
	return values
}

func paritySparsePattern(rows, cols int) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		for x := range cols {
			if x == y || x+y == cols-1 || (x+2*y)%7 == 0 || (2*x+y)%11 == 0 {
				values[y][x] = 1
			}
		}
	}
	return values
}

func parityFloatPtr(v float64) *float64 {
	return &v
}

func parityUseMatrixTopAxis(ax *core.Axes) {
	if ax.XAxis != nil {
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
	}
	top := ax.TopAxis()
	top.ShowSpine = true
	top.ShowTicks = true
	top.ShowLabels = true
	_ = ax.SetXLabelPosition("top")
}

func paritySetMatrixTicks(ax *core.Axes, rows, cols int) {
	xTicks := parityArange(cols)
	yTicks := parityArange(rows)
	for _, axis := range []*core.Axis{ax.XAxis, ax.XAxisTop} {
		if axis != nil {
			axis.Locator = core.FixedLocator{TicksList: xTicks}
			axis.Formatter = core.ScalarFormatter{Prec: 0}
		}
	}
	if ax.YAxis != nil {
		ax.YAxis.Locator = core.FixedLocator{TicksList: yTicks}
		ax.YAxis.Formatter = core.ScalarFormatter{Prec: 0}
	}
}

func parityArange(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		values[i] = float64(i)
	}
	return values
}
