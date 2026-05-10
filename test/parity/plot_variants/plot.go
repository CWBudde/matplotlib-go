// Package plot_variants is the parity-test wrapper for the plot_variants showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/plot_variants;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package plot_variants

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/plot_variants"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
