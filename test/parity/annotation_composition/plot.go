// Package annotation_composition is the parity-test wrapper for the annotation_composition showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/annotation_composition;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package annotation_composition

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/annotation_composition"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
