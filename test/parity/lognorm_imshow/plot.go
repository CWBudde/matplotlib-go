// Package lognorm_imshow is the parity-test wrapper for the lognorm_imshow showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/lognorm_imshow;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package lognorm_imshow

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/lognorm_imshow"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
