// Package twoslope_norm_image is the parity-test wrapper for the twoslope_norm_image showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/twoslope_norm_image;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package twoslope_norm_image

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/twoslope_norm_image"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
