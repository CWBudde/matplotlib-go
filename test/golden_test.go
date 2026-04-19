package test

import (
	"flag"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/test/imagecmp"
	"matplotlib-go/transform"
)

var updateGolden = flag.Bool("update-golden", false, "Update golden images instead of comparing")

func TestBasicLine_Golden(t *testing.T) {
	runGoldenTest(t, "basic_line", renderBasicLine)
}

func TestJoinsCaps_Golden(t *testing.T) {
	runGoldenTest(t, "joins_caps", renderJoinsCaps)
}

func TestDashes_Golden(t *testing.T) {
	runGoldenTest(t, "dashes", renderDashes)
}

func TestScatterBasic_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_basic", renderScatterBasic)
}

func TestScatterMarkerTypes_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_marker_types", renderScatterMarkerTypes)
}

func TestScatterAdvanced_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_advanced", renderScatterAdvanced)
}

func TestBarBasicFrame_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_frame", renderBarBasicFrame)
}

func TestBarBasicTicks_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_ticks", renderBarBasicTicks)
}

func TestBarBasicTickLabels_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_tick_labels", renderBarBasicTickLabels)
}

func TestBarBasicTitle_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_title", renderBarBasicTitle)
}

func TestBarBasic_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic", renderBarBasic)
}

func TestBarHorizontal_Golden(t *testing.T) {
	runGoldenTest(t, "bar_horizontal", renderBarHorizontal)
}

func TestBarGrouped_Golden(t *testing.T) {
	runGoldenTest(t, "bar_grouped", renderBarGrouped)
}

func TestFillBasic_Golden(t *testing.T) {
	runGoldenTest(t, "fill_basic", renderFillBasic)
}

func TestFillBetween_Golden(t *testing.T) {
	runGoldenTest(t, "fill_between", renderFillBetween)
}

func TestFillStacked_Golden(t *testing.T) {
	runGoldenTest(t, "fill_stacked", renderFillStacked)
}

func TestMultiSeriesBasic_Golden(t *testing.T) {
	runGoldenTest(t, "multi_series_basic", renderMultiSeriesBasic)
}

func TestMultiSeriesColorCycle_Golden(t *testing.T) {
	runGoldenTest(t, "multi_series_color_cycle", renderMultiSeriesColorCycle)
}

func TestHistBasic_Golden(t *testing.T) {
	runGoldenTest(t, "hist_basic", renderHistBasic)
}

func TestHistDensity_Golden(t *testing.T) {
	runGoldenTest(t, "hist_density", renderHistDensity)
}

func TestHistStrategies_Golden(t *testing.T) {
	runGoldenTest(t, "hist_strategies", renderHistStrategies)
}

func TestBoxPlotBasic_Golden(t *testing.T) {
	runGoldenTest(t, "boxplot_basic", renderBoxPlotBasic)
}

func TestErrorBars_Golden(t *testing.T) {
	runGoldenTest(t, "errorbar_basic", renderErrorBars)
}

func TestTextLabelsStrict_Golden(t *testing.T) {
	runGoldenTest(t, "text_labels_strict", renderTextLabelsStrict)
}

func TestTitleStrict_Golden(t *testing.T) {
	runGoldenTest(t, "title_strict", renderTitleStrict)
}

func TestImageHeatmap_Golden(t *testing.T) {
	runGoldenTest(t, "image_heatmap", renderImageHeatmap)
}

