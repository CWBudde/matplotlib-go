// Package mplot3d_terrain is the parity-test wrapper for the mplot3d_terrain showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/mplot3d_terrain;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package mplot3d_terrain

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/mplot3d_terrain"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
