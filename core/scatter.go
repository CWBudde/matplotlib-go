package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// MarkerType defines the shape of markers in scatter plots.
type MarkerType uint8

const (
	MarkerCircle MarkerType = iota
	MarkerSquare
	MarkerTriangle
	MarkerDiamond
	MarkerPlus
	MarkerCross
)

// Scatter2D renders points with configurable markers.
type Scatter2D struct {
	XY         []geom.Pt      // data space points
	Sizes      []float64      // marker sizes (radius in pixels), if nil uses Size
	Colors     []render.Color // marker colors, if nil uses Color
	EdgeColors []render.Color // edge colors for marker outlines, if nil uses EdgeColor
	Size       float64        // default marker size (radius in pixels)
	Color      render.Color   // default marker color
	EdgeColor  render.Color   // default edge color for marker outlines
	EdgeWidth  float64        // edge width in pixels (0 means no edge)
	Alpha      float64        // alpha transparency (0-1), applied to both fill and edge
	Marker     MarkerType     // marker shape
	Label      string         // series label for legend
	z          float64        // z-order
}

// Draw renders scatter points by creating filled paths for each marker.
func (s *Scatter2D) Draw(r render.Renderer, ctx *DrawContext) {
	if len(s.XY) == 0 {
		return // nothing to draw
	}

	for i, pt := range s.XY {
		// Transform to pixel coordinates
		pixelPt := ctx.DataToPixel.Apply(pt)

		// Get size for this point
		size := s.Size
		if s.Sizes != nil && i < len(s.Sizes) {
			size = s.Sizes[i]
		}

		// Get fill color for this point
		fillColor := s.Color
		if s.Colors != nil && i < len(s.Colors) {
			fillColor = s.Colors[i]
		}

		// Get edge color for this point
		edgeColor := s.EdgeColor
		if s.EdgeColors != nil && i < len(s.EdgeColors) {
			edgeColor = s.EdgeColors[i]
		}

		// Apply alpha transparency
		alpha := s.Alpha
		if alpha <= 0 {
			alpha = 1.0 // default to fully opaque
		}
		if alpha > 1 {
			alpha = 1.0 // clamp to maximum opacity
		}

		// Apply alpha to colors
		fillColor.A *= alpha
		edgeColor.A *= alpha

		// Create marker path
		markerPath := s.createMarkerPath(pixelPt, size)
		if len(markerPath.C) == 0 {
			continue // skip invalid markers
		}

		isLineMarker := s.Marker == MarkerPlus || s.Marker == MarkerCross
		lineWidth := s.EdgeWidth
		if isLineMarker && lineWidth <= 0 && ctx != nil && ctx.RC.LineWidth > 0 {
			lineWidth = ctx.RC.LineWidth
		}

		// Create paint for marker
		paint := render.Paint{
			Fill: fillColor,
		}
		if isLineMarker && edgeColor.A <= 0 {
			edgeColor = fillColor
		}
		if isLineMarker {
			paint.Fill.A = 0
		}

		// Add stroke if edge width is specified (or for line-only markers)
		if lineWidth > 0 && edgeColor.A > 0 {
			paint.Stroke = edgeColor
			paint.LineWidth = lineWidth
			paint.LineJoin = render.JoinRound
			paint.LineCap = render.CapRound
		}

		// Draw marker
		r.Path(markerPath, &paint)
	}
}

// createMarkerPath creates a filled path for the given marker type at the specified position and size.
func (s *Scatter2D) createMarkerPath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}

	scale := radius * math.Sqrt(math.Pi)

	switch s.Marker {
	case MarkerCircle:
		return s.createCirclePath(center, scale)
	case MarkerSquare:
		return s.createSquarePath(center, scale)
	case MarkerTriangle:
		return s.createTrianglePath(center, scale)
	case MarkerDiamond:
		return s.createDiamondPath(center, scale)
	case MarkerPlus:
		return s.createPlusPath(center, scale)
	case MarkerCross:
		return s.createCrossPath(center, scale)
	default:
		return s.createCirclePath(center, scale) // default to circle
	}
}

