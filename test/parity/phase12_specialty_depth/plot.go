// Package phase12_specialty_depth is the parity-test wrapper for the phase12_specialty_depth showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/phase12_specialty_depth;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package phase12_specialty_depth

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/phase12_specialty_depth"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
