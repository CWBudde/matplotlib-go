package core

import (
	"testing"

	"matplotlib-go/internal/geom"
)

func TestWidgetConstructorsPrepareAxesAndStoreState(t *testing.T) {
	fig := NewFigure(800, 600)
	axButton := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.3, Y: 0.2}})
	axSlider := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.25}, Max: geom.Pt{X: 0.5, Y: 0.38}})
	axCheck := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.1}, Max: geom.Pt{X: 0.8, Y: 0.32}})
	axRadio := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.82, Y: 0.1}, Max: geom.Pt{X: 0.95, Y: 0.32}})
	axText := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.42}, Max: geom.Pt{X: 0.55, Y: 0.56}})

	pressed := true
	active := true
	button := axButton.Button("Run", ButtonOptions{Pressed: &pressed})
	slider := axSlider.Slider("gain", 0, 10, 12)
	checks := axCheck.CheckButtons([]string{"A", "B"}, []bool{true, false})
	radios := axRadio.RadioButtons([]string{"x", "y", "z"}, 5)
	text := axText.TextBox("Query", "", TextBoxOptions{Placeholder: "type...", Active: &active})

	if button == nil || slider == nil || checks == nil || radios == nil || text == nil {
		t.Fatal("expected widget constructors to return artists")
	}
	if !button.Pressed {
		t.Fatal("button should store pressed state")
	}
	if slider.Value != 10 {
		t.Fatalf("slider value = %v, want clamped max 10", slider.Value)
	}
	if radios.Active != 2 {
		t.Fatalf("radio active index = %d, want 2", radios.Active)
	}
	if text.Placeholder != "type..." || !text.Active {
		t.Fatal("text box should store placeholder and active state")
	}
	for _, ax := range []*Axes{axButton, axSlider, axCheck, axRadio, axText} {
		if ax.XAxis.ShowTicks || ax.YAxis.ShowTicks || ax.XAxis.ShowLabels || ax.YAxis.ShowLabels {
			t.Fatal("widget constructors should hide axis ticks and labels")
		}
	}
}
