// Package hist_density is the parity-test wrapper for the hist_density showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/hist_density;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package hist_density

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/hist_density"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
