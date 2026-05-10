// Package bar_grouped is the parity-test wrapper for the bar_grouped showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/bar_grouped;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package bar_grouped

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/bar_grouped"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
