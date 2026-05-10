// Package fill_basic is the parity-test wrapper for the fill_basic showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/fill_basic;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package fill_basic

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/fill_basic"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
