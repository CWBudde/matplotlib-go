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

type capabilityRendererWithBatches struct {
	capabilityRendererWithPNG
}

func (r *capabilityRendererWithBatches) DrawMarkers(_ render.MarkerBatch) bool {
	return true
}

func (r *capabilityRendererWithBatches) DrawPathCollection(_ render.PathCollectionBatch) bool {
	return true
}

func (r *capabilityRendererWithBatches) DrawQuadMesh(_ render.QuadMeshBatch) bool {
	return true
}

func (r *capabilityRendererWithBatches) DrawGouraudTriangles(_ render.GouraudTriangleBatch) bool {
	return true
}

func TestSupportsRendererCapability(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("contract"), &BackendInfo{
		Name:         "Contract Backend",
		Available:    true,
		Capabilities: []Capability{TextShaping, FontHinting, VectorOutput, MarkerBatch, PathCollectionBatch, QuadMeshBatch, GouraudTriangleBatch},
		Factory:      func(Config) (render.Renderer, error) { return &capabilityRendererWithBatches{}, nil },
	})
	withDefaultRegistry(t, reg)

	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, TextShaping) {
		t.Fatal("expected TextShaping capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, FontHinting) {
		t.Fatal("expected FontHinting capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, VectorOutput) {
		t.Fatal("expected VectorOutput capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, MarkerBatch) {
		t.Fatal("expected MarkerBatch capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, PathCollectionBatch) {
		t.Fatal("expected PathCollectionBatch capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, QuadMeshBatch) {
		t.Fatal("expected QuadMeshBatch capability to be supported")
	}
	if !SupportsRendererCapability(Backend("contract"), &capabilityRendererWithBatches{}, GouraudTriangleBatch) {
		t.Fatal("expected GouraudTriangleBatch capability to be supported")
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
	if SupportsRendererCapability(Backend("contract"), &capabilityRendererWithPNG{}, MarkerBatch) {
		t.Fatal("expected MarkerBatch capability to be unsupported without batch interface")
	}
	if SupportsRendererCapability(Backend("contract"), nil, TextShaping) {
		t.Fatal("expected nil renderer to be unsupported")
	}
}

func TestRendererCapabilityStatusDistinguishesFallback(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("native"), &BackendInfo{
		Name:         "Native Backend",
		Available:    true,
		Capabilities: []Capability{MarkerBatch},
		Factory:      func(Config) (render.Renderer, error) { return &capabilityRendererWithBatches{}, nil },
	})
	reg.Register(Backend("fallback"), &BackendInfo{
		Name:                 "Fallback Backend",
		Available:            true,
		FallbackCapabilities: []Capability{MarkerBatch},
		Factory:              func(Config) (render.Renderer, error) { return &capabilityRendererBase{}, nil },
	})
	withDefaultRegistry(t, reg)

	if got := RendererCapabilityStatus(Backend("native"), &capabilityRendererWithBatches{}, MarkerBatch); got != CapabilityNative {
		t.Fatalf("native status = %s, want %s", got, CapabilityNative)
	}
	if got := RendererCapabilityStatus(Backend("fallback"), &capabilityRendererBase{}, MarkerBatch); got != CapabilityFallback {
		t.Fatalf("fallback status = %s, want %s", got, CapabilityFallback)
	}
	if got := RendererCapabilityStatus(Backend("native"), &capabilityRendererBase{}, MarkerBatch); got != CapabilityUnsupported {
		t.Fatalf("unsupported status = %s, want %s", got, CapabilityUnsupported)
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
