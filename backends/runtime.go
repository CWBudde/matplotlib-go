package backends

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"matplotlib-go/canvas"
	"matplotlib-go/core"
	"matplotlib-go/render"
)

type headlessCanvas struct {
	mu         sync.Mutex
	figure     *core.Figure
	backend    Backend
	config     Config
	factory    Factory
	renderer   render.Renderer
	dispatcher canvas.Dispatcher
	closed     bool
}

type defaultManager struct {
	canvas  canvas.FigureCanvas
	tools   *canvas.ToolManager
	title   string
	home    figureHomeState
	saveFn  func(string) error
	homeFn  func() error
	drawFn  func() error
	closeFn func() error
}

type figureHomeState struct {
	width  int
	height int
	axes   []axesHomeState
}

type axesHomeState struct {
	axes       *core.Axes
	xMin, xMax float64
	yMin, yMax float64
}

func NewCanvas(choice string, config Config, fig *core.Figure, required []Capability) (canvas.FigureCanvas, Backend, error) {
	backend, info, err := resolveBackendInfo(choice, required)
	if err != nil {
		return nil, "", err
	}
	if info.CanvasFactory != nil {
		out, err := info.CanvasFactory(config, fig)
		return out, backend, err
	}
	return newHeadlessCanvas(fig, backend, config, info.Factory), backend, nil
}

func NewCanvasFromEnv(config Config, fig *core.Figure, required []Capability) (canvas.FigureCanvas, Backend, error) {
	return NewCanvas(strings.TrimSpace(getBackendEnv()), config, fig, required)
}

func NewManager(choice string, config Config, fig *core.Figure, required []Capability) (canvas.FigureManager, Backend, error) {
	backend, info, err := resolveBackendInfo(choice, required)
	if err != nil {
		return nil, "", err
	}
	if info.ManagerFactory != nil {
		out, err := info.ManagerFactory(config, fig)
		return out, backend, err
	}
	if info.CanvasFactory != nil {
		out, err := info.CanvasFactory(config, fig)
		if err != nil {
			return nil, "", err
		}
		return newDefaultManager(out, nil), backend, nil
	}
	headless := newHeadlessCanvas(fig, backend, config, info.Factory)
	return newDefaultManager(headless, headless.save), backend, nil
}

func NewManagerFromEnv(config Config, fig *core.Figure, required []Capability) (canvas.FigureManager, Backend, error) {
	return NewManager(strings.TrimSpace(getBackendEnv()), config, fig, required)
}

func resolveBackendInfo(choice string, required []Capability) (Backend, *BackendInfo, error) {
	backend, err := ResolveBackend(choice, required)
	if err != nil {
		return "", nil, err
	}
	info, ok := DefaultRegistry.Get(backend)
	if !ok {
		return "", nil, fmt.Errorf("unknown backend: %s", backend)
	}
	return backend, info, nil
}

func newHeadlessCanvas(fig *core.Figure, backend Backend, config Config, factory Factory) *headlessCanvas {
	return &headlessCanvas{
		figure:  fig,
		backend: backend,
		config:  config,
		factory: factory,
	}
}

func (c *headlessCanvas) Figure() *core.Figure {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.figure
}

func (c *headlessCanvas) Draw() error {
	event, err := c.draw(false)
	if err != nil {
		return err
	}
	return c.dispatcher.Emit(event)
}

func (c *headlessCanvas) Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.New("backends: resize dimensions must be positive")
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("backends: canvas is closed")
	}
	c.config.Width = width
	c.config.Height = height
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	event := c.normalizeEvent(canvas.Event{
		Type:   canvas.EventResize,
		Width:  width,
		Height: height,
	})
	c.mu.Unlock()

	if err := c.dispatcher.Emit(event); err != nil {
		return err
	}
	_, err := c.draw(false)
	return err
}

func (c *headlessCanvas) Connect(eventType canvas.EventType, handler canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *headlessCanvas) Disconnect(id canvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *headlessCanvas) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.renderer = nil
	event := c.normalizeEvent(canvas.Event{Type: canvas.EventClose})
	c.mu.Unlock()
	return c.dispatcher.Emit(event)
}

func (c *headlessCanvas) save(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("backends: save path is required")
	}
	_, err := c.draw(true)
	if err != nil {
		return err
	}

	c.mu.Lock()
	renderer := c.renderer
	c.mu.Unlock()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		exporter, ok := renderer.(render.PNGExporter)
		if !ok {
			return fmt.Errorf("backends: backend %s does not support PNG export", c.backend)
		}
		return exporter.SavePNG(path)
	case ".svg":
		exporter, ok := renderer.(render.SVGExporter)
		if !ok {
			return fmt.Errorf("backends: backend %s does not support SVG export", c.backend)
		}
		return exporter.SaveSVG(path)
	default:
		return fmt.Errorf("backends: unsupported save extension %q", filepath.Ext(path))
	}
}

