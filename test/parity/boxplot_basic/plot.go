// Package boxplot_basic is the parity-test wrapper for the boxplot_basic showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/boxplot_basic;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package boxplot_basic

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/boxplot_basic"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
