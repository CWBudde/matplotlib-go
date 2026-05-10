// Package multi_series_basic is the parity-test wrapper for the multi_series_basic showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/multi_series_basic;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package multi_series_basic

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/multi_series_basic"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
