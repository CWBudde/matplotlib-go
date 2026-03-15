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
			backends.GradientFill,
			backends.PathClip,
		},
		Factory: func(config backends.Config) (render.Renderer, error) {
			return New(config.Width, config.Height, config.Background), nil
		},
		Available: true,
	})
}
