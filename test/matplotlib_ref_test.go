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
// This requires a Python interpreter with matplotlib installed. Prefer a system
// Python linked against system FreeType so reference glyph rasterization uses the
// same FreeType stack as the Go cgo backend.

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cwbudde/matplotlib-go/examples/parity"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
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

func matplotlibPythonPath() (string, error) {
	candidates := []string{}
	if env := os.Getenv("MATPLOTLIB_GO_PYTHON"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates, "/usr/bin/python3")
	if pyPath, err := exec.LookPath("python3"); err == nil {
		candidates = append(candidates, pyPath)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		cmd := exec.Command(candidate, "-c", "import matplotlib.ft2font")
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no Python interpreter with matplotlib found; set MATPLOTLIB_GO_PYTHON")
}

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

		pyPath, err := matplotlibPythonPath()
		if err != nil {
			mplErr = err
			return
		}
		cmd := exec.Command(pyPath, script, "--output-dir", outDir)
		if repoRoot, err := filepath.Abs(".."); err == nil {
			cmd.Env = append(os.Environ(), "PYTHONPATH="+prependEnvPath(repoRoot, os.Getenv("PYTHONPATH")))
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		mplErr = cmd.Run()
	})

	if mplErr != nil {
		t.Fatalf("matplotlib reference generation failed: %v", mplErr)
	}
}

func prependEnvPath(path, existing string) string {
	if existing == "" {
		return path
	}
	return path + string(os.PathListSeparator) + existing
}

