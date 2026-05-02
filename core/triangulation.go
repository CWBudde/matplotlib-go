package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Triangulation stores an unstructured triangular mesh.
type Triangulation struct {
	X         []float64
	Y         []float64
	Triangles [][3]int
	Mask      []bool
}

// TriPlotOptions configures triplot rendering.
type TriPlotOptions struct {
	Color     *render.Color
	LineWidth *float64
	Alpha     *float64
	Label     string
}

// TriColorOptions configures tripcolor rendering.
type TriColorOptions struct {
	Colormap  *string
	VMin      *float64
	VMax      *float64
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Label     string
}

// Validate verifies that the triangulation references valid point indices.
func (t Triangulation) Validate() error {
	if len(t.X) == 0 || len(t.Y) == 0 {
		return fmt.Errorf("triangulation requires coordinates")
	}
	if len(t.X) != len(t.Y) {
		return fmt.Errorf("triangulation X/Y lengths differ")
	}
	for triIdx, tri := range t.Triangles {
		for _, idx := range tri {
			if idx < 0 || idx >= len(t.X) {
				return fmt.Errorf("triangle %d references point %d outside 0..%d", triIdx, idx, len(t.X)-1)
			}
		}
	}
	if len(t.Mask) > 0 && len(t.Mask) != len(t.Triangles) {
		return fmt.Errorf("triangulation mask length %d does not match triangles %d", len(t.Mask), len(t.Triangles))
	}
	return nil
}

// TriPlot draws the unique edges of the supplied triangulation.
func (a *Axes) TriPlot(tri Triangulation, opts ...TriPlotOptions) *LineCollection {
	if err := tri.Validate(); err != nil || len(tri.Triangles) == 0 {
		return nil
	}

	var opt TriPlotOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	color := a.NextColor()
	if opt.Color != nil {
		color = *opt.Color
	}
	alpha := meshAlpha(opt.Alpha)
	color.A *= alpha

	width := 1.0
	if opt.LineWidth != nil {
		width = *opt.LineWidth
	}

	edgeSet := map[[2]int]struct{}{}
	segments := make([][]geom.Pt, 0, len(tri.Triangles)*3)
	for triIdx, triangle := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		edges := [][2]int{
			sortedEdge(triangle[0], triangle[1]),
			sortedEdge(triangle[1], triangle[2]),
			sortedEdge(triangle[2], triangle[0]),
		}
		for _, edge := range edges {
			if _, exists := edgeSet[edge]; exists {
				continue
			}
			edgeSet[edge] = struct{}{}
			segments = append(segments, []geom.Pt{tri.point(edge[0]), tri.point(edge[1])})
		}
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  opt.Label,
			Alpha:  1,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: width,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	return collection
}

// TriColor draws per-triangle colored polygons over a triangulation. Values
// may be provided per triangle or per point; point values are averaged onto
// each triangle.
func (a *Axes) TriColor(tri Triangulation, values []float64, opts ...TriColorOptions) *PolyCollection {
	if err := tri.Validate(); err != nil || len(tri.Triangles) == 0 {
		return nil
	}

	var opt TriColorOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	triangleValues, ok := triColorValues(tri, values)
	if !ok {
		return nil
	}

	cmap := ""
	if opt.Colormap != nil {
		cmap = *opt.Colormap
	}
	mapping := resolveScalarMapValues(triangleValues, cmap, opt.VMin, opt.VMax)
	alpha := meshAlpha(opt.Alpha)

	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}
	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	polygons := make([][]geom.Pt, 0, len(tri.Triangles))
	faceColors := make([]render.Color, 0, len(tri.Triangles))
	edgeColors := make([]render.Color, 0, len(tri.Triangles))
	for triIdx, triangle := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		polygons = append(polygons, []geom.Pt{
			tri.point(triangle[0]),
			tri.point(triangle[1]),
			tri.point(triangle[2]),
		})
		value := triangleValues[triIdx]
		if !isFinite(value) {
			faceColors = append(faceColors, render.Color{})
			edgeColors = append(edgeColors, render.Color{})
			continue
		}
		faceColors = append(faceColors, mapping.Color(value, alpha))
		edgeColors = append(edgeColors, edgeColor)
	}

	collection := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:   Coords(CoordData),
				Label:    opt.Label,
				Alpha:    1,
				Colormap: mapping.Colormap,
				VMin:     mapping.VMin,
				VMax:     mapping.VMax,
			},
			FaceColors: faceColors,
			EdgeColors: edgeColors,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
		Polygons: polygons,
	}
	a.Add(collection)
	return collection
}

func sortedEdge(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}

func (t Triangulation) point(idx int) geom.Pt {
	return geom.Pt{X: t.X[idx], Y: t.Y[idx]}
}

func (t Triangulation) masked(triIdx int) bool {
	return len(t.Mask) > 0 && triIdx < len(t.Mask) && t.Mask[triIdx]
}

func triColorValues(tri Triangulation, values []float64) ([]float64, bool) {
	switch len(values) {
	case len(tri.Triangles):
		out := make([]float64, len(values))
		copy(out, values)
		return out, true
	case len(tri.X):
		out := make([]float64, len(tri.Triangles))
		for i, triangle := range tri.Triangles {
			out[i] = meshValueAverage([]float64{
				values[triangle[0]],
				values[triangle[1]],
				values[triangle[2]],
			})
		}
		return out, true
	default:
		return nil, false
	}
}
