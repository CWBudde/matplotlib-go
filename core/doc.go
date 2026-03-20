// Package core contains the Artist tree (Figure, Axes, Artist) and traversal.
//
// Core types:
//   - Artist: Interface for drawable elements with z-order and bounds
//   - Figure: Root container with pixel dimensions and styling
//   - Axes: Plot region with coordinate transforms and child artists
//   - DrawContext: Per-draw state including transforms and styling
//
// Artists:
//   - Line2D: Polyline artist for stroke-only line plots
//   - BoxPlot2D: Statistical box plot artist with quartiles, whiskers, and fliers
//
// Helpers:
//   - DrawFigure: Traverses and renders a figure with proper z-ordering
//   - SavePNG: Placeholder for PNG export (requires backend renderer)
package core
