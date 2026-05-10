// Package title_strict is the parity-test wrapper for the title_strict showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/title_strict;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package title_strict

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/title_strict"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
