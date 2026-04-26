package test

import (
	"flag"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"matplotlib-go/backends/agg"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/test/imagecmp"
	"matplotlib-go/transform"
)

var updateGolden = flag.Bool("update-golden", false, "Update golden images instead of comparing")

type testDistanceKM float64

type testDistanceConverter struct{}

var registerTestDistanceUnitsOnce sync.Once

func (testDistanceConverter) Convert(value any) (float64, error) {
	v, ok := value.(testDistanceKM)
	if !ok {
		return 0, fmt.Errorf("unexpected distance value %T", value)
	}
	return float64(v), nil
}

func (testDistanceConverter) AxisInfo([]float64) core.AxisInfo {
	return core.AxisInfo{
		Formatter: core.FormatStrFormatter{Pattern: "%.0f km"},
	}
}

func registerTestDistanceUnits() {
	registerTestDistanceUnitsOnce.Do(func() {
		core.MustRegisterUnitConverter(testDistanceKM(0), func() core.UnitsConverter {
			return testDistanceConverter{}
		})
	})
}

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
	requireOptionalVisualTests(t)
	runGoldenTest(t, "boxplot_basic", renderBoxPlotBasic)
}

func TestErrorBars_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "errorbar_basic", renderErrorBars)
}

func TestTextLabelsStrict_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "text_labels_strict", renderTextLabelsStrict)
}

func TestTitleStrict_Golden(t *testing.T) {
	runGoldenTest(t, "title_strict", renderTitleStrict)
}

func TestImageHeatmap_Golden(t *testing.T) {
	runGoldenTest(t, "image_heatmap", renderImageHeatmap)
}

func TestAxesTopRightInverted_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_top_right_inverted", renderAxesTopRightInverted)
}

func TestAxesControlSurface_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_control_surface", renderAxesControlSurface)
}

func TestTransformCoordinates_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "transform_coordinates", renderTransformCoordinates)
}

func TestGridSpecComposition_Golden(t *testing.T) {
	runGoldenTest(t, "gridspec_composition", renderGridSpecComposition)
}

func TestFigureLabelsComposition_Golden(t *testing.T) {
	runGoldenTest(t, "figure_labels_composition", renderFigureLabelsComposition)
}

func TestColorbarComposition_Golden(t *testing.T) {
	runGoldenTest(t, "colorbar_composition", renderColorbarComposition)
}

func TestAnnotationComposition_Golden(t *testing.T) {
	runGoldenTest(t, "annotation_composition", renderAnnotationComposition)
}

func TestPlotVariants_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "plot_variants", renderPlotVariants)
}

func TestStatVariants_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "stat_variants", renderStatVariants)
}

func TestUnitsOverview_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_overview", renderUnitsOverview)
}

func TestUnitsDates_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_dates", renderUnitsDates)
}

func TestUnitsCategories_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_categories", renderUnitsCategories)
}

func TestUnitsCustomConverter_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_custom_converter", renderUnitsCustomConverter)
}

func TestPatchShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "patch_showcase", renderPatchShowcase)
}

func TestMeshContourTri_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mesh_contour_tri", renderMeshContourTri)
}

func TestStemPlot_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "stem_plot", renderStemPlot)
}

func TestSpecialtyArtists_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "specialty_artists", renderSpecialtyArtists)
}

func TestVectorFields_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "vector_fields", renderVectorFields)
}

func TestPolarAxes_Golden(t *testing.T) {
	runGoldenTest(t, "polar_axes", renderPolarAxes)
}

