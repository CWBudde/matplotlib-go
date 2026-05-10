// Package pcolormesh_gouraud is the parity-test wrapper for the pcolormesh_gouraud showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/pcolormesh_gouraud;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package pcolormesh_gouraud

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/pcolormesh_gouraud"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
