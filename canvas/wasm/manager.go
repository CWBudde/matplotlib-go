//go:build js && wasm

package wasm

import (
	"errors"
	"fmt"
	"syscall/js"

	"matplotlib-go/backends/gobasic"
	plotcanvas "matplotlib-go/canvas"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
)

type listener struct {
	target js.Value
	event  string
	fn     js.Func
}

type manager struct {
	canvas *figureCanvas
	tools  *plotcanvas.ToolManager
	home   figureHomeState
}

type figureCanvas struct {
	figure     *core.Figure
	element    js.Value
	context    js.Value
	dispatcher plotcanvas.Dispatcher
	listeners  []listener
	closed     bool
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

func NewGoBasicManager(elementID string, fig *core.Figure) (plotcanvas.FigureManager, error) {
	if fig == nil {
		return nil, errors.New("canvas/wasm: nil figure")
	}

	document := js.Global().Get("document")
	if document.IsUndefined() || document.IsNull() {
		return nil, errors.New("canvas/wasm: document is unavailable")
	}
	element := document.Call("getElementById", elementID)
	if element.IsUndefined() || element.IsNull() {
		return nil, fmt.Errorf("canvas/wasm: canvas element %q not found", elementID)
	}
	context := element.Call("getContext", "2d")
	if context.IsUndefined() || context.IsNull() {
		return nil, errors.New("canvas/wasm: 2d context is unavailable")
	}

	c := &figureCanvas{
		figure:  fig,
		element: element,
		context: context,
	}
	if tabIndex := element.Get("tabIndex"); tabIndex.IsUndefined() || tabIndex.Int() < 0 {
		element.Set("tabIndex", 0)
	}
	c.installListeners()

	m := &manager{
		canvas: c,
		tools:  plotcanvas.NewToolManager(),
		home:   snapshotFigureHome(fig),
	}
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "home",
		Run: func(plotcanvas.ToolArgs) error {
			return restoreFigureHome(m.home, c)
		},
	})
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "redraw",
		Run: func(plotcanvas.ToolArgs) error {
			return c.Draw()
		},
	})
	m.tools.Register(plotcanvas.ToolFunc{
		Name: "save",
		Run: func(plotcanvas.ToolArgs) error {
			return errors.New("canvas/wasm: save is not implemented for the browser host")
		},
	})

	return m, nil
}

func (m *manager) Canvas() plotcanvas.FigureCanvas { return m.canvas }

func (m *manager) Show() error { return m.canvas.Draw() }

func (m *manager) Close() error { return m.canvas.Close() }

func (m *manager) SetTitle(title string) {
	document := js.Global().Get("document")
	if document.IsUndefined() || document.IsNull() {
		return
	}
	document.Set("title", title)
}

func (m *manager) ToolManager() *plotcanvas.ToolManager { return m.tools }

func (c *figureCanvas) Figure() *core.Figure { return c.figure }

func (c *figureCanvas) Draw() error {
	if c.closed {
		return errors.New("canvas/wasm: canvas is closed")
	}
	width, height := c.currentSize()
	if width <= 0 || height <= 0 {
		return errors.New("canvas/wasm: invalid canvas size")
	}

	c.element.Set("width", width)
	c.element.Set("height", height)
	c.figure.SizePx.X = float64(width)
	c.figure.SizePx.Y = float64(height)

	renderer := gobasic.New(width, height, c.figure.RC.FigureBackground())
	if renderer == nil {
		return errors.New("canvas/wasm: failed to create GoBasic renderer")
	}
	core.DrawFigure(c.figure, renderer)

	img := renderer.GetImage()
	pixels := js.Global().Get("Uint8ClampedArray").New(len(img.Pix))
	js.CopyBytesToJS(pixels, img.Pix)
	imageData := js.Global().Get("ImageData").New(pixels, width, height)
	c.context.Call("putImageData", imageData, 0, 0)

	return c.dispatcher.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventDraw,
		Figure: c.figure,
		Width:  width,
		Height: height,
	})
}

func (c *figureCanvas) Resize(width, height int) error {
	if c.closed {
		return errors.New("canvas/wasm: canvas is closed")
	}
	if width <= 0 || height <= 0 {
		return errors.New("canvas/wasm: resize dimensions must be positive")
	}

	c.element.Set("width", width)
	c.element.Set("height", height)
	c.figure.SizePx.X = float64(width)
	c.figure.SizePx.Y = float64(height)

	if err := c.dispatcher.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventResize,
		Figure: c.figure,
		Width:  width,
		Height: height,
	}); err != nil {
		return err
	}
	return c.Draw()
}

func (c *figureCanvas) Connect(eventType plotcanvas.EventType, handler plotcanvas.Handler) plotcanvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *figureCanvas) Disconnect(id plotcanvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *figureCanvas) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	for _, listener := range c.listeners {
		listener.target.Call("removeEventListener", listener.event, listener.fn)
		listener.fn.Release()
	}
	c.listeners = nil
	return c.dispatcher.Emit(plotcanvas.Event{
		Type:   plotcanvas.EventClose,
		Figure: c.figure,
		Width:  int(c.figure.SizePx.X + 0.5),
		Height: int(c.figure.SizePx.Y + 0.5),
	})
}

