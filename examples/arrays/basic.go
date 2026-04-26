package main

import (
	"flag"
	"fmt"
	"math"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

const (
	widthPX  = 1240
	heightPX = 620
	dpi      = 100
)

type anchoredTextTarget interface {
	AddAnchoredText(string, ...core.AnchoredTextOptions) *core.AnchoredTextBox
}

// axesRect keeps the example coordinates readable as lower-left and upper-right
// figure fractions, matching the Python helper.
func axesRect(x0, y0, x1, y1 float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y0},
		Max: geom.Pt{X: x1, Y: y1},
	}
}

// ptr is local example glue for option fields where nil means "use default".
func ptr[T any](v T) *T {
	return &v
}

// arange mirrors numpy.arange for integer tick and edge coordinates.
func arange(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		values[i] = float64(i)
	}
	return values
}

// annotatedData is the scalar matrix shown in the first panel.
func annotatedData() [][]float64 {
	return [][]float64{
		{0.12, 0.28, 0.46, 0.64, 0.82},
		{0.18, 0.34, 0.58, 0.74, 0.88},
		{0.24, 0.42, 0.63, 0.79, 0.91},
		{0.16, 0.38, 0.61, 0.83, 0.97},
	}
}

// waveGrid creates the shared scalar field for pcolormesh and contour.
func waveGrid(rows, cols int, phase float64) [][]float64 {
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

// sparsePattern builds a deterministic non-zero pattern for the spy panel.
func sparsePattern(rows, cols int) [][]float64 {
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

// useMatrixTopAxis moves column ticks and the x label to the top, as matrix
// displays conventionally do.
func useMatrixTopAxis(ax *core.Axes) {
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

// setMatrixTicks pins matrix axes to integer row and column indices.
func setMatrixTicks(ax *core.Axes, rows, cols int) {
	xTicks := arange(cols)
	yTicks := arange(rows)
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

// sparseIndices mirrors np.where(data > precision), returning x and y
// coordinate slices for scatter.
func sparseIndices(data [][]float64, precision float64) ([]float64, []float64) {
	xx := []float64{}
	yy := []float64{}
	for y, row := range data {
		for x, value := range row {
			if value > precision {
				xx = append(xx, float64(x))
				yy = append(yy, float64(y))
			}
		}
	}
	return xx, yy
}

// addAnchoredText centralizes the boxed-note style used at axes and figure level.
func addAnchoredText(target anchoredTextTarget, text string, location core.LegendLocation) {
	target.AddAnchoredText(text, core.AnchoredTextOptions{
		Location:        location,
		Padding:         4,
		Inset:           6,
		CornerRadius:    4,
		BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		BorderColor:     render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
		FontSize:        10,
	})
}

func drawAnnotatedHeatmap(fig *core.Figure) {
	data := annotatedData()
	rows, cols := len(data), len(data[0])

	// Create the first axes and label it exactly like the Python version.
	ax := fig.AddAxes(axesRect(0.05, 0.14, 0.31, 0.88))
	ax.SetTitle("MatShow Annotated Heatmap")
	ax.SetXLabel("column index")
	ax.SetYLabel("row")

	// Render the matrix with imshow-like extents centered on integer cells.
	ax.Image(data, core.ImageOptions{
		Colormap: ptr("viridis"),
		XMin:     ptr(-0.5),
		XMax:     ptr(float64(cols) - 0.5),
		YMin:     ptr(-0.5),
		YMax:     ptr(float64(rows) - 0.5),
		Origin:   core.ImageOriginUpper,
	})
	ax.SetXLim(-0.5, float64(cols)-0.5)
	ax.SetYLim(-0.5, float64(rows)-0.5)
	ax.InvertY()
	_ = ax.SetAspect("equal")
	setMatrixTicks(ax, rows, cols)
	useMatrixTopAxis(ax)

	// Add one centered numeric label per cell, switching text color above the threshold.
	threshold := 0.5 * (0.12 + 0.97)
	for row := range rows {
		for col := range cols {
			textColor := render.Color{R: 0.12, G: 0.12, B: 0.14, A: 1}
			if data[row][col] >= threshold {
				textColor = render.Color{R: 1, G: 1, B: 1, A: 1}
			}
			ax.Text(float64(col), float64(row), fmt.Sprintf("%.2f", data[row][col]), core.TextOptions{
				HAlign:   core.TextAlignCenter,
				VAlign:   core.TextVAlignMiddle,
				FontSize: 10,
				Color:    textColor,
			})
		}
	}
}

func drawMeshAndContour(fig *core.Figure) {
	rows, cols := 8, 10
	data := waveGrid(rows, cols, 0.35)

	// Create the middle axes for the quad mesh and its contour overlay.
	ax := fig.AddAxes(axesRect(0.37, 0.14, 0.63, 0.88))
	ax.SetTitle("PColorMesh + Contour")
	ax.SetXLabel("x bin")
	ax.SetYLabel("y bin")

	// Draw filled cells with explicit edge coordinates, matching pcolormesh.
	ax.PColorMesh(data, core.MeshOptions{
		XEdges:    arange(cols + 1),
		YEdges:    arange(rows + 1),
		Colormap:  ptr("plasma"),
		EdgeColor: ptr(render.Color{R: 1, G: 1, B: 1, A: 0.48}),
		EdgeWidth: ptr(0.65),
		Label:     "pcolormesh",
	})

	// Draw contour lines and inline labels over the same scalar field.
	contourColor := ptr(render.Color{R: 0.14, G: 0.10, B: 0.16, A: 0.95})
	ax.Contour(data, core.ContourOptions{
		Color:         contourColor,
		LineWidth:     ptr(1.1),
		LevelCount:    6,
		LabelLines:    true,
		LabelColor:    contourColor,
		LabelFontSize: ptr(10.0),
	})
	// Keep mesh tick labels in the same 0..N coordinate space as Python.
	ax.SetXLim(0, float64(cols))
	ax.SetYLim(0, float64(rows))
}

func drawSpyMatrix(fig *core.Figure) {
	data := sparsePattern(18, 18)

	// Create the third axes for the sparse matrix view.
	ax := fig.AddAxes(axesRect(0.69, 0.14, 0.95, 0.88))
	ax.SetTitle("Spy Matrix")
	ax.SetXLabel("column index")
	ax.SetYLabel("row")

	// Convert non-zero matrix entries into square scatter markers.
	xx, yy := sparseIndices(data, 0.1)
	ax.Scatter(xx, yy, core.ScatterOptions{
		Size:      ptr(10.0),
		Color:     ptr(render.Color{R: 0.16, G: 0.38, B: 0.72, A: 1}),
		Marker:    ptr(core.MarkerSquare),
		EdgeWidth: ptr(0.0),
		Label:     "spy",
	})
	// Matrix coordinates increase downward, so the y-axis is inverted.
	ax.SetXLim(-0.5, 17.5)
	ax.SetYLim(17.5, -0.5)
	_ = ax.SetAspect("equal")
	setMatrixTicks(ax, 18, 18)
	useMatrixTopAxis(ax)

	// Add the small explanatory note inside the lower-right corner.
	addAnchoredText(ax, "sparse structure view", core.LegendLowerRight)
}

func saveFigure(fig *core.Figure, out string) error {
	// The Go backend path is explicit; Python hides the equivalent inside savefig.
	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      widthPX,
		Height:     heightPX,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        dpi,
	}, backends.TextCapabilities)
	if err != nil {
		return err
	}
	return core.SavePNG(fig, r, out)
}

func main() {
	out := flag.String("out", "arrays_basic.png", "output PNG path")
	flag.Parse()

	fig := core.NewFigure(widthPX, heightPX)
	drawAnnotatedHeatmap(fig)
	drawMeshAndContour(fig)
	drawSpyMatrix(fig)
	// Add the figure-level gallery label after all panels are in place.
	addAnchoredText(fig, "arrays gallery family\nheatmap, quad mesh, sparse matrix", core.LegendUpperRight)

	if err := saveFigure(fig, *out); err != nil {
		fmt.Printf("error saving figure: %v\n", err)
		return
	}
	fmt.Printf("saved %s\n", *out)
}
