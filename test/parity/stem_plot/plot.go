// Package stem_plot is the parity-test wrapper for the stem_plot showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/stem_plot;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package stem_plot

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/stem_plot"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
