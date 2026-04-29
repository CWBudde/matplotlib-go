package agg

import (
	"matplotlib-go/backends"
	"matplotlib-go/render"
)

func init() {
	backends.Register(backends.AGG, &backends.BackendInfo{
		Name:        "AGG",
		Description: "Anti-Grain Geometry renderer with high-quality anti-aliasing",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,
			backends.SubPixel,
			backends.TextShaping,
			backends.FontHinting,
			backends.MarkerBatch,
			backends.PathCollectionBatch,
			backends.QuadMeshBatch,
			backends.GouraudTriangleBatch,
		},
		SaveFormats: map[string]backends.SaveHandler{
			".png": backends.SavePNG,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background)
		},
		Available: true,
	})
}
