// Package multi_series_color_cycle is the parity-test wrapper for the multi_series_color_cycle showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/multi_series_color_cycle;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package multi_series_color_cycle

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/multi_series_color_cycle"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
