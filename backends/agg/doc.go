// Package agg implements the AGG backend for matplotlib-go.
//
// Phase 14.1 task (AGG reference parity) status
// ----------------------------------------------
//
// This package is treated as the AGG reference backend and should track
// parity against:
//
//	third_party/matplotlib/lib/matplotlib/backends/backend_agg.py
//	third_party/matplotlib/src/_backend_agg.h
//
// Native coverage currently implemented
// - Path and marker drawing, including clip rect/path stacking
// - Path collections, quad mesh, and Gouraud triangle rasterization
// - Image drawing and affine image drawing
// - Text raster and path rendering, text measurement, font metrics
// - PNG export
// - `copy_from_bbox` / `restore_region` equivalent (`CopyFromBBox` / `RestoreRegion`)
// - `start_filter` / `stop_filter` equivalent (`StartFilter` / `StopFilter`)
// 
// Intentionally unsupported for now
// - Full upstream `draw_mathtext` / `draw_tex` pipeline parity; MatplotlibTeX/MathText
//   glyph shaping is handled by existing native text fallback plumbing where used.
// - Raw pixel-readback helpers (`buffer_rgba`, `tostring_argb`).
//
// These intentional gaps should be revisited as part of ongoing 14.1 sub-tasks.
package agg
