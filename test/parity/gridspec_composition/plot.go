// Package gridspec_composition is the parity-test wrapper for the gridspec_composition showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/gridspec_composition;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package gridspec_composition

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/gridspec_composition"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
