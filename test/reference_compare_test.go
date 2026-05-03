package test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/imagecmp"
)

const (
	referenceCompareTolerance = 1
	// Titled parity fixtures now exercise text rendering on every case, so the
	// matplotlib cross-check needs text-aware tolerances instead of shape-only
	// thresholds. Golden images remain exact.
	referenceCompareMinPSNR    = 44.0
	referenceCompareMaxMeanAbs = 2.50
)

type referenceCompareCase struct {
	name       string
	render     func() image.Image
	minPSNR    float64
	maxMeanAbs float64
}

var referenceCompareCases = []referenceCompareCase{
	{name: "basic_line", render: renderBasicLine},
	{name: "joins_caps", render: renderJoinsCaps},
	{name: "dashes", render: renderDashes},
	{name: "scatter_basic", render: renderScatterBasic},
	{name: "scatter_marker_types", render: renderScatterMarkerTypes},
	{name: "scatter_advanced", render: renderScatterAdvanced},
	{name: "bar_basic_frame", render: renderBarBasicFrame},
	{name: "bar_basic_ticks", render: renderBarBasicTicks},
	{name: "bar_basic_tick_labels", render: renderBarBasicTickLabels},
	{name: "bar_basic_title", render: renderBarBasicTitle},
	{name: "bar_basic", render: renderBarBasic},
	{name: "bar_horizontal", render: renderBarHorizontal},
	{name: "bar_grouped", render: renderBarGrouped},
	{name: "fill_basic", render: renderFillBasic, minPSNR: 45.0, maxMeanAbs: 6.0},
	{name: "fill_between", render: renderFillBetween},
	{name: "fill_stacked", render: renderFillStacked},
	{name: "multi_series_basic", render: renderMultiSeriesBasic},
	{name: "multi_series_color_cycle", render: renderMultiSeriesColorCycle},
	{name: "hist_basic", render: renderHistBasic},
	{name: "hist_density", render: renderHistDensity},
	{name: "hist_strategies", render: renderHistStrategies},
	{name: "boxplot_basic", render: renderBoxPlotBasic, minPSNR: 44.0, maxMeanAbs: 2.0},
	{name: "axes_top_right_inverted", render: renderAxesTopRightInverted},
	{name: "axes_control_surface", render: renderAxesControlSurface, minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "transform_coordinates", render: renderTransformCoordinates, minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "gridspec_composition", render: renderGridSpecComposition, minPSNR: 35.0, maxMeanAbs: 8.0},
	{name: "figure_labels_composition", render: renderFigureLabelsComposition, minPSNR: 32.0, maxMeanAbs: 9.0},
	{name: "colorbar_composition", render: renderColorbarComposition, minPSNR: 32.0, maxMeanAbs: 16.0},
	{name: "annotation_composition", render: renderAnnotationComposition, minPSNR: 35.0, maxMeanAbs: 7.0},
	{name: "patch_showcase", render: renderPatchShowcase, minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "mesh_contour_tri", render: renderMeshContourTri, minPSNR: 37.5, maxMeanAbs: 7.5},
	{name: "plot_variants", render: renderPlotVariants, minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "stat_variants", render: renderStatVariants, minPSNR: 32.0, maxMeanAbs: 9.0},
	{name: "stem_plot", render: renderStemPlot},
	{name: "specialty_artists", render: renderSpecialtyArtists},
	{name: "units_overview", render: renderUnitsOverview, minPSNR: 43.5},
	{name: "units_dates", render: renderUnitsDates, minPSNR: 45.0, maxMeanAbs: 1.6},
	{name: "units_categories", render: renderUnitsCategories, minPSNR: 41.0, maxMeanAbs: 3.2},
	{name: "units_custom_converter", render: renderUnitsCustomConverter, minPSNR: 40.0, maxMeanAbs: 3.5},
	{name: "vector_fields", render: renderVectorFields, minPSNR: 41.5, maxMeanAbs: 3.0},
	{name: "polar_axes", render: renderPolarAxes, minPSNR: 32.0, maxMeanAbs: 9.0},
	{name: "geo_mollweide_axes", render: renderGeoMollweideAxes, minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "unstructured_showcase", render: renderUnstructuredShowcase, minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "arrays_showcase", render: renderArraysShowcase, minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "axisartist_showcase", render: renderAxisArtistShowcase, minPSNR: 28.0, maxMeanAbs: 12.0},
	{name: "axes_grid1_showcase", render: renderAxesGrid1Showcase, minPSNR: 28.0, maxMeanAbs: 12.0},
	{name: "rendereragg_large_scatter", render: renderRendererAggLargeScatter, minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "rendereragg_mixed_collection", render: renderRendererAggMixedCollection, minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "rendereragg_quad_mesh", render: renderRendererAggQuadMesh, minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "rendereragg_gouraud_triangles", render: renderRendererAggGouraudTriangles, minPSNR: 25.0, maxMeanAbs: 18.0},
}

func TestReferenceImages_GoldenVsMatplotlibRef(t *testing.T) {
	for _, tc := range referenceCompareCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runReferenceCompareTest(t, tc)
		})
	}
}

