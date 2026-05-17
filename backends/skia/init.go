package skia

import (
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/render"
)

func init() {
	available := isAvailable()
	capabilities := []backends.Capability(nil)
	fallbackCapabilities := []backends.Capability(nil)
	saveFormats := map[string]backends.SaveHandler(nil)
	if available {
		capabilities = []backends.Capability{
			backends.AntiAliasing,
			backends.SubPixel,
			backends.GradientFill,
			backends.PathClip,
			backends.TextShaping,
			backends.FontHinting,
			backends.PNGExport,
		}
		fallbackCapabilities = []backends.Capability{
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
		}
		saveFormats = map[string]backends.SaveHandler{
			".png": backends.SavePNG,
		}
	}

	// Register Skia backend with the global registry
	backends.Register(backends.Skia, &backends.BackendInfo{
		Name:                 "Skia",
		Description:          "Reserved Skia renderer backend; unavailable until the Phase 14.4 implementation lands",
		Capabilities:         capabilities,
		FallbackCapabilities: fallbackCapabilities,
		SaveFormats:          saveFormats,
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config)
		},
		Available: available,
	})
}

// isAvailable checks if Skia dependencies are available at runtime.
func isAvailable() bool {
	// TODO: Check for Skia shared library
	// TODO: Check for required graphics drivers
	// For now, return false since it's not implemented
	return false
}
