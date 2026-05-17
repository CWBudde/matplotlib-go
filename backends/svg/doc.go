// Package svg implements a native SVG renderer backend for matplotlib-go.
//
// The backend emits vector paths, text, clip paths, transformed images, marker
// batches, path collections, and hatch fills directly as SVG. Gouraud-shaded
// triangle meshes are intentionally not native yet; core code keeps using the
// renderer-neutral raster fallback for that case until an SVG gradient strategy
// is worth the added complexity.
package svg
