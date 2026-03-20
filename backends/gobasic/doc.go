// Package gobasic provides a pure Go renderer backend using golang.org/x/image/vector.
//
// This backend uses image.RGBA as the drawing surface and vector.Rasterizer for
// path filling and stroking. It is designed to be deterministic and work without
// CGO dependencies.
//
// The GoBasic renderer supports:
//   - Fill and stroke operations (no dashes in Phase B)
//   - Rectangular clipping
//   - State stack for Save/Restore operations
//   - PNG export via image/png package
//
// This backend is kept as a pure-Go reference implementation and compatibility
// fallback, while AGG is the default renderer in the current port.
package gobasic
