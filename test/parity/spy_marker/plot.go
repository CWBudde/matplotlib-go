// Package spy_marker is the parity-test wrapper for the spy_marker showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/spy_marker;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package spy_marker

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/spy_marker"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
