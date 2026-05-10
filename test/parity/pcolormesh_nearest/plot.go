// Package pcolormesh_nearest is the parity-test wrapper for the pcolormesh_nearest showcase.
// The canonical rendering body lives in github.com/cwbudde/matplotlib-go/examples/pcolormesh_nearest;
// this file imports it so the parity registry and golden tests share that single
// source of truth.
package pcolormesh_nearest

import (
	"image"

	showcase "github.com/cwbudde/matplotlib-go/examples/pcolormesh_nearest"
)

// Render returns the parity image, identical to the showcase output.
func Render() image.Image {
	return showcase.Render()
}
