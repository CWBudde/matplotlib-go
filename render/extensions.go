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
	// DrawTextRotated matches Matplotlib's default y-axis label anchoring:
	// the point is the bottom-center anchor of the unrotated text box, and the
	// text is then rotated around that anchor.
	DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor Color)
}

// TextBounder is implemented by renderers that can report the actual ink bounds
// of text relative to the baseline origin used for DrawText.
type TextBounder interface {
	MeasureTextBounds(text string, size float64, fontKey string) (TextBounds, bool)
}

// TextFontMetricer is implemented by renderers that can report font-wide line
// metrics separately from the ink bounds of a particular string.
type TextFontMetricer interface {
	MeasureFontHeights(size float64, fontKey string) (FontHeightMetrics, bool)
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

// SVGExporter is implemented by renderers that can export their output to SVG.
type SVGExporter interface {
	SaveSVG(path string) error
}
