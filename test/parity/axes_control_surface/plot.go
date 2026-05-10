// Package axes_control_surface is the parity-test wrapper for the axes_control_surface showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/axes_control_surface;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package axes_control_surface

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/axes_control_surface"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
