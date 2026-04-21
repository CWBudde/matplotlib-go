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
	MarkerPath geom.Path      // optional custom marker path in normalized collection space
	Size       float64        // default marker size (radius in pixels)
	Color      render.Color   // default marker color
	EdgeColor  render.Color   // default edge color for marker outlines
	EdgeWidth  float64        // edge width in pixels (0 means no edge)
	Alpha      float64        // alpha transparency (0-1), applied to both fill and edge
	Marker     MarkerType     // marker shape
	Label      string         // series label for legend
	z          float64        // z-order
}

var stemMarkerScale = math.Sqrt(math.Pi)

// Draw renders scatter points by creating filled paths for each marker.
func (s *Scatter2D) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || len(s.XY) == 0 {
		return
	}
	s.toPathCollection(ctx).Draw(r, ctx)
}

// createMarkerPath creates a filled path for the given marker type at the specified position and size.
func (s *Scatter2D) createMarkerPath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}
	return scaleAndTranslatePath(s.markerPrototypePath(), radius*stemMarkerScale, center)
}

func (s *Scatter2D) markerPrototypePath() geom.Path {
	if len(s.MarkerPath.C) > 0 {
		return s.MarkerPath
	}

	switch s.Marker {
	case MarkerCircle:
		return s.createCirclePath(geom.Pt{}, 1)
	case MarkerSquare:
		return s.createSquarePath(geom.Pt{}, 1)
	case MarkerTriangle:
		return s.createTrianglePath(geom.Pt{}, 1)
	case MarkerDiamond:
		return s.createDiamondPath(geom.Pt{}, 1)
	case MarkerPlus:
		return s.createPlusPath(geom.Pt{}, 1)
	case MarkerCross:
		return s.createCrossPath(geom.Pt{}, 1)
	default:
		return s.createCirclePath(geom.Pt{}, 1)
	}
}

func (s *Scatter2D) toPathCollection(ctx *DrawContext) *PathCollection {
	alpha := s.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	lineOnly := s.Marker == MarkerPlus || s.Marker == MarkerCross
	lineWidth := s.EdgeWidth
	if lineOnly && lineWidth <= 0 && ctx != nil && ctx.RC.LineWidth > 0 {
		lineWidth = ctx.RC.LineWidth
	}

	sizes := make([]float64, len(s.XY))
	for i := range s.XY {
		size := s.Size
		if len(s.Sizes) > 0 && i < len(s.Sizes) {
			size = s.Sizes[i]
		}
		sizes[i] = size * stemMarkerScale
	}

	faceColor := s.Color
	edgeColor := s.EdgeColor
	if lineOnly && edgeColor.A <= 0 {
		edgeColor = faceColor
	}

	return &PathCollection{
		Collection: Collection{
			Label: s.Label,
			Alpha: alpha,
			z:     s.z,
		},
		Path:          s.markerPrototypePath(),
		Offsets:       append([]geom.Pt(nil), s.XY...),
		Sizes:         sizes,
		PathInDisplay: true,
		FaceColors:    append([]render.Color(nil), s.Colors...),
		FaceColor:     faceColor,
		EdgeColors:    append([]render.Color(nil), s.EdgeColors...),
		EdgeColor:     edgeColor,
		EdgeWidth:     lineWidth,
		LineOnly:      lineOnly,
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