// runMplTest renders our version, loads the matplotlib reference, and
// compares them. It always writes artefacts for visual inspection and fails
// only when PSNR drops below mplMinPSNR (indicating a structural error).
func runMplTest(t *testing.T, name string) {
	t.Helper()
	ensureRefs(t)

	got, _, err := parity.Render(name)
	if err != nil {
		t.Fatalf("render parity example %s: %v", name, err)
	}

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
	artifactsDir := matplotlibArtifactsDir(t)
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

func matplotlibArtifactsDir(t *testing.T) string {
	t.Helper()

	artifactsDir := filepath.Join("..", "testdata", "_artifacts", "matplotlib_ref")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Logf("could not create artifacts directory %s: %v; using temp dir", artifactsDir, err)
		return t.TempDir()
	}
	probe, err := os.CreateTemp(artifactsDir, ".write-probe-*")
	if err != nil {
		t.Logf("artifacts directory %s is not writable: %v; using temp dir", artifactsDir, err)
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

// Individual tests
// Each test renders the canonical example from examples/parity.

func TestMpl_BasicLine(t *testing.T)    { runMplTest(t, "basic_line") }
func TestMpl_JoinsCaps(t *testing.T)    { runMplTest(t, "joins_caps") }
func TestMpl_Dashes(t *testing.T)       { runMplTest(t, "dashes") }
func TestMpl_ScatterBasic(t *testing.T) { runMplTest(t, "scatter_basic") }
func TestMpl_ScatterMarkerTypes(t *testing.T) {
	runMplTest(t, "scatter_marker_types")
}
func TestMpl_ScatterAdvanced(t *testing.T) { runMplTest(t, "scatter_advanced") }
func TestMpl_BarBasicFrame(t *testing.T)   { runMplTest(t, "bar_basic_frame") }
func TestMpl_BarBasicTicks(t *testing.T)   { runMplTest(t, "bar_basic_ticks") }
func TestMpl_BarBasicTickLabels(t *testing.T) {
	runMplTest(t, "bar_basic_tick_labels")
}
func TestMpl_BarBasicTitle(t *testing.T) { runMplTest(t, "bar_basic_title") }
func TestMpl_BarBasic(t *testing.T)      { runMplTest(t, "bar_basic") }
func TestMpl_BarHorizontal(t *testing.T) { runMplTest(t, "bar_horizontal") }
func TestMpl_BarGrouped(t *testing.T)    { runMplTest(t, "bar_grouped") }
func TestMpl_FillBasic(t *testing.T)     { runMplTest(t, "fill_basic") }
func TestMpl_FillBetween(t *testing.T)   { runMplTest(t, "fill_between") }
func TestMpl_FillStacked(t *testing.T)   { runMplTest(t, "fill_stacked") }
func TestMpl_MultiSeriesBasic(t *testing.T) {
	runMplTest(t, "multi_series_basic")
}

func TestMpl_MultiSeriesColorCycle(t *testing.T) {
	runMplTest(t, "multi_series_color_cycle")
}

func TestMpl_HistBasic(t *testing.T)      { runMplTest(t, "hist_basic") }
func TestMpl_HistDensity(t *testing.T)    { runMplTest(t, "hist_density") }
func TestMpl_HistStrategies(t *testing.T) { runMplTest(t, "hist_strategies") }
func TestMpl_BoxPlotBasic(t *testing.T)   { runMplTest(t, "boxplot_basic") }
func TestMpl_Phase12SpecialtyDepth(t *testing.T) {
	runMplTest(t, "phase12_specialty_depth")
}
func TestMpl_AxesTopRightInverted(t *testing.T) {
	runMplTest(t, "axes_top_right_inverted")
}

func TestMpl_AxesControlSurface(t *testing.T) {
	runMplTest(t, "axes_control_surface")
}

func TestMpl_TransformCoordinates(t *testing.T) {
	runMplTest(t, "transform_coordinates")
}

func TestMpl_GridSpecComposition(t *testing.T) {
	runMplTest(t, "gridspec_composition")
}

func TestMpl_FigureLabelsComposition(t *testing.T) {
	runMplTest(t, "figure_labels_composition")
}

func TestMpl_ColorbarComposition(t *testing.T) {
	runMplTest(t, "colorbar_composition")
}

func TestMpl_AnnotationComposition(t *testing.T) {
	runMplTest(t, "annotation_composition")
}
func TestMpl_PatchShowcase(t *testing.T)  { runMplTest(t, "patch_showcase") }
func TestMpl_MeshContourTri(t *testing.T) { runMplTest(t, "mesh_contour_tri") }
func TestMpl_PlotVariants(t *testing.T)   { runMplTest(t, "plot_variants") }
func TestMpl_SpectrumVariants(t *testing.T) {
	runMplTest(t, "spectrum_variants")
}
func TestMpl_StatVariants(t *testing.T) { runMplTest(t, "stat_variants") }
func TestMpl_StemPlot(t *testing.T)     { runMplTest(t, "stem_plot") }
func TestMpl_SpecialtyArtists(t *testing.T) {
	runMplTest(t, "specialty_artists")
}
func TestMpl_UnitsOverview(t *testing.T) { runMplTest(t, "units_overview") }
func TestMpl_VectorFields(t *testing.T)  { runMplTest(t, "vector_fields") }
func TestMpl_PolarAxes(t *testing.T)     { runMplTest(t, "polar_axes") }
func TestMpl_UnstructuredShowcase(t *testing.T) {
	runMplTest(t, "unstructured_showcase")
}

func TestMpl_ArraysShowcase(t *testing.T) {
	runMplTest(t, "arrays_showcase")
}

func TestMpl_AxisArtistShowcase(t *testing.T) {
	runMplTest(t, "axisartist_showcase")
}

func TestMpl_AxesGrid1Showcase(t *testing.T) {
	runMplTest(t, "axes_grid1_showcase")
}

func TestMpl_PColorFlat(t *testing.T) {
	runMplTest(t, "pcolor_flat")
}

func TestMpl_PColorMeshNearest(t *testing.T) {
	runMplTest(t, "pcolormesh_nearest")
}

func TestMpl_PColorMeshGouraud(t *testing.T) {
	runMplTest(t, "pcolormesh_gouraud")
}

func TestMpl_PColorMeshMasked(t *testing.T) {
	runMplTest(t, "pcolormesh_masked")
}

func TestMpl_Hist2DWeightedDensity(t *testing.T) {
	runMplTest(t, "hist2d_weighted_density")
}

func TestMpl_BoundaryNormPColorMesh(t *testing.T) {
	runMplTest(t, "boundarynorm_pcolormesh")
}

func TestMpl_LogNormImshow(t *testing.T) {
	runMplTest(t, "lognorm_imshow")
}

func TestMpl_TwoSlopeNormImage(t *testing.T) {
	runMplTest(t, "twoslope_norm_image")
}

func TestMpl_ColorbarExtensions(t *testing.T) {
	runMplTest(t, "colorbar_extensions")
}

func TestMpl_LargeScatter(t *testing.T) {
	runMplTest(t, "large_scatter")
}

func TestMpl_MixedCollection(t *testing.T) {
	runMplTest(t, "mixed_collection")
}

func TestMpl_QuadMesh(t *testing.T) {
	runMplTest(t, "quad_mesh")
}

func TestMpl_GouraudTriangles(t *testing.T) {
	runMplTest(t, "gouraud_triangles")
}

func TestMpl_ClipPathBatch(t *testing.T) {
	runMplTest(t, "clip_path_batch")
}

func TestMpl_ErrorBars(t *testing.T)    { runMplTest(t, "errorbar_basic") }
func TestMpl_TitleStrict(t *testing.T)  { runMplTest(t, "title_strict") }
func TestMpl_ImageHeatmap(t *testing.T) { runMplTest(t, "image_heatmap") }
func TestMpl_ImshowClipped(t *testing.T) {
	runMplTest(t, "imshow_clipped")
}
func TestMpl_ImshowTransformed(t *testing.T) {
	runMplTest(t, "imshow_transformed")
}
func TestMpl_ImageAlpha(t *testing.T)   { runMplTest(t, "image_alpha") }
func TestMpl_MatshowBasic(t *testing.T) { runMplTest(t, "matshow_basic") }
func TestMpl_SpyMarker(t *testing.T)    { runMplTest(t, "spy_marker") }
func TestMpl_SpyImage(t *testing.T)     { runMplTest(t, "spy_image") }

func TestMpl_Mplot3DPlot(t *testing.T) {
	runMplTest(t, "mplot3d_plot3d")
}

func TestMpl_Mplot3DScatter(t *testing.T) {
	runMplTest(t, "mplot3d_scatter3d")
}

func TestMpl_Mplot3DSurface(t *testing.T) {
	runMplTest(t, "mplot3d_surface3d")
}

func TestMpl_Mplot3DWire(t *testing.T) {
	runMplTest(t, "mplot3d_wire3d")
}

func TestMpl_Mplot3DTrisurf(t *testing.T) {
	runMplTest(t, "mplot3d_trisurf3d")
}

func TestMpl_Mplot3DBar3d(t *testing.T) {
	runMplTest(t, "mplot3d_bar3d")
}

func TestMpl_Mplot3DVoxels(t *testing.T) {
	runMplTest(t, "mplot3d_voxels")
}

func TestMpl_Mplot3DQuiver(t *testing.T) {
	runMplTest(t, "mplot3d_quiver3d")
}

func TestMpl_Mplot3DStem(t *testing.T) {
	runMplTest(t, "mplot3d_stem3d")
}

func TestMpl_Mplot3DFillBetween(t *testing.T) {
	runMplTest(t, "mplot3d_fill_between3d")
}
