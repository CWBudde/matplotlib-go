// Package unstructured_showcase is the parity-test wrapper for the unstructured_showcase showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/unstructured_showcase;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package unstructured_showcase

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/unstructured_showcase"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
