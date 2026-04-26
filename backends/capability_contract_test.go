package backends

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

type capabilityRendererBase struct {
	render.NullRenderer
}

type capabilityRendererWithTextDrawer struct {
	capabilityRendererBase
}

func (r *capabilityRendererWithTextDrawer) DrawText(_ string, _ geom.Pt, _ float64, _ render.Color) {
	_ = r // no-op
}

type capabilityRendererWithFontMetrics struct {
	capabilityRendererWithTextDrawer
}

func (r *capabilityRendererWithFontMetrics) MeasureFontHeights(_ float64, _ string) (render.FontHeightMetrics, bool) {
	return render.FontHeightMetrics{}, true
}

type capabilityRendererWithPNG struct {
	capabilityRendererWithFontMetrics
}

func (r *capabilityRendererWithPNG) SavePNG(_ string) error {
	return nil
}

func (r *capabilityRendererWithPNG) ImageTransformed(_ render.Image, _ geom.Rect, _ geom.Affine) {}

func TestSupportsRendererCapability(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("contract"), &BackendInfo{
		Name:         "Contract Backend",
		Available:    true,
		Capabilities: []Capability{TextShaping, FontHinting, VectorOutput},
		Factory:      func(Config) (render.Renderer, error) { return &capabilityRendererWithPNG{}, nil },
	})
	withDefaultRegistry(t, reg)

	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithPNG{}, TextShaping) {
		t.Fatal("expected TextShaping capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithPNG{}, FontHinting) {
		t.Fatal("expected FontHinting capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithPNG{}, VectorOutput) {
		t.Fatal("expected VectorOutput capability to be supported")
	}

	if SupportsRendererCapability(Backend("contract"), &capabilityRendererBase{}, TextShaping) {
		t.Fatal("expected TextShaping capability to be unsupported without DrawText")
	}
	if SupportsRendererCapability(Backend("contract"), &capabilityRendererWithTextDrawer{}, FontHinting) {
		t.Fatal("expected FontHinting capability to be unsupported without TextFontMetricer")
	}
	if SupportsRendererCapability(Backend("contract"), &capabilityRendererWithFontMetrics{}, VectorOutput) {
		t.Fatal("expected VectorOutput capability to be unsupported without export interface")
	}
	if SupportsRendererCapability(Backend("contract"), nil, TextShaping) {
		t.Fatal("expected nil renderer to be unsupported")
	}
}

func TestVerifyRendererCapabilities(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("good"), &BackendInfo{
		Name:         "Good Backend",
		Available:    true,
		Capabilities: []Capability{TextShaping, FontHinting},
		Factory:      func(Config) (render.Renderer, error) { return &capabilityRendererWithFontMetrics{}, nil },
	})
	reg.Register(Backend("bad"), &BackendInfo{
		Name:         "Bad Backend",
		Available:    true,
		Capabilities: []Capability{FontHinting},
		Factory:      func(Config) (render.Renderer, error) { return &capabilityRendererWithTextDrawer{}, nil },
	})
	withDefaultRegistry(t, reg)

	if err := VerifyRendererCapabilities(Backend("good"), &capabilityRendererWithFontMetrics{}); err != nil {
		t.Fatalf("VerifyRendererCapabilities(good) error = %v", err)
	}
	if err := VerifyRendererCapabilities(Backend("bad"), &capabilityRendererWithTextDrawer{}); err == nil {
		t.Fatal("expected VerifyRendererCapabilities(bad) to fail")
	}
	if err := VerifyRendererCapabilities(Backend("unknown"), &capabilityRendererWithPNG{}); err == nil {
		t.Fatal("expected VerifyRendererCapabilities(unknown) to fail")
	}
	if err := VerifyRendererCapabilities(Backend("good"), nil); err == nil {
		t.Fatal("expected VerifyRendererCapabilities(nil renderer) to fail")
	}
}