func TestColorbarCompositionImageOriginMatchesMatplotlibRef(t *testing.T) {
	got := renderColorbarComposition()
	matplotlibRef, err := imagecmp.LoadPNG(filepath.Join("..", "testdata", "matplotlib_ref", "colorbar_composition.png"))
	if err != nil {
		t.Fatalf("load matplotlib reference: %v", err)
	}

	gotRect, ok := largestChromaComponentBounds(got)
	if !ok {
		t.Fatal("rendered colorbar composition has no chromatic component")
	}
	wantRect, ok := largestChromaComponentBounds(matplotlibRef)
	if !ok {
		t.Fatal("matplotlib colorbar composition has no chromatic component")
	}

	if !rectWithinPixels(gotRect, wantRect, 1) {
		t.Fatalf("main image component bounds = %v, want within 1 px of matplotlib %v", gotRect, wantRect)
	}
}

func runReferenceCompareTest(t *testing.T, tc referenceCompareCase) {
	t.Helper()

	got := tc.render()

	goldenPath := filepath.Join("..", "testdata", "golden", tc.name+".png")
	matplotlibPath := filepath.Join("..", "testdata", "matplotlib_ref", tc.name+".png")

	golden, err := imagecmp.LoadPNG(goldenPath)
	if err != nil {
		t.Fatalf("load golden image %s: %v", goldenPath, err)
	}

	matplotlibRef, err := imagecmp.LoadPNG(matplotlibPath)
	if err != nil {
		t.Fatalf("load matplotlib reference %s: %v", matplotlibPath, err)
	}

	artifactsDir := filepath.Join("..", "testdata", "_artifacts", "reference_compare")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("create artifacts directory %s: %v", artifactsDir, err)
	}

	savePNGOrFail(t, got, filepath.Join(artifactsDir, tc.name+"_rendered.png"))
	savePNGOrFail(t, golden, filepath.Join(artifactsDir, tc.name+"_golden.png"))
	savePNGOrFail(t, matplotlibRef, filepath.Join(artifactsDir, tc.name+"_matplotlib_ref.png"))

	diff, err := imagecmp.ComparePNG(golden, matplotlibRef, referenceCompareTolerance)
	if err != nil {
		t.Fatalf("compare %s and %s: %v", goldenPath, matplotlibPath, err)
	}

	diffPath := filepath.Join(artifactsDir, tc.name+"_golden_vs_matplotlib_ref_diff.png")
	if err := imagecmp.SaveDiffImage(golden, matplotlibRef, referenceCompareTolerance, diffPath); err != nil {
		t.Fatalf("save diff image %s: %v", diffPath, err)
	}

	t.Logf(
		"%s: MaxDiff=%d MeanAbs=%.2f PSNR=%.2fdB",
		tc.name,
		diff.MaxDiff,
		diff.MeanAbs,
		diff.PSNR,
	)

	minPSNR := referenceCompareMinPSNR
	if tc.minPSNR > 0 {
		minPSNR = tc.minPSNR
	}
	maxMeanAbs := referenceCompareMaxMeanAbs
	if tc.maxMeanAbs > 0 {
		maxMeanAbs = tc.maxMeanAbs
	}
	if diff.PSNR < minPSNR || diff.MeanAbs > maxMeanAbs {
		t.Fatalf(
			"reference mismatch for %s: PSNR=%.2f (min %.2f), MeanAbs=%.2f (max %.2f)",
			tc.name,
			diff.PSNR,
			minPSNR,
			diff.MeanAbs,
			maxMeanAbs,
		)
	}
}

func savePNGOrFail(t *testing.T, img image.Image, path string) {
	t.Helper()
	if err := imagecmp.SavePNG(img, path); err != nil {
		t.Fatalf("save PNG %s: %v", path, err)
	}
}

func largestChromaComponentBounds(img image.Image) (image.Rectangle, bool) {
	if img == nil {
		return image.Rectangle{}, false
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return image.Rectangle{}, false
	}

	mask := make([]bool, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if isChromaticPixel(img.At(bounds.Min.X+x, bounds.Min.Y+y)) {
				mask[y*width+x] = true
			}
		}
	}

	visited := make([]bool, len(mask))
	queue := make([]int, 0, len(mask)/8)
	bestArea := 0
	best := image.Rectangle{}

	for idx, chromatic := range mask {
		if !chromatic || visited[idx] {
			continue
		}
		visited[idx] = true
		queue = append(queue[:0], idx)
		area := 0
		minX, maxX := idx%width, idx%width
		minY, maxY := idx/width, idx/width

		for len(queue) > 0 {
			cur := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			x := cur % width
			y := cur / width
			area++
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}

			neighbors := [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}}
			for _, neighbor := range neighbors {
				nx, ny := neighbor[0], neighbor[1]
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					continue
				}
				next := ny*width + nx
				if !mask[next] || visited[next] {
					continue
				}
				visited[next] = true
				queue = append(queue, next)
			}
		}

		if area > bestArea {
			bestArea = area
			best = image.Rect(bounds.Min.X+minX, bounds.Min.Y+minY, bounds.Min.X+maxX+1, bounds.Min.Y+maxY+1)
		}
	}

	return best, bestArea > 0
}

func isChromaticPixel(c color.Color) bool {
	r16, g16, b16, a16 := c.RGBA()
	if a16 == 0 {
		return false
	}
	r := int(r16 >> 8)
	g := int(g16 >> 8)
	b := int(b16 >> 8)
	lo := min(r, min(g, b))
	hi := max(r, max(g, b))
	return hi-lo > 20 && hi < 252
}

func rectWithinPixels(got, want image.Rectangle, tolerance int) bool {
	return absInt(got.Min.X-want.Min.X) <= tolerance &&
		absInt(got.Min.Y-want.Min.Y) <= tolerance &&
		absInt(got.Max.X-want.Max.X) <= tolerance &&
		absInt(got.Max.Y-want.Max.Y) <= tolerance
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
