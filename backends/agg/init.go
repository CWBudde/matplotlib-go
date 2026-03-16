package agg

import (
	"matplotlib-go/backends"
	"matplotlib-go/render"
)

// init registers the AGG backend with the global registry.
func init() {
	backends.Register(backends.AGG, &backends.BackendInfo{
		Name:        "AGG (Anti-Grain Geometry)",
		Description: "High-quality anti-aliased rendering with sub-pixel precision",
		Capabilities: []backends.Capability{
			backends.AntiAliasing,  // High-quality anti-aliasing
			backends.SubPixel,      // Sub-pixel accuracy
			backends.PathClip,      // Complex path-based clipping
			// TODO: Add more capabilities as they're implemented
			// backends.GradientFill,  // Gradient fills (when AGG port supports)
			// backends.TextShaping,   // Advanced text rendering
		},
		Factory:   createRenderer,
		Available: true, // AGG is always available once compiled
	})
}

// createRenderer creates a new AGG renderer instance with the given configuration.
func createRenderer(config backends.Config) (render.Renderer, error) {
	// Extract AGG-specific configuration if provided
	var aggConfig *AggConfig
	if config.Options != nil {
		if cfg, ok := config.Options.(*AggConfig); ok {
			aggConfig = cfg
		}
	}
	
	// Apply default configuration if none provided
	if aggConfig == nil {
		aggConfig = &AggConfig{
			AntiAliasing:    true,
			SubPixelPrec:    true,
			GammaCorrection: 2.2, // Standard gamma for anti-aliasing
		}
	}
	
	// Create the renderer
	renderer, err := New(config.Width, config.Height)
	if err != nil {
		return nil, err
	}
	
	// Apply AGG-specific configuration
	renderer.config = *aggConfig
	
	return renderer, nil
}

