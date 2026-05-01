package agg

import (
	"testing"

	"matplotlib-go/backends"
)

func TestAGGBackend_PopulatesSaveFormats(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.AGG)
	if !ok {
		t.Fatal("AGG backend not registered")
	}
	if info.SaveFormats == nil {
		t.Fatal("AGG SaveFormats map is nil")
	}
	if _, ok := info.SaveFormats[".png"]; !ok {
		t.Fatalf("AGG SaveFormats missing .png handler; got keys: %v", mapKeys(info.SaveFormats))
	}
}

func mapKeys(m map[string]backends.SaveHandler) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
