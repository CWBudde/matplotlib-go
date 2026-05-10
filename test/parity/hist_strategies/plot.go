// Package hist_strategies is the parity-test wrapper for the hist_strategies showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/hist_strategies;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package hist_strategies

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/hist_strategies"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
