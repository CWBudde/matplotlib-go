package pyplot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"matplotlib-go/backends"
	"matplotlib-go/canvas"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/render"
	"matplotlib-go/style"
)

const (
	// DefaultFigureWidth matches Matplotlib's default 6.4in figure width at
	// the repository's default 100 DPI.
	DefaultFigureWidth = 640
	// DefaultFigureHeight matches Matplotlib's default 4.8in figure height at
	// the repository's default 100 DPI.
	DefaultFigureHeight = 480
)

// ShowHandler renders or presents a figure when Show or Pause is called.
type ShowHandler func(*core.Figure) error

// ManagerFactory creates a figure manager for pyplot Show/Pause lifecycle calls.
type ManagerFactory func(*core.Figure) (canvas.FigureManager, error)

type registryState struct {
	mu          sync.Mutex
	current     *core.Figure
	figures     []*core.Figure
	currentAxes map[*core.Figure]*core.Axes
	subplotAxes map[*core.Figure]map[string]*core.Axes
	managers    map[*core.Figure]canvas.FigureManager
	managerFactory ManagerFactory
}

var registry = registryState{
	currentAxes: make(map[*core.Figure]*core.Axes),
	subplotAxes: make(map[*core.Figure]map[string]*core.Axes),
	managers:    make(map[*core.Figure]canvas.FigureManager),
	managerFactory: defaultManagerFactory,
}

// Figure creates a new current figure using Matplotlib-like default dimensions.
func Figure(opts ...style.Option) *core.Figure {
	return FigureSized(DefaultFigureWidth, DefaultFigureHeight, opts...)
}

// FigureSized creates a new current figure with explicit pixel dimensions.
func FigureSized(width, height int, opts ...style.Option) *core.Figure {
	if width <= 0 {
		width = DefaultFigureWidth
	}
	if height <= 0 {
		height = DefaultFigureHeight
	}

	fig := core.NewFigure(width, height, opts...)
	registry.mu.Lock()
	registry.figures = append(registry.figures, fig)
	registry.current = fig
	registry.mu.Unlock()
	return fig
}

// GCF returns the current figure, creating one if necessary.
func GCF() *core.Figure {
	registry.mu.Lock()
	fig := registry.current
	registry.mu.Unlock()
	if fig != nil {
		return fig
	}
	return Figure()
}

// GCA returns the current axes for the current figure, creating a default
// 1x1 subplot if necessary.
func GCA() *core.Axes {
	fig := GCF()

	registry.mu.Lock()
	ax := registry.currentAxes[fig]
	registry.mu.Unlock()
	if ax != nil {
		return ax
	}
	return ensureDefaultAxes(fig)
}

// Subplot returns the requested subplot axes in the current figure.
func Subplot(nRows, nCols, index int) *core.Axes {
	fig := GCF()
	key := subplotKey(nRows, nCols, index)

	registry.mu.Lock()
	if ax := registry.subplotAxes[fig][key]; ax != nil {
		registry.current = fig
		registry.currentAxes[fig] = ax
		registry.mu.Unlock()
		return ax
	}
	registry.mu.Unlock()

	ax := fig.AddSubplot(nRows, nCols, index)
	if ax == nil {
		return nil
	}

	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, key)
	registry.mu.Unlock()
	return ax
}

// Subplots creates a new figure and subplot grid, then marks the first axes as
// current.
func Subplots(nRows, nCols int, opts ...core.SubplotOption) (*core.Figure, [][]*core.Axes) {
	fig := Figure()
	grid := fig.Subplots(nRows, nCols, opts...)
	if len(grid) == 0 || len(grid[0]) == 0 {
		return fig, grid
	}

	registry.mu.Lock()
	for row := range grid {
		for col, ax := range grid[row] {
			registry.rememberAxesLocked(fig, ax, subplotKey(nRows, nCols, row*nCols+col+1))
		}
	}
	registry.current = fig
	registry.currentAxes[fig] = grid[0][0]
	registry.mu.Unlock()

	return fig, grid
}

