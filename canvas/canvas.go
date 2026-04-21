package canvas

import (
	"sync/atomic"

	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
)

// Figure aliases the plotting figure type for runtime-facing APIs.
type Figure = core.Figure

// Axes aliases the plotting axes type for runtime-facing APIs.
type Axes = core.Axes

// EventType identifies a canvas runtime event.
type EventType string

const (
	EventDraw         EventType = "draw"
	EventResize       EventType = "resize"
	EventClose        EventType = "close"
	EventMousePress   EventType = "mouse_press"
	EventMouseRelease EventType = "mouse_release"
	EventMouseMove    EventType = "mouse_move"
	EventScroll       EventType = "scroll"
	EventKeyPress     EventType = "key_press"
	EventKeyRelease   EventType = "key_release"
)

// MouseButton identifies a mouse button in a runtime event.
type MouseButton uint8

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

// Modifier tracks active key modifiers for an event.
type Modifier uint8

const (
	ModifierShift Modifier = 1 << iota
	ModifierControl
	ModifierAlt
	ModifierMeta
)

// Event carries normalized runtime input and lifecycle information.
type Event struct {
	Type            EventType
	Figure          *Figure
	Axes            *Axes
	Position        geom.Pt
	DataPosition    geom.Pt
	HasDataPosition bool
	Width           int
	Height          int
	Button          MouseButton
	Key             string
	Modifiers       Modifier
	DeltaX          float64
	DeltaY          float64
	Native          any
}

// ConnectionID identifies a registered event handler.
type ConnectionID int64

// Handler is invoked for a normalized runtime event.
type Handler func(Event) error

// FigureCanvas exposes drawing, sizing, and event subscription for one figure.
type FigureCanvas interface {
	Figure() *Figure
	Draw() error
	Resize(width, height int) error
	Connect(EventType, Handler) ConnectionID
	Disconnect(ConnectionID)
	Close() error
}

// FigureManager exposes presentation and tooling for one figure canvas.
type FigureManager interface {
	Canvas() FigureCanvas
	Show() error
	Close() error
	SetTitle(string)
	ToolManager() *ToolManager
}

// ResolveEventTarget resolves the topmost axes under a figure-pixel position.
func ResolveEventTarget(fig *Figure, position geom.Pt) (*Axes, geom.Pt, bool) {
	if fig == nil {
		return nil, geom.Pt{}, false
	}

	for i := len(fig.Children) - 1; i >= 0; i-- {
		ax := fig.Children[i]
		if ax == nil || !ax.ContainsDisplayPoint(position) {
			continue
		}
		data, ok := ax.PixelToData(position)
		return ax, data, ok
	}

	return nil, geom.Pt{}, false
}

func nextConnectionID(counter *atomic.Int64) ConnectionID {
	return ConnectionID(counter.Add(1))
}
