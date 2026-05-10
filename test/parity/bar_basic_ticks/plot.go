// Package bar_basic_ticks is the parity-test wrapper for the bar_basic_ticks showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/bar_basic_ticks;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package bar_basic_ticks

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/bar_basic_ticks"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
