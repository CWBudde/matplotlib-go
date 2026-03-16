// Package agg provides a high-quality rendering backend using Anti-Grain Geometry.
//
// AGG (Anti-Grain Geometry) is a professional 2D graphics library that provides
// exceptional anti-aliasing, sub-pixel accuracy, and advanced path operations.
// This backend offers superior visual quality compared to the basic GoBasic renderer,
// making it ideal for scientific visualization and publication-quality plots.
//
// Key features:
//   - High-quality anti-aliasing with gamma correction
//   - Sub-pixel precision for accurate scientific plots
//   - Advanced path operations and complex clipping
//   - Professional rendering quality suitable for publications
//   - Cross-platform consistent output
//
// The AGG backend implements the render.Renderer interface and can be selected
// through the backend registry system:
//
//	// Explicit selection
//	renderer, err := backends.Create(backends.AGG, config)
//
//	// Automatic selection for scientific use
//	backend, _ := backends.GetRecommendedBackend("scientific")
//	renderer, err := backends.Create(backend, config)
//
// Performance considerations:
//   - AGG prioritizes quality over speed
//   - Expect 1.5-2.5x slower than GoBasic for complex plots
//   - Memory usage is higher due to anti-aliasing buffers
//   - Best suited for final output rather than interactive previews
package agg