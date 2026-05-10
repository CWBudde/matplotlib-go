package test

import (
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/test/parity"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
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

func referenceDateNumber(t time.Time) float64 {
	t = t.UTC()
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

func referencePointsToPixels(points float64) float64 {
	return points * style.Default.DPI / 72.0
}

func TestBasicLine_Golden(t *testing.T) {
	runGoldenTest(t, "basic_line")
}

func TestJoinsCaps_Golden(t *testing.T) {
	runGoldenTest(t, "joins_caps")
}

func TestDashes_Golden(t *testing.T) {
	runGoldenTest(t, "dashes")
}

func TestScatterBasic_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_basic")
}

func TestScatterMarkerTypes_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_marker_types")
}

func TestScatterAdvanced_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_advanced")
}

func TestBarBasicFrame_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_frame")
}

func TestBarBasicTicks_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_ticks")
}

func TestBarBasicTickLabels_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_tick_labels")
}

func TestBarBasicTitle_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic_title")
}

func TestBarBasic_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic")
}

func TestBarHorizontal_Golden(t *testing.T) {
	runGoldenTest(t, "bar_horizontal")
}

func TestBarGrouped_Golden(t *testing.T) {
	runGoldenTest(t, "bar_grouped")
}

func TestFillBasic_Golden(t *testing.T) {
	runGoldenTest(t, "fill_basic")
}

func TestFillBetween_Golden(t *testing.T) {
	runGoldenTest(t, "fill_between")
}

func TestFillStacked_Golden(t *testing.T) {
	runGoldenTest(t, "fill_stacked")
}

func TestMultiSeriesBasic_Golden(t *testing.T) {
	runGoldenTest(t, "multi_series_basic")
}

func TestMultiSeriesColorCycle_Golden(t *testing.T) {
	runGoldenTest(t, "multi_series_color_cycle")
}

func TestHistBasic_Golden(t *testing.T) {
	runGoldenTest(t, "hist_basic")
}

func TestHistDensity_Golden(t *testing.T) {
	runGoldenTest(t, "hist_density")
}

func TestHistStrategies_Golden(t *testing.T) {
	runGoldenTest(t, "hist_strategies")
}

func TestBoxPlotBasic_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "boxplot_basic")
}

func TestPhase12SpecialtyDepth_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "phase12_specialty_depth")
}

func TestErrorBars_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "errorbar_basic")
}

func TestTextLabelsStrict_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "text_labels_strict")
}

func TestTitleStrict_Golden(t *testing.T) {
	runGoldenTest(t, "title_strict")
}

func TestImageHeatmap_Golden(t *testing.T) {
	runGoldenTest(t, "image_heatmap")
}

func TestAxesTopRightInverted_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_top_right_inverted")
}

func TestAxesControlSurface_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_control_surface")
}

func TestTransformCoordinates_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "transform_coordinates")
}

func TestGridSpecComposition_Golden(t *testing.T) {
	runGoldenTest(t, "gridspec_composition")
}

func TestFigureLabelsComposition_Golden(t *testing.T) {
	runGoldenTest(t, "figure_labels_composition")
}

func TestColorbarComposition_Golden(t *testing.T) {
	runGoldenTest(t, "colorbar_composition")
}

func TestAnnotationComposition_Golden(t *testing.T) {
	runGoldenTest(t, "annotation_composition")
}

func TestPlotVariants_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "plot_variants")
}

func TestSpectrumVariants_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "spectrum_variants")
}

func TestStatVariants_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "stat_variants")
}

func TestUnitsOverview_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_overview")
}

func TestUnitsDates_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_dates")
}

func TestUnitsCategories_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_categories")
}

func TestUnitsCustomConverter_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "units_custom_converter")
}

func TestPatchShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "patch_showcase")
}

func TestMeshContourTri_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mesh_contour_tri")
}

func TestStemPlot_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "stem_plot")
}

func TestSpecialtyArtists_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "specialty_artists")
}

func TestVectorFields_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "vector_fields")
}

func TestPolarAxes_Golden(t *testing.T) {
	runGoldenTest(t, "polar_axes")
}

func TestGeoMollweideAxes_Golden(t *testing.T) {
	runGoldenTest(t, "geo_mollweide_axes")
}

func TestGeoAitoffAxes_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "geo_aitoff_axes")
}

func TestGeoHammerAxes_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "geo_hammer_axes")
}

func TestGeoLambertAxes_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "geo_lambert_axes")
}

