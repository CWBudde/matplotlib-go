// Package bar_basic is the parity-test wrapper for the bar_basic showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/bar_basic;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package bar_basic

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/bar_basic"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
