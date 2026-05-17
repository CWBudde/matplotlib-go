package svg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
)

func TestSVGBackend_PopulatesSaveFormats(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.SVG)
	if !ok {
		t.Fatal("SVG backend not registered")
	}
	if info.SaveFormats == nil {
		t.Fatal("SVG SaveFormats map is nil")
	}
	if _, ok := info.SaveFormats[".svg"]; !ok {
		t.Fatalf("SVG SaveFormats missing .svg handler; got keys: %v", mapKeys(info.SaveFormats))
	}
}

func TestSVGBackend_AdvertisedCapabilitiesAreImplemented(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.SVG)
	if !ok {
		t.Fatal("SVG backend not registered")
	}
	renderer, err := info.Factory(backends.Config{Width: 10, Height: 10})
	if err != nil {
		t.Fatalf("Factory failed: %v", err)
	}
	if err := backends.DefaultRegistry.VerifyRendererCapabilities(backends.SVG, renderer); err != nil {
		t.Fatalf("SVG backend advertises a capability it does not implement: %v", err)
	}
}

func mapKeys(m map[string]backends.SaveHandler) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