// createCirclePath creates a circular marker using a polygon approximation.
func (s *Scatter2D) createCirclePath(center geom.Pt, radius float64) geom.Path {
	const numSegments = 26 // Match matplotlib's default unit circle approximation.
	path := geom.Path{}

	for i := 0; i < numSegments; i++ {
		angle := 2 * math.Pi * float64(i) / numSegments
		x := center.X + radius*0.5*math.Cos(angle)
		y := center.Y + radius*0.5*math.Sin(angle)

		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, geom.Pt{X: x, Y: y})
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createSquarePath creates a square marker centered at the given point.
func (s *Scatter2D) createSquarePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Square vertices
	half := 0.5 * radius
	vertices := []geom.Pt{
		{X: center.X - half, Y: center.Y - half}, // bottom-left
		{X: center.X + half, Y: center.Y - half}, // bottom-right
		{X: center.X + half, Y: center.Y + half}, // top-right
		{X: center.X - half, Y: center.Y + half}, // top-left
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createTrianglePath creates an upward-pointing triangle marker.
func (s *Scatter2D) createTrianglePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Triangle vertices matching matplotlib's '^' marker geometry.
	half := 0.5 * radius
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y - half}, // top
		{X: center.X - half, Y: center.Y + half},
		{X: center.X + half, Y: center.Y + half},
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createDiamondPath creates a diamond (rotated square) marker.
func (s *Scatter2D) createDiamondPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Diamond vertices matching matplotlib's 'D' marker geometry.
	half := radius / math.Sqrt2
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y - half}, // top
		{X: center.X + half, Y: center.Y}, // right
		{X: center.X, Y: center.Y + half}, // bottom
		{X: center.X - half, Y: center.Y}, // left
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createPlusPath creates a plus sign marker.
func (s *Scatter2D) createPlusPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Plus is a line marker in matplotlib.
	half := 0.5 * radius
	hBar := []geom.Pt{
		{X: center.X - half, Y: center.Y},
		{X: center.X + half, Y: center.Y},
	}

	for i, v := range hBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	// Vertical bar
	vBar := []geom.Pt{
		{X: center.X, Y: center.Y - half},
		{X: center.X, Y: center.Y + half},
	}

	for i, v := range vBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	return path
}

// createCrossPath creates a cross (X) marker.
func (s *Scatter2D) createCrossPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// First diagonal bar (\)
	diag1 := []geom.Pt{
		{X: center.X - 0.5*radius, Y: center.Y - 0.5*radius},
		{X: center.X + 0.5*radius, Y: center.Y + 0.5*radius},
	}

	for i, v := range diag1 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	// Second diagonal bar (/)
	diag2 := []geom.Pt{
		{X: center.X - 0.5*radius, Y: center.Y + 0.5*radius},
		{X: center.X + 0.5*radius, Y: center.Y - 0.5*radius},
	}

	for i, v := range diag2 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	return path
}

// Z returns the z-order for sorting.
func (s *Scatter2D) Z() float64 {
	return s.z
}

// Bounds returns the bounding box of all points, including marker size.
func (s *Scatter2D) Bounds(*DrawContext) geom.Rect {
	if len(s.XY) == 0 {
		return geom.Rect{}
	}

	// Find the maximum size for bounds calculation
	maxSize := s.Size
	if s.Sizes != nil {
		for _, size := range s.Sizes {
			if size > maxSize {
				maxSize = size
			}
		}
	}

	// Initialize bounds with first point
	bounds := geom.Rect{
		Min: s.XY[0],
		Max: s.XY[0],
	}

	// Expand bounds to include all points
	for _, pt := range s.XY[1:] {
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}

	// Expand bounds by marker size (in data space)
	// Note: This is an approximation since marker size is in pixels
	// A more accurate implementation would need the transform context
	sizeInData := maxSize * 0.01 // rough approximation
	bounds.Min.X -= sizeInData
	bounds.Min.Y -= sizeInData
	bounds.Max.X += sizeInData
	bounds.Max.Y += sizeInData

	return bounds
}
