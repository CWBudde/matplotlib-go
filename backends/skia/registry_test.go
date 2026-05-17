package skia

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
)

func TestUnavailableSkiaScaffoldDoesNotClaimNativeCapabilities(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.Skia)
	if !ok {
		t.Fatal("skia backend should be registered")
	}
	if info.Available {
		t.Fatal("skia scaffold should stay unavailable until dependencies and renderer are implemented")
	}
	if len(info.Capabilities) != 0 {
		t.Fatalf("unavailable skia scaffold should not claim native capabilities: %v", info.Capabilities)
	}
	if len(info.SaveFormats) != 0 {
		t.Fatalf("unavailable skia scaffold should not claim save formats: %v", info.SaveFormats)
	}
}