// runGoldenTest is a helper function for golden image testing
func runGoldenTest(t *testing.T, testName string, renderFunc func() image.Image) {
	// Render the plot
	img := renderFunc()

	goldenPath := "../testdata/golden/" + testName + ".png"

	if *updateGolden {
		// Update the golden image
		err := imagecmp.SavePNG(img, goldenPath)
		if err != nil {
			t.Fatalf("Failed to update golden image: %v", err)
		}
		t.Skip("Updated golden image")
		return
	}

	// Load the expected golden image
	want, err := imagecmp.LoadPNG(goldenPath)
	if err != nil {
		t.Fatalf("Failed to load golden image %s: %v", goldenPath, err)
	}

	// Compare with tolerance
	diff, err := imagecmp.ComparePNG(img, want, 1) // ≤1 LSB tolerance
	if err != nil {
		t.Fatalf("Image comparison failed: %v", err)
	}

	// Check if images are within tolerance
	if !diff.Identical {
		// Save debug images
		artifactsDir := "../testdata/_artifacts"
		if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
			t.Fatalf("Could not create artifacts directory %s: %v", artifactsDir, err)
		} else {
			// Save the rendered image
			gotPath := filepath.Join(artifactsDir, testName+"_got.png")
			if err := imagecmp.SavePNG(img, gotPath); err != nil {
				t.Fatalf("Could not save got image %s: %v", gotPath, err)
			}

			// Save the diff image
			diffPath := filepath.Join(artifactsDir, testName+"_diff.png")
			if err := imagecmp.SaveDiffImage(img, want, 1, diffPath); err != nil {
				t.Fatalf("Could not save diff image %s: %v", diffPath, err)
			}

			t.Logf("Debug images saved to %s/", artifactsDir)
		}

		t.Fatalf("Golden image mismatch: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB",
			diff.MaxDiff, diff.MeanAbs, diff.PSNR)
	}

	t.Logf("Golden image match: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB",
		diff.MaxDiff, diff.MeanAbs, diff.PSNR)
}

// renderBasicLine creates the same basic line plot as examples/lines/basic.go
func renderBasicLine() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})
	ax.SetTitle("Basic Line")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)

	line := &core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1},
	}
	ax.Add(line)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderJoinsCaps creates a plot demonstrating different line joins and caps
func renderJoinsCaps() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Line Joins and Caps")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 6)

	joinPath := []geom.Pt{
		{X: 1, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: 3}, {X: 5, Y: 3},
	}
	miterLine := &core.Line2D{
		XY:  joinPath,
		W:   8.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
	}
	ax.Add(miterLine)

	capPath := []geom.Pt{
		{X: 7, Y: 5}, {X: 9, Y: 5},
	}
	capLine := &core.Line2D{
		XY:  capPath,
		W:   8.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
	}
	ax.Add(capLine)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderDashes creates a plot demonstrating dash patterns
func renderDashes() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Dash Patterns")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 5)

	lines := []struct {
		y      float64
		dashes []float64
		color  render.Color
	}{
		{4, []float64{}, render.Color{R: 0, G: 0, B: 0, A: 1}},
		{3, []float64{10, 4}, render.Color{R: 0.8, G: 0, B: 0, A: 1}},
		{2, []float64{6, 2, 2, 2}, render.Color{R: 0, G: 0.6, B: 0, A: 1}},
		{1, []float64{2, 2}, render.Color{R: 0, G: 0, B: 0.8, A: 1}},
	}

	for _, lineSpec := range lines {
		path := []geom.Pt{
			{X: 1, Y: lineSpec.y}, {X: 9, Y: lineSpec.y},
		}
		line := &core.Line2D{
			XY:     path,
			W:      3.0,
			Col:    lineSpec.color,
			Dashes: lineSpec.dashes,
		}
		ax.Add(line)
	}

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderScatterBasic creates a basic scatter plot for golden testing
func renderScatterBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Basic Scatter")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 10)

	basicPoints := []geom.Pt{
		{X: 2, Y: 3},
		{X: 4, Y: 6},
		{X: 6, Y: 4},
		{X: 8, Y: 7},
		{X: 3, Y: 8},
		{X: 7, Y: 2},
	}
	scatter := &core.Scatter2D{
		XY:     basicPoints,
		Size:   8.0,
		Color:  render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
		Marker: core.MarkerCircle,
		Alpha:  1.0,
	}
	ax.Add(scatter)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderScatterMarkerTypes creates a plot showing all marker types
