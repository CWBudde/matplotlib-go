// Package scatter_marker_types is the parity-test wrapper for the scatter_marker_types showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/scatter_marker_types;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package scatter_marker_types

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/scatter_marker_types"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
