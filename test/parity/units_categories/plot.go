// Package units_categories is the parity-test wrapper for the units_categories showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/units_categories;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package units_categories

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/units_categories"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