func TestRadarBasic_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "radar_basic")
}

func TestSkewTBasic_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "skewt_basic")
}

func TestMplot3DBasic_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_basic")
}

func TestMplot3DTerrain_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_terrain")
}

func TestMplot3DPlot_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_plot3d")
}

func TestMplot3DScatter_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_scatter3d")
}

func TestMplot3DSurface_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_surface3d")
}

func TestMplot3DWire_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_wire3d")
}

func TestMplot3DTrisurf_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_trisurf3d")
}

func TestMplot3DBar3d_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_bar3d")
}

func TestMplot3DVoxels_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_voxels")
}

func TestMplot3DQuiver_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_quiver3d")
}

func TestMplot3DStem_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_stem3d")
}

func TestMplot3DFillBetween_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "mplot3d_fill_between3d")
}

// runGoldenTest is a helper function for golden image testing.
func runGoldenTest(t *testing.T, testName string) {
	// Render the plot
	img, _, err := parity.Render(testName)
	if err != nil {
		t.Fatalf("Failed to render parity example %s: %v", testName, err)
	}

	goldenPath := "../testdata/" + goldenDirName() + "/" + testName + ".png"

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

// renderJoinsCaps creates a plot demonstrating different line joins and caps

// renderDashes creates a plot demonstrating dash patterns

// renderScatterBasic creates a basic scatter plot for golden testing

// renderScatterMarkerTypes creates a plot showing all marker types

// renderScatterAdvanced creates an advanced scatter plot with edges, alpha, and variable sizes

// renderBarBasic creates a basic vertical bar chart for golden testing

// renderBarHorizontal creates a horizontal bar chart for golden testing

// renderBarGrouped creates a grouped bar chart with variable colors and edges

// renderFillBasic creates a basic fill to baseline for golden testing

// renderFillBetween creates a fill between two curves for golden testing

// renderFillStacked creates a stacked area chart for golden testing

// renderMultiSeriesBasic creates a plot with multiple series using different plot types

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

// renderHistDensity creates a density-normalized histogram for golden testing.

// renderHistStrategies creates a histogram comparing two datasets for golden testing.

// renderBoxPlotBasic creates a multi-series box plot for golden testing.

// renderMultiSeriesColorCycle creates a plot demonstrating automatic color cycling

func applyMatplotlibGridSpecStyle(fig *core.Figure) {
	fig.RC = style.Apply(fig.RC, style.WithFont("DejaVu Sans", 10))
	fig.RC.TitleFontSize = 12
	fig.RC.AxisLabelFontSize = 10
	fig.RC.XTickLabelFontSize = 10
	fig.RC.YTickLabelFontSize = 10
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

func addReferenceYGrid(ax *core.Axes) {
	grid := ax.AddGrid(core.AxisLeft)
	grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	grid.LineWidth = 0.5
}

func addReferenceXYGrid(ax *core.Axes) {
	xGrid := ax.AddGrid(core.AxisBottom)
	xGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	xGrid.LineWidth = 0.5
	addReferenceYGrid(ax)
}

func lambertLongitudeTicks() []float64 {
	ticks := make([]float64, 0, 9)
	for deg := -120; deg <= 120; deg += 30 {
		ticks = append(ticks, float64(deg)*math.Pi/180)
	}
	return ticks
}

func gridMin(values [][]float64) float64 {
	minValue := math.Inf(1)
	for _, row := range values {
		for _, value := range row {
			if value < minValue {
				minValue = value
			}
		}
	}
	return minValue
}

func sinusoidalTerrain(xCount, yCount int) ([]float64, []float64, [][]float64) {
	if xCount < 2 {
		xCount = 2
	}
	if yCount < 2 {
		yCount = 2
	}
	x := make([]float64, xCount)
	y := make([]float64, yCount)
	z := make([][]float64, yCount)

	for yi := 0; yi < yCount; yi++ {
		y[yi] = -math.Pi + 2*math.Pi*float64(yi)/float64(yCount-1)
	}
	for xi := 0; xi < xCount; xi++ {
		x[xi] = -math.Pi + 2*math.Pi*float64(xi)/float64(xCount-1)
	}
	for yi := 0; yi < yCount; yi++ {
		row := make([]float64, xCount)
		for xi := 0; xi < xCount; xi++ {
			angleX := x[xi]
			angleY := y[yi]
			row[xi] = 0.5*math.Sin(angleX)*math.Cos(angleY) +
				0.35*math.Sin(2*angleX+0.6)*math.Cos(angleY/2) +
				0.15*math.Cos(3*angleY-angleX)
		}
		z[yi] = row
	}
	return x, y, z
}

func disableMplot3DTickLabels(ax *core.Axes3D) {
	if ax == nil {
		return
	}
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowLabels = false
	ax.SetShowZTickLabels(false)
}

func get3DWireframeTestData(delta float64) (x []float64, y []float64, z [][]float64) {
	if delta <= 0 {
		delta = 0.05
	}

	x = append(x, -3.0)
	for x[len(x)-1] < 3.0-delta {
		x = append(x, x[len(x)-1]+delta)
	}
	y = make([]float64, len(x))
	copy(y, x)

	yCount := len(y)
	xCount := len(x)
	z = make([][]float64, yCount)
	for yi := range y {
		yRow := make([]float64, xCount)
		for xi := range x {
			gridX := x[xi] * 10
			gridY := y[yi] * 10
			z1 := math.Exp(-(gridX*gridX+gridY*gridY)/2) / (2 * math.Pi)
			z2 := (math.Exp(-(((gridX-1)/1.5)*((gridX-1)/1.5)+((gridY-1)/0.5)*((gridY-1)/0.5))/2) / (2 * math.Pi * 0.5 * 1.5))
			yRow[xi] = (z2 - z1) * 500
		}
		z[yi] = yRow
	}

	return x, y, z
}

func mplot3DScatter3DPoints() ([]float64, []float64, []float64) {
	x := []float64{
		23.445374307849796, 24.12194086355045, 23.819529536124904, 24.00420727443508, 25.163187193449946, 24.295416940034634, 30.91130691389642, 24.605786721711933, 27.21139885293976, 27.986781163292047, 31.89192226239924, 26.298999938978977, 29.635299649362583, 23.906336659263545, 25.095071084764612, 29.459978845052547, 24.193875247596218, 25.542156954393818, 24.032675548990003, 30.46612707082184, 28.34036606828759, 23.279382886171128, 24.945607674429596, 31.87570266522563, 29.517913381654594, 27.868155221203335, 28.397102977106236, 23.644471484021, 23.882053814715775, 24.52445349928731, 23.212418078695983, 26.639680340341172, 29.83475945947286, 30.603547876550454, 26.230654568636293, 27.56211058454415, 29.847148543854843, 26.965685062731566, 29.482291156335634, 30.11650737649944, 28.517902593942317, 23.530972320001112, 23.225112299107412, 27.899278440398664, 30.46826279097253, 28.31231362198466, 28.577546999668883, 30.906285221181864, 28.346559611236383, 25.412836408863605, 31.02502964640599, 25.838621147007252, 30.2695752017394, 27.259331993205212, 28.44442014010872, 24.832138051438253, 24.772237782626977, 31.13679360966941, 24.063084952446, 30.71782734066047, 24.1020675998253, 26.407089598800855, 26.073524740342314, 28.247337040080218, 28.52664630941329, 29.703277253079385, 23.51766543170498, 26.84174807837552, 31.89768786849023, 30.269704026803602, 25.822943366549396, 26.05741098865235, 27.455351420313217, 28.205788288765213, 27.575667962815753, 31.17131795801531, 31.724537257046894, 31.31136870635791, 24.864967466347135, 27.003051120655385, 25.081221861565915, 27.242215510078502, 29.246630208236752, 23.32681288685234, 24.695596557933705, 26.754750582763073, 30.46128890826422, 27.67783726413593, 31.376752583189585, 23.683526640813607, 29.84172357895877, 25.644964523091826, 23.0707980857386, 25.95315705307221, 25.27885172160532, 30.92098933398311, 30.909130791941855, 26.27857489720248, 29.77550802400433, 26.740355393097172,
	}
	y := []float64{
		89.61883854011774, 96.77743632842744, 28.157105387783755, 60.011862983483155, 50.27973260563567, 59.58611289697421, 68.26196589481192, 66.73500607782337, 67.89520298909866, 87.78005909692455, 55.90228301720692, 32.88014938342748, 89.97415203662578, 71.68172364424767, 81.61739008854082, 31.120813187076656, 72.91709365732572, 91.62415061387192, 51.5130996240686, 76.2690864324446, 57.93388477056373, 73.60353427143525, 57.060535302854596, 36.12824011228716, 29.977493234592444, 52.656275264665375, 32.52641907894169, 82.27351993645613, 10.757834757137019, 41.588172080714834, 50.65257873376535, 80.23210616089744, 45.834810172011885, 99.46739050855098, 98.16767381127598, 26.59797987212249, 57.81523756904051, 60.599446993554594, 78.80749105261128, 24.67844325571106, 99.9542535613643, 64.00386683404848, 55.01621741204809, 7.066634165562524, 77.68317602155433, 86.89455412309032, 85.51673027005023, 91.69055695060099, 5.877318874843295, 38.42391319901706, 42.44900889766391, 42.512769237215394, 9.168219240795361, 70.74011894662621, 62.81565474918824, 84.91864535059966, 84.35847524321538, 10.67104149274558, 35.98257538341652, 56.20354307974449, 29.226815795480366, 88.2185386315846, 19.667559587893834, 88.62367632848353, 33.64564571287989, 7.303882324772459, 73.13000138787797, 42.02603513269971, 65.97217312105327, 75.01176572737438, 54.29507834568107, 63.788472548222266, 78.8522131219556, 34.18898303269541, 63.7342289526038, 30.361940713543834, 60.09161357327015, 17.099846839943964, 59.05087945692124, 55.18103114595721, 83.46490643618304, 61.61212648243952, 29.62479879099057, 17.41752052501766, 38.960031151814114, 68.6147244267844, 20.755361547523275, 40.362009969485314, 58.28602545821371, 35.79089514400056, 22.10473173821892, 36.88407117873733, 49.02305607083998, 73.34531482348302, 41.72325076515475, 22.163067859293196, 65.62211504414263, 99.55079489898242, 45.199920930142426, 84.50320262512236,
	}
	z := []float64{
		-48.4767154236838, -33.586641900572545, -44.00071235430737, -39.70665691078236, -44.92804393396085, -34.233123745550806, -29.072855638585203, -43.565752277569224, -48.01789653521801, -33.1518337197717, -37.644732126268664, -48.126192981827806, -36.32252360799759, -43.84079453378656, -49.38959837544273, -34.08125115264693, -30.984893040214764, -33.25045160725494, -36.99164596623909, -28.85450939811091, -27.481038624211397, -41.15948666701077, -42.04208820195398, -39.612953968569244, -47.45528371165212, -27.69983016083391, -46.2879923393594, -25.301674620928914, -38.373599258385184, -32.10817063472631, -39.03817647462475, -29.503394969218473, -44.66257359710358, -48.236771520787215, -29.884421768629725, -34.7040098468824, -44.381150826984005, -38.93777347779819, -40.99598063080647, -44.189033637840744, -40.43370686189671, -37.84281069930992, -27.08352158896705, -36.25839112345035, -37.594213588818306, -44.66379905948732, -34.616967173476034, -47.02981531051485, -35.77604595776252, -35.362807195488614, -25.656063886969623, -34.369380733795666, -43.96613722009599, -49.78135322953642, -30.444420990090908, -41.762454148965645, -46.53133316451258, -41.74519181502462, -34.65825654681202, -31.98659208375311, -39.461768207302484, -39.461775075012085, -41.26997820699918, -35.15274151752209, -40.236873144208275, -39.214116932138914, -49.84312736772681, -45.71654949394924, -38.102994824254864, -42.32542932869035, -37.40241622996856, -26.028541012185865, -34.90855626044588, -42.08248043989314, -38.90362923447811, -31.00021948118308, -34.511007438650196, -25.744016891437898, -30.800657255504547, -42.90583047337795, -45.619698983737834, -45.7861265337756, -38.56953877824531, -25.225706120348136, -45.41004740011793, -30.10574251394272, -32.48594999403221, -34.00087799764117, -34.62955453447974, -42.496063305947544, -36.252056904967205, -28.086508009869927, -47.367499252184246, -31.24941106150154, -42.487837438082316, -42.2677745058031, -45.14083423306369, -27.389198544705586, -32.928193861607255, -35.50995788997248,
	}
	return x, y, z
}

func minInSlice(values []float64) float64 {
	minValue := math.Inf(1)
	for _, value := range values {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

func boolPtr(v bool) *bool {
	return &v
}

func minInGrid(values [][]float64) float64 {
	minValue := math.Inf(1)
	for _, row := range values {
		for _, value := range row {
			if value < minValue {
				minValue = value
			}
		}
	}
	return minValue
}

func floatPtr(v float64) *float64 {
	return &v
}
