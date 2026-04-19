package core

import (
	"errors"

	"matplotlib-go/render"
)

// SaveSVG saves a figure to an SVG file using the provided renderer.
// This function draws the figure using the renderer and then exports to SVG.
func SaveSVG(fig *Figure, r render.Renderer, path string) error {
	// Draw the figure using the renderer.
	DrawFigure(fig, r)

	// Check if this renderer supports SVG export.
	if exporter, ok := r.(render.SVGExporter); ok {
		return exporter.SaveSVG(path)
	}

	// For other renderer types, we don't have SVG export support yet.
	return errors.New("SVG export not supported for this renderer type")
}
