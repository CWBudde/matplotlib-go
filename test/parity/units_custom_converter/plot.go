// Package units_custom_converter is the parity-test wrapper for the units_custom_converter showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/units_custom_converter;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package units_custom_converter

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/units_custom_converter"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
