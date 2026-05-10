// Package dashes is the parity-test wrapper for the dashes showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/dashes;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package dashes

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/dashes"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
