package pyplot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"matplotlib-go/canvas"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/style"
)

func TestFigureRegistryTracksCurrentFigureAndAxes(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	if got := GCF(); got != fig1 {
		t.Fatalf("GCF() = %p, want %p", got, fig1)
	}

	ax1 := GCA()
	if ax1 == nil {
		t.Fatal("GCA() returned nil")
	}
	if len(fig1.Children) != 1 {
		t.Fatalf("len(fig1.Children) = %d, want 1", len(fig1.Children))
	}

	fig2 := FigureSized(900, 700)
	if got := GCF(); got != fig2 {
		t.Fatalf("after FigureSized, GCF() = %p, want %p", got, fig2)
	}
	if fig2.SizePx.X != 900 || fig2.SizePx.Y != 700 {
		t.Fatalf("FigureSized dimensions = %.0fx%.0f, want 900x700", fig2.SizePx.X, fig2.SizePx.Y)
	}
}

func TestSubplotReusesAxesForSameSlot(t *testing.T) {
	resetForTests()

	fig := Figure()
	ax1 := Subplot(2, 2, 3)
	ax2 := Subplot(2, 2, 3)
	if ax1 == nil || ax2 == nil {
		t.Fatal("Subplot returned nil axes")
	}
	if ax1 != ax2 {
		t.Fatalf("Subplot did not reuse axes: %p != %p", ax1, ax2)
	}
	if got := len(fig.Children); got != 1 {
		t.Fatalf("len(fig.Children) = %d, want 1", got)
	}
	if got := GCA(); got != ax1 {
		t.Fatalf("GCA() = %p, want %p", got, ax1)
	}
}

func TestSubplotsCreatesNewFigureAndCurrentAxes(t *testing.T) {
	resetForTests()

	fig, grid := Subplots(2, 2, core.WithSubplotShareX())
	if fig == nil {
		t.Fatal("Subplots returned nil figure")
	}
	if len(grid) != 2 || len(grid[0]) != 2 {
		t.Fatalf("Subplots grid dimensions = %dx%d, want 2x2", len(grid), len(grid[0]))
	}
	if got := GCF(); got != fig {
		t.Fatalf("GCF() = %p, want %p", got, fig)
	}
	if got := GCA(); got != grid[0][0] {
		t.Fatalf("GCA() = %p, want %p", got, grid[0][0])
	}
}

func TestStatefulHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	Plot([]float64{0, 1, 2}, []float64{1, 2, 3}, core.PlotOptions{Label: "line"})
	Title("Demo")
	XLabel("time")
	YLabel("value")
	legend := Legend()
	if legend == nil {
		t.Fatal("Legend() returned nil")
	}

	ax := GCA()
	if ax.Title != "Demo" {
		t.Fatalf("ax.Title = %q, want %q", ax.Title, "Demo")
	}
	if ax.XLabel != "time" || ax.YLabel != "value" {
		t.Fatalf("axis labels = (%q, %q), want (%q, %q)", ax.XLabel, ax.YLabel, "time", "value")
	}
	if len(ax.Artists) != 2 {
		t.Fatalf("len(ax.Artists) = %d, want 2", len(ax.Artists))
	}
}

func TestColorbarUsesCurrentAxesAndFigure(t *testing.T) {
	resetForTests()

	img := Image([][]float64{
		{0, 1},
		{2, 3},
	})
	cb := Colorbar(img, core.ColorbarOptions{Label: "Intensity"})
	if cb == nil {
		t.Fatal("Colorbar() returned nil")
	}

	fig := GCF()
	if len(fig.Children) != 2 {
		t.Fatalf("len(fig.Children) = %d, want 2", len(fig.Children))
	}
	if cb.YLabel != "Intensity" {
		t.Fatalf("colorbar label = %q, want %q", cb.YLabel, "Intensity")
	}
}

func TestVectorFieldHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	q := Quiver(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 0.5},
		[]float64{0.25, 0.75},
		core.QuiverOptions{Label: "q"},
	)
	if q == nil {
		t.Fatal("Quiver() returned nil")
	}
	key := QuiverKey(q, 0.8, 0.2, 1, "1 unit")
	if key == nil {
		t.Fatal("QuiverKey() returned nil")
	}
	barbs := Barbs(
		[]float64{0.5},
		[]float64{0.5},
		[]float64{12},
		[]float64{3},
	)
	if barbs == nil {
		t.Fatal("Barbs() returned nil")
	}
	stream := Streamplot(
		[]float64{0, 1, 2},
		[]float64{0, 1},
		[][]float64{{1, 1, 1}, {1, 1, 1}},
		[][]float64{{0, 0.2, 0.2}, {0, 0.2, 0.2}},
		core.StreamplotOptions{StartPoints: []geom.Pt{{X: 0.2, Y: 0.4}}},
	)
	if stream == nil {
		t.Fatal("Streamplot() returned nil")
	}

	ax := GCA()
	if len(ax.Artists) != 4 {
		t.Fatalf("len(ax.Artists) = %d, want 4", len(ax.Artists))
	}
}

func TestSavefigWritesPNGAndSVG(t *testing.T) {
	resetForTests()
	t.Setenv("MATPLOTLIB_BACKEND", "gobasic")

	Plot([]float64{0, 1, 2}, []float64{2, 1, 3})
	Title("Savefig")

	dir := t.TempDir()
	pngPath := filepath.Join(dir, "plot.png")
	if err := Savefig(pngPath); err != nil {
		t.Fatalf("Savefig(%q) failed: %v", pngPath, err)
	}
	if info, err := os.Stat(pngPath); err != nil || info.Size() == 0 {
		t.Fatalf("PNG output missing or empty: info=%v err=%v", info, err)
	}

	svgPath := filepath.Join(dir, "plot.svg")
	if err := Savefig(svgPath); err != nil {
		t.Fatalf("Savefig(%q) failed: %v", svgPath, err)
	}
	data, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) failed: %v", svgPath, err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Fatalf("SVG output does not start with <svg: %q", string(data))
	}
}

