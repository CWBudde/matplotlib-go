// Package joins_caps is the parity-test wrapper for the joins_caps showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/joins_caps;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package joins_caps

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/joins_caps"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
