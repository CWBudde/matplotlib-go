// Package scatter_advanced is the parity-test wrapper for the scatter_advanced showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/scatter_advanced;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package scatter_advanced

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/scatter_advanced"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
