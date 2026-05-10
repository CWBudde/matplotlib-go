package test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/parity"
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
	minPSNR    float64
	maxMeanAbs float64
	maxRMSE    float64
}

var referenceCompareCases = []referenceCompareCase{
	{name: "basic_line"},
	{name: "joins_caps"},
	{name: "dashes"},
	{name: "scatter_basic"},
	{name: "scatter_marker_types"},
	{name: "scatter_advanced"},
	{name: "bar_basic_frame"},
	{name: "bar_basic_ticks"},
	{name: "bar_basic_tick_labels"},
	{name: "bar_basic_title"},
	{name: "bar_basic"},
	{name: "bar_horizontal"},
	{name: "bar_grouped"},
	{name: "fill_basic", minPSNR: 45.0, maxMeanAbs: 6.0},
	{name: "fill_between"},
	{name: "fill_stacked"},
	{name: "multi_series_basic"},
	{name: "multi_series_color_cycle"},
	{name: "hist_basic"},
	{name: "hist_density"},
	{name: "hist_strategies"},
	{name: "boxplot_basic", minPSNR: 44.0, maxMeanAbs: 2.0},
	{name: "axes_top_right_inverted"},
	{name: "axes_control_surface", minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "transform_coordinates", minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "gridspec_composition", minPSNR: 35.0, maxMeanAbs: 8.0},
	{name: "figure_labels_composition", minPSNR: 32.0, maxMeanAbs: 9.0},
	{name: "colorbar_composition", minPSNR: 32.0, maxMeanAbs: 16.0},
	{name: "annotation_composition", minPSNR: 35.0, maxMeanAbs: 7.0},
	{name: "patch_showcase", minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "mesh_contour_tri", minPSNR: 37.5, maxMeanAbs: 7.5},
	{name: "plot_variants", minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "spectrum_variants", minPSNR: 35.0, maxMeanAbs: 6.5},
	{name: "stat_variants", minPSNR: 32.0, maxMeanAbs: 9.0},
	{name: "phase12_specialty_depth", minPSNR: 22.0, maxMeanAbs: 20.0, maxRMSE: 35.0},
	{name: "stem_plot"},
	{name: "specialty_artists"},
	{name: "units_overview", minPSNR: 43.5},
	{name: "units_dates", minPSNR: 45.0, maxMeanAbs: 1.6},
	{name: "units_categories", minPSNR: 41.0, maxMeanAbs: 3.2},
	{name: "units_custom_converter", minPSNR: 40.0, maxMeanAbs: 3.5},
	{name: "vector_fields", minPSNR: 41.5, maxMeanAbs: 3.0},
	{name: "imshow_clipped", minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "imshow_transformed", minPSNR: 24.0, maxMeanAbs: 18.0, maxRMSE: 30.0},
	{name: "imshow_bilinear", minPSNR: 30.0, maxMeanAbs: 16.0},
	{name: "image_alpha", minPSNR: 30.0, maxMeanAbs: 16.0, maxRMSE: 10.0},
	{name: "matshow_basic", minPSNR: 30.0, maxMeanAbs: 10.0, maxRMSE: 10.0},
	{name: "spy_marker", minPSNR: 28.0, maxMeanAbs: 12.0},
	{name: "spy_image", minPSNR: 27.0, maxMeanAbs: 22.0, maxRMSE: 30.0},
	{name: "polar_axes", minPSNR: 32.0, maxMeanAbs: 9.0},
	{name: "geo_mollweide_axes", minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "geo_aitoff_axes", minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "geo_hammer_axes", minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "geo_lambert_axes", minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "radar_basic", minPSNR: 45.0, maxMeanAbs: 2.0},
	{name: "skewt_basic", minPSNR: 24.0, maxMeanAbs: 18.0},
	{name: "mplot3d_basic", minPSNR: 39.0, maxMeanAbs: 5.0},
	{name: "mplot3d_terrain", minPSNR: 38.0, maxMeanAbs: 5.0},
	{name: "mplot3d_plot3d", minPSNR: 38.0, maxMeanAbs: 8.0},
	{name: "mplot3d_scatter3d", minPSNR: 35.0, maxMeanAbs: 8.0},
	{name: "mplot3d_surface3d", minPSNR: 35.0, maxMeanAbs: 10.0},
	{name: "mplot3d_wire3d", minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "mplot3d_trisurf3d", minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "mplot3d_bar3d", minPSNR: 30.0, maxMeanAbs: 8.0},
	{name: "mplot3d_voxels", minPSNR: 30.0, maxMeanAbs: 12.0},
	{name: "mplot3d_quiver3d", minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "mplot3d_stem3d", minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "mplot3d_fill_between3d", minPSNR: 35.0, maxMeanAbs: 10.0},
	{name: "unstructured_showcase", minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "arrays_showcase", minPSNR: 30.0, maxMeanAbs: 10.0},
	{name: "axisartist_showcase", minPSNR: 28.0, maxMeanAbs: 12.0},
	{name: "axes_grid1_showcase", minPSNR: 28.0, maxMeanAbs: 12.0},
	{name: "pcolor_flat", minPSNR: 28.0, maxMeanAbs: 15.0},
	{name: "pcolormesh_nearest", minPSNR: 28.0, maxMeanAbs: 15.0},
	{name: "pcolormesh_gouraud", minPSNR: 20.0, maxMeanAbs: 22.0, maxRMSE: 30.0},
	{name: "pcolormesh_masked", minPSNR: 28.0, maxMeanAbs: 15.0},
	{name: "hist2d_weighted_density", minPSNR: 28.0, maxMeanAbs: 16.0, maxRMSE: 30.0},
	{name: "boundarynorm_pcolormesh", minPSNR: 28.0, maxMeanAbs: 16.0},
	{name: "lognorm_imshow", minPSNR: 28.0, maxMeanAbs: 16.0},
	{name: "twoslope_norm_image", minPSNR: 28.0, maxMeanAbs: 16.0},
	{name: "colorbar_extensions", minPSNR: 28.0, maxMeanAbs: 16.0},
	{name: "large_scatter", minPSNR: 55.0, maxMeanAbs: 0.5, maxRMSE: 4.0},
	{name: "mixed_collection", minPSNR: 60.0, maxMeanAbs: 0.5, maxRMSE: 2.0},
	{name: "quad_mesh", minPSNR: 48.0, maxMeanAbs: 1.0, maxRMSE: 4.0},
	{name: "gouraud_triangles", minPSNR: 25.0, maxMeanAbs: 18.0},
	{name: "clip_path_batch", minPSNR: 45.0, maxMeanAbs: 1.0, maxRMSE: 6.0},
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
	got, _, err := parity.Render("colorbar_composition")
	if err != nil {
		t.Fatalf("render parity example colorbar_composition: %v", err)
	}
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

	if !rectWithinPixels(gotRect, wantRect, 2) {
		t.Fatalf("main image component bounds = %v, want within 2 px of matplotlib %v", gotRect, wantRect)
	}
}

