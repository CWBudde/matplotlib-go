// Package mixed_collection is the parity-test wrapper for the mixed_collection showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/mixed_collection;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package mixed_collection

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/mixed_collection"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