// Plot delegates to the current axes.
func Plot(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().Plot(x, y, opts...)
}

// Scatter delegates to the current axes.
func Scatter(x, y []float64, opts ...core.ScatterOptions) *core.Scatter2D {
	return GCA().Scatter(x, y, opts...)
}

// Bar delegates to the current axes.
func Bar(x, heights []float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().Bar(x, heights, opts...)
}

// FillBetween delegates to the current axes.
func FillBetween(x, y1, y2 []float64, opts ...core.FillOptions) *core.Fill2D {
	return GCA().FillBetween(x, y1, y2, opts...)
}

// Hist delegates to the current axes.
func Hist(data []float64, opts ...core.HistOptions) *core.Hist2D {
	return GCA().Hist(data, opts...)
}

// ErrorBar delegates to the current axes.
func ErrorBar(x, y, xErr, yErr []float64, opts ...core.ErrorBarOptions) *core.ErrorBar {
	return GCA().ErrorBar(x, y, xErr, yErr, opts...)
}

// Image delegates to the current axes.
func Image(data [][]float64, opts ...core.ImageOptions) *core.Image2D {
	return GCA().Image(data, opts...)
}

// Title sets the current axes title.
func Title(label string) {
	GCA().SetTitle(label)
}

// XLabel sets the current axes x-axis label.
func XLabel(label string) {
	GCA().SetXLabel(label)
}

// YLabel sets the current axes y-axis label.
func YLabel(label string) {
	GCA().SetYLabel(label)
}

// Legend adds a legend to the current axes.
func Legend() *core.Legend {
	return GCA().AddLegend()
}

// Colorbar adds a figure-level colorbar for the current axes.
func Colorbar(mappable *core.Image2D, opts ...core.ColorbarOptions) *core.Axes {
	ax := GCA()
	fig := GCF()
	if cb := fig.AddColorbar(ax, mappable, opts...); cb != nil {
		return cb
	}
	makeColorbarRoom(ax, opts...)
	return fig.AddColorbar(ax, mappable, opts...)
}

// RCParams returns the active rcParam-style defaults.
func RCParams() style.Params {
	return style.CurrentParams()
}

// RC applies rcParam-style overrides to the active defaults. When group is non-empty,
// keys in params are prefixed with "group." unless already fully qualified.
func RC(group string, params style.Params) error {
	_, err := style.UpdateParams(qualifyRCParams(group, params))
	return err
}

// RCContext applies temporary rcParam overrides and returns a restore function.
func RCContext(params style.Params) (func(), error) {
	restore, _, err := style.PushContext(params)
	if err != nil {
		return nil, err
	}
	return restore, nil
}

// RCDefaults restores the active defaults to the library baseline.
func RCDefaults() {
	style.ResetDefaults()
}

// LoadRCFile loads a Matplotlib-style rc file into the active defaults.
func LoadRCFile(path string) error {
	_, err := style.LoadRCFile(path)
	return err
}

// LoadDefaultRCFile loads the first rc file found in the default search path.
func LoadDefaultRCFile() (string, error) {
	path, _, err := style.LoadDefaultRCFile()
	return path, err
}

// Savefig renders the current figure to a file selected by extension.
func Savefig(path string) error {
	return saveFigure(GCF(), path)
}

// SetManagerFactory overrides how pyplot creates managers for Show and Pause.
// Passing nil restores the default backend-driven manager selection.
func SetManagerFactory(factory ManagerFactory) {
	if factory == nil {
		factory = defaultManagerFactory
	}

	registry.mu.Lock()
	existing := registry.managers
	registry.managers = make(map[*core.Figure]canvas.FigureManager)
	registry.managerFactory = factory
	registry.mu.Unlock()

	for _, manager := range existing {
		if manager != nil {
			_ = manager.Close()
		}
	}
}

// SetShowHandler overrides how Show and Pause present figures. Passing nil
// restores the default manager-backed behavior.
func SetShowHandler(handler ShowHandler) {
	if handler == nil {
		SetManagerFactory(nil)
		return
	}
	SetManagerFactory(func(fig *core.Figure) (canvas.FigureManager, error) {
		return newShowHandlerManager(fig, handler), nil
	})
}