func TestShowAndPauseUseConfiguredHandler(t *testing.T) {
	resetForTests()

	fig1 := Figure()
	fig2 := Figure()
	var shown []*core.Figure

	SetShowHandler(func(fig *core.Figure) error {
		shown = append(shown, fig)
		return nil
	})

	if err := Show(); err != nil {
		t.Fatalf("Show() failed: %v", err)
	}
	if len(shown) != 2 || shown[0] != fig1 || shown[1] != fig2 {
		t.Fatalf("Show() figures = %v, want [%p %p]", shown, fig1, fig2)
	}

	shown = shown[:0]
	if err := Pause(5 * time.Millisecond); err != nil {
		t.Fatalf("Pause() failed: %v", err)
	}
	if len(shown) != 2 {
		t.Fatalf("Pause() show count = %d, want 2", len(shown))
	}
}

func TestSetManagerFactoryCachesManagerPerFigure(t *testing.T) {
	resetForTests()

	fig := Figure()
	factoryCalls := 0
	showCalls := 0

	SetManagerFactory(func(got *core.Figure) (canvas.FigureManager, error) {
		factoryCalls++
		if got != fig {
			t.Fatalf("factory figure = %p, want %p", got, fig)
		}
		return &testFigureManager{
			canvas: &testFigureCanvas{figure: got},
			onShow: func() { showCalls++ },
			tools:  canvas.NewToolManager(),
		}, nil
	})

	if err := Show(); err != nil {
		t.Fatalf("Show() error = %v", err)
	}
	if err := Show(); err != nil {
		t.Fatalf("Show() second call error = %v", err)
	}

	if factoryCalls != 1 {
		t.Fatalf("factory calls = %d, want 1", factoryCalls)
	}
	if showCalls != 2 {
		t.Fatalf("show calls = %d, want 2", showCalls)
	}
}

func TestRCUpdatesActiveDefaultsForNewFigures(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "144"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}
	if err := RC("axes", style.Params{"facecolor": "#ddeeff"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	fig := Figure()
	if fig.RC.DPI != 144 {
		t.Fatalf("figure DPI = %v, want 144", fig.RC.DPI)
	}
	if got := fig.RC.AxesBackground; got.R != 0xdd/255.0 || got.G != 0xee/255.0 || got.B != 0xff/255.0 {
		t.Fatalf("axes facecolor = %+v", got)
	}
}

func TestRCContextTemporarilyOverridesDefaults(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "120"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	restore, err := RCContext(style.Params{"figure.dpi": "220"})
	if err != nil {
		t.Fatalf("RCContext() error = %v", err)
	}

	if got := Figure().RC.DPI; got != 220 {
		t.Fatalf("context figure DPI = %v, want 220", got)
	}

	restore()
	if got := Figure().RC.DPI; got != 120 {
		t.Fatalf("restored figure DPI = %v, want 120", got)
	}
}

func TestRCDefaultsResetsActiveDefaults(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{"dpi": "144"}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}
	RCDefaults()

	if got := Figure().RC.DPI; got != style.Default.DPI {
		t.Fatalf("figure DPI = %v, want default %v", got, style.Default.DPI)
	}
}

func TestLoadRCFileUpdatesDefaults(t *testing.T) {
	resetForTests()

	path := filepath.Join(t.TempDir(), "matplotlibrc")
	if err := os.WriteFile(path, []byte("figure.dpi: 175\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := LoadRCFile(path); err != nil {
		t.Fatalf("LoadRCFile() error = %v", err)
	}
	if got := Figure().RC.DPI; got != 175 {
		t.Fatalf("figure DPI = %v, want 175", got)
	}
}

func TestFigureUsesConfiguredFigureSize(t *testing.T) {
	resetForTests()

	if err := RC("figure", style.Params{
		"dpi":     "120",
		"figsize": "7.5, 5",
	}); err != nil {
		t.Fatalf("RC() error = %v", err)
	}

	fig := Figure()
	if fig.SizePx.X != 900 || fig.SizePx.Y != 600 {
		t.Fatalf("figure size = %.0fx%.0f, want 900x600", fig.SizePx.X, fig.SizePx.Y)
	}
}

type testFigureManager struct {
	canvas canvas.FigureCanvas
	onShow func()
	tools  *canvas.ToolManager
}

func (m *testFigureManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *testFigureManager) Show() error {
	if m.onShow != nil {
		m.onShow()
	}
	return nil
}

func (m *testFigureManager) Close() error { return nil }

func (m *testFigureManager) SetTitle(string) {}

func (m *testFigureManager) ToolManager() *canvas.ToolManager { return m.tools }

type testFigureCanvas struct {
	figure *core.Figure
}

func (c *testFigureCanvas) Figure() *core.Figure { return c.figure }

func (c *testFigureCanvas) Draw() error { return nil }

func (c *testFigureCanvas) Resize(width, height int) error {
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	return nil
}

func (c *testFigureCanvas) Connect(canvas.EventType, canvas.Handler) canvas.ConnectionID { return 0 }

func (c *testFigureCanvas) Disconnect(canvas.ConnectionID) {}

func (c *testFigureCanvas) Close() error { return nil }