func renderScatterMarkerTypes() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Scatter Marker Types")
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 8)

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
			lineWidth = 1.44 // lw(2) at DPI=100 used by the reference generator
		}
		scatter := &core.Scatter2D{
			XY:     []geom.Pt{{X: float64(1 + i), Y: 4}},
			Size:   12.0,
			Color:  colors[i],
			EdgeColor: colors[i],
			EdgeWidth: lineWidth,
			Marker: markerType,
			Alpha:  1.0,
		}
		ax.Add(scatter)
	}

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderScatterAdvanced creates an advanced scatter plot with edges, alpha, and variable sizes
func renderScatterAdvanced() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Advanced Scatter")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 10)

	points := []geom.Pt{
		{X: 2, Y: 2},
		{X: 4, Y: 4},
		{X: 6, Y: 6},
		{X: 8, Y: 8},
		{X: 2, Y: 8},
		{X: 4, Y: 6},
		{X: 6, Y: 4},
		{X: 8, Y: 2},
	}
	sizes := []float64{6, 10, 14, 18, 8, 12, 16, 20}
	fillColors := []render.Color{
		{R: 1, G: 0.5, B: 0.5, A: 1},
		{R: 0.5, G: 1, B: 0.5, A: 1},
		{R: 0.5, G: 0.5, B: 1, A: 1},
		{R: 1, G: 1, B: 0.5, A: 1},
		{R: 1, G: 0.5, B: 1, A: 1},
		{R: 0.5, G: 1, B: 1, A: 1},
		{R: 0.8, G: 0.8, B: 0.8, A: 1},
		{R: 0.3, G: 0.3, B: 0.3, A: 1},
	}
	edgeColors := []render.Color{
		{R: 0.5, G: 0, B: 0, A: 1},
		{R: 0, G: 0.5, B: 0, A: 1},
		{R: 0, G: 0, B: 0.5, A: 1},
		{R: 0.5, G: 0.5, B: 0, A: 1},
		{R: 0.5, G: 0, B: 0.5, A: 1},
		{R: 0, G: 0.5, B: 0.5, A: 1},
		{R: 0.4, G: 0.4, B: 0.4, A: 1},
		{R: 0, G: 0, B: 0, A: 1},
	}

	scatter := &core.Scatter2D{
		XY:         points,
		Sizes:      sizes,
		Colors:     fillColors,
		EdgeColors: edgeColors,
		EdgeWidth:  2.0,
		Alpha:      0.8,
		Marker:     core.MarkerCircle,
	}
	ax.Add(scatter)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderBarBasicScaffold(showTicks, showTickLabels, showTitle bool) image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XScale = transform.NewLinear(0, 6)
	ax.YScale = transform.NewLinear(0, 10)

	ax.XAxis.ShowTicks = showTicks
	ax.YAxis.ShowTicks = showTicks
	ax.XAxis.ShowLabels = showTickLabels
	ax.YAxis.ShowLabels = showTickLabels

	if showTitle {
		ax.SetTitle("Basic Bars")
	}

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderBarBasicFrame() image.Image {
	return renderBarBasicScaffold(false, false, false)
}

func renderBarBasicTicks() image.Image {
	return renderBarBasicScaffold(true, false, false)
}

func renderBarBasicTickLabels() image.Image {
	return renderBarBasicScaffold(true, true, false)
}

func renderBarBasicTitle() image.Image {
	return renderBarBasicScaffold(true, true, true)
}

// renderBarBasic creates a basic vertical bar chart for golden testing
func renderBarBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Basic Bars")
	ax.XScale = transform.NewLinear(0, 6)
	ax.YScale = transform.NewLinear(0, 10)

	bar := &core.Bar2D{
		X:           []float64{1, 2, 3, 4, 5},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.6,
		Color:       render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1},
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderBarHorizontal creates a horizontal bar chart for golden testing
func renderBarHorizontal() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Horizontal Bars")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 6)

	bar := &core.Bar2D{
		X:           []float64{1, 2, 3, 4, 5},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.6,
		Color:       render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1},
		Baseline:    0,
		Orientation: core.BarHorizontal,
	}
	ax.Add(bar)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderBarGrouped creates a grouped bar chart with variable colors and edges
