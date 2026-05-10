// Package units_overview is the parity-test wrapper for the units_overview showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/units_overview;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package units_overview

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/units_overview"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
