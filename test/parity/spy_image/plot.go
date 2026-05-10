// Package spy_image is the parity-test wrapper for the spy_image showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/spy_image;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package spy_image

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/spy_image"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
