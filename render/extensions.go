package render

import "matplotlib-go/internal/geom"

// DPIAware is implemented by renderers that adapt text/layout behavior to DPI.
type DPIAware interface {
	SetResolution(dpi uint)
}

// TextDrawer is implemented by renderers that support direct text drawing.
type TextDrawer interface {
	DrawText(text string, origin geom.Pt, size float64, textColor Color)
}

// RotatedTextDrawer is implemented by renderers that support rotated text.
type RotatedTextDrawer interface {
	TextDrawer
	DrawTextRotated(text string, center geom.Pt, size, angle float64, textColor Color)
}

// VerticalTextDrawer is implemented by renderers that support vertical text.
type VerticalTextDrawer interface {
	TextDrawer
	DrawTextVertical(text string, center geom.Pt, size float64, textColor Color)
}

// ImageTransformer is implemented by renderers that support affine image transforms.
type ImageTransformer interface {
	ImageTransformed(img Image, dst geom.Rect, transform geom.Affine)
}

// PNGExporter is implemented by renderers that can export their output to PNG.
type PNGExporter interface {
	SavePNG(path string) error
}
