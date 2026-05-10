// Package bar_horizontal is the parity-test wrapper for the bar_horizontal showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/bar_horizontal;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package bar_horizontal

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/bar_horizontal"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
