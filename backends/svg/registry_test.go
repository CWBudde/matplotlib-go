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

func mapKeys(m map[string]backends.SaveHandler) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
