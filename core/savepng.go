package core

import (
	"errors"

	"matplotlib-go/render"
)

// SavePNG saves a figure to a PNG file using the provided renderer.
// This function draws the figure using the renderer and then exports to PNG.
func SavePNG(fig *Figure, r render.Renderer, path string) error {
	// Draw the figure using the renderer
	DrawFigure(fig, r)

	// Check if this renderer supports PNG export
	if exporter, ok := r.(render.PNGExporter); ok {
		return exporter.SavePNG(path)
	}

	// For other renderer types, we don't have PNG export support yet
	return errors.New("PNG export not supported for this renderer type")
}
