// Package transform_coordinates is the parity-test wrapper for the transform_coordinates showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/transform_coordinates;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package transform_coordinates

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/transform_coordinates"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