// Show renders all registered figures through the configured show handler.
func Show() error {
	registry.mu.Lock()
	figures := append([]*core.Figure(nil), registry.figures...)
	registry.mu.Unlock()

	for _, fig := range figures {
		if fig == nil {
			continue
		}
		manager, err := ensureManager(fig)
		if err != nil {
			return err
		}
		if err := manager.Show(); err != nil {
			return err
		}
	}
	return nil
}

// Pause renders open figures and then blocks for the requested interval.
func Pause(interval time.Duration) error {
	if err := Show(); err != nil {
		return err
	}
	if interval > 0 {
		time.Sleep(interval)
	}
	return nil
}

func ensureDefaultAxes(fig *core.Figure) *core.Axes {
	registry.mu.Lock()
	if ax := registry.currentAxes[fig]; ax != nil {
		registry.mu.Unlock()
		return ax
	}
	registry.mu.Unlock()
	return Subplot(1, 1, 1)
}

func subplotKey(nRows, nCols, index int) string {
	return fmt.Sprintf("%d:%d:%d", nRows, nCols, index)
}

func qualifyRCParams(group string, params style.Params) style.Params {
	if len(params) == 0 {
		return nil
	}

	qualified := make(style.Params, len(params))
	group = strings.ToLower(strings.TrimSpace(group))
	for key, value := range params {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if group != "" && !strings.Contains(normalizedKey, ".") {
			normalizedKey = group + "." + normalizedKey
		}
		qualified[normalizedKey] = value
	}
	return qualified
}

func (r *registryState) rememberAxesLocked(fig *core.Figure, ax *core.Axes, key string) {
	if fig == nil || ax == nil {
		return
	}
	if r.subplotAxes[fig] == nil {
		r.subplotAxes[fig] = make(map[string]*core.Axes)
	}
	if key != "" {
		r.subplotAxes[fig][key] = ax
	}
	r.current = fig
	r.currentAxes[fig] = ax
}

