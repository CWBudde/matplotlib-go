package svg

import (
	"matplotlib-go/backends"
	"matplotlib-go/render"
)

func init() {
	backends.Register(backends.SVG, &backends.BackendInfo{
		Name:        "SVG",
		Description: "Pure Go SVG backend with path recording and native text output",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,
			backends.VectorOutput,
			backends.PathClip,
			backends.TextShaping,
			backends.FontHinting,
		},
		FallbackCapabilities: []backends.Capability{
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".svg": backends.SaveSVG,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
