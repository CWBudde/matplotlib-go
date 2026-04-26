package webdemo

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"math"
	"math/rand/v2"
	"slices"
	"time"

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
		ID:          "variants",
		Title:       "Plot Variants",
		Description: "Step, stairs, reference lines, spans, broken bars, and stacked bars.",
	},
	{
		ID:          "axes",
		Title:       "Axes, Scales, and Twins",
		Description: "Minor ticks, top/right axes, aspect controls, log scale, twin axes, and secondary axes.",
	},
	{
		ID:          "histogram",
		Title:       "Latency Histogram",
		Description: "A density-normalized histogram built from deterministic sample data.",
	},
	{
		ID:          "statistics",
		Title:       "Statistical Views",
		Description: "Box plots, violin plots, empirical CDFs, and stack plots.",
	},
	{
		ID:          "errorbars",
		Title:       "Measured Trend With Error Bars",
		Description: "Line, scatter, and symmetric uncertainty bars combined on one axes.",
	},
	{
		ID:          "units",
		Title:       "Dates and Categories",
		Description: "Time-aware axes, categorical bars, and horizontal categorical bars.",
	},
	{
		ID:          "heatmap",
		Title:       "Heatmap Surface",
		Description: "An image-based heatmap rendered through the plotting API.",
	},
	{
		ID:          "matrix",
		Title:       "Matrix Helpers",
		Description: "MatShow, sparsity spy plots, annotated heatmaps, and colorbars.",
	},
	{
		ID:          "mesh",
		Title:       "Meshes and Contours",
		Description: "PColorMesh, contour/contourf, Hist2D, triplot, tripcolor, and tricontour.",
	},
	{
		ID:          "vectors",
		Title:       "Vector Fields",
		Description: "Quiver, quiver keys, barbs, streamplots, and grid-based vector input.",
	},
	{
		ID:          "specialty",
		Title:       "Specialty Artists",
		Description: "Event plots, hexbin, pie charts, stem plots, tables, and Sankey-style flows.",
	},
	{
		ID:          "patches",
		Title:       "Patch Showcase",
		Description: "Rectangle, circle, ellipse, polygon, and arrow patches with a legend.",
	},
	{
		ID:          "annotations",
		Title:       "Text and Annotations",
		Description: "Data, axes, and figure coordinate text plus arrow annotations and anchored boxes.",
	},
	{
		ID:          "composition",
		Title:       "Figure Composition",
		Description: "GridSpec spans, figure-level labels, figure legends, anchored text, and colorbars.",
	},
	{
		ID:          "polar",
		Title:       "Polar Wave",
		Description: "A filled polar curve with custom radial and angular grid styling.",
	},
	{
		ID:          "phase7",
		Title:       "Projections and Insets",
		Description: "Phase 7 features: Mollweide geo projection plus a zoomed inset axes.",
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
		case "variants":
			fig = buildPlotVariantsDemo(width, height)
		case "axes":
			fig = buildAxesDemo(width, height)
		case "histogram":
			fig = buildHistogramDemo(width, height)
		case "statistics":
			fig = buildStatisticsDemo(width, height)
		case "errorbars":
			fig = buildErrorBarsDemo(width, height)
		case "units":
			fig = buildUnitsDemo(width, height)
		case "heatmap":
			fig = buildHeatmapDemo(width, height)
		case "matrix":
			fig = buildMatrixDemo(width, height)
		case "mesh":
			fig = buildMeshDemo(width, height)
		case "vectors":
			fig = buildVectorFieldsDemo(width, height)
		case "specialty":
			fig = buildSpecialtyDemo(width, height)
		case "patches":
			fig = buildPatchesDemo(width, height)
		case "annotations":
			fig = buildAnnotationsDemo(width, height)
		case "composition":
			fig = buildCompositionDemo(width, height)
		case "polar":
			fig = buildPolarDemo(width, height)
		case "phase7":
			fig = buildPhase7Demo(width, height)
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

func RenderPNG(id string, width, height int) ([]byte, Descriptor, error) {
	img, descriptor, err := Render(id, width, height)
	if err != nil {
		return nil, Descriptor{}, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, Descriptor{}, fmt.Errorf("webdemo: encode PNG: %w", err)
	}
	return buf.Bytes(), descriptor, nil
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

	ax.Image(data, core.ImageOptions{Colormap: strPtr("inferno"), Origin: core.ImageOriginLower})
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

func buildPhase7Demo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)

	geoAx, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.06, Y: 0.16},
		Max: geom.Pt{X: 0.48, Y: 0.84},
	}, "mollweide")
	if err == nil {
		geoAx.SetTitle("Mollweide Projection")
		geoAx.SetXLabel("longitude")
		geoAx.SetYLabel("latitude")
		gridColor := render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
		lonGrid := geoAx.AddGrid(core.AxisBottom)
		lonGrid.Color = gridColor
		lonGrid.LineWidth = 0.8
		latGrid := geoAx.AddGrid(core.AxisLeft)
		latGrid.Color = gridColor
		latGrid.LineWidth = 0.8

		lon := linspace(-math.Pi, math.Pi, 241)
		lat := make([]float64, len(lon))
		for i, v := range lon {
			lat[i] = 0.35 * math.Sin(3*v)
		}
		lineColor := render.Color{R: 0.14, G: 0.34, B: 0.70, A: 1}
		geoAx.Plot(lon, lat, core.PlotOptions{Color: &lineColor, LineWidth: floatPtr(2.0)})
	}

	mainAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.57, Y: 0.16},
		Max: geom.Pt{X: 0.96, Y: 0.84},
	})
	mainAx.SetTitle("Zoomed Inset")
	mainAx.SetXLabel("x")
	mainAx.SetYLabel("response")
	mainAx.SetXLim(0, 10)
	mainAx.SetYLim(-1.2, 1.2)
	mainAx.AddXGrid()
	mainAx.AddYGrid()

	x := linspace(0, 10, 320)
	y := make([]float64, len(x))
	for i, v := range x {
		y[i] = math.Sin(v) * (0.75 + 0.20*math.Cos(2*v))
	}
	lineColor := render.Color{R: 0.12, G: 0.36, B: 0.72, A: 1}
	mainAx.Plot(x, y, core.PlotOptions{Color: &lineColor, LineWidth: floatPtr(2.0)})

	inset, _ := mainAx.ZoomedInset(
		geom.Rect{Min: geom.Pt{X: 0.52, Y: 0.52}, Max: geom.Pt{X: 0.95, Y: 0.92}},
		[2]float64{2, 4},
		[2]float64{-0.2, 1.05},
	)
	if inset != nil {
		inset.SetTitle("detail")
		inset.Plot(x, y, core.PlotOptions{Color: &lineColor, LineWidth: floatPtr(1.6)})
		inset.AddXGrid()
		inset.AddYGrid()
	}

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

func buildPlotVariantsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.08, 0.97, 0.11, 0.91),
		core.WithSubplotSpacing(0.10, 0.16),
	)

	stepAx := grid[0][0]
	stepAx.SetTitle("Step + Stairs")
	stepAx.SetXLim(0, 6)
	stepAx.SetYLim(0, 5.2)
	stepAx.AddYGrid()
	where := core.StepWherePost
	stepAx.Step([]float64{0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4}, []float64{1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6}, core.StepOptions{
		Where:     &where,
		Color:     &render.Color{R: 0.15, G: 0.39, B: 0.78, A: 1},
		LineWidth: floatPtr(2.0),
		Label:     "step",
	})
	fill := true
	baseline := 0.35
	stepAx.Stairs([]float64{0.9, 1.7, 1.4, 2.6, 1.8, 2.2}, []float64{0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5}, core.StairsOptions{
		Fill:      &fill,
		Baseline:  &baseline,
		Color:     &render.Color{R: 0.91, G: 0.49, B: 0.20, A: 0.72},
		EdgeColor: &render.Color{R: 0.58, G: 0.26, B: 0.08, A: 1},
		LineWidth: floatPtr(1.5),
		Label:     "stairs",
	})
	stepAx.AddLegend()

	fillAx := grid[0][1]
	fillAx.SetTitle("FillBetweenX + Refs")
	fillAx.SetXLim(0, 7)
	fillAx.SetYLim(0, 6)
	fillAx.AddXGrid()
	fillAx.FillBetweenX(
		[]float64{0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2},
		[]float64{1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6},
		[]float64{3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3},
		core.FillOptions{
			Color:     &render.Color{R: 0.24, G: 0.68, B: 0.54, A: 0.72},
			EdgeColor: &render.Color{R: 0.12, G: 0.38, B: 0.28, A: 1},
			EdgeWidth: floatPtr(1.2),
		},
	)
	fillAx.AxVSpan(2.2, 3.1, core.VSpanOptions{
		Color: &render.Color{R: 0.92, G: 0.75, B: 0.18, A: 1},
		Alpha: floatPtr(0.20),
	})
	fillAx.AxHLine(4.0, core.HLineOptions{
		Color:     &render.Color{R: 0.52, G: 0.18, B: 0.18, A: 1},
		LineWidth: floatPtr(1.2),
		Dashes:    []float64{4, 3},
	})
	fillAx.AxLine(geom.Pt{X: 0.9, Y: 0.3}, geom.Pt{X: 6.4, Y: 5.6}, core.ReferenceLineOptions{
		Color:     &render.Color{R: 0.22, G: 0.22, B: 0.22, A: 1},
		LineWidth: floatPtr(1.1),
	})

	brokenAx := grid[1][0]
	brokenAx.SetTitle("Broken BarH")
	brokenAx.SetXLim(0, 10)
	brokenAx.SetYLim(0, 4.4)
	brokenAx.AddXGrid()
	first := brokenAx.BrokenBarH([][2]float64{{0.8, 1.6}, {3.1, 2.2}, {6.5, 1.3}}, [2]float64{0.7, 0.9}, core.BarOptions{
		Color: &render.Color{R: 0.21, G: 0.51, B: 0.76, A: 1},
	})
	second := brokenAx.BrokenBarH([][2]float64{{1.6, 1.0}, {4.0, 1.4}, {7.1, 1.7}}, [2]float64{2.1, 0.9}, core.BarOptions{
		Color: &render.Color{R: 0.86, G: 0.38, B: 0.16, A: 1},
	})
	brokenAx.BarLabel(first, []string{"prep", "run", "cool"}, core.BarLabelOptions{Position: "center", Color: render.Color{R: 1, G: 1, B: 1, A: 1}, FontSize: 10})
	brokenAx.BarLabel(second, []string{"IO", "fit", "ship"}, core.BarLabelOptions{Position: "center", Color: render.Color{R: 1, G: 1, B: 1, A: 1}, FontSize: 10})

	stackAx := grid[1][1]
	stackAx.SetTitle("Stacked Bars")
	stackAx.SetXLim(0.4, 4.6)
	stackAx.SetYLim(0, 7.6)
	stackAx.AddYGrid()
	x := []float64{1, 2, 3, 4}
	base := []float64{0, 0, 0, 0}
	seriesA := []float64{1.4, 2.2, 1.8, 2.5}
	seriesB := []float64{2.1, 1.6, 2.4, 1.7}
	bottom := stackAx.Bar(x, seriesA, core.BarOptions{Baselines: base, Color: &render.Color{R: 0.16, G: 0.59, B: 0.49, A: 1}})
	top := stackAx.Bar(x, seriesB, core.BarOptions{Baselines: seriesA, Color: &render.Color{R: 0.88, G: 0.47, B: 0.16, A: 1}})
	stackAx.BarLabel(bottom, []string{"A1", "A2", "A3", "A4"}, core.BarLabelOptions{Position: "center", Color: render.Color{R: 1, G: 1, B: 1, A: 1}, FontSize: 10})
	stackAx.BarLabel(top, nil, core.BarLabelOptions{Format: "%.1f", Color: render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1}})
	return fig
}

func buildAxesDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)

	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.14}, Max: geom.Pt{X: 0.42, Y: 0.86}})
	left.SetTitle("Top/Right + Equal Aspect")
	left.SetXLabel("top x")
	left.SetYLabel("right y")
	left.SetXLim(-1, 5)
	left.SetYLim(-1, 5)
	left.AddXGrid()
	left.AddYGrid()
	_ = left.SetXLabelPosition("top")
	_ = left.SetYLabelPosition("right")
	left.TopAxis().ShowLabels = true
	left.TopAxis().ShowTicks = true
	left.RightAxis().ShowLabels = true
	left.RightAxis().ShowTicks = true
	left.XAxis.ShowLabels = false
	left.XAxis.ShowTicks = false
	left.YAxis.ShowLabels = false
	left.YAxis.ShowTicks = false
	left.SetAxisEqual()
	_ = left.SetBoxAspect(1)
	_ = left.MinorticksOn("both")
	_ = left.LocatorParams(core.LocatorParams{Axis: "both", MajorCount: 6, MinorCount: 24})
	_ = left.TickParams(core.TickParams{Axis: "both", Which: "major", Length: floatPtr(7), Width: floatPtr(1.2)})
	_ = left.TickParams(core.TickParams{Axis: "both", Which: "minor", Length: floatPtr(4), Width: floatPtr(0.9)})
	left.Plot([]float64{-0.5, 0.8, 2.2, 4.2}, []float64{-0.2, 1.0, 2.1, 4.4}, core.PlotOptions{
		Color:     &render.Color{R: 0.10, G: 0.32, B: 0.76, A: 1},
		LineWidth: floatPtr(2),
	})
	left.Scatter([]float64{0, 1.5, 3.5, 4.5}, []float64{0, 1.8, 3.2, 4.6}, core.ScatterOptions{
		Color:     &render.Color{R: 0.92, G: 0.48, B: 0.20, A: 0.92},
		EdgeColor: &render.Color{R: 0.52, G: 0.22, B: 0.08, A: 1},
		EdgeWidth: floatPtr(1),
		Size:      floatPtr(8),
	})

	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.56, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.86}})
	right.SetTitle("Log, Twin, Secondary")
	right.SetXLabel("seconds")
	right.SetYLabel("count")
	right.SetXLim(0, 10)
	right.SetYLim(1, 100)
	_ = right.SetYScale("log")
	right.AddXGrid()
	right.AddYGrid()
	right.Plot([]float64{0, 2, 4, 6, 8, 10}, []float64{2, 6, 9, 18, 40, 82}, core.PlotOptions{
		Color:     &render.Color{R: 0.12, G: 0.45, B: 0.72, A: 1},
		LineWidth: floatPtr(2),
		Label:     "log series",
	})
	twin := right.TwinX()
	twin.SetYLim(0, 100)
	twinLineColor := render.Color{R: 0.80, G: 0.22, B: 0.22, A: 1}
	if axis := twin.RightAxis(); axis != nil {
		axis.MinorLocator = nil
	}
	twin.Plot([]float64{0, 2, 4, 6, 8, 10}, []float64{10, 22, 38, 58, 81, 96}, core.PlotOptions{
		Color:     &twinLineColor,
		LineWidth: floatPtr(1.8),
		Label:     "twin",
	})
	if sec, err := right.SecondaryXAxis(core.AxisTop, func(x float64) float64 { return x * 10 }, func(x float64) (float64, bool) { return x / 10, true }); err == nil {
		if axis := sec.TopAxis(); axis != nil {
			axis.MinorLocator = nil
		}
	}
	right.AddLegend()
	return fig
}

func buildStatisticsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(2, 2, core.WithSubplotPadding(0.08, 0.97, 0.10, 0.91), core.WithSubplotSpacing(0.10, 0.16))

	boxAx := grid[0][0]
	boxAx.SetTitle("Box + Violin")
	boxAx.SetXLim(0.4, 3.6)
	boxAx.SetYLim(0.6, 5.4)
	boxAx.AddYGrid()
	for i, data := range [][]float64{
		{1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2},
		{1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0},
		{2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6},
	} {
		pos := float64(i + 1)
		boxAx.BoxPlot(data, core.BoxPlotOptions{Position: &pos, Width: floatPtr(0.42), Color: &render.Color{R: 0.39, G: 0.62, B: 0.84, A: 0.38}})
	}
	boxAx.Violinplot([][]float64{
		{1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2},
		{1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0},
		{2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6},
	}, core.ViolinOptions{ShowMeans: boolPtr(true), ShowMedians: boolPtr(true), Alpha: 0.22})

	ecdfAx := grid[0][1]
	ecdfAx.SetTitle("ECDF")
	ecdfAx.SetXLim(0, 8)
	ecdfAx.SetYLim(0, 1.05)
	ecdfAx.AddYGrid()
	ecdfAx.ECDF([]float64{1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0}, core.ECDFOptions{
		Color:     &render.Color{R: 0.18, G: 0.36, B: 0.75, A: 1},
		LineWidth: floatPtr(2),
		Compress:  true,
	})

	stackAx := grid[1][0]
	stackAx.SetTitle("StackPlot")
	stackAx.SetXLim(0, 5)
	stackAx.SetYLim(0, 7)
	stackAx.AddYGrid()
	stackAx.StackPlot([]float64{0, 1, 2, 3, 4, 5}, [][]float64{
		{1.0, 1.4, 1.3, 1.8, 1.6, 2.0},
		{0.8, 1.1, 1.4, 1.2, 1.6, 1.8},
		{0.5, 0.8, 1.0, 1.4, 1.1, 1.5},
	}, core.StackPlotOptions{
		Colors: []render.Color{{R: 0.20, G: 0.55, B: 0.75, A: 1}, {R: 0.90, G: 0.48, B: 0.18, A: 1}, {R: 0.35, G: 0.66, B: 0.42, A: 1}},
		Alpha:  floatPtr(0.76),
	})

	histAx := grid[1][1]
	histAx.SetTitle("Cumulative Multi-Hist")
	histAx.SetXLim(0, 6)
	histAx.SetYLim(0, 6)
	histAx.AddYGrid()
	histAx.HistMulti([][]float64{
		{0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2},
		{0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0},
		{1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4},
	}, core.MultiHistOptions{
		BinEdges: []float64{0, 1, 2, 3, 4, 5, 6},
		Stacked:  true,
		Colors:   []render.Color{{R: 0.22, G: 0.55, B: 0.70, A: 0.8}, {R: 0.86, G: 0.42, B: 0.19, A: 0.8}, {R: 0.36, G: 0.62, B: 0.36, A: 0.8}},
	})
	return fig
}

func buildUnitsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(1, 3, core.WithSubplotPadding(0.06, 0.98, 0.17, 0.86), core.WithSubplotSpacing(0.10, 0.08))

	dateAx := grid[0][0]
	dateAx.SetTitle("Dates")
	dateAx.SetYLabel("requests")
	dateAx.AddYGrid()
	_, _ = dateAx.PlotUnits([]time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
	}, []float64{12, 18, 9, 21}, core.PlotOptions{
		Color:     &render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1},
		LineWidth: floatPtr(2),
	})
	_ = dateAx.TickParams(core.TickParams{Axis: "x", Which: "major", LabelRotation: floatPtr(30)})
	dateAx.AutoScale(0.05)

	categoryAx := grid[0][1]
	categoryAx.SetTitle("Categories")
	categoryAx.SetYLabel("count")
	categoryAx.AddYGrid()
	_, _ = categoryAx.BarUnits([]string{"draft", "review", "ship", "watch"}, []float64{3, 8, 6, 4}, core.BarOptions{
		Color:     &render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1},
		EdgeColor: &render.Color{R: 0.60, G: 0.30, B: 0.03, A: 1},
		EdgeWidth: floatPtr(1),
	})
	categoryAx.AutoScale(0.10)

	horizontalAx := grid[0][2]
	horizontalAx.SetTitle("Categorical Y")
	horizontalAx.SetXLabel("hours")
	horizontalAx.AddXGrid()
	orientation := core.BarHorizontal
	_, _ = horizontalAx.BarUnits([]string{"north", "south", "east"}, []float64{4, 7, 5}, core.BarOptions{
		Orientation: &orientation,
		Color:       &render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1},
		EdgeColor:   &render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1},
		EdgeWidth:   floatPtr(1),
	})
	horizontalAx.AutoScale(0.10)
	return fig
}

func buildMatrixDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(1, 3, core.WithSubplotPadding(0.07, 0.92, 0.14, 0.86), core.WithSubplotSpacing(0.10, 0.06))

	matAx := grid[0][0]
	matAx.SetTitle("MatShow")
	matAx.MatShow([][]float64{{0.1, 0.5, 0.9}, {0.7, 0.3, 0.2}, {0.4, 0.8, 0.6}}, core.MatShowOptions{Colormap: strPtr("viridis")})

	spyAx := grid[0][1]
	spyAx.SetTitle("Spy")
	spyAx.Spy([][]float64{{1, 0, 0, 2, 0}, {0, 0, 3, 0, 0}, {4, 0, 0, 0, 5}, {0, 6, 0, 0, 0}}, core.SpyOptions{
		Color:      &render.Color{R: 0.13, G: 0.43, B: 0.72, A: 1},
		MarkerSize: 8,
	})

	heatAx := grid[0][2]
	heatAx.SetTitle("Annotated Heatmap")
	img := heatAx.AnnotatedHeatmap([][]float64{{0.1, 0.7, 0.4}, {0.9, 0.2, 0.5}, {0.3, 0.8, 0.6}}, core.AnnotatedHeatmapOptions{
		MatShowOptions: core.MatShowOptions{Colormap: strPtr("magma")},
		Format:         "%.1f",
		FontSize:       9,
		TextColor:      render.Color{R: 0.05, G: 0.05, B: 0.05, A: 1},
		TextColorHigh:  render.Color{R: 1, G: 1, B: 1, A: 1},
	})
	if img != nil && img.Image != nil {
		fig.AddColorbar(heatAx, img.Image, core.ColorbarOptions{Label: "value"})
	}
	return fig
}

func buildMeshDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(2, 2, core.WithSubplotPadding(0.08, 0.97, 0.10, 0.91), core.WithSubplotSpacing(0.12, 0.16))

	meshEdgeWidth := 0.8
	meshEdgeColor := render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1}
	meshAx := grid[0][0]
	meshAx.SetTitle("PColorMesh")
	meshAx.SetXLim(0, 4)
	meshAx.SetYLim(0, 3)
	meshAx.PColorMesh([][]float64{{0.2, 0.6, 0.3, 0.9}, {0.4, 0.8, 0.5, 0.7}, {0.1, 0.3, 0.9, 0.6}}, core.MeshOptions{
		XEdges: []float64{0, 1, 2, 3, 4}, YEdges: []float64{0, 1, 2, 3}, EdgeColor: &meshEdgeColor, EdgeWidth: &meshEdgeWidth,
	})

	contourAx := grid[0][1]
	contourAx.SetTitle("Contour + Contourf")
	contourAx.SetXLim(0, 4)
	contourAx.SetYLim(0, 4)
	contourData := [][]float64{{0, 0.4, 0.8, 0.4, 0}, {0.2, 0.8, 1.3, 0.8, 0.2}, {0.3, 1.0, 1.7, 1.0, 0.3}, {0.2, 0.8, 1.3, 0.8, 0.2}, {0, 0.4, 0.8, 0.4, 0}}
	contourAx.Contourf(contourData, core.ContourOptions{Levels: []float64{0.2, 0.6, 1.0, 1.4, 1.8}})
	contourAx.Contour(contourData, core.ContourOptions{Levels: []float64{0.4, 0.8, 1.2, 1.6}, Color: &render.Color{R: 0.18, G: 0.18, B: 0.18, A: 1}})

	histAx := grid[1][0]
	histAx.SetTitle("Hist2D")
	histAx.SetXLim(0, 4)
	histAx.SetYLim(0, 4)
	histAx.Hist2D([]float64{0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6}, []float64{0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4}, core.Hist2DOptions{
		XBinEdges: []float64{0, 1, 2, 3, 4}, YBinEdges: []float64{0, 1, 2, 3, 4},
	})

	triAx := grid[1][1]
	triAx.SetTitle("Triangulation")
	triAx.SetXLim(0, 4)
	triAx.SetYLim(0, 4)
	tri := core.Triangulation{X: []float64{0.4, 1.6, 3.0, 0.8, 2.1, 3.5}, Y: []float64{0.5, 0.4, 0.7, 2.2, 2.8, 2.1}, Triangles: [][3]int{{0, 1, 3}, {1, 4, 3}, {1, 2, 4}, {2, 5, 4}}}
	values := []float64{0.2, 0.8, 1.0, 1.5, 1.1, 0.6}
	triAx.TriColor(tri, values)
	triAx.TriPlot(tri, core.TriPlotOptions{Color: &render.Color{R: 0.15, G: 0.15, B: 0.15, A: 1}, LineWidth: floatPtr(1)})
	triAx.TriContour(tri, values, core.ContourOptions{Levels: []float64{0.7, 1.1}, Color: &render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1}})
	return fig
}

func buildVectorFieldsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(2, 2, core.WithSubplotPadding(0.08, 0.97, 0.10, 0.91), core.WithSubplotSpacing(0.10, 0.16))

	quiverAx := grid[0][0]
	quiverAx.SetTitle("Quiver + Key")
	quiverAx.SetXLim(0, 6)
	quiverAx.SetYLim(0, 5)
	quiverAx.AddXGrid()
	quiverAx.AddYGrid()
	var qx, qy, qu, qv []float64
	for row := 0; row < 4; row++ {
		for col := 0; col < 5; col++ {
			x := 0.8 + float64(col)
			y := 0.8 + float64(row)*0.95
			qx, qy = append(qx, x), append(qy, y)
			qu, qv = append(qu, 0.55+0.08*math.Sin(y*0.9)), append(qv, 0.22*math.Cos(x*0.8))
		}
	}
	quiver := quiverAx.Quiver(qx, qy, qu, qv, core.QuiverOptions{Color: &render.Color{R: 0.14, G: 0.42, B: 0.73, A: 1}, Scale: floatPtr(10), ScaleUnits: "width", Units: "dots", Width: floatPtr(2.2)})
	if quiver != nil {
		quiverAx.QuiverKey(quiver, 0.78, 0.12, 0.5, "0.5", core.QuiverKeyOptions{Coords: core.Coords(core.CoordAxes), LabelPos: "E"})
	}

	barbAx := grid[0][1]
	barbAx.SetTitle("Barbs")
	barbAx.SetXLim(0, 6)
	barbAx.SetYLim(0, 5)
	barbAx.AddXGrid()
	barbAx.AddYGrid()
	var bx, by, bu, bv []float64
	for row := 0; row < 4; row++ {
		for col := 0; col < 5; col++ {
			x := 0.9 + float64(col)*0.95
			y := 0.8 + float64(row)*0.95
			bx, by = append(bx, x), append(by, y)
			bu, bv = append(bu, 14+5*math.Sin(y*0.8)), append(bv, 8*math.Cos(x*0.7))
		}
	}
	barbAx.Barbs(bx, by, bu, bv, core.BarbsOptions{BarbColor: &render.Color{R: 0.47, G: 0.23, B: 0.12, A: 1}, FlagColor: &render.Color{R: 0.86, G: 0.52, B: 0.24, A: 1}, Length: floatPtr(16), Units: "dots"})

	streamAx := grid[1][0]
	streamAx.SetTitle("Streamplot")
	streamAx.SetXLim(0, 6)
	streamAx.SetYLim(0, 5)
	sx := []float64{0, 1, 2, 3, 4, 5, 6}
	sy := []float64{0, 1, 2, 3, 4, 5}
	su, sv := vectorGrid(sx, sy)
	arrows := 2
	streamFalse := false
	streamAx.Streamplot(sx, sy, su, sv, core.StreamplotOptions{StartPoints: []geom.Pt{{X: 0.4, Y: 0.8}, {X: 0.4, Y: 2.2}, {X: 0.4, Y: 3.6}}, BrokenStreamlines: &streamFalse, IntegrationDirection: "forward", ArrowCount: &arrows, Color: &render.Color{R: 0.13, G: 0.53, B: 0.39, A: 1}})

	xyAx := grid[1][1]
	xyAx.SetTitle("Quiver Grid XY")
	xyAx.SetXLim(0, 6)
	xyAx.SetYLim(0, 5)
	xg := []float64{0.8, 1.8, 2.8, 3.8, 4.8}
	yg := []float64{0.8, 1.8, 2.8, 3.8}
	ugu := make([][]float64, len(yg))
	ugv := make([][]float64, len(yg))
	for yi, y := range yg {
		ugu[yi], ugv[yi] = make([]float64, len(xg)), make([]float64, len(xg))
		for xi, x := range xg {
			ugu[yi][xi] = -(y - 2.4) * 0.35
			ugv[yi][xi] = (x - 2.8) * 0.35
		}
	}
	xyAx.QuiverGrid(xg, yg, ugu, ugv, core.QuiverOptions{Color: &render.Color{R: 0.74, G: 0.23, B: 0.27, A: 1}, Pivot: "middle", Angles: "xy", Scale: floatPtr(9), ScaleUnits: "width", Units: "dots", Width: floatPtr(1.9)})
	return fig
}

func buildSpecialtyDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	grid := fig.Subplots(2, 3, core.WithSubplotPadding(0.07, 0.98, 0.09, 0.91), core.WithSubplotSpacing(0.10, 0.14))

	eventAx := grid[0][0]
	eventAx.SetTitle("Eventplot")
	eventAx.SetXLim(0, 10)
	eventAx.SetYLim(0.4, 3.6)
	eventAx.Eventplot([][]float64{{0.8, 1.4, 3.1, 4.6, 7.3}, {1.2, 2.9, 4.0, 6.4, 8.6}, {0.5, 2.2, 5.4, 6.8, 9.1}}, core.EventPlotOptions{
		LineOffsets: []float64{1, 2, 3},
		LineLengths: []float64{0.6, 0.7, 0.8},
		Colors:      []render.Color{{R: 0.18, G: 0.44, B: 0.74, A: 1}, {R: 0.84, G: 0.38, B: 0.16, A: 1}, {R: 0.20, G: 0.63, B: 0.42, A: 1}},
	})

	hexAx := grid[0][1]
	hexAx.SetTitle("Hexbin")
	hexAx.SetXLim(0, 1)
	hexAx.SetYLim(0, 1)
	hexAx.Hexbin([]float64{0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88}, []float64{0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58}, core.HexbinOptions{
		GridSizeX: 7, C: []float64{1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1}, Reduce: "mean",
	})

	pieAx := grid[0][2]
	pieAx.SetTitle("Pie")
	pieAx.Pie([]float64{28, 22, 18, 32}, core.PieOptions{Labels: []string{"Core", "I/O", "Render", "Docs"}, AutoPct: "%.0f%%", StartAngle: 90, LabelDistance: 1.08, Explode: []float64{0, 0.04, 0, 0.02}})

	stemAx := grid[1][0]
	stemAx.SetTitle("Stem")
	stemAx.SetXLim(0.5, 7.5)
	stemAx.SetYLim(-0.2, 4.2)
	stemAx.AddYGrid()
	stemAx.Stem([]float64{1, 2, 3, 4, 5, 6, 7}, []float64{0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1}, core.StemOptions{Color: &render.Color{R: 0.15, G: 0.42, B: 0.73, A: 1}, Baseline: floatPtr(0.3), MarkerSize: floatPtr(7)})

	tableAx := grid[1][1]
	tableAx.SetTitle("Table")
	hideAxes(tableAx)
	tableAx.Table(core.TableOptions{ColLabels: []string{"Metric", "Q1", "Q2"}, RowLabels: []string{"A", "B"}, CellText: [][]string{{"Latency", "18ms", "14ms"}, {"Throughput", "220/s", "265/s"}}, BBox: geom.Rect{Min: geom.Pt{X: 0.04, Y: 0.18}, Max: geom.Pt{X: 0.96, Y: 0.82}}})

	sankeyAx := grid[1][2]
	sankeyAx.SetTitle("Sankey")
	hideAxes(sankeyAx)
	builder := core.NewSankey(sankeyAx, core.SankeyOptions{Center: geom.Pt{X: 0.18, Y: 0.5}})
	builder.Add([]float64{-2, 3, 1.5}, core.SankeyAddOptions{Labels: []string{"Waste", "CPU", "Cache"}, Orientations: []int{-1, 1, 1}})
	return fig
}

func buildAnnotationsDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	ax := fig.AddAxes(defaultAxesRect())
	ax.SetTitle("Coordinate Text + Arrow Annotation")
	ax.SetXLabel("x")
	ax.SetYLabel("response")
	ax.AddXGrid()
	ax.AddYGrid()
	x := linspace(0, 8, 120)
	y := make([]float64, len(x))
	peakX, peakY := 0.0, -math.MaxFloat64
	for i, xv := range x {
		y[i] = math.Sin(xv) * math.Exp(-xv/8)
		if y[i] > peakY {
			peakX, peakY = xv, y[i]
		}
	}
	ax.Plot(x, y, core.PlotOptions{Color: &render.Color{R: 0.13, G: 0.43, B: 0.72, A: 1}, LineWidth: floatPtr(2.2), Label: "signal"})
	ax.Annotate("peak", peakX, peakY, core.AnnotationOptions{OffsetX: 44, OffsetY: -34, FontSize: 12})
	ax.Text(0.03, 0.94, "axes coords", core.TextOptions{Coords: core.Coords(core.CoordAxes), HAlign: core.TextAlignLeft, VAlign: core.TextVAlignTop, FontSize: 11})
	ax.Text(0.50, 0.10, "figure coords", core.TextOptions{Coords: core.Coords(core.CoordFigure), HAlign: core.TextAlignCenter, VAlign: core.TextVAlignBottom, FontSize: 11})
	ax.Text(6.0, -0.55, "data coords", core.TextOptions{FontSize: 11, Color: render.Color{R: 0.56, G: 0.22, B: 0.18, A: 1}})
	ax.AddAnchoredText("anchored\ntext box", core.AnchoredTextOptions{Location: core.LegendLowerRight})
	ax.SetXLim(0, 8)
	ax.SetYLim(-0.8, 1.1)
	ax.AddLegend()
	return fig
}

func buildCompositionDemo(width, height int) *core.Figure {
	fig := core.NewFigure(width, height)
	fig.SetSuptitle("GridSpec, Figure Labels, Legend, Colorbar")
	fig.SetSupXLabel("shared figure x")
	fig.SetSupYLabel("shared figure y")
	gs := fig.GridSpec(2, 3, core.WithGridSpecPadding(0.08, 0.92, 0.14, 0.86), core.WithGridSpecSpacing(0.08, 0.12), core.WithGridSpecWidthRatios(1.3, 1, 0.9))

	left := gs.Span(0, 0, 2, 1).AddAxes()
	left.SetTitle("spanning axes")
	left.AddYGrid()
	left.Plot([]float64{0, 1, 2, 3, 4}, []float64{1.0, 1.6, 1.2, 2.2, 1.8}, core.PlotOptions{Color: &render.Color{R: 0.16, G: 0.42, B: 0.82, A: 1}, LineWidth: floatPtr(2), Label: "left"})
	left.Scatter([]float64{0, 1, 2, 3, 4}, []float64{1.0, 1.6, 1.2, 2.2, 1.8}, core.ScatterOptions{Color: &render.Color{R: 0.91, G: 0.45, B: 0.16, A: 1}, Size: floatPtr(6), Label: "points"})
	left.AutoScale(0.10)

	top := gs.Cell(0, 1).AddAxes(core.WithSharedX(left))
	top.SetTitle("shared x")
	top.Plot([]float64{0, 1, 2, 3, 4}, []float64{2, 1, 2.4, 1.7, 2.8}, core.PlotOptions{Color: &render.Color{R: 0.23, G: 0.62, B: 0.34, A: 1}, LineWidth: floatPtr(1.8), Label: "top"})
	top.AutoScale(0.10)

	bottom := gs.Cell(1, 1).AddAxes()
	bottom.SetTitle("anchored")
	bottom.AddXGrid()
	bottom.AddYGrid()
	bottom.Plot([]float64{0, 1, 2, 3, 4}, []float64{3.0, 2.6, 2.9, 2.1, 1.9}, core.PlotOptions{Color: &render.Color{R: 0.69, G: 0.27, B: 0.67, A: 1}, LineWidth: floatPtr(1.8), Label: "bottom"})
	bottom.AddAnchoredText("axes note", core.AnchoredTextOptions{Location: core.LegendUpperLeft})
	bottom.AutoScale(0.10)

	heat := gs.Span(0, 2, 2, 1).AddAxes()
	heat.SetTitle("colorbar")
	img := heat.MatShow([][]float64{{0.2, 0.5, 0.7}, {0.9, 0.4, 0.1}, {0.3, 0.8, 0.6}}, core.MatShowOptions{Colormap: strPtr("inferno")})
	if img != nil {
		fig.AddColorbar(heat, img, core.ColorbarOptions{Label: "intensity"})
	}
	fig.AddLegend()
	fig.AddAnchoredText("figure note", core.AnchoredTextOptions{Location: core.LegendLowerRight})
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

func vectorGrid(x, y []float64) ([][]float64, [][]float64) {
	u := make([][]float64, len(y))
	v := make([][]float64, len(y))
	for yi, yv := range y {
		u[yi] = make([]float64, len(x))
		v[yi] = make([]float64, len(x))
		for xi, xv := range x {
			u[yi][xi] = 1.0 + 0.12*math.Cos(yv*0.7)
			v[yi][xi] = 0.35*math.Sin((xv-3)*0.8) - 0.10*(yv-2.5)
		}
	}
	return u, v
}

func hideAxes(ax *core.Axes) {
	if ax == nil {
		return
	}
	ax.ShowFrame = false
	if ax.XAxis != nil {
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
	}
	if ax.YAxis != nil {
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func strPtr(s string) *string {
	return &s
}