func TestGeoMollweideAxes_Golden(t *testing.T) {
	runGoldenTest(t, "geo_mollweide_axes", renderGeoMollweideAxes)
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
			XY:        []geom.Pt{{X: float64(1 + i), Y: 4}},
			Size:      12.0,
			Color:     colors[i],
			EdgeColor: colors[i],
			EdgeWidth: lineWidth,
			Marker:    markerType,
			Alpha:     1.0,
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
	const (
		w = 320
		h = 280
	)

	fig := core.NewFigure(w, h)
	titles := []string{
		"Histogram Strategies",
		"Fill to Baseline",
		"Dash Patterns",
		"Box Plots",
		"Text Labels",
	}
	rows := []geom.Rect{
		{Min: geom.Pt{X: 0.05, Y: 0.20}, Max: geom.Pt{X: 0.95, Y: 0.28}},
		{Min: geom.Pt{X: 0.05, Y: 0.36}, Max: geom.Pt{X: 0.95, Y: 0.44}},
		{Min: geom.Pt{X: 0.05, Y: 0.52}, Max: geom.Pt{X: 0.95, Y: 0.60}},
		{Min: geom.Pt{X: 0.05, Y: 0.68}, Max: geom.Pt{X: 0.95, Y: 0.76}},
		{Min: geom.Pt{X: 0.05, Y: 0.84}, Max: geom.Pt{X: 0.95, Y: 0.92}},
	}
	for i, title := range titles {
		ax := fig.AddAxes(rows[i])
		ax.SetTitle(title)
		ax.SetXLim(0, 1)
		ax.SetYLim(0, 1)
		ax.XAxis.ShowSpine = false
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
		ax.YAxis.ShowSpine = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		ax.ShowFrame = false
	}

	r, err := agg.New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
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

func renderAxesTopRightInverted() image.Image {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.12},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetTitle("Top/Right Axes + Inversion")
	ax.SetXLabel("Bottom X")
	ax.SetYLabel("Left Y")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	top := ax.TopAxis()
	top.MinorLocator = nil
	right := ax.RightAxis()
	right.MinorLocator = nil

	line := &core.Line2D{
		XY: []geom.Pt{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
			{X: 6, Y: 6.5},
			{X: 8.5, Y: 8},
		},
		W:   2.2,
		Col: render.Color{R: 0.15, G: 0.35, B: 0.75, A: 1},
	}
	ax.Add(line)

	scatter := &core.Scatter2D{
		XY: []geom.Pt{
			{X: 2, Y: 8},
			{X: 5, Y: 5},
			{X: 8, Y: 2},
		},
		Size:      9,
		Color:     render.Color{R: 0.85, G: 0.35, B: 0.20, A: 0.9},
		EdgeColor: render.Color{R: 0.45, G: 0.15, B: 0.05, A: 1},
		EdgeWidth: 1.0,
		Marker:    core.MarkerDiamond,
		Alpha:     1.0,
	}
	ax.Add(scatter)

	ax.InvertX()
	ax.InvertY()

	r, err := agg.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderAxesControlSurface() image.Image {
	fig := core.NewFigure(760, 360)

	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.07, Y: 0.14},
		Max: geom.Pt{X: 0.47, Y: 0.90},
	})
	left.SetTitle("Moved Axes + Aspect")
	left.SetXLabel("Top X")
	left.SetYLabel("Right Y")
	left.SetXLim(-1, 5)
	left.SetYLim(-1, 5)
	if err := left.SetXLabelPosition("top"); err != nil {
		panic(err)
	}
	if err := left.SetYLabelPosition("right"); err != nil {
		panic(err)
	}
	top := left.TopAxis()
	top.ShowLabels = true
	top.ShowTicks = true
	rightAxis := left.RightAxis()
	rightAxis.ShowLabels = true
	rightAxis.ShowTicks = true
	left.XAxis.ShowLabels = false
	left.XAxis.ShowTicks = false
	left.YAxis.ShowLabels = false
	left.YAxis.ShowTicks = false
	left.SetAxisEqual()
	if err := left.SetBoxAspect(1); err != nil {
		panic(err)
	}
	if err := left.MinorticksOn("both"); err != nil {
		panic(err)
	}
	if err := left.LocatorParams(core.LocatorParams{
		Axis:       "both",
		MajorCount: 6,
		MinorCount: 24,
	}); err != nil {
		panic(err)
	}
	majorLen := 7.0
	minorLen := 4.0
	majorWidth := 1.2
	minorWidth := 0.9
	tickColor := render.Color{R: 0.18, G: 0.42, B: 0.55, A: 1}
	if err := left.TickParams(core.TickParams{
		Axis:   "both",
		Which:  "major",
		Color:  &tickColor,
		Length: &majorLen,
		Width:  &majorWidth,
	}); err != nil {
		panic(err)
	}
	if err := left.TickParams(core.TickParams{
		Axis:   "both",
		Which:  "minor",
		Color:  &tickColor,
		Length: &minorLen,
		Width:  &minorWidth,
	}); err != nil {
		panic(err)
	}
	left.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: -0.5, Y: -0.2},
			{X: 0.8, Y: 1.0},
			{X: 2.2, Y: 2.1},
			{X: 4.2, Y: 4.4},
		},
		W:   2.0,
		Col: render.Color{R: 0.10, G: 0.32, B: 0.76, A: 1},
	})
	left.Add(&core.Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1.5, Y: 1.8},
			{X: 3.5, Y: 3.2},
			{X: 4.5, Y: 4.6},
		},
		Size:      8,
		Color:     render.Color{R: 0.92, G: 0.48, B: 0.20, A: 0.92},
		EdgeColor: render.Color{R: 0.52, G: 0.22, B: 0.08, A: 1},
		EdgeWidth: 1.0,
		Marker:    core.MarkerCircle,
		Alpha:     1,
	})

	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.58, Y: 0.14},
		Max: geom.Pt{X: 0.95, Y: 0.90},
	})
	right.SetTitle("Twin + Secondary")
	right.SetXLim(0, 10)
	right.SetYLim(0, 20)
	right.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 2},
			{X: 2, Y: 6},
			{X: 4, Y: 9},
			{X: 6, Y: 13},
			{X: 8, Y: 16},
			{X: 10, Y: 19},
		},
		W:   2.0,
		Col: render.Color{R: 0.12, G: 0.45, B: 0.72, A: 1},
	})

	twin := right.TwinX()
	twin.SetYLim(0, 100)
	twinLineColor := render.Color{R: 0.80, G: 0.22, B: 0.22, A: 1}
	if axis := twin.RightAxis(); axis != nil {
		axis.Color = twinLineColor
		axis.MinorLocator = nil
	}
	twin.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 10},
			{X: 2, Y: 22},
			{X: 4, Y: 38},
			{X: 6, Y: 58},
			{X: 8, Y: 81},
			{X: 10, Y: 96},
		},
		W:   1.8,
		Col: twinLineColor,
	})

	sec, err := right.SecondaryXAxis(core.AxisTop,
		func(x float64) float64 { return x * 10 },
		func(x float64) (float64, bool) { return x / 10, true },
	)
	if err != nil {
		panic(err)
	}
	if axis := sec.TopAxis(); axis != nil {
		axis.Color = render.Color{R: 0.16, G: 0.42, B: 0.30, A: 1}
		axis.MinorLocator = nil
	}

	r, err := agg.New(760, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderTransformCoordinates() image.Image {
	fig := core.NewFigure(720, 420)

	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.13, Y: 0.16},
		Max: geom.Pt{X: 0.90, Y: 0.84},
	})
	ax.SetTitle("Transform Coordinates")
	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	ax.AddXGrid()
	ax.AddYGrid()

	lineColor := render.Color{R: 0.14, G: 0.37, B: 0.74, A: 1}
	pointColor := render.Color{R: 0.88, G: 0.42, B: 0.16, A: 0.92}
	textColor := render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1}

	ax.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 1.0, Y: 1.5},
			{X: 2.5, Y: 3.2},
			{X: 4.5, Y: 5.6},
			{X: 7.0, Y: 6.4},
			{X: 8.8, Y: 8.2},
		},
		W:   2.2,
		Col: lineColor,
	})
	ax.Add(&core.Scatter2D{
		XY: []geom.Pt{
			{X: 2.5, Y: 3.2},
			{X: 7.0, Y: 6.4},
			{X: 8.8, Y: 8.2},
		},
		Size:      8,
		Color:     pointColor,
		EdgeColor: render.Color{R: 0.45, G: 0.18, B: 0.05, A: 1},
		EdgeWidth: 1.0,
		Marker:    core.MarkerDiamond,
		Alpha:     1.0,
	})

	ax.Text(1.3, 1.1, "data", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		Coords:   core.Coords(core.CoordData),
	})
	ax.Text(0.03, 0.97, "axes", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		Coords:   core.Coords(core.CoordAxes),
	})
	ax.Text(0.07, 0.08, "figure", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignBottom,
		Coords:   core.Coords(core.CoordFigure),
	})
	ax.Text(0.50, 0.22, "blend", core.TextOptions{
		FontSize: 11,
		Color:    textColor,
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignBottom,
		Coords:   core.BlendCoords(core.CoordFigure, core.CoordAxes),
		OffsetY:  6,
	})
	ax.Annotate("axes note", 0.82, 0.78, core.AnnotationOptions{
		Coords:   core.Coords(core.CoordAxes),
		OffsetX:  -48,
		OffsetY:  -26,
		FontSize: 10,
		Color:    textColor,
	})

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderGridSpecComposition() image.Image {
	fig := core.NewFigure(960, 640)
	applyMatplotlibGridSpecStyle(fig)

	outer := fig.GridSpec(
		2,
		2,
		core.WithGridSpecPadding(0.08, 0.96, 0.10, 0.92),
		core.WithGridSpecSpacing(0.06/(2+0.06), 0.28/(2+0.28)),
		core.WithGridSpecWidthRatios(2, 1),
	)

	mainAx := outer.Span(0, 0, 2, 1).AddAxes()
	configureCompositionAxes(mainAx, "Main Span", []float64{0, 1, 2, 3, 4}, []float64{1.2, 2.8, 2.1, 3.6, 3.1}, render.Color{R: 0.15, G: 0.35, B: 0.72, A: 1})
	configureCompositionTicks(mainAx, []float64{0, 1, 2, 3, 4}, []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5}, "%.1f")

	nested := outer.Cell(0, 1).GridSpec(2, 1, core.WithGridSpecSpacing(0, 0.75/(2+0.75)))
	topRight := nested.Cell(0, 0).AddAxes()
	configureCompositionAxes(topRight, "Nested Top", []float64{0, 1, 2, 3}, []float64{3.4, 2.6, 2.9, 1.8}, render.Color{R: 0.72, G: 0.32, B: 0.18, A: 1})
	configureCompositionTicks(topRight, []float64{0, 1, 2, 3}, []float64{2, 3}, "%.0f")

	bottomRight := nested.Cell(1, 0).AddAxes(core.WithSharedX(topRight))
	configureCompositionAxes(bottomRight, "Nested Bottom", []float64{0, 1, 2, 3}, []float64{1.0, 1.6, 1.3, 2.2}, render.Color{R: 0.18, G: 0.55, B: 0.34, A: 1})
	configureCompositionTicks(bottomRight, []float64{0, 1, 2, 3}, []float64{1, 2}, "%.0f")

	sub := outer.Cell(1, 1).SubFigure()
	inset := sub.AddSubplot(1, 1, 1)
	configureCompositionAxes(inset, "SubFigure", []float64{0, 1, 2, 3}, []float64{2.0, 2.4, 1.9, 2.7}, render.Color{R: 0.55, G: 0.22, B: 0.50, A: 1})
	configureCompositionTicks(inset, []float64{0, 1, 2, 3}, []float64{2.0, 2.2, 2.4, 2.6}, "%.1f")

	r, err := agg.New(960, 640, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func applyMatplotlibGridSpecStyle(fig *core.Figure) {
	fig.RC = style.Apply(fig.RC, style.WithFont("DejaVu Sans", 10))
	fig.RC.TitleFontSize = 12 * fig.RC.DPI / 72
	fig.RC.AxisLabelFontSize = 10 * fig.RC.DPI / 72
	fig.RC.XTickLabelFontSize = 10 * fig.RC.DPI / 72
	fig.RC.YTickLabelFontSize = 10 * fig.RC.DPI / 72
}

func configureCompositionAxes(ax *core.Axes, title string, x, y []float64, c render.Color) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{
		Color:     &c,
		LineWidth: &width,
		Label:     title,
	})
	ax.AutoScale(0.10)
}

