// Package polar_axes is the parity-test wrapper for the polar_axes showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/polar_axes;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package polar_axes

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/polar_axes"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