func (c *headlessCanvas) draw(skipEmit bool) (canvas.Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return canvas.Event{}, errors.New("backends: canvas is closed")
	}
	if c.factory == nil {
		return canvas.Event{}, errors.New("backends: nil renderer factory")
	}

	renderer, err := c.factory(c.config)
	if err != nil {
		return canvas.Event{}, err
	}
	core.DrawFigure(c.figure, renderer)
	c.renderer = renderer

	event := c.normalizeEvent(canvas.Event{
		Type:   canvas.EventDraw,
		Width:  c.config.Width,
		Height: c.config.Height,
	})
	if skipEmit {
		return event, nil
	}
	return event, nil
}

func (c *headlessCanvas) normalizeEvent(event canvas.Event) canvas.Event {
	if event.Figure == nil {
		event.Figure = c.figure
	}
	if event.Width == 0 {
		event.Width = c.config.Width
	}
	if event.Height == 0 {
		event.Height = c.config.Height
	}
	if event.Axes == nil && event.Figure != nil {
		if ax, data, ok := canvas.ResolveEventTarget(event.Figure, event.Position); ax != nil {
			event.Axes = ax
			event.DataPosition = data
			event.HasDataPosition = ok
		}
	}
	return event
}

func newDefaultManager(figCanvas canvas.FigureCanvas, saveFn func(string) error) canvas.FigureManager {
	manager := &defaultManager{
		canvas: figCanvas,
		tools:  canvas.NewToolManager(),
		saveFn: saveFn,
	}
	manager.drawFn = figCanvas.Draw
	manager.closeFn = figCanvas.Close
	manager.home = snapshotFigureHome(figCanvas.Figure())
	manager.homeFn = func() error {
		return restoreFigureHome(manager.home, figCanvas)
	}
	manager.tools.Register(canvas.ToolFunc{
		Name: "home",
		Run: func(canvas.ToolArgs) error {
			return manager.homeFn()
		},
	})
	manager.tools.Register(canvas.ToolFunc{
		Name: "redraw",
		Run: func(canvas.ToolArgs) error {
			return manager.drawFn()
		},
	})
	manager.tools.Register(canvas.ToolFunc{
		Name: "save",
		Run: func(args canvas.ToolArgs) error {
			if manager.saveFn == nil {
				return errors.New("backends: active canvas does not support save")
			}
			return manager.saveFn(args.Path)
		},
	})
	return manager
}

func (m *defaultManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *defaultManager) Show() error {
	if m == nil || m.drawFn == nil {
		return nil
	}
	return m.drawFn()
}

func (m *defaultManager) Close() error {
	if m == nil || m.closeFn == nil {
		return nil
	}
	return m.closeFn()
}

func (m *defaultManager) SetTitle(title string) { m.title = title }

func (m *defaultManager) ToolManager() *canvas.ToolManager { return m.tools }

func snapshotFigureHome(fig *core.Figure) figureHomeState {
	state := figureHomeState{}
	if fig == nil {
		return state
	}
	state.width = int(fig.SizePx.X + 0.5)
	state.height = int(fig.SizePx.Y + 0.5)
	state.axes = make([]axesHomeState, 0, len(fig.Children))
	for _, ax := range fig.Children {
		if ax == nil {
			continue
		}
		xMin, xMax := 0.0, 1.0
		if ax.XScale != nil {
			xMin, xMax = ax.XScale.Domain()
		}
		yMin, yMax := 0.0, 1.0
		if ax.YScale != nil {
			yMin, yMax = ax.YScale.Domain()
		}
		state.axes = append(state.axes, axesHomeState{
			axes: ax,
			xMin: xMin,
			xMax: xMax,
			yMin: yMin,
			yMax: yMax,
		})
	}
	return state
}

func restoreFigureHome(state figureHomeState, figCanvas canvas.FigureCanvas) error {
	fig := figCanvas.Figure()
	if fig == nil {
		return nil
	}
	for _, axState := range state.axes {
		if axState.axes == nil {
			continue
		}
		axState.axes.SetXLim(axState.xMin, axState.xMax)
		axState.axes.SetYLim(axState.yMin, axState.yMax)
	}
	width := int(fig.SizePx.X + 0.5)
	height := int(fig.SizePx.Y + 0.5)
	if state.width > 0 && state.height > 0 && (width != state.width || height != state.height) {
		return figCanvas.Resize(state.width, state.height)
	}
	return figCanvas.Draw()
}

func getBackendEnv() string {
	return getenv("MATPLOTLIB_BACKEND")
}