func configureCompositionTicks(ax *core.Axes, xTicks, yTicks []float64, yFormat string) {
	ax.XAxis.Locator = core.FixedLocator{TicksList: xTicks}
	ax.YAxis.Locator = core.FixedLocator{TicksList: yTicks}
	ax.YAxis.Formatter = core.FormatStrFormatter{Pattern: yFormat}
}

func renderFigureLabelsComposition() image.Image {
	fig := core.NewFigure(1100, 720)
	fig.SetSuptitle("Shared-Axis Figure Labels")
	fig.SetSupXLabel("time [s]")
	fig.SetSupYLabel("amplitude")

	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.083, 0.996, 0.0986, 0.9333),
		core.WithSubplotSpacing(0.067, 0.100),
	)
	for row := range grid {
		for col, ax := range grid[row] {
			x := make([]float64, 180)
			y := make([]float64, 180)
			for i := range x {
				xv := 2 * math.Pi * float64(i) / float64(len(x)-1)
				x[i] = xv
				y[i] = math.Sin(xv+float64(row)*0.5) * (1 + float64(col)*0.2)
			}
			ax.Plot(x, y, core.PlotOptions{Label: fmt.Sprintf("series %d", row*2+col+1)})
			ax.SetTitle(fmt.Sprintf("Panel %d", row*2+col+1))
			ax.SetXLabel("local x")
			ax.SetYLabel("local y")
			ax.SetXLim(0, 2*math.Pi)
			ax.SetYLim(-1.6, 1.6)
			ax.AddXGrid()
			ax.AddYGrid()
		}
	}

	grid[0][0].AddAnchoredText("upper-left\nnote", core.AnchoredTextOptions{
		Locator: core.RelativeAnchoredBoxLocator{
			X: 0.02, Y: 0.08,
			HAlign: core.BoxAlignLeft,
			VAlign: core.BoxAlignTop,
		},
	})
	grid[1][1].AddAnchoredText("lower-right", core.AnchoredTextOptions{
		Locator: core.RelativeAnchoredBoxLocator{
			X: 0.98, Y: 0.92,
			HAlign: core.BoxAlignRight,
			VAlign: core.BoxAlignBottom,
		},
	})
	fig.AddAnchoredText("Figure note", core.AnchoredTextOptions{
		Locator: core.RelativeAnchoredBoxLocator{
			X: 0.985, Y: 0.06,
			HAlign: core.BoxAlignRight,
			VAlign: core.BoxAlignTop,
		},
	})
	legend := fig.AddLegend()
	legend.SetLocator(core.RelativeAnchoredBoxLocator{
		X: 0.99, Y: 0.10,
		HAlign: core.BoxAlignRight,
		VAlign: core.BoxAlignTop,
	})

	r, err := agg.New(1100, 720, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderColorbarComposition() image.Image {
	fig := core.NewFigure(1000, 700)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.052, Y: 0.066},
		Max: geom.Pt{X: 0.846, Y: 0.963},
	})

	const (
		rows = 80
		cols = 120
	)
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			x := (float64(col)/float64(cols-1))*4 - 2
			y := (float64(row)/float64(rows-1))*4 - 2
			radius := math.Hypot(x, y)
			data[row][col] = math.Sin(3*radius) * math.Exp(-0.6*radius)
		}
	}

	cmap := "inferno"
	img := ax.Image(data, core.ImageOptions{Colormap: &cmap})
	ax.SetTitle("Heatmap with Colorbar")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, cols)
	ax.SetYLim(0, rows)
	ax.AddXGrid()
	ax.AddYGrid()
	fig.AddColorbar(ax, img, core.ColorbarOptions{
		Width:   0.030,
		Padding: 0.054,
		Label:   "Intensity",
	})

	r, err := agg.New(1000, 700, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderAnnotationComposition() image.Image {
	fig := core.NewFigure(1040, 720)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.90, Y: 0.88},
	})
	ax.SetTitle("Text and Arrow Annotations")
	ax.SetXLabel("phase")
	ax.SetYLabel("response")
	ax.AddXGrid()
	ax.AddYGrid()

	x := make([]float64, 240)
	y := make([]float64, 240)
	for i := range x {
		xv := 6 * math.Pi * float64(i) / float64(len(x)-1)
		x[i] = xv
		y[i] = math.Sin(xv)*math.Exp(-0.015*xv) + 0.2*math.Cos(0.5*xv)
	}
	ax.Plot(x, y, core.PlotOptions{Label: "signal"})
	ax.SetXLim(0, 6*math.Pi)
	ax.SetYLim(-1.2, 1.2)
	ax.AddLegend()

	peakX := math.Pi / 2
	peakY := math.Sin(peakX)*math.Exp(-0.015*peakX) + 0.2*math.Cos(0.5*peakX)
	ax.Annotate("Peak\n= 0.42", peakX, peakY, core.AnnotationOptions{
		OffsetX:  48,
		OffsetY:  -42,
		FontSize: 12,
	})
	ax.Text(0.20, 0.90, "m∫T  φ x =  λ/4", core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		FontSize: 12,
	})

	r, err := agg.New(1040, 720, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderPlotVariants() image.Image {
	fig := core.NewFigure(840, 620)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.08, 0.97, 0.10, 0.93),
		core.WithSubplotSpacing(0.10, 0.14),
	)

	stepAx := grid[0][0]
	stepAx.SetTitle("Step + Stairs")
	stepAx.SetXLim(0, 6)
	stepAx.SetYLim(0, 5.2)
	stepAx.AddYGrid()
	stepWhere := core.StepWherePost
	stepAx.Step(
		[]float64{0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4},
		[]float64{1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6},
		core.StepOptions{
			Where:     &stepWhere,
			Color:     &render.Color{R: 0.15, G: 0.39, B: 0.78, A: 1},
			LineWidth: floatPtr(2.0),
		},
	)
	fillTrue := true
	stairsBaseline := 0.35
	stepAx.Stairs(
		[]float64{0.9, 1.7, 1.4, 2.6, 1.8, 2.2},
		[]float64{0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5},
		core.StairsOptions{
			Fill:      &fillTrue,
			Baseline:  &stairsBaseline,
			Color:     &render.Color{R: 0.91, G: 0.49, B: 0.20, A: 0.72},
			EdgeColor: &render.Color{R: 0.58, G: 0.26, B: 0.08, A: 1},
			LineWidth: floatPtr(1.5),
		},
	)

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
	fillAx.AxVLine(5.3, core.VLineOptions{
		Color:     &render.Color{R: 0.18, G: 0.22, B: 0.55, A: 1},
		LineWidth: floatPtr(1.2),
		Dashes:    []float64{2, 2},
	})
	fillAx.AxLine(
		geom.Pt{X: 0.9, Y: 0.3},
		geom.Pt{X: 6.4, Y: 5.6},
		core.ReferenceLineOptions{
			Color:     &render.Color{R: 0.22, G: 0.22, B: 0.22, A: 1},
			LineWidth: floatPtr(1.1),
		},
	)

	brokenAx := grid[1][0]
	brokenAx.SetTitle("broken_barh")
	brokenAx.SetXLim(0, 10)
	brokenAx.SetYLim(0, 4.4)
	brokenAx.AddXGrid()
	firstTrack := brokenAx.BrokenBarH(
		[][2]float64{{0.8, 1.6}, {3.1, 2.2}, {6.5, 1.3}},
		[2]float64{0.7, 0.9},
		core.BarOptions{Color: &render.Color{R: 0.21, G: 0.51, B: 0.76, A: 1}},
	)
	secondTrack := brokenAx.BrokenBarH(
		[][2]float64{{1.6, 1.0}, {4.0, 1.4}, {7.1, 1.7}},
		[2]float64{2.1, 0.9},
		core.BarOptions{Color: &render.Color{R: 0.86, G: 0.38, B: 0.16, A: 1}},
	)
	brokenAx.BarLabel(firstTrack, []string{"prep", "run", "cool"}, core.BarLabelOptions{
		Position: "center",
		Color:    render.Color{R: 1, G: 1, B: 1, A: 1},
		FontSize: 10,
	})
	brokenAx.BarLabel(secondTrack, []string{"IO", "fit", "ship"}, core.BarLabelOptions{
		Position: "center",
		Color:    render.Color{R: 1, G: 1, B: 1, A: 1},
		FontSize: 10,
	})

	stackAx := grid[1][1]
	stackAx.SetTitle("Stacked Bars + Labels")
	stackAx.SetXLim(0.4, 4.6)
	stackAx.SetYLim(0, 7.6)
	stackAx.AddYGrid()
	x := []float64{1, 2, 3, 4}
	base := []float64{0, 0, 0, 0}
	seriesA := []float64{1.4, 2.2, 1.8, 2.5}
	seriesB := []float64{2.1, 1.6, 2.4, 1.7}
	bottom := stackAx.Bar(x, seriesA, core.BarOptions{
		Baselines: base,
		Color:     &render.Color{R: 0.16, G: 0.59, B: 0.49, A: 1},
	})
	top := stackAx.Bar(x, seriesB, core.BarOptions{
		Baselines: seriesA,
		Color:     &render.Color{R: 0.88, G: 0.47, B: 0.16, A: 1},
	})
	stackAx.BarLabel(bottom, []string{"A1", "A2", "A3", "A4"}, core.BarLabelOptions{
		Position: "center",
		Color:    render.Color{R: 1, G: 1, B: 1, A: 1},
		FontSize: 10,
	})
	stackAx.BarLabel(top, nil, core.BarLabelOptions{
		Format: "%.1f",
		Color:  render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1},
	})

	r, err := agg.New(840, 620, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderStatVariants() image.Image {
	fig := core.NewFigure(840, 620)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.08, 0.97, 0.10, 0.93),
		core.WithSubplotSpacing(0.10, 0.14),
	)

	stackAx := grid[0][0]
	stackAx.SetTitle("StackPlot")
	stackAx.SetXLim(0, 5)
	stackAx.SetYLim(0, 7)
	stackAx.AddYGrid()
	stackAx.StackPlot(
		[]float64{0, 1, 2, 3, 4, 5},
		[][]float64{
			{1.0, 1.4, 1.3, 1.8, 1.6, 2.0},
			{0.8, 1.1, 1.4, 1.2, 1.6, 1.8},
			{0.5, 0.8, 1.0, 1.4, 1.1, 1.5},
		},
		core.StackPlotOptions{
			Colors: []render.Color{
				{R: 0.20, G: 0.55, B: 0.75, A: 1},
				{R: 0.90, G: 0.48, B: 0.18, A: 1},
				{R: 0.35, G: 0.66, B: 0.42, A: 1},
			},
			Alpha: floatPtr(0.76),
		},
	)

	ecdfAx := grid[0][1]
	ecdfAx.SetTitle("ECDF")
	ecdfAx.SetXLim(0, 8)
	ecdfAx.SetYLim(0, 1.05)
	ecdfAx.AddYGrid()
	ecdfAx.ECDF(
		[]float64{1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0},
		core.ECDFOptions{
			Color:     &render.Color{R: 0.18, G: 0.36, B: 0.75, A: 1},
			LineWidth: floatPtr(2),
			Compress:  true,
		},
	)

	cumulativeAx := grid[1][0]
	cumulativeAx.SetTitle("Cumulative Step Hist")
	cumulativeAx.SetXLim(0, 6)
	cumulativeAx.SetYLim(0, 1.05)
	cumulativeAx.AddYGrid()
	cumulativeAx.Hist(
		[]float64{0.4, 0.7, 1.2, 1.4, 2.1, 2.6, 3.1, 3.2, 4.0, 4.8, 5.2},
		core.HistOptions{
			BinEdges:   []float64{0, 1, 2, 3, 4, 5, 6},
			Norm:       core.HistNormProbability,
			Cumulative: true,
			HistType:   core.HistTypeStepFilled,
			Color:      &render.Color{R: 0.42, G: 0.62, B: 0.90, A: 0.55},
			EdgeColor:  &render.Color{R: 0.12, G: 0.25, B: 0.55, A: 1},
			EdgeWidth:  floatPtr(1.4),
		},
	)

	multiAx := grid[1][1]
	multiAx.SetTitle("Stacked Multi-Hist")
	multiAx.SetXLim(0, 6)
	multiAx.SetYLim(0, 6)
	multiAx.AddYGrid()
	multiAx.HistMulti(
		[][]float64{
			{0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2},
			{0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0},
			{1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4},
		},
		core.MultiHistOptions{
			BinEdges: []float64{0, 1, 2, 3, 4, 5, 6},
			Stacked:  true,
			Colors: []render.Color{
				{R: 0.22, G: 0.55, B: 0.70, A: 0.8},
				{R: 0.86, G: 0.42, B: 0.19, A: 0.8},
				{R: 0.36, G: 0.62, B: 0.36, A: 0.8},
			},
			EdgeColor: &render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1},
			EdgeWidth: floatPtr(0.7),
		},
	)

	r, err := agg.New(840, 620, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderUnitsOverview() image.Image {
	registerTestDistanceUnits()

	fig := core.NewFigure(1200, 420)

	dateAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.06, Y: 0.18},
		Max: geom.Pt{X: 0.32, Y: 0.86},
	})
	dateAx.SetTitle("Dates")
	dateAx.SetYLabel("Requests")
	dateAx.AddYGrid()
	_, err := dateAx.PlotUnits(
		[]time.Time{
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
		[]float64{12, 18, 9, 21},
		core.PlotOptions{
			Color:     &render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1},
			LineWidth: floatPtr(2.0),
		},
	)
	if err != nil {
		panic(err)
	}
	dateAx.AutoScale(0.05)

	categoryAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.38, Y: 0.18},
		Max: geom.Pt{X: 0.64, Y: 0.86},
	})
	categoryAx.SetTitle("Categories")
	categoryAx.SetYLabel("Count")
	categoryAx.AddYGrid()
	_, err = categoryAx.BarUnits(
		[]string{"alpha", "beta", "gamma", "delta"},
		[]float64{4, 9, 6, 7},
		core.BarOptions{
			Color:     &render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1},
			EdgeColor: &render.Color{R: 0.60, G: 0.30, B: 0.03, A: 1},
			EdgeWidth: floatPtr(1.0),
		},
	)
	if err != nil {
		panic(err)
	}
	categoryAx.AutoScale(0.10)

	unitAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.70, Y: 0.18},
		Max: geom.Pt{X: 0.96, Y: 0.86},
	})
	unitAx.SetTitle("Custom Units")
	unitAx.SetXLabel("Distance")
	unitAx.SetYLabel("Pace")
	unitAx.AddXGrid()
	unitAx.AddYGrid()
	_, err = unitAx.ScatterUnits(
		[]testDistanceKM{5, 10, 21.1, 42.2},
		[]float64{6.4, 5.8, 5.2, 5.5},
		core.ScatterOptions{
			Color:     &render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.92},
			EdgeColor: &render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1},
			EdgeWidth: floatPtr(1.0),
			Size:      floatPtr(8.0),
		},
	)
	if err != nil {
		panic(err)
	}
	unitAx.AutoScale(0.08)

	r, err := agg.New(1200, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderUnitsDates() image.Image {
	fig := core.NewFigure(720, 380)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Date Units")
	ax.SetXLabel("Date")
	ax.SetYLabel("Requests")
	ax.AddYGrid()

	dates := []time.Time{
		time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 20, 0, 0, 0, 0, time.UTC),
	}
	lower := []float64{6, 7, 5, 8, 7}
	upper := []float64{10, 15, 13, 18, 16}
	_, err := ax.FillBetweenUnits(dates, lower, upper, core.FillOptions{
		Color: &render.Color{R: 0.75, G: 0.86, B: 0.93, A: 1},
	})
	if err != nil {
		panic(err)
	}
	_, err = ax.PlotUnits(dates, []float64{8, 12, 9, 15, 13}, core.PlotOptions{
		Color:     &render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1},
		LineWidth: floatPtr(2.0),
	})
	if err != nil {
		panic(err)
	}
	ax.AutoScale(0.06)

	r, err := agg.New(720, 380, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderUnitsCategories() image.Image {
	fig := core.NewFigure(760, 360)
	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.20}, Max: geom.Pt{X: 0.47, Y: 0.86}})
	left.SetTitle("Categorical X")
	left.SetYLabel("Count")
	left.YAxis.Locator = core.MultipleLocator{Base: 2}
	left.AddYGrid()
	_, err := left.BarUnits(
		[]string{"draft", "review", "ship", "watch"},
		[]float64{3, 8, 6, 4},
		core.BarOptions{
			Color:     &render.Color{R: 1.0, G: 0.50, B: 0.05, A: 1},
			EdgeColor: &render.Color{R: 0.60, G: 0.30, B: 0.03, A: 1},
			EdgeWidth: floatPtr(1.0),
		},
	)
	if err != nil {
		panic(err)
	}
	left.AutoScale(0.10)
	left.SetYLim(0, 8.8)

	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.58, Y: 0.20}, Max: geom.Pt{X: 0.94, Y: 0.86}})
	right.SetTitle("Categorical Y")
	right.SetXLabel("Hours")
	right.XAxis.Locator = core.MultipleLocator{Base: 2}
	right.AddXGrid()
	orientation := core.BarHorizontal
	_, err = right.BarUnits(
		[]string{"north", "south", "east"},
		[]float64{4, 7, 5},
		core.BarOptions{
			Orientation: &orientation,
			Color:       &render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1},
			EdgeColor:   &render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1},
			EdgeWidth:   floatPtr(1.0),
		},
	)
	if err != nil {
		panic(err)
	}
	right.AutoScale(0.10)
	right.SetXLim(0, 7.7)

	r, err := agg.New(760, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderUnitsCustomConverter() image.Image {
	registerTestDistanceUnits()

	fig := core.NewFigure(680, 380)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.18}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Custom Distance Units")
	ax.SetXLabel("Distance")
	ax.SetYLabel("Pace")
	ax.AddXGrid()
	ax.AddYGrid()

	distances := []testDistanceKM{5, 10, 21.1, 30, 42.2}
	pace := []float64{6.4, 5.9, 5.3, 5.1, 5.4}
	_, err := ax.PlotUnits(distances, pace, core.PlotOptions{
		Color:     &render.Color{R: 0.55, G: 0.34, B: 0.29, A: 1},
		LineWidth: floatPtr(1.4),
	})
	if err != nil {
		panic(err)
	}
	_, err = ax.ScatterUnits(distances, pace, core.ScatterOptions{
		Color:     &render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.92},
		EdgeColor: &render.Color{R: 0.09, G: 0.36, B: 0.09, A: 1},
		EdgeWidth: floatPtr(1.0),
		Size:      floatPtr(8.0),
	})
	if err != nil {
		panic(err)
	}
	ax.AutoScale(0.08)
	ax.SetXLim(2.024, 45.176)
	ax.SetYLim(4.996, 6.504)
	ax.XAxis.Locator = core.MultipleLocator{Base: 5}
	ax.YAxis.Locator = core.MultipleLocator{Base: 0.2}
	ax.YAxis.Formatter = core.FormatStrFormatter{Pattern: "%.1f"}

	r, err := agg.New(680, 380, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderPatchShowcase() image.Image {
	fig := core.NewFigure(930, 340)

	left := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.05, Y: 0.16}, Max: geom.Pt{X: 0.31, Y: 0.88}})
	left.SetTitle("Patch Primitives")
	left.SetXLim(0, 6)
	left.SetYLim(0, 4)
	left.AddPatch(&core.Rectangle{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.95, G: 0.70, B: 0.23, A: 0.86},
			EdgeColor: render.Color{R: 0.48, G: 0.27, B: 0.08, A: 1},
			EdgeWidth: 1.1,
			Hatch:     "/",
		},
		XY:     geom.Pt{X: 0.6, Y: 0.7},
		Width:  1.5,
		Height: 1.0,
	})
	left.AddPatch(&core.Circle{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.22, G: 0.57, B: 0.82, A: 0.82},
			EdgeColor: render.Color{R: 0.11, G: 0.29, B: 0.44, A: 1},
			EdgeWidth: 1.0,
		},
		Center: geom.Pt{X: 3.0, Y: 1.25},
		Radius: 0.56,
	})
	left.AddPatch(&core.Ellipse{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.23, G: 0.72, B: 0.51, A: 0.80},
			EdgeColor: render.Color{R: 0.10, G: 0.36, B: 0.24, A: 1},
			EdgeWidth: 1.0,
		},
		Center: geom.Pt{X: 4.8, Y: 2.75},
		Width:  1.55,
		Height: 0.95,
		Angle:  28,
	})
	left.AddPatch(&core.Polygon{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.84, G: 0.34, B: 0.34, A: 0.82},
			EdgeColor: render.Color{R: 0.48, G: 0.14, B: 0.14, A: 1},
			EdgeWidth: 1.0,
		},
		XY: []geom.Pt{
			{X: 2.15, Y: 3.2},
			{X: 2.85, Y: 2.25},
			{X: 1.35, Y: 2.45},
		},
	})

	middle := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.37, Y: 0.16}, Max: geom.Pt{X: 0.63, Y: 0.88}})
	middle.SetTitle("Fancy Arrow + Path")
	middle.SetXLim(0, 6)
	middle.SetYLim(0, 4)
	middle.AddPatch(&core.FancyArrow{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.91, G: 0.42, B: 0.22, A: 0.88},
			EdgeColor: render.Color{R: 0.58, G: 0.22, B: 0.10, A: 1},
			EdgeWidth: 1.0,
		},
		XY:         geom.Pt{X: 0.9, Y: 3.2},
		DX:         2.2,
		DY:         -1.0,
		Width:      0.18,
		HeadWidth:  0.62,
		HeadLength: 0.62,
	})
	star := geom.Path{}
	star.MoveTo(geom.Pt{X: 4.15, Y: 0.95})
	star.LineTo(geom.Pt{X: 4.45, Y: 1.70})
	star.LineTo(geom.Pt{X: 5.22, Y: 1.75})
	star.LineTo(geom.Pt{X: 4.63, Y: 2.22})
	star.LineTo(geom.Pt{X: 4.84, Y: 2.96})
	star.LineTo(geom.Pt{X: 4.15, Y: 2.54})
	star.LineTo(geom.Pt{X: 3.46, Y: 2.96})
	star.LineTo(geom.Pt{X: 3.67, Y: 2.22})
	star.LineTo(geom.Pt{X: 3.08, Y: 1.75})
	star.LineTo(geom.Pt{X: 3.85, Y: 1.70})
	star.Close()
	middle.AddPatch(&core.PathPatch{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.76, G: 0.76, B: 0.86, A: 0.72},
			EdgeColor: render.Color{R: 0.29, G: 0.29, B: 0.45, A: 1},
			EdgeWidth: 1.0,
			Hatch:     "x",
		},
		Path: star,
	})

	right := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.69, Y: 0.16}, Max: geom.Pt{X: 0.95, Y: 0.88}})
	right.SetTitle("Fancy Boxes")
	right.SetXLim(0, 6)
	right.SetYLim(0, 4)
	right.AddPatch(&core.FancyBboxPatch{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.29, G: 0.67, B: 0.78, A: 0.28},
			EdgeColor: render.Color{R: 0.10, G: 0.37, B: 0.45, A: 1},
			EdgeWidth: 1.0,
			Hatch:     "/",
		},
		XY:           geom.Pt{X: 0.9, Y: 0.8},
		Width:        2.1,
		Height:       1.25,
		Pad:          0.14,
		BoxStyle:     core.BoxStyleRound,
		RoundingSize: 0.24,
	})
	right.AddPatch(&core.FancyBboxPatch{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.96, G: 0.87, B: 0.60, A: 0.82},
			EdgeColor: render.Color{R: 0.50, G: 0.39, B: 0.12, A: 1},
			EdgeWidth: 1.0,
		},
		XY:       geom.Pt{X: 3.35, Y: 1.55},
		Width:    1.65,
		Height:   1.05,
		Pad:      0.10,
		BoxStyle: core.BoxStyleSquare,
	})

	r, err := agg.New(930, 340, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderMeshContourTri() image.Image {
	fig := core.NewFigure(980, 620)
	halfTicks := func(ax *core.Axes) {
		if ax == nil {
			return
		}
		ax.XAxis.Locator = core.MultipleLocator{Base: 0.5}
		ax.YAxis.Locator = core.MultipleLocator{Base: 0.5}
		ax.XAxis.Formatter = core.FormatStrFormatter{Pattern: "%.1f"}
		ax.YAxis.Formatter = core.FormatStrFormatter{Pattern: "%.1f"}
	}

	meshAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.57}, Max: geom.Pt{X: 0.46, Y: 0.93}})
	meshAx.SetTitle("PColorMesh")
	meshAx.SetXLim(0, 4)
	meshAx.SetYLim(0, 3)
	halfTicks(meshAx)
	meshEdgeWidth := 0.8
	meshEdgeColor := render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1}
	meshAx.PColorMesh([][]float64{
		{0.2, 0.6, 0.3, 0.9},
		{0.4, 0.8, 0.5, 0.7},
		{0.1, 0.3, 0.9, 0.6},
	}, core.MeshOptions{
		XEdges:    []float64{0, 1, 2, 3, 4},
		YEdges:    []float64{0, 1, 2, 3},
		EdgeColor: &meshEdgeColor,
		EdgeWidth: &meshEdgeWidth,
	})

	contourAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.57}, Max: geom.Pt{X: 0.96, Y: 0.93}})
	contourAx.SetTitle("Contour + Contourf")
	contourAx.SetXLim(0, 4)
	contourAx.SetYLim(0, 4)
	halfTicks(contourAx)
	contourData := [][]float64{
		{0.0, 0.4, 0.8, 0.4, 0.0},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.3, 1.0, 1.7, 1.0, 0.3},
		{0.2, 0.8, 1.3, 0.8, 0.2},
		{0.0, 0.4, 0.8, 0.4, 0.0},
	}
	contourLevels := []float64{0.2, 0.6, 1.0, 1.4, 1.8}
	contourAx.Contourf(contourData, core.ContourOptions{
		Levels: contourLevels,
	})
	contourAx.Contour(contourData, core.ContourOptions{
		Levels:     []float64{0.4, 0.8, 1.2, 1.6},
		LabelLines: true,
		Color:      &render.Color{R: 0.18, G: 0.18, B: 0.18, A: 1},
	})

	histAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.10}, Max: geom.Pt{X: 0.46, Y: 0.46}})
	histAx.SetTitle("Hist2D")
	histAx.SetXLim(0, 4)
	histAx.SetYLim(0, 4)
	halfTicks(histAx)
	histAx.Hist2D(
		[]float64{0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6},
		[]float64{0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4},
		core.Hist2DOptions{
			XBinEdges: []float64{0, 1, 2, 3, 4},
			YBinEdges: []float64{0, 1, 2, 3, 4},
			EdgeColor: &meshEdgeColor,
			EdgeWidth: &meshEdgeWidth,
		},
	)

	triAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.57, Y: 0.10}, Max: geom.Pt{X: 0.96, Y: 0.46}})
	triAx.SetTitle("Triangulation")
	triAx.SetXLim(0, 4)
	triAx.SetYLim(0, 4)
	halfTicks(triAx)
	tri := core.Triangulation{
		X:         []float64{0.4, 1.6, 3.0, 0.8, 2.1, 3.5},
		Y:         []float64{0.5, 0.4, 0.7, 2.2, 2.8, 2.1},
		Triangles: [][3]int{{0, 1, 3}, {1, 4, 3}, {1, 2, 4}, {2, 5, 4}},
	}
	triAx.TriColor(tri, []float64{0.2, 0.8, 1.0, 1.5, 1.1, 0.6})
	triLineWidth := 1.0
	triAx.TriPlot(tri, core.TriPlotOptions{
		Color:     &render.Color{R: 0.15, G: 0.15, B: 0.15, A: 1},
		LineWidth: &triLineWidth,
	})
	triAx.TriContour(tri, []float64{0.2, 0.8, 1.0, 1.5, 1.1, 0.6}, core.ContourOptions{
		Levels: []float64{0.7, 1.1},
		Color:  &render.Color{R: 0.98, G: 0.98, B: 0.98, A: 1},
	})

	r, err := agg.New(980, 620, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderStemPlot() image.Image {
	fig := core.NewFigure(720, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.86}})
	ax.SetTitle("Stem")
	ax.SetXLabel("Sample")
	ax.SetYLabel("Amplitude")
	ax.SetXLim(0.5, 7.5)
	ax.SetYLim(-0.2, 4.2)
	ax.AddYGrid()
	stemColor := render.Color{R: 0.15, G: 0.42, B: 0.73, A: 1}
	baseline := 0.3
	markerSize := 7.0
	ax.Stem(
		[]float64{1, 2, 3, 4, 5, 6, 7},
		[]float64{0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1},
		core.StemOptions{
			Color:         &stemColor,
			Baseline:      &baseline,
			MarkerSize:    &markerSize,
			Label:         "samples",
			BaselineColor: &render.Color{R: 0.32, G: 0.32, B: 0.32, A: 1},
		},
	)

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderSpecialtyArtists() image.Image {
	fig := core.NewFigure(980, 720)

	eventAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.57}, Max: geom.Pt{X: 0.34, Y: 0.94}})
	eventAx.SetTitle("Eventplot")
	eventAx.SetXLim(0, 10)
	eventAx.SetYLim(0.4, 3.6)
	eventAx.AddXGrid()
	eventAx.Eventplot([][]float64{
		{0.8, 1.4, 3.1, 4.6, 7.3},
		{1.2, 2.9, 4.0, 6.4, 8.6},
		{0.5, 2.2, 5.4, 6.8, 9.1},
	}, core.EventPlotOptions{
		LineOffsets: []float64{1, 2, 3},
		LineLengths: []float64{0.6, 0.7, 0.8},
		Colors: []render.Color{
			{R: 0.18, G: 0.44, B: 0.74, A: 1},
			{R: 0.84, G: 0.38, B: 0.16, A: 1},
			{R: 0.20, G: 0.63, B: 0.42, A: 1},
		},
	})

	hexAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.39, Y: 0.57}, Max: geom.Pt{X: 0.66, Y: 0.94}})
	hexAx.SetTitle("Hexbin")
	hexAx.SetXLim(0, 1)
	hexAx.SetYLim(0, 1)
	hexExtent := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}
	hexAx.Hexbin(
		[]float64{0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88},
		[]float64{0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58},
		core.HexbinOptions{
			GridSizeX: 7,
			Extent:    &hexExtent,
			C:         []float64{1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1},
			Reduce:    "mean",
			Label:     "clusters",
		},
	)

	pieAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.73, Y: 0.57}, Max: geom.Pt{X: 0.98, Y: 0.94}})
	pieAx.SetTitle("Pie")
	pie := pieAx.Pie([]float64{28, 22, 18, 32}, core.PieOptions{
		Labels:        []string{"Core", "I/O", "Render", "Docs"},
		AutoPct:       "%.0f%%",
		StartAngle:    90,
		LabelDistance: 0.98,
		Explode:       []float64{0, 0.04, 0, 0.02},
		Colors: []render.Color{
			{R: 0.12, G: 0.47, B: 0.71, A: 1},
			{R: 1.00, G: 0.50, B: 0.05, A: 1},
			{R: 0.17, G: 0.63, B: 0.17, A: 1},
			{R: 0.84, G: 0.15, B: 0.16, A: 1},
		},
	})
	if pie != nil {
		for _, txt := range pie.AutoText {
			txt.Color = render.Color{R: 0, G: 0, B: 0, A: 1}
		}
	}

	violinAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.08}, Max: geom.Pt{X: 0.34, Y: 0.45}})
	violinAx.SetTitle("Violin")
	violinAx.SetXLim(0.4, 3.6)
	violinAx.SetYLim(0.8, 5.2)
	violinAx.AddYGrid()
	violinBlue := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	violinEdge := render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1}
	violinShowMeans := true
	violinShowMedians := false
	violinShowExtrema := true
	violinPoints := 100
	violinWidths := []float64{0.5, 0.5, 0.5}
	violinAx.Violinplot([][]float64{
		{1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2},
		{1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0},
		{2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6},
	}, core.ViolinOptions{
		Colors:      []render.Color{violinBlue, violinBlue, violinBlue},
		EdgeColor:   &violinEdge,
		LineColor:   &violinBlue,
		Alpha:       0.45,
		Points:      violinPoints,
		Widths:      violinWidths,
		ShowMeans:   &violinShowMeans,
		ShowMedians: &violinShowMedians,
		ShowExtrema: &violinShowExtrema,
		Label:       "distribution",
	})

	tableAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.39, Y: 0.08}, Max: geom.Pt{X: 0.66, Y: 0.45}})
	tableAx.SetTitle("Table")
	tableAx.ShowFrame = false
	if tableAx.XAxis != nil {
		tableAx.XAxis.ShowSpine = false
		tableAx.XAxis.ShowTicks = false
		tableAx.XAxis.ShowLabels = false
	}
	if tableAx.YAxis != nil {
		tableAx.YAxis.ShowSpine = false
		tableAx.YAxis.ShowTicks = false
		tableAx.YAxis.ShowLabels = false
	}
	tableAx.Table(core.TableOptions{
		ColLabels: []string{"Metric", "Q1", "Q2"},
		CellText: [][]string{
			{"Latency", "18ms", "14ms"},
			{"Throughput", "220/s", "265/s"},
		},
		BBox: geom.Rect{
			Min: geom.Pt{X: 0.04, Y: 0.18},
			Max: geom.Pt{X: 0.96, Y: 0.82},
		},
		FontSize:        10,
		TextColor:       render.Color{R: 0, G: 0, B: 0, A: 1},
		HeaderTextColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		HeaderFillColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		CellFillColor:   render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor:       render.Color{R: 0, G: 0, B: 0, A: 1},
	})
	rowLabelCells := []struct {
		label string
		y     float64
	}{
		{label: "A", y: 0.3933},
		{label: "B", y: 0.1800},
	}
	for _, cell := range rowLabelCells {
		tableAx.AddPatch(&core.Rectangle{
			Patch: core.Patch{
				FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
				EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
				EdgeWidth: 1,
			},
			XY:     geom.Pt{X: 0, Y: cell.y},
			Width:  0.04,
			Height: 0.2133,
			Coords: core.Coords(core.CoordAxes),
		})
		tableAx.Text(0.012, cell.y+0.10665, cell.label, core.TextOptions{
			Coords:   core.Coords(core.CoordAxes),
			HAlign:   core.TextAlignLeft,
			VAlign:   core.TextVAlignMiddle,
			FontSize: 10,
			Color:    render.Color{R: 0, G: 0, B: 0, A: 1},
		})
	}

	sankeyAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.73, Y: 0.08}, Max: geom.Pt{X: 0.98, Y: 0.45}})
	sankeyAx.SetTitle("Sankey")
	sankeyAx.ShowFrame = false
	if sankeyAx.XAxis != nil {
		sankeyAx.XAxis.ShowSpine = false
		sankeyAx.XAxis.ShowTicks = false
		sankeyAx.XAxis.ShowLabels = false
	}
	if sankeyAx.YAxis != nil {
		sankeyAx.YAxis.ShowSpine = false
		sankeyAx.YAxis.ShowTicks = false
		sankeyAx.YAxis.ShowLabels = false
	}
	sankeyAx.SetXLim(0, 1)
	sankeyAx.SetYLim(0, 1)
	sankeyAx.AddPatch(&core.Rectangle{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.12, G: 0.47, B: 0.71, A: 0.75},
			EdgeColor: render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1},
			EdgeWidth: 1,
		},
		XY:     geom.Pt{X: 0.18, Y: 0.47},
		Width:  0.18,
		Height: 0.06,
		Coords: core.Coords(core.CoordAxes),
	})
	sankeyFlows := []struct {
		label string
		flow  float64
		color render.Color
	}{
		{label: "Waste", flow: -2, color: render.Color{R: 0.84, G: 0.15, B: 0.16, A: 0.75}},
		{label: "CPU", flow: 3, color: render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.75}},
		{label: "Cache", flow: 1.5, color: render.Color{R: 1.00, G: 0.50, B: 0.05, A: 0.75}},
	}
	for idx, flow := range sankeyFlows {
		width := math.Abs(flow.flow) * 0.018
		y := 0.40 + float64(idx)*0.095
		x0 := 0.36
		x1 := 0.66
		path := geom.Path{}
		path.MoveTo(geom.Pt{X: x0, Y: 0.50 - width/2})
		path.LineTo(geom.Pt{X: x0 + 0.10, Y: y - width/2})
		path.LineTo(geom.Pt{X: x1, Y: y - width/2})
		path.LineTo(geom.Pt{X: x1, Y: y + width/2})
		path.LineTo(geom.Pt{X: x0 + 0.10, Y: y + width/2})
		path.LineTo(geom.Pt{X: x0, Y: 0.50 + width/2})
		path.Close()
		sankeyAx.AddPatch(&core.PathPatch{
			Patch: core.Patch{
				FaceColor: flow.color,
				EdgeColor: render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1},
				EdgeWidth: 1,
			},
			Path:   path,
			Coords: core.Coords(core.CoordAxes),
		})
		sankeyAx.Text(0.70, y, flow.label, core.TextOptions{
			Coords:   core.Coords(core.CoordAxes),
			VAlign:   core.TextVAlignMiddle,
			FontSize: 10,
		})
	}

	r, err := agg.New(980, 720, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderVectorFields() image.Image {
	fig := core.NewFigure(919, 620)
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.07, 0.97, 0.10, 0.92),
		core.WithSubplotSpacing(0.10, 0.16),
	)

	quiverAx := grid[0][0]
	quiverAx.SetTitle("Quiver + Key")
	quiverAx.SetXLim(0, 6)
	quiverAx.SetYLim(0, 5)
	quiverAx.AddXGrid()
	quiverAx.AddYGrid()
	var qx, qy, qu, qv []float64
	for row := 0; row < 4; row++ {
		for col := 0; col < 5; col++ {
			x := 0.8 + float64(col)*1.0
			y := 0.8 + float64(row)*0.95
			qx = append(qx, x)
			qy = append(qy, y)
			qu = append(qu, 0.55+0.08*math.Sin(y*0.9))
			qv = append(qv, 0.22*math.Cos(x*0.8))
		}
	}
	scaleWidth := 10.0
	widthDots := 2.2
	quiver := quiverAx.Quiver(qx, qy, qu, qv, core.QuiverOptions{
		Color:      &render.Color{R: 0.14, G: 0.42, B: 0.73, A: 1},
		Scale:      &scaleWidth,
		ScaleUnits: "width",
		Units:      "dots",
		Width:      &widthDots,
	})
	if quiver != nil {
		quiverAx.QuiverKey(quiver, 0.78, 0.12, 0.5, "0.5", core.QuiverKeyOptions{
			Coords:   core.Coords(core.CoordAxes),
			LabelPos: "E",
			LabelSep: 10,
		})
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
			bx = append(bx, x)
			by = append(by, y)
			bu = append(bu, 14+5*math.Sin(y*0.8))
			bv = append(bv, 8*math.Cos(x*0.7))
		}
	}
	barbLen := 28.0
	barbAx.Barbs(bx, by, bu, bv, core.BarbsOptions{
		BarbColor: &render.Color{R: 0.47, G: 0.23, B: 0.12, A: 1},
		FlagColor: &render.Color{R: 0.86, G: 0.52, B: 0.24, A: 1},
		Length:    &barbLen,
		Units:     "dots",
	})

	streamAx := grid[1][0]
	streamAx.SetTitle("Streamplot")
	streamAx.SetXLim(0, 6)
	streamAx.SetYLim(0, 5)
	streamAx.AddXGrid()
	streamAx.AddYGrid()
	sx := []float64{0, 1, 2, 3, 4, 5, 6}
	sy := []float64{0, 1, 2, 3, 4, 5}
	su := make([][]float64, len(sy))
	sv := make([][]float64, len(sy))
	for yi, y := range sy {
		su[yi] = make([]float64, len(sx))
		sv[yi] = make([]float64, len(sx))
		for xi, x := range sx {
			su[yi][xi] = 1.0 + 0.12*math.Cos(y*0.7)
			sv[yi][xi] = 0.35*math.Sin((x-3)*0.8) - 0.10*(y-2.5)
		}
	}
	streamFalse := false
	streamArrows := 1
	streamAx.Streamplot(sx, sy, su, sv, core.StreamplotOptions{
		StartPoints:          []geom.Pt{{X: 0.4, Y: 0.8}, {X: 0.4, Y: 2.2}, {X: 0.4, Y: 3.6}},
		BrokenStreamlines:    &streamFalse,
		IntegrationDirection: "forward",
		ArrowCount:           &streamArrows,
		Color:                &render.Color{R: 0.13, G: 0.53, B: 0.39, A: 1},
	})

	xyAx := grid[1][1]
	xyAx.SetTitle("Quiver XY")
	xyAx.SetXLim(0, 6)
	xyAx.SetYLim(0, 5)
	xyAx.AddXGrid()
	xyAx.AddYGrid()
	xg := []float64{0.8, 1.8, 2.8, 3.8, 4.8}
	yg := []float64{0.8, 1.8, 2.8, 3.8}
	ugu := make([][]float64, len(yg))
	ugv := make([][]float64, len(yg))
	for yi, y := range yg {
		ugu[yi] = make([]float64, len(xg))
		ugv[yi] = make([]float64, len(xg))
		for xi, x := range xg {
			ugu[yi][xi] = -(y - 2.4) * 0.35
			ugv[yi][xi] = (x - 2.8) * 0.35
		}
	}
	xyScale := 9.0
	xyWidth := 1.9
	xyAx.QuiverGrid(xg, yg, ugu, ugv, core.QuiverOptions{
		Color:      &render.Color{R: 0.74, G: 0.23, B: 0.27, A: 1},
		Pivot:      "middle",
		Angles:     "xy",
		Scale:      &xyScale,
		ScaleUnits: "width",
		Units:      "dots",
		Width:      &xyWidth,
	})

	r, err := agg.New(919, 620, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderPolarAxes() image.Image {
	fig := core.NewFigure(720, 720)
	ax := fig.AddPolarAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	ax.SetTitle("Polar Axes")
	ax.SetXLabel("theta")
	ax.SetYLabel("radius")
	ax.YScale = transform.NewLinear(0, 1.1)

	thetaGrid := ax.AddGrid(core.AxisBottom)
	thetaGrid.Color = render.Color{R: 0.8, G: 0.82, B: 0.86, A: 1}
	thetaGrid.LineWidth = 0.9

	radiusGrid := ax.AddGrid(core.AxisLeft)
	radiusGrid.Color = render.Color{R: 0.82, G: 0.84, B: 0.88, A: 0.9}
	radiusGrid.LineWidth = 0.8

	const n = 720
	theta := make([]float64, n)
	radius := make([]float64, n)
	for i := range n {
		theta[i] = 2 * math.Pi * float64(i) / float64(n-1)
		radius[i] = 0.55 + 0.35*math.Cos(5*theta[i])
	}

	lineColor := render.Color{R: 0.16, G: 0.33, B: 0.73, A: 1}
	fillColor := render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2}
	lineWidth := 2.2

	ax.Plot(theta, radius, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
		Label:     "r = 0.55 + 0.35 cos(5theta)",
	})
	ax.FillToBaseline(theta, radius, core.FillOptions{
		Color: &fillColor,
	})

	r, err := agg.New(720, 720, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderGeoMollweideAxes() image.Image {
	fig := core.NewFigure(720, 420)
	ax, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.92, Y: 0.86},
	}, "mollweide")
	if err != nil {
		panic(err)
	}
	ax.SetTitle("Mollweide Projection")
	ax.SetXLabel("longitude")
	ax.SetYLabel("latitude")

	gridColor := render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	lonGrid := ax.AddGrid(core.AxisBottom)
	lonGrid.Color = gridColor
	lonGrid.LineWidth = 0.8
	latGrid := ax.AddGrid(core.AxisLeft)
	latGrid.Color = gridColor
	latGrid.LineWidth = 0.8

	const n = 361
	lon := make([]float64, n)
	lat := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		lon[i] = -math.Pi + 2*math.Pi*t
		lat[i] = 0.35 * math.Sin(3*lon[i])
	}

	lineColor := render.Color{R: 0.14, G: 0.34, B: 0.70, A: 1}
	lineWidth := 2.0
	ax.Plot(lon, lat, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
	})

	r, err := agg.New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func floatPtr(v float64) *float64 {
	return &v
}
