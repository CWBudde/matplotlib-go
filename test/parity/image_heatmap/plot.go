// Package image_heatmap is the parity-test wrapper for the image_heatmap showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/image_heatmap;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package image_heatmap

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/image_heatmap"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
