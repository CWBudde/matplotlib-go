// Package specialty_artists is the parity-test wrapper for the specialty_artists showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/specialty_artists;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package specialty_artists

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/specialty_artists"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
