// Package fill_between is the parity-test wrapper for the fill_between showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/fill_between;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package fill_between

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/fill_between"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
