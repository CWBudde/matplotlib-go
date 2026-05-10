// Package fill_stacked is the parity-test wrapper for the fill_stacked showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/fill_stacked;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package fill_stacked

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/fill_stacked"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
