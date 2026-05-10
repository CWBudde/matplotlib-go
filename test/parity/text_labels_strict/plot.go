// Package text_labels_strict is the parity-test wrapper for the text_labels_strict showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/text_labels_strict;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package text_labels_strict

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/text_labels_strict"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
