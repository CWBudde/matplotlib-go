package backends

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNewManagerFallsBackToHeadlessCanvas(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("runtime"), &BackendInfo{
		Name:      "Runtime",
		Available: true,
		Factory:   testRendererFactory(&render.NullRenderer{}, nil),
	})
	withDefaultRegistry(t, reg)

	fig := core.NewFigure(200, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 20)

	manager, backend, err := NewManager("runtime", SimpleConfig(200, 100, render.Color{A: 1}), fig, nil)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	if backend != Backend("runtime") {
		t.Fatalf("backend = %q, want %q", backend, Backend("runtime"))
	}

	var events []canvas.EventType
	manager.Canvas().Connect(canvas.EventResize, func(canvas.Event) error {
		events = append(events, canvas.EventResize)
		return nil
	})
	manager.Canvas().Connect(canvas.EventDraw, func(canvas.Event) error {
		events = append(events, canvas.EventDraw)
		return nil
	})

	if err := manager.Canvas().Resize(320, 180); err != nil {
		t.Fatalf("Resize() error = %v", err)
	}
	if fig.SizePx.X != 320 || fig.SizePx.Y != 180 {
		t.Fatalf("figure size = %.0fx%.0f, want 320x180", fig.SizePx.X, fig.SizePx.Y)
	}
	if len(events) != 2 || events[0] != canvas.EventResize || events[1] != canvas.EventDraw {
		t.Fatalf("events = %v, want [resize draw]", events)
	}
}

func TestHeadlessManagerHomeToolRestoresLimits(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("runtime"), &BackendInfo{
		Name:      "Runtime",
		Available: true,
		Factory:   testRendererFactory(&render.NullRenderer{}, nil),
	})
	withDefaultRegistry(t, reg)

	fig := core.NewFigure(200, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 10)
	ax.SetYLim(-5, 5)

	manager, _, err := NewManager("runtime", SimpleConfig(200, 100, render.Color{A: 1}), fig, nil)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	ax.SetXLim(3, 4)
	ax.SetYLim(7, 9)
	if err := manager.ToolManager().Execute("home", canvas.ToolArgs{}); err != nil {
		t.Fatalf("home tool error = %v", err)
	}

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("x limits = (%v, %v), want (0, 10)", xMin, xMax)
	}
	if yMin != -5 || yMax != 5 {
		t.Fatalf("y limits = (%v, %v), want (-5, 5)", yMin, yMax)
	}
}
