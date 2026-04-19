package test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"matplotlib-go/test/imagecmp"
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
}

func TestReferenceImages_GoldenVsMatplotlibRef(t *testing.T) {
	for _, tc := range referenceCompareCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runReferenceCompareTest(t, tc)
		})
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
