package webdemo

import (
	"errors"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"slices"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

const (
	DefaultWidth  = 960
	DefaultHeight = 540
)

type Descriptor struct {
	ID          string
	Title       string
	Description string
}

var descriptors = []Descriptor{
	{
		ID:          "lines",
		Title:       "Signal Comparison",
		Description: "Multiple line series with shared axes, legend, and grids.",
	},
	{
		ID:          "scatter",
		Title:       "Scatter Clusters",
		Description: "Three deterministic point clouds with marker and alpha styling.",
	},
	{
		ID:          "bars",
		Title:       "Grouped Revenue Bars",
		Description: "Grouped bars with labels, grid lines, and category-style spacing.",
	},
	{
		ID:          "fills",
		Title:       "Filled Signals",
		Description: "An area band between two curves with outline styling and a legend.",
	},
	{
		ID:          "histogram",
		Title:       "Latency Histogram",
		Description: "A density-normalized histogram built from deterministic sample data.",
	},
	{
		ID:          "errorbars",
		Title:       "Measured Trend With Error Bars",
		Description: "Line, scatter, and symmetric uncertainty bars combined on one axes.",
	},
	{
		ID:          "heatmap",
		Title:       "Heatmap Surface",
		Description: "An image-based heatmap rendered through the plotting API.",
	},
	{
		ID:          "patches",
		Title:       "Patch Showcase",
		Description: "Rectangle, circle, ellipse, polygon, and arrow patches with a legend.",
	},
	{
		ID:          "polar",
		Title:       "Polar Wave",
		Description: "A filled polar curve with custom radial and angular grid styling.",
	},
	{
		ID:          "subplots",
		Title:       "Small Multiples",
		Description: "A compact 2×2 subplot layout showing shared limits and styling.",
	},
}

func Catalog() []Descriptor {
	out := make([]Descriptor, len(descriptors))
	copy(out, descriptors)
	return out
}

func Build(id string, width, height int) (*core.Figure, Descriptor, error) {
	if width <= 0 {
		width = DefaultWidth
	}
	if height <= 0 {
		height = DefaultHeight
	}

	for _, descriptor := range descriptors {
		if descriptor.ID != id {
			continue
		}

		var fig *core.Figure
		switch id {
		case "lines":
			fig = buildLinesDemo(width, height)
		case "scatter":
			fig = buildScatterDemo(width, height)
		case "bars":
			fig = buildBarsDemo(width, height)
		case "fills":
			fig = buildFillDemo(width, height)
		case "histogram":
			fig = buildHistogramDemo(width, height)
		case "errorbars":
			fig = buildErrorBarsDemo(width, height)
		case "heatmap":
			fig = buildHeatmapDemo(width, height)
		case "patches":
			fig = buildPatchesDemo(width, height)
		case "polar":
			fig = buildPolarDemo(width, height)
		case "subplots":
			fig = buildSubplotsDemo(width, height)
		default:
			return nil, Descriptor{}, fmt.Errorf("webdemo: unsupported demo %q", id)
		}

		return fig, descriptor, nil
	}

	return nil, Descriptor{}, fmt.Errorf("webdemo: unknown demo %q", id)
}

func Render(id string, width, height int) (*image.RGBA, Descriptor, error) {
	fig, descriptor, err := Build(id, width, height)
	if err != nil {
		return nil, Descriptor{}, err
	}

	renderWidth := width
	renderHeight := height
	if fig != nil {
		if w := int(fig.SizePx.X); w > 0 {
			renderWidth = w
		}
		if h := int(fig.SizePx.Y); h > 0 {
			renderHeight = h
		}
	}

	r := gobasic.New(renderWidth, renderHeight, render.Color{R: 1, G: 1, B: 1, A: 1})
	if r == nil {
		return nil, Descriptor{}, errors.New("webdemo: failed to create gobasic renderer")
	}

	core.DrawFigure(fig, r)
	return r.GetImage(), descriptor, nil
}

func DefaultDemoID() string {
	return descriptors[0].ID
}

func ValidDemoID(id string) bool {
	return slices.ContainsFunc(descriptors, func(descriptor Descriptor) bool {
		return descriptor.ID == id
	})
}

func buildLinesDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Signal Comparison")
	ax.SetXLabel("t")
	ax.SetYLabel("amplitude")
	ax.AddXGrid()
	ax.AddYGrid()

	const n = 160
	x := linspace(0, 12, n)
	sinY := make([]float64, n)
	cosY := make([]float64, n)
	dampedY := make([]float64, n)
	for i, xv := range x {
		sinY[i] = math.Sin(xv)
		cosY[i] = 0.7 * math.Cos(0.7*xv+0.3)
		dampedY[i] = math.Sin(1.6*xv) * math.Exp(-xv/11)
	}

	blue := render.Color{R: 0.16, G: 0.42, B: 0.82, A: 1}
	orange := render.Color{R: 0.91, G: 0.45, B: 0.16, A: 1}
	green := render.Color{R: 0.13, G: 0.62, B: 0.38, A: 1}
	lwWide := 3.0
	lw := 2.2
	ax.Plot(x, sinY, core.PlotOptions{Color: &blue, LineWidth: &lwWide, Label: "sin(t)"})
	ax.Plot(x, cosY, core.PlotOptions{Color: &orange, LineWidth: &lw, Dashes: []float64{10, 6}, Label: "0.7 cos(0.7t + 0.3)"})
	ax.Plot(x, dampedY, core.PlotOptions{Color: &green, LineWidth: &lw, Label: "damped"})
	ax.SetXLim(0, 12)
	ax.SetYLim(-1.4, 1.4)
	ax.AddLegend()
	return fig
}

func buildScatterDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Scatter Clusters")
	ax.SetXLabel("feature x")
	ax.SetYLabel("feature y")
	ax.AddXGrid()
	ax.AddYGrid()

	edge := render.Color{R: 1, G: 1, B: 1, A: 0.8}
	edgeWidth := 1.25
	alpha := 0.8

	xA, yA := scatterCluster(1, 11, -1.2, 0.5, 64)
	xB, yB := scatterCluster(2, 22, 1.0, 1.4, 64)
	xC, yC := scatterCluster(3, 33, 2.4, -0.8, 64)

	sizeA := 10.0
	sizeB := 12.0
	sizeC := 11.0
	diamond := core.MarkerDiamond
	triangle := core.MarkerTriangle
	square := core.MarkerSquare

	colA := render.Color{R: 0.13, G: 0.49, B: 0.92, A: 1}
	colB := render.Color{R: 0.93, G: 0.39, B: 0.26, A: 1}
	colC := render.Color{R: 0.24, G: 0.72, B: 0.42, A: 1}

	ax.Scatter(xA, yA, core.ScatterOptions{
		Color:     &colA,
		Size:      &sizeA,
		Marker:    &diamond,
		EdgeColor: &edge,
		EdgeWidth: &edgeWidth,
		Alpha:     &alpha,
		Label:     "cluster a",
	})
	ax.Scatter(xB, yB, core.ScatterOptions{
		Color:     &colB,
		Size:      &sizeB,
		Marker:    &triangle,
		EdgeColor: &edge,
		EdgeWidth: &edgeWidth,
		Alpha:     &alpha,
		Label:     "cluster b",
	})
	ax.Scatter(xC, yC, core.ScatterOptions{
		Color:     &colC,
		Size:      &sizeC,
		Marker:    &square,
		EdgeColor: &edge,
		EdgeWidth: &edgeWidth,
		Alpha:     &alpha,
		Label:     "cluster c",
	})
	ax.SetXLim(-3.2, 4.2)
	ax.SetYLim(-3.0, 3.4)
	ax.AddLegend()
	return fig
}

func buildBarsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Quarterly Revenue")
	ax.SetXLabel("quarter")
	ax.SetYLabel("EUR million")
	ax.AddYGrid()

	groupA := []float64{18, 24, 29, 34}
	groupB := []float64{14, 20, 27, 31}
	xA := []float64{-0.18, 0.82, 1.82, 2.82}
	xB := []float64{0.18, 1.18, 2.18, 3.18}
	widthBar := 0.34
	edgeWidth := 1.0
	edge := render.Color{R: 0.18, G: 0.18, B: 0.22, A: 0.7}
	blue := render.Color{R: 0.16, G: 0.42, B: 0.82, A: 0.9}
	orange := render.Color{R: 0.91, G: 0.45, B: 0.16, A: 0.9}

	seriesA := ax.Bar(xA, groupA, core.BarOptions{
		Width:     &widthBar,
		Color:     &blue,
		EdgeColor: &edge,
		EdgeWidth: &edgeWidth,
		Label:     "Product A",
	})
	seriesB := ax.Bar(xB, groupB, core.BarOptions{
		Width:     &widthBar,
		Color:     &orange,
		EdgeColor: &edge,
		EdgeWidth: &edgeWidth,
		Label:     "Product B",
	})

	ax.BarLabel(seriesA, []string{"18", "24", "29", "34"})
	ax.BarLabel(seriesB, []string{"14", "20", "27", "31"})
	centered := core.TextOptions{HAlign: core.TextAlignCenter}
	ax.Text(0, -2.5, "Q1", centered)
	ax.Text(1, -2.5, "Q2", centered)
	ax.Text(2, -2.5, "Q3", centered)
	ax.Text(3, -2.5, "Q4", centered)
	ax.SetXLim(-0.75, 3.75)
	ax.SetYLim(-4, 38)
	ax.AddLegend()
	return fig
}

func buildFillDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Filled Signals")
	ax.SetXLabel("t")
	ax.SetYLabel("value")
	ax.AddXGrid()
	ax.AddYGrid()

	const n = 180
	x := linspace(0, 2*math.Pi, n)
	upper := make([]float64, n)
	lower := make([]float64, n)
	for i, xv := range x {
		upper[i] = 0.85*math.Sin(xv) + 0.22*math.Cos(2*xv-0.4)
		lower[i] = -0.45*math.Cos(xv-0.2) - 0.18*math.Sin(2.4*xv)
	}

	fillColor := render.Color{R: 0.22, G: 0.60, B: 0.88, A: 1}
	fillEdge := render.Color{R: 0.09, G: 0.30, B: 0.48, A: 1}
	upperColor := render.Color{R: 0.10, G: 0.24, B: 0.62, A: 1}
	lowerColor := render.Color{R: 0.86, G: 0.34, B: 0.18, A: 1}
	fillAlpha := 0.30
	fillEdgeWidth := 1.1
	lineWidth := 2.2

	ax.FillBetween(x, upper, lower, core.FillOptions{
		Color:     &fillColor,
		EdgeColor: &fillEdge,
		EdgeWidth: &fillEdgeWidth,
		Alpha:     &fillAlpha,
		Label:     "band",
	})
	ax.Plot(x, upper, core.PlotOptions{
		Color:     &upperColor,
		LineWidth: &lineWidth,
		Label:     "upper",
	})
	ax.Plot(x, lower, core.PlotOptions{
		Color:     &lowerColor,
		LineWidth: &lineWidth,
		Dashes:    []float64{9, 5},
		Label:     "lower",
	})
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.25, 1.25)
	ax.AddLegend()
	return fig
}

func buildHistogramDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Latency Distribution")
	ax.SetXLabel("latency (ms)")
	ax.SetYLabel("density")
	ax.AddXGrid()
	ax.AddYGrid()

	data := deterministicNormal(400, 47.0, 8.5)
	bins := 24
	edgeWidth := 0.8
	fill := render.Color{R: 0.42, G: 0.23, B: 0.77, A: 0.7}
	edge := render.Color{R: 0.17, G: 0.12, B: 0.33, A: 1}
	ax.Hist(data, core.HistOptions{
		Bins:      bins,
		Norm:      core.HistNormDensity,
		Color:     &fill,
		EdgeColor: &edge,
		EdgeWidth: &edgeWidth,
		Label:     "requests",
	})
	ax.AutoScale(0.05)
	ax.AddLegend()
	return fig
}

func buildErrorBarsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Measured Trend With Error Bars")
	ax.SetXLabel("sample")
	ax.SetYLabel("response")
	ax.AddXGrid()
	ax.AddYGrid()

	x := []float64{1, 2, 3, 4, 5, 6}
	y := []float64{1.8, 2.5, 2.2, 3.1, 2.8, 3.7}
	xErr := []float64{0.20, 0.25, 0.15, 0.22, 0.30, 0.18}
	yErr := []float64{0.28, 0.20, 0.35, 0.24, 0.30, 0.22}

	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	pointColor := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	errorColor := render.Color{R: 0.10, G: 0.12, B: 0.16, A: 1}
	lineWidth := 2.1
	errorWidth := 1.2
	pointSize := 5.0
	capSize := 7.0

	ax.Plot(x, y, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "trend",
	})
	ax.Scatter(x, y, core.ScatterOptions{
		Color: &pointColor,
		Size:  &pointSize,
		Label: "samples",
	})
	ax.ErrorBar(x, y, xErr, yErr, core.ErrorBarOptions{
		Color:     &errorColor,
		LineWidth: &errorWidth,
		CapSize:   &capSize,
		Label:     "1sigma",
	})
	ax.SetXLim(0.4, 6.6)
	ax.SetYLim(1.0, 4.3)
	ax.AddLegend()
	return fig
}

func buildHeatmapDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Heatmap Surface")
	ax.SetXLabel("x")
	ax.SetYLabel("y")

	rows := 28
	cols := 36
	data := make([][]float64, rows)
	for row := range rows {
		data[row] = make([]float64, cols)
		y := -1 + 2*float64(row)/float64(rows-1)
		for col := range cols {
			x := -1 + 2*float64(col)/float64(cols-1)
			r1 := math.Hypot(x+0.35, y-0.15)
			r2 := math.Hypot(x-0.4, y+0.2)
			data[row][col] = math.Sin(8*r1)/(1+3*r1) + 0.8*math.Cos(7*r2)
		}
	}

	ax.Image(data, core.ImageOptions{Colormap: strPtr("inferno")})
	ax.SetXLim(0, float64(cols))
	ax.SetYLim(0, float64(rows))
	return fig
}

func buildPatchesDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Patch Showcase")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 4)

	ax.AddPatch(&core.Rectangle{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.95, G: 0.70, B: 0.23, A: 0.86},
			EdgeColor: render.Color{R: 0.48, G: 0.27, B: 0.08, A: 1},
			EdgeWidth: 1.1,
			Label:     "rectangle",
		},
		XY:     geom.Pt{X: 0.5, Y: 0.6},
		Width:  1.5,
		Height: 1.0,
	})
	ax.AddPatch(&core.Circle{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.22, G: 0.57, B: 0.82, A: 0.82},
			EdgeColor: render.Color{R: 0.11, G: 0.29, B: 0.44, A: 1},
			EdgeWidth: 1.0,
			Label:     "circle",
		},
		Center: geom.Pt{X: 2.8, Y: 1.25},
		Radius: 0.56,
	})
	ax.AddPatch(&core.Ellipse{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.23, G: 0.72, B: 0.51, A: 0.80},
			EdgeColor: render.Color{R: 0.10, G: 0.36, B: 0.24, A: 1},
			EdgeWidth: 1.0,
			Label:     "ellipse",
		},
		Center: geom.Pt{X: 4.9, Y: 2.85},
		Width:  1.4,
		Height: 0.92,
		Angle:  24,
	})
	ax.AddPatch(&core.Polygon{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.84, G: 0.34, B: 0.34, A: 0.82},
			EdgeColor: render.Color{R: 0.48, G: 0.14, B: 0.14, A: 1},
			EdgeWidth: 1.0,
			Label:     "polygon",
		},
		XY: []geom.Pt{
			{X: 1.6, Y: 3.0},
			{X: 2.2, Y: 2.1},
			{X: 0.9, Y: 2.3},
		},
	})
	ax.AddPatch(&core.FancyArrow{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.91, G: 0.42, B: 0.22, A: 0.88},
			EdgeColor: render.Color{R: 0.58, G: 0.22, B: 0.10, A: 1},
			EdgeWidth: 1.0,
			Label:     "arrow",
		},
		XY:         geom.Pt{X: 3.4, Y: 0.8},
		DX:         1.4,
		DY:         1.1,
		Width:      0.16,
		HeadWidth:  0.48,
		HeadLength: 0.42,
	})
	ax.AddLegend()
	return fig
}

func buildPolarDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddPolarAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	ax.SetTitle("Polar Wave")
	ax.SetXLabel("theta")
	ax.SetYLabel("radius")
	ax.YScale = transform.NewLinear(0, 1.1)

	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.80, G: 0.82, B: 0.86, A: 1}
	thetaGrid.LineWidth = 0.9

	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.82, G: 0.84, B: 0.88, A: 0.9}
	radiusGrid.LineWidth = 0.8

	const n = 720
	theta := make([]float64, n)
	radius := make([]float64, n)
	for i := range n {
		theta[i] = 2 * math.Pi * float64(i) / float64(n-1)
		radius[i] = 0.55 + 0.28*math.Cos(4*theta[i]) + 0.08*math.Sin(9*theta[i])
	}

	lineColor := render.Color{R: 0.16, G: 0.33, B: 0.73, A: 1}
	fillColor := render.Color{R: 0.36, G: 0.56, B: 0.92, A: 1}
	fillEdge := render.Color{R: 0.14, G: 0.25, B: 0.52, A: 1}
	lineWidth := 2.2
	fillAlpha := 0.24
	fillEdgeWidth := 1.0

	ax.FillToBaseline(theta, radius, core.FillOptions{
		Color:     &fillColor,
		EdgeColor: &fillEdge,
		EdgeWidth: &fillEdgeWidth,
		Alpha:     &fillAlpha,
		Label:     "filled area",
	})
	ax.Plot(theta, radius, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "r(theta)",
	})
	ax.AddLegend()
	return fig
}

func buildSubplotsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.08, 0.97, 0.10, 0.92),
		core.WithSubplotSpacing(0.08, 0.12),
		core.WithSubplotShareX(),
		core.WithSubplotShareY(),
	)

	palette := []render.Color{
		{R: 0.16, G: 0.42, B: 0.82, A: 1},
		{R: 0.91, G: 0.45, B: 0.16, A: 1},
		{R: 0.24, G: 0.72, B: 0.42, A: 1},
		{R: 0.80, G: 0.20, B: 0.42, A: 1},
	}

	x := linspace(0, 10, 120)
	idx := 0
	for row, rowAxes := range grid {
		for col, ax := range rowAxes {
			colr := palette[idx]
			ax.SetTitle(fmt.Sprintf("Panel %d", idx+1))
			ax.SetXLabel("x")
			ax.SetYLabel("y")
			ax.AddXGrid()
			ax.AddYGrid()

			y := make([]float64, len(x))
			for i, xv := range x {
				decay := math.Exp(-0.12 * float64(row+1) * xv / 10)
				y[i] = math.Sin(float64(col+1)*xv+float64(idx)*0.4) * decay
			}

			lw := 2.4
			ax.Plot(x, y, core.PlotOptions{Color: &colr, LineWidth: &lw})
			idx++
		}
	}

	grid[0][0].SetXLim(0, 10)
	grid[0][0].SetYLim(-1.25, 1.25)
	return fig
}

func defaultAxesRect() geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.96, Y: 0.90},
	}
}

func linspace(min, max float64, n int) []float64 {
	if n <= 1 {
		return []float64{min}
	}
	out := make([]float64, n)
	step := (max - min) / float64(n-1)
	for i := range n {
		out[i] = min + float64(i)*step
	}
	return out
}

func scatterCluster(seed1, seed2 uint64, centerX, centerY float64, n int) ([]float64, []float64) {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	x := make([]float64, n)
	y := make([]float64, n)
	for i := range n {
		x[i] = centerX + 0.65*normalSample(rng)
		y[i] = centerY + 0.55*normalSample(rng)
	}
	return x, y
}

func deterministicNormal(n int, mean, sigma float64) []float64 {
	rng := rand.New(rand.NewPCG(42, 7))
	out := make([]float64, n)
	for i := range n {
		out[i] = mean + sigma*normalSample(rng)
	}
	return out
}

func normalSample(rng *rand.Rand) float64 {
	u1 := rng.Float64()
	if u1 == 0 {
		u1 = math.SmallestNonzeroFloat64
	}
	u2 := rng.Float64()
	return math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
}

func strPtr(s string) *string {
	return &s
}
