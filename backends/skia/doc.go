// Package skia is the reserved Skia renderer backend for matplotlib-go.
//
// # Strategy
//
// Skia is developed under PLAN.md 14.4, not as a Phase 8 side task. The chosen
// integration strategy is recorded by BackendStrategy:
//   - build tag: skia
//   - binding: a small external C ABI wrapper around Skia
//   - first implementation target: CPU raster output
//   - GPU mode: deferred until the CPU path and capability reporting are stable
//   - default CI: non-Skia stub builds; dependency-enabled Skia tests are gated
//
// The package deliberately does not depend on an unstable Go Skia binding. The
// production backend should call a narrow C ABI that this repository controls,
// keeping build failures and platform policy localized to the skia package.
//
// # Dependencies
//
// A functional Skia build will require CGO_ENABLED=1, a Skia shared library, and
// the matplotlib-go C ABI wrapper library. GPU mode will additionally require
// platform-specific graphics drivers and context setup.
//
// # Current Status
//
// Default builds compile the unavailable stub in skia_stub.go. The skia build
// tag currently compiles a scaffold only; it is not registered as available
// until the renderer implements the base render.Renderer contract and PNG
// export. Capability claims must remain aligned with runtime interfaces.
//
// # Configuration
//
// Use SkiaConfig to configure GPU usage, color formats, and quality settings:
//
//	config := backends.Config{
//		Width: 800, Height: 600,
//		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
//		Options: backends.SkiaConfig{
//			UseGPU: true,
//			SampleCount: 4, // 4x MSAA
//		},
//	}
package skia
