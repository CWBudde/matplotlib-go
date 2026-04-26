package canvas

import (
	"testing"

	"matplotlib-go/internal/geom"
)

func TestFirstClassEventWrappersPreserveCanonicalTypes(t *testing.T) {
	fig := &Figure{}

	draw := NewDrawEvent(fig, 640, 480)
	if draw.Type != EventDraw || draw.Figure != fig || draw.Width != 640 || draw.Height != 480 {
		t.Fatalf("draw event = %+v", draw.Event)
	}

	mouse := NewMouseEvent(EventMousePress, fig, geom.Pt{X: 12, Y: 34}, MouseButtonLeft)
	if mouse.Type != EventMousePress || mouse.Position.X != 12 || mouse.Button != MouseButtonLeft {
		t.Fatalf("mouse event = %+v", mouse.Event)
	}

	key := NewKeyEvent(EventKeyPress, fig, "ctrl+s", ModifierControl)
	if key.Type != EventKeyPress || key.Key != "ctrl+s" || key.Modifiers != ModifierControl {
		t.Fatalf("key event = %+v", key.Event)
	}
}

func TestDispatcherConnectionLifecycle(t *testing.T) {
	var dispatcher Dispatcher
	called := 0
	id := dispatcher.Connect(EventDraw, func(Event) error {
		called++
		return nil
	})
	if id == 0 {
		t.Fatal("connection id = 0, want non-zero")
	}

	if err := dispatcher.Emit(Event{Type: EventDraw}); err != nil {
		t.Fatal(err)
	}
	dispatcher.Disconnect(id)
	if err := dispatcher.Emit(Event{Type: EventDraw}); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("called = %d, want 1", called)
	}
}
