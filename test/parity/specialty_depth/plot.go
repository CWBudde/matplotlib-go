// Package specialty_depth is the parity-test wrapper for the specialty_depth showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/specialty_depth;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package specialty_depth

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/specialty_depth"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
