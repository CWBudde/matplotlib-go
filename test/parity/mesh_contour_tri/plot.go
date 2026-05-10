// Package mesh_contour_tri is the parity-test wrapper for the mesh_contour_tri showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/mesh_contour_tri;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package mesh_contour_tri

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/mesh_contour_tri"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
