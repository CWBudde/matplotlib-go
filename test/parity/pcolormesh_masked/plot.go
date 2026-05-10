// Package pcolormesh_masked is the parity-test wrapper for the pcolormesh_masked showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/pcolormesh_masked;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package pcolormesh_masked

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/pcolormesh_masked"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