func renderBarGrouped() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Grouped Bars")
	ax.XScale = transform.NewLinear(0, 7)
	ax.YScale = transform.NewLinear(0, 10)

	bar1 := &core.Bar2D{
		X:           []float64{1.2, 2.2, 3.2, 4.2, 5.2},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.35,
		Color:       render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
		EdgeColor:   render.Color{R: 0.5, G: 0, B: 0, A: 1},
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar1)

	bar2 := &core.Bar2D{
		X:           []float64{1.8, 2.8, 3.8, 4.8, 5.8},
		Heights:     []float64{5, 4, 6, 3, 7},
		Width:       0.35,
		Color:       render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1},
		EdgeColor:   render.Color{R: 0, G: 0.5, B: 0, A: 1},
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar2)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderFillBasic creates a basic fill to baseline for golden testing
func renderFillBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Fill to Baseline")
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1, 3)

	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	y := []float64{0.5, 1.8, 2.3, 1.2, 2.8, 1.9, 2.1, 1.5, 0.8}

	fill := &core.Fill2D{
		X:         x,
		Y1:        y,
		Baseline:  0,
		Color:     render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7},
		EdgeColor: render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1.0},
		EdgeWidth: 2.0,
		Alpha:     1.0,
	}
	ax.Add(fill)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderFillBetween creates a fill between two curves for golden testing
func renderFillBetween() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Fill Between Curves")
	ax.XScale = transform.NewLinear(0, 6.28)
	ax.YScale = transform.NewLinear(-1.5, 1.5)

	n := 50
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		t := 6.28 * float64(i) / float64(n-1)
		x[i] = t
		y1[i] = math.Sin(t)
		y2[i] = 0.8 * math.Cos(t)
	}

	fill := core.FillBetween(x, y1, y2, render.Color{R: 0.8, G: 0.3, B: 0.3, A: 0.6})
	fill.EdgeColor = render.Color{R: 0.5, G: 0.1, B: 0.1, A: 1.0}
	fill.EdgeWidth = 1.5
	ax.Add(fill)

	sineLine := &core.Line2D{XY: make([]geom.Pt, n), W: 2.0, Col: render.Color{R: 1, G: 0, B: 0, A: 1}}
	cosLine := &core.Line2D{XY: make([]geom.Pt, n), W: 2.0, Col: render.Color{R: 0, G: 0, B: 1, A: 1}}
	for i := 0; i < n; i++ {
		sineLine.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
		cosLine.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}
	ax.Add(sineLine)
	ax.Add(cosLine)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderFillStacked creates a stacked area chart for golden testing
func renderFillStacked() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Stacked Fills")
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 8)

	x := []float64{1, 2, 3, 4, 5, 6, 7}
	layer1 := []float64{1, 1.5, 2, 1.8, 2.2, 1.9, 1.6}
	layer2 := make([]float64, len(layer1))
	layer3 := make([]float64, len(layer1))
	for i := range layer1 {
		layer2[i] = layer1[i] + 1.5 + 0.3*math.Sin(float64(i))
		layer3[i] = layer2[i] + 1.2 + 0.4*math.Cos(float64(i))
	}

	fill1 := core.FillToBaseline(x, layer1, 0, render.Color{R: 0.8, G: 0.2, B: 0.2, A: 0.8})
	fill1.EdgeColor = render.Color{R: 0.5, G: 0, B: 0, A: 1}
	fill1.EdgeWidth = 1.0

	fill2 := core.FillBetween(x, layer1, layer2, render.Color{R: 0.2, G: 0.8, B: 0.2, A: 0.8})
	fill2.EdgeColor = render.Color{R: 0, G: 0.5, B: 0, A: 1}
	fill2.EdgeWidth = 1.0

	fill3 := core.FillBetween(x, layer2, layer3, render.Color{R: 0.2, G: 0.2, B: 0.8, A: 0.8})
	fill3.EdgeColor = render.Color{R: 0, G: 0, B: 0.5, A: 1}
	fill3.EdgeWidth = 1.0

	ax.Add(fill1)
	ax.Add(fill2)
	ax.Add(fill3)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderMultiSeriesBasic creates a plot with multiple series using different plot types
func renderMultiSeriesBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Multi-Series Plot")
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 6)

	x1 := []float64{1, 2, 3, 4, 5, 6}
	y1 := []float64{1.5, 2.8, 2.2, 3.5, 3.8, 4.2}
	x2 := []float64{1.5, 2.5, 3.5, 4.5, 5.5}
	y2 := []float64{2.2, 3.1, 2.9, 4.1, 4.5}
	x3 := []float64{2, 3, 4, 5}
	y3 := []float64{3.8, 2.5, 4.8, 3.2}

	ax.Plot(x1, y1, core.PlotOptions{Label: "Series 1"})
	ax.Scatter(x2, y2, core.ScatterOptions{Label: "Series 2"})
	width := 0.4
	ax.Bar(x3, y3, core.BarOptions{Label: "Series 3", Width: &width})

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// normalData generates a seeded normally-distributed sample using Box-Muller.
func normalData(seed1, seed2 uint64, n int, mean, stddev float64) []float64 {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	data := make([]float64, n)
	for i := range data {
		u1 := rng.Float64()
		u2 := rng.Float64()
		data[i] = math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*stddev + mean
	}
	return data
}

// renderHistBasic creates a basic count histogram for golden testing.
func renderHistBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Basic Histogram")

	data := normalData(42, 0, 500, 5.0, 1.5)

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	ax.Hist(data, core.HistOptions{
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.AutoScale(0.05)
	_, yMax := ax.YScale.Domain()
	ax.SetYLim(0, yMax)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderHistDensity creates a density-normalized histogram for golden testing.
func renderHistDensity() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Density Histogram")

	data := normalData(42, 0, 500, 5.0, 1.5)

	green := render.Color{R: 0.20, G: 0.65, B: 0.30, A: 0.8}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.8
	bins := 20
	ax.Hist(data, core.HistOptions{
		Bins:      bins,
		Norm:      core.HistNormDensity,
		Color:     &green,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.AutoScale(0.05)
	_, yMax := ax.YScale.Domain()
	ax.SetYLim(0, yMax)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderHistStrategies creates a histogram comparing two datasets for golden testing.
func renderHistStrategies() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	ax.SetTitle("Histogram Strategies")

	data1 := normalData(42, 0, 300, 4.0, 1.0)
	data2 := normalData(7, 0, 300, 7.0, 1.2)

	blue := render.Color{R: 0.26, G: 0.53, B: 0.80, A: 0.6}
	orange := render.Color{R: 0.90, G: 0.50, B: 0.10, A: 0.6}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	ew := 0.5
	bins := 15
	prob := core.HistNormProbability

	ax.Hist(data1, core.HistOptions{
		Bins:      bins,
		Norm:      prob,
		Color:     &blue,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.Hist(data2, core.HistOptions{
		Bins:      bins,
		Norm:      prob,
		Color:     &orange,
		EdgeColor: &black,
		EdgeWidth: &ew,
	})
	ax.AutoScale(0.05)
	_, yMax := ax.YScale.Domain()
	ax.SetYLim(0, yMax)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

// renderBoxPlotBasic creates a multi-series box plot for golden testing.
func renderBoxPlotBasic() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XScale = transform.NewLinear(0, 4)
	ax.YScale = transform.NewLinear(0, 10)
	ax.SetTitle("Box Plots")
	ax.SetXLabel("Group")
	ax.SetYLabel("Value")
	ax.AddYGrid()

	alpha := 0.75
	boxWidth := 0.55
	edgeWidth := 1.2
	whiskerWidth := 1.2
	medianWidth := 1.8
	capWidth := 0.35
	flierSize := 4.0

	boxSpecs := []struct {
		data     []float64
		position float64
		color    render.Color
	}{
		{
			data:     []float64{0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8},
			position: 1.0,
			color:    render.Color{R: 0.25, G: 0.55, B: 0.82, A: 1},
		},
		{
			data:     []float64{4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4},
			position: 2.0,
			color:    render.Color{R: 0.80, G: 0.45, B: 0.20, A: 1},
		},
		{
			data:     []float64{2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8},
			position: 3.0,
			color:    render.Color{R: 0.35, G: 0.70, B: 0.35, A: 1},
		},
	}

	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	for _, spec := range boxSpecs {
		pos := spec.position
		ax.BoxPlot(spec.data, core.BoxPlotOptions{
			Position:     &pos,
			Width:        &boxWidth,
			Color:        &spec.color,
			EdgeColor:    &black,
			MedianColor:  &black,
			WhiskerColor: &black,
			CapColor:     &black,
			FlierColor:   &black,
			EdgeWidth:    &edgeWidth,
			WhiskerWidth: &whiskerWidth,
			MedianWidth:  &medianWidth,
			CapWidth:     &capWidth,
			FlierSize:    &flierSize,
			Alpha:        &alpha,
			ShowFliers:   boolPtr(true),
		})
	}

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderTextLabelsStrict() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.SetTitle("Text Labels")
	ax.SetXLabel("Group")
	ax.SetYLabel("Value")

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderTitleStrict() image.Image {
	fig := core.NewFigure(320, 80)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.40},
		Max: geom.Pt{X: 0.95, Y: 0.85},
	})
	ax.SetTitle("Histogram Strategies")
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	r, err := agg.New(320, 80, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func boolPtr(v bool) *bool {
	return &v
}

// renderMultiSeriesColorCycle creates a plot demonstrating automatic color cycling
func renderMultiSeriesColorCycle() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Color Cycle")
	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(-1.2, 1.2)

	nPoints := 50
	x := make([]float64, nPoints)
	for i := 0; i < nPoints; i++ {
		x[i] = 2 * math.Pi * float64(i) / float64(nPoints-1)
	}

	for freq := 1; freq <= 4; freq++ {
		y := make([]float64, nPoints)
		for i := 0; i < nPoints; i++ {
			y[i] = math.Sin(float64(freq) * x[i])
		}
		label := fmt.Sprintf("f=%d", freq)
		ax.Plot(x, y, core.PlotOptions{Label: label})
	}

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderErrorBars() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Error Bars")
	ax.XScale = transform.NewLinear(0, 7)
	ax.YScale = transform.NewLinear(0, 6)

	x := []float64{1, 2, 3, 4, 5, 6}
	y := []float64{1.8, 2.5, 2.2, 3.1, 2.8, 3.7}
	xErr := []float64{0.20, 0.25, 0.15, 0.22, 0.30, 0.18}
	yErr := []float64{0.28, 0.20, 0.35, 0.24, 0.30, 0.22}

	lineColor := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	black := render.Color{R: 0, G: 0, B: 0, A: 1}
	pointColor := render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	lineWidth := 2.0
	errorWidth := 1.2
	pointSize := 4.5
	errorCap := 6.0

	ax.Plot(x, y, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "Trend",
	})
	ax.Scatter(x, y, core.ScatterOptions{
		Color: &pointColor,
		Size:  &pointSize,
		Label: "Samples",
	})
	ax.ErrorBar(x, y, xErr, yErr, core.ErrorBarOptions{
		Color:     &black,
		LineWidth: &errorWidth,
		CapSize:   &errorCap,
		Label:     "1σ Uncertainty",
	})

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderImageHeatmap() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})
	ax.SetTitle("Image Heatmap")
	ax.XScale = transform.NewLinear(0, 3)
	ax.YScale = transform.NewLinear(0, 3)

	data := [][]float64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
	}

	ax.Image(data)

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}