func runReferenceCompareTest(t *testing.T, tc referenceCompareCase) {
	t.Helper()

	got, _, err := parity.Render(tc.name)
	if err != nil {
		t.Fatalf("render parity example %s: %v", tc.name, err)
	}

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

	artifactsDir := referenceCompareArtifactsDir(t)

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
		"%s: MaxDiff=%d MeanAbs=%.2f RMSE=%.2f PSNR=%.2fdB",
		tc.name,
		diff.MaxDiff,
		diff.MeanAbs,
		diff.RMSE,
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
	if diff.PSNR < minPSNR || diff.MeanAbs > maxMeanAbs || (tc.maxRMSE > 0 && diff.RMSE > tc.maxRMSE) {
		t.Fatalf(
			"reference mismatch for %s: PSNR=%.2f (min %.2f), MeanAbs=%.2f (max %.2f), RMSE=%.2f (max %.2f)",
			tc.name,
			diff.PSNR,
			minPSNR,
			diff.MeanAbs,
			maxMeanAbs,
			diff.RMSE,
			tc.maxRMSE,
		)
	}
}

func savePNGOrFail(t *testing.T, img image.Image, path string) {
	t.Helper()
	if err := imagecmp.SavePNG(img, path); err != nil {
		t.Fatalf("save PNG %s: %v", path, err)
	}
}

func referenceCompareArtifactsDir(t *testing.T) string {
	t.Helper()

	artifactsDir := filepath.Join("..", "testdata", "_artifacts", "reference_compare")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Logf("could not create reference artifacts directory %s: %v; using temp dir", artifactsDir, err)
		return t.TempDir()
	}

	probe, err := os.CreateTemp(artifactsDir, ".write-probe-*")
	if err != nil {
		t.Logf("reference artifacts directory %s is not writable: %v; using temp dir", artifactsDir, err)
		return t.TempDir()
	}
	probeName := probe.Name()
	if err := probe.Close(); err != nil {
		t.Logf("could not close write probe %s: %v; using temp dir", probeName, err)
		return t.TempDir()
	}
	if err := os.Remove(probeName); err != nil {
		t.Logf("could not remove write probe %s: %v", probeName, err)
	}
	return artifactsDir
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
