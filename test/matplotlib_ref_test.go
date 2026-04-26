package test

// Matplotlib reference tests compare our renderer against real matplotlib output.
//
// Reference images live in testdata/matplotlib_ref/ and are committed to the repo.
// The skip path is only a local bootstrap fallback when refs are intentionally absent.
//
// To (re)generate the reference images:
//
//	go test ./test/... -run TestMpl -update-matplotlib
//
// This requires either `uv` (recommended) or `python3` with matplotlib installed.
// uv will automatically fetch matplotlib into a temporary virtual environment.

import (
	"flag"
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"matplotlib-go/test/imagecmp"
)

var updateMatplotlib = flag.Bool("update-matplotlib", false,
	"Regenerate matplotlib reference images via `uv run` (or python3) before comparing")

const (
	mplRefDir  = "../testdata/matplotlib_ref"
	mplMinPSNR = 10.0 // dB; below this indicates a fundamental rendering error
)

var (
	mplOnce sync.Once
	mplErr  error
)

// ensureRefs generates reference images when -update-matplotlib is set,
// or skips the calling test if the committed ref directory is missing locally.
func ensureRefs(t *testing.T) {
	t.Helper()
	if !*updateMatplotlib {
		if _, err := os.Stat(mplRefDir); os.IsNotExist(err) {
			t.Skip("matplotlib refs not found – run with -update-matplotlib to generate")
		}
		return
	}

	mplOnce.Do(func() {
		script := filepath.Join("matplotlib_ref", "generate.py")
		outDir, _ := filepath.Abs(mplRefDir)

		uvPath, err := exec.LookPath("uv")
		var cmd *exec.Cmd
		if err == nil {
			cmd = exec.Command(uvPath, "run", script, "--output-dir", outDir)
		} else {
			pyPath, err2 := exec.LookPath("python3")
			if err2 != nil {
				mplErr = fmt.Errorf("need uv or python3: uv: %v; python3: %v", err, err2)
				return
			}
			cmd = exec.Command(pyPath, script, "--output-dir", outDir)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		mplErr = cmd.Run()
	})

	if mplErr != nil {
		t.Fatalf("matplotlib reference generation failed: %v", mplErr)
	}
}

// runMplTest renders our version, loads the matplotlib reference, and
// compares them. It always writes artefacts for visual inspection and fails
// only when PSNR drops below mplMinPSNR (indicating a structural error).
func runMplTest(t *testing.T, name string, renderFunc func() image.Image) {
	t.Helper()
	ensureRefs(t)

	got := renderFunc()

	refPath := filepath.Join(mplRefDir, name+".png")
	want, err := imagecmp.LoadPNG(refPath)
	if err != nil {
		t.Fatalf("load matplotlib ref %s: %v", refPath, err)
	}

	// tolerance=255 so Identical is always true; we rely on PSNR for the verdict.
	diff, err := imagecmp.ComparePNG(got, want, 255)
	if err != nil {
		t.Fatalf("image comparison failed: %v", err)
	}

	t.Logf("PSNR=%.1f dB  MeanAbs=%.2f  MaxDiff=%d", diff.PSNR, diff.MeanAbs, diff.MaxDiff)

	// Always write artefacts so humans can inspect visually.
	artifactsDir := filepath.Join("..", "testdata", "_artifacts", "matplotlib_ref")
	if mkErr := os.MkdirAll(artifactsDir, 0o755); mkErr != nil {
		t.Fatalf("could not create artifacts directory %s: %v", artifactsDir, mkErr)
	}
	if err := imagecmp.SavePNG(got, filepath.Join(artifactsDir, name+"_go.png")); err != nil {
		t.Fatalf("could not save rendered image: %v", err)
	}
	if err := imagecmp.SavePNG(want, filepath.Join(artifactsDir, name+"_mpl.png")); err != nil {
		t.Fatalf("could not save matplotlib image: %v", err)
	}
	if err := imagecmp.SaveDiffImage(got, want, 10, filepath.Join(artifactsDir, name+"_diff.png")); err != nil {
		t.Fatalf("could not save diff image: %v", err)
	}

	if !math.IsInf(diff.PSNR, 1) && diff.PSNR < mplMinPSNR {
		t.Errorf("PSNR %.1f dB < %.1f dB: likely fundamental rendering mismatch with matplotlib",
			diff.PSNR, mplMinPSNR)
	}
}

// ─── Individual tests ────────────────────────────────────────────────────────
// Each test reuses the render* helper from golden_test.go.

func TestMpl_BasicLine(t *testing.T)    { runMplTest(t, "basic_line", renderBasicLine) }
func TestMpl_JoinsCaps(t *testing.T)    { runMplTest(t, "joins_caps", renderJoinsCaps) }
func TestMpl_Dashes(t *testing.T)       { runMplTest(t, "dashes", renderDashes) }
func TestMpl_ScatterBasic(t *testing.T) { runMplTest(t, "scatter_basic", renderScatterBasic) }
func TestMpl_ScatterMarkerTypes(t *testing.T) {
	runMplTest(t, "scatter_marker_types", renderScatterMarkerTypes)
}
func TestMpl_ScatterAdvanced(t *testing.T) { runMplTest(t, "scatter_advanced", renderScatterAdvanced) }
func TestMpl_BarBasicFrame(t *testing.T)   { runMplTest(t, "bar_basic_frame", renderBarBasicFrame) }
func TestMpl_BarBasicTicks(t *testing.T)   { runMplTest(t, "bar_basic_ticks", renderBarBasicTicks) }
func TestMpl_BarBasicTickLabels(t *testing.T) {
	runMplTest(t, "bar_basic_tick_labels", renderBarBasicTickLabels)
}
func TestMpl_BarBasicTitle(t *testing.T) { runMplTest(t, "bar_basic_title", renderBarBasicTitle) }
func TestMpl_BarBasic(t *testing.T)      { runMplTest(t, "bar_basic", renderBarBasic) }
func TestMpl_BarHorizontal(t *testing.T) { runMplTest(t, "bar_horizontal", renderBarHorizontal) }
func TestMpl_BarGrouped(t *testing.T)    { runMplTest(t, "bar_grouped", renderBarGrouped) }
func TestMpl_FillBasic(t *testing.T)     { runMplTest(t, "fill_basic", renderFillBasic) }
func TestMpl_FillBetween(t *testing.T)   { runMplTest(t, "fill_between", renderFillBetween) }
func TestMpl_FillStacked(t *testing.T)   { runMplTest(t, "fill_stacked", renderFillStacked) }
func TestMpl_MultiSeriesBasic(t *testing.T) {
	runMplTest(t, "multi_series_basic", renderMultiSeriesBasic)
}

func TestMpl_MultiSeriesColorCycle(t *testing.T) {
	runMplTest(t, "multi_series_color_cycle", renderMultiSeriesColorCycle)
}

func TestMpl_HistBasic(t *testing.T)      { runMplTest(t, "hist_basic", renderHistBasic) }
func TestMpl_HistDensity(t *testing.T)    { runMplTest(t, "hist_density", renderHistDensity) }
func TestMpl_HistStrategies(t *testing.T) { runMplTest(t, "hist_strategies", renderHistStrategies) }
func TestMpl_BoxPlotBasic(t *testing.T)   { runMplTest(t, "boxplot_basic", renderBoxPlotBasic) }
func TestMpl_AxesTopRightInverted(t *testing.T) {
	runMplTest(t, "axes_top_right_inverted", renderAxesTopRightInverted)
}
func TestMpl_AxesControlSurface(t *testing.T) {
	runMplTest(t, "axes_control_surface", renderAxesControlSurface)
}
func TestMpl_TransformCoordinates(t *testing.T) {
	runMplTest(t, "transform_coordinates", renderTransformCoordinates)
}
func TestMpl_GridSpecComposition(t *testing.T) {
	runMplTest(t, "gridspec_composition", renderGridSpecComposition)
}
func TestMpl_FigureLabelsComposition(t *testing.T) {
	runMplTest(t, "figure_labels_composition", renderFigureLabelsComposition)
}
func TestMpl_ColorbarComposition(t *testing.T) {
	runMplTest(t, "colorbar_composition", renderColorbarComposition)
}
func TestMpl_AnnotationComposition(t *testing.T) {
	runMplTest(t, "annotation_composition", renderAnnotationComposition)
}
func TestMpl_PatchShowcase(t *testing.T)  { runMplTest(t, "patch_showcase", renderPatchShowcase) }
func TestMpl_MeshContourTri(t *testing.T) { runMplTest(t, "mesh_contour_tri", renderMeshContourTri) }
func TestMpl_PlotVariants(t *testing.T)   { runMplTest(t, "plot_variants", renderPlotVariants) }
func TestMpl_StatVariants(t *testing.T)   { runMplTest(t, "stat_variants", renderStatVariants) }
func TestMpl_StemPlot(t *testing.T)       { runMplTest(t, "stem_plot", renderStemPlot) }
func TestMpl_SpecialtyArtists(t *testing.T) {
	runMplTest(t, "specialty_artists", renderSpecialtyArtists)
}
func TestMpl_UnitsOverview(t *testing.T) { runMplTest(t, "units_overview", renderUnitsOverview) }
func TestMpl_VectorFields(t *testing.T)  { runMplTest(t, "vector_fields", renderVectorFields) }
func TestMpl_PolarAxes(t *testing.T)     { runMplTest(t, "polar_axes", renderPolarAxes) }
func TestMpl_UnstructuredShowcase(t *testing.T) {
	runMplTest(t, "unstructured_showcase", renderUnstructuredShowcase)
}
func TestMpl_ArraysShowcase(t *testing.T) {
	runMplTest(t, "arrays_showcase", renderArraysShowcase)
}
func TestMpl_ErrorBars(t *testing.T)     { runMplTest(t, "errorbar_basic", renderErrorBars) }
func TestMpl_TitleStrict(t *testing.T)   { runMplTest(t, "title_strict", renderTitleStrict) }
func TestMpl_ImageHeatmap(t *testing.T)  { runMplTest(t, "image_heatmap", renderImageHeatmap) }
