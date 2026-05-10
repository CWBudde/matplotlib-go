// Package basic_line is the parity-test wrapper for the basic_line showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/basic_line;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package basic_line

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/basic_line"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
