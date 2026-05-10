// Package bar_basic_title is the parity-test wrapper for the bar_basic_title showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/bar_basic_title;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package bar_basic_title

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/bar_basic_title"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
