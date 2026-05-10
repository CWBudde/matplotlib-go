// Package geo_lambert_axes is the parity-test wrapper for the geo_lambert_axes showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/geo_lambert_axes;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package geo_lambert_axes

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/geo_lambert_axes"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