func defaultManagerFactory(fig *core.Figure) (canvas.FigureManager, error) {
	manager, _, err := backends.NewManagerFromEnv(rendererConfig(fig), fig, backends.TextCapabilities)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func saveFigure(fig *core.Figure, path string) error {
	if fig == nil {
		return errors.New("pyplot: nil figure")
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		renderer, err := newPNGRenderer(fig)
		if err != nil {
			return err
		}
		return core.SavePNG(fig, renderer, path)
	case ".svg":
		renderer, _, err := backends.NewRenderer(string(backends.SVG), rendererConfig(fig), nil)
		if err != nil {
			return err
		}
		return core.SaveSVG(fig, renderer, path)
	default:
		return fmt.Errorf("pyplot: unsupported savefig extension %q", filepath.Ext(path))
	}
}

func newPNGRenderer(fig *core.Figure) (render.Renderer, error) {
	cfg := rendererConfig(fig)
	if choice := os.Getenv("MATPLOTLIB_BACKEND"); strings.TrimSpace(choice) != "" {
		renderer, _, err := backends.NewRenderer(choice, cfg, backends.TextCapabilities)
		if err == nil {
			if _, ok := renderer.(render.PNGExporter); ok {
				return renderer, nil
			}
		}
	}

	for _, backend := range backends.Available() {
		renderer, err := backends.Create(backend, cfg)
		if err != nil {
			continue
		}
		if _, ok := renderer.(render.PNGExporter); ok {
			return renderer, nil
		}
	}
	return nil, errors.New("pyplot: no PNG-capable backend available")
}

func rendererConfig(fig *core.Figure) backends.Config {
	width := DefaultFigureWidth
	height := DefaultFigureHeight
	defaults := style.CurrentDefaults()
	background := defaults.FigureBackground()
	dpi := defaults.DPI

	if fig != nil {
		if fig.SizePx.X > 0 {
			width = int(fig.SizePx.X + 0.5)
		}
		if fig.SizePx.Y > 0 {
			height = int(fig.SizePx.Y + 0.5)
		}
		background = fig.RC.FigureBackground()
		dpi = fig.RC.DPI
	}

	return backends.Config{
		Width:      width,
		Height:     height,
		Background: background,
		DPI:        dpi,
	}
}

func makeColorbarRoom(ax *core.Axes, opts ...core.ColorbarOptions) {
	if ax == nil {
		return
	}

	width := 0.035
	padding := 0.02
	if len(opts) > 0 {
		if opts[0].Width > 0 {
			width = opts[0].Width
		}
		if opts[0].Padding > 0 {
			padding = opts[0].Padding
		}
	}

	maxRight := 1 - padding - width
	if ax.RectFraction.Max.X <= maxRight {
		return
	}
	if maxRight <= ax.RectFraction.Min.X {
		return
	}
	ax.RectFraction.Max.X = maxRight
}

func resetForTests() {
	registry.mu.Lock()
	registry.current = nil
	registry.figures = nil
	registry.currentAxes = make(map[*core.Figure]*core.Axes)
	registry.subplotAxes = make(map[*core.Figure]map[string]*core.Axes)
	registry.managers = make(map[*core.Figure]canvas.FigureManager)
	registry.managerFactory = defaultManagerFactory
	registry.mu.Unlock()
	style.ResetDefaults()
}

func ensureManager(fig *core.Figure) (canvas.FigureManager, error) {
	registry.mu.Lock()
	if manager := registry.managers[fig]; manager != nil {
		registry.mu.Unlock()
		return manager, nil
	}
	factory := registry.managerFactory
	registry.mu.Unlock()

	manager, err := factory(fig)
	if err != nil {
		return nil, err
	}

	registry.mu.Lock()
	if existing := registry.managers[fig]; existing != nil {
		registry.mu.Unlock()
		_ = manager.Close()
		return existing, nil
	}
	registry.managers[fig] = manager
	registry.mu.Unlock()
	return manager, nil
}

type showHandlerManager struct {
	canvas *showHandlerCanvas
	tools  *canvas.ToolManager
}

type showHandlerCanvas struct {
	figure     *core.Figure
	handler    ShowHandler
	dispatcher canvas.Dispatcher
	closed     bool
}

func newShowHandlerManager(fig *core.Figure, handler ShowHandler) canvas.FigureManager {
	c := &showHandlerCanvas{figure: fig, handler: handler}
	manager := &showHandlerManager{
		canvas: c,
		tools:  canvas.NewToolManager(),
	}
	manager.tools.Register(canvas.ToolFunc{
		Name: "redraw",
		Run: func(canvas.ToolArgs) error {
			return c.Draw()
		},
	})
	return manager
}

func (m *showHandlerManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *showHandlerManager) Show() error { return m.canvas.Draw() }

func (m *showHandlerManager) Close() error { return m.canvas.Close() }

func (m *showHandlerManager) SetTitle(string) {}

func (m *showHandlerManager) ToolManager() *canvas.ToolManager { return m.tools }

func (c *showHandlerCanvas) Figure() *core.Figure { return c.figure }

func (c *showHandlerCanvas) Draw() error {
	if c.closed {
		return nil
	}
	if c.handler == nil {
		return nil
	}
	if err := c.handler(c.figure); err != nil {
		return err
	}
	return c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventDraw,
		Figure: c.figure,
		Width:  int(c.figure.SizePx.X + 0.5),
		Height: int(c.figure.SizePx.Y + 0.5),
	})
}

func (c *showHandlerCanvas) Resize(width, height int) error {
	if c.closed {
		return nil
	}
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	if err := c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventResize,
		Figure: c.figure,
		Width:  width,
		Height: height,
	}); err != nil {
		return err
	}
	return c.Draw()
}

func (c *showHandlerCanvas) Connect(eventType canvas.EventType, handler canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *showHandlerCanvas) Disconnect(id canvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *showHandlerCanvas) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventClose,
		Figure: c.figure,
	})
}
