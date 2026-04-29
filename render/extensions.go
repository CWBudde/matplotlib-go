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

// TextPather is implemented by renderers that can convert text to vector paths.
type TextPather interface {
	TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool)
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

// MarkerItem is one positioned marker instance in a repeated marker batch.
// Transform is applied to Marker first, then Offset is added in display space.
type MarkerItem struct {
	Offset      geom.Pt
	Transform   geom.Affine
	Paint       Paint
	Antialiased bool
}

// MarkerBatch describes one marker path rendered at many display-space
// positions.
type MarkerBatch struct {
	Marker geom.Path
	Items  []MarkerItem
}

// MarkerDrawer is implemented by renderers with a native repeated-marker path.
type MarkerDrawer interface {
	DrawMarkers(batch MarkerBatch) bool
}

// PathCollectionItem is one display-space path with its paint state.
type PathCollectionItem struct {
	Path        geom.Path
	Paint       Paint
	Hatch       string
	HatchColor  Color
	HatchWidth  float64
	Antialiased bool
}

// PathCollectionBatch describes many display-space paths rendered as one
// collection operation.
type PathCollectionBatch struct {
	Items []PathCollectionItem
}

// PathCollectionDrawer is implemented by renderers with a native collection
// path.
type PathCollectionDrawer interface {
	DrawPathCollection(batch PathCollectionBatch) bool
}

// QuadMeshCell is one display-space quadrilateral cell.
type QuadMeshCell struct {
	Quad        [4]geom.Pt
	Face        Color
	Edge        Color
	LineWidth   float64
	Dashes      []float64
	Hatch       string
	HatchColor  Color
	HatchWidth  float64
	Antialiased bool
}

// QuadMeshBatch describes pcolor/pcolormesh-style quadrilateral cells.
type QuadMeshBatch struct {
	Cells []QuadMeshCell
}

// QuadMeshDrawer is implemented by renderers with a native quad mesh path.
type QuadMeshDrawer interface {
	DrawQuadMesh(batch QuadMeshBatch) bool
}

// GouraudTriangle describes one triangle with per-vertex display-space colors.
type GouraudTriangle struct {
	P     [3]geom.Pt
	Color [3]Color
}

// GouraudTriangleBatch describes interpolated-color triangles.
type GouraudTriangleBatch struct {
	Triangles   []GouraudTriangle
	Antialiased bool
}

// GouraudTriangleDrawer is implemented by renderers with native Gouraud
// triangle shading.
type GouraudTriangleDrawer interface {
	DrawGouraudTriangles(batch GouraudTriangleBatch) bool
}

// PNGExporter is implemented by renderers that can export their output to PNG.
type PNGExporter interface {
	SavePNG(path string) error
}

// SVGExporter is implemented by renderers that can export their output to SVG.
type SVGExporter interface {
	SaveSVG(path string) error
}
