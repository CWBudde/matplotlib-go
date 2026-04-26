package backends

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

type contractRenderer struct {
	render.NullRenderer
	pngPath string
}

func (r *contractRenderer) DrawText(string, geom.Pt, float64, render.Color) {}

func (r *contractRenderer) SavePNG(path string) error {
	r.pngPath = path
	return nil
}

func TestRegistryVerifiesDeclaredRendererCapabilities(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("contract")
	registry.Register(backend, &BackendInfo{
		Name:         "Contract",
		Available:    true,
		Capabilities: []Capability{TextShaping, VectorOutput},
		Factory: func(Config) (render.Renderer, error) {
			return &contractRenderer{}, nil
		},
	})

	if registry.SupportsRendererCapability(backend, &render.NullRenderer{}, TextShaping) {
		t.Fatal("null renderer unexpectedly supports text shaping")
	}
	if err := registry.VerifyRendererCapabilities(backend, &render.NullRenderer{}); err == nil {
		t.Fatal("VerifyRendererCapabilities accepted renderer missing declared optional interfaces")
	}
	if err := registry.VerifyRendererCapabilities(backend, &contractRenderer{}); err != nil {
		t.Fatalf("VerifyRendererCapabilities rejected contract renderer: %v", err)
	}
}

func TestRegistrySaveViaExtensionUsesBackendFormatHandlers(t *testing.T) {
	registry := NewRegistry()
	backend := Backend("contract")
	registry.Register(backend, &BackendInfo{
		Name:      "Contract",
		Available: true,
		SaveFormats: map[string]SaveHandler{
			".png": SavePNG,
		},
	})

	renderer := &contractRenderer{}
	if err := registry.SaveViaExtension(backend, renderer, "plot.png"); err != nil {
		t.Fatalf("SaveViaExtension returned error: %v", err)
	}
	if renderer.pngPath != "plot.png" {
		t.Fatalf("pngPath = %q, want plot.png", renderer.pngPath)
	}
	if err := registry.SaveViaExtension(backend, renderer, "plot.pdf"); err == nil {
		t.Fatal("SaveViaExtension accepted unsupported format")
	}
}