func (c *figureCanvas) installListeners() {
	c.on(c.element, "mousedown", func(this js.Value, args []js.Value) any {
		event := c.mouseEvent(plotcanvas.EventMousePress, args[0])
		c.focus()
		return c.emit(event)
	})
	c.on(c.element, "mouseup", func(this js.Value, args []js.Value) any {
		return c.emit(c.mouseEvent(plotcanvas.EventMouseRelease, args[0]))
	})
	c.on(c.element, "mousemove", func(this js.Value, args []js.Value) any {
		return c.emit(c.mouseEvent(plotcanvas.EventMouseMove, args[0]))
	})
	c.on(c.element, "wheel", func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		return c.emit(c.scrollEvent(args[0]))
	})

	window := js.Global().Get("window")
	c.on(window, "keydown", func(this js.Value, args []js.Value) any {
		return c.emit(c.keyEvent(plotcanvas.EventKeyPress, args[0]))
	})
	c.on(window, "keyup", func(this js.Value, args []js.Value) any {
		return c.emit(c.keyEvent(plotcanvas.EventKeyRelease, args[0]))
	})
	c.on(window, "resize", func(this js.Value, _ []js.Value) any {
		width, height := c.elementClientSize()
		if width <= 0 || height <= 0 {
			return nil
		}
		if err := c.Resize(width, height); err != nil {
			js.Global().Get("console").Call("error", err.Error())
		}
		return nil
	})
}

func (c *figureCanvas) on(target js.Value, event string, fn func(this js.Value, args []js.Value) any) {
	callback := js.FuncOf(fn)
	target.Call("addEventListener", event, callback)
	c.listeners = append(c.listeners, listener{target: target, event: event, fn: callback})
}

func (c *figureCanvas) emit(event plotcanvas.Event) any {
	if c.closed {
		return nil
	}
	event.Figure = c.figure
	if event.Axes == nil {
		if ax, data, ok := plotcanvas.ResolveEventTarget(c.figure, event.Position); ax != nil {
			event.Axes = ax
			event.DataPosition = data
			event.HasDataPosition = ok
		}
	}
	if err := c.dispatcher.Emit(event); err != nil {
		js.Global().Get("console").Call("error", err.Error())
	}
	return nil
}

func (c *figureCanvas) mouseEvent(eventType plotcanvas.EventType, domEvent js.Value) plotcanvas.Event {
	position := elementPosition(c.element, domEvent)
	return plotcanvas.Event{
		Type:      eventType,
		Position:  position,
		Button:    mouseButton(domEvent.Get("button").Int()),
		Modifiers: modifiers(domEvent),
		Native:    domEvent,
	}
}

func (c *figureCanvas) scrollEvent(domEvent js.Value) plotcanvas.Event {
	position := elementPosition(c.element, domEvent)
	return plotcanvas.Event{
		Type:      plotcanvas.EventScroll,
		Position:  position,
		DeltaX:    domEvent.Get("deltaX").Float(),
		DeltaY:    domEvent.Get("deltaY").Float(),
		Modifiers: modifiers(domEvent),
		Native:    domEvent,
	}
}

func (c *figureCanvas) keyEvent(eventType plotcanvas.EventType, domEvent js.Value) plotcanvas.Event {
	return plotcanvas.Event{
		Type:      eventType,
		Figure:    c.figure,
		Key:       domEvent.Get("key").String(),
		Modifiers: modifiers(domEvent),
		Native:    domEvent,
	}
}

func (c *figureCanvas) currentSize() (int, int) {
	width, height := c.elementClientSize()
	if width > 0 && height > 0 {
		return width, height
	}
	return int(c.figure.SizePx.X + 0.5), int(c.figure.SizePx.Y + 0.5)
}

func (c *figureCanvas) elementClientSize() (int, int) {
	width := c.element.Get("clientWidth").Int()
	height := c.element.Get("clientHeight").Int()
	return width, height
}

func (c *figureCanvas) focus() {
	if c.element.IsUndefined() || c.element.IsNull() {
		return
	}
	if focus := c.element.Get("focus"); focus.Type() == js.TypeFunction {
		c.element.Call("focus")
	}
}

func elementPosition(element, event js.Value) geom.Pt {
	rect := element.Call("getBoundingClientRect")
	width := float64(element.Get("width").Int())
	height := float64(element.Get("height").Int())
	rectWidth := rect.Get("width").Float()
	rectHeight := rect.Get("height").Float()
	if rectWidth <= 0 {
		rectWidth = 1
	}
	if rectHeight <= 0 {
		rectHeight = 1
	}
	x := (event.Get("clientX").Float() - rect.Get("left").Float()) * width / rectWidth
	y := (event.Get("clientY").Float() - rect.Get("top").Float()) * height / rectHeight
	return geom.Pt{X: x, Y: y}
}

func mouseButton(button int) plotcanvas.MouseButton {
	switch button {
	case 1:
		return plotcanvas.MouseButtonMiddle
	case 2:
		return plotcanvas.MouseButtonRight
	default:
		return plotcanvas.MouseButtonLeft
	}
}

func modifiers(event js.Value) plotcanvas.Modifier {
	var out plotcanvas.Modifier
	if event.Get("shiftKey").Bool() {
		out |= plotcanvas.ModifierShift
	}
	if event.Get("ctrlKey").Bool() {
		out |= plotcanvas.ModifierControl
	}
	if event.Get("altKey").Bool() {
		out |= plotcanvas.ModifierAlt
	}
	if event.Get("metaKey").Bool() {
		out |= plotcanvas.ModifierMeta
	}
	return out
}

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

func restoreFigureHome(state figureHomeState, c *figureCanvas) error {
	for _, axState := range state.axes {
		if axState.axes == nil {
			continue
		}
		axState.axes.SetXLim(axState.xMin, axState.xMax)
		axState.axes.SetYLim(axState.yMin, axState.yMax)
	}
	width := int(c.figure.SizePx.X + 0.5)
	height := int(c.figure.SizePx.Y + 0.5)
	if width != state.width || height != state.height {
		return c.Resize(state.width, state.height)
	}
	return c.Draw()
}
