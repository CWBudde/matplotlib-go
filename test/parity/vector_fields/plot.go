// Package vector_fields is the parity-test wrapper for the vector_fields showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/vector_fields;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package vector_fields

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/vector_fields"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
