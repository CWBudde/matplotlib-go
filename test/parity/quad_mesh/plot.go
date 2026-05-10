// Package quad_mesh is the parity-test wrapper for the quad_mesh showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/quad_mesh;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package quad_mesh

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/quad_mesh"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
