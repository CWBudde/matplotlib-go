package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	default3DAzimuthDeg   = -60
	default3DElevationDeg = 30
	default3DDistance     = 10
)

// Axes3D represents an Axes with basic 3D projection helpers.
//
// The underlying artist model is still 2D (`*Axes`) with pre-projected
// 3D coordinates converted into data-space 2D points before drawing.
type Axes3D struct {
	*Axes
	azimuthDeg   float64
	elevationDeg float64
	distance     float64
}

// NewAxes3D wraps an existing axes and configures 3D default view settings.
func NewAxes3D(ax *Axes) *Axes3D {
	if ax == nil {
		return nil
	}
	return &Axes3D{
		Axes:         ax,
		azimuthDeg:   default3DAzimuthDeg,
		elevationDeg: default3DElevationDeg,
		distance:     default3DDistance,
	}
}

// Plot3D projects x/y/z values and draws a line through projected points.
func (a *Axes3D) Plot3D(x, y, z []float64, opts ...PlotOptions) *Line2D {
	projected := a.projectedData(x, y, z)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	if len(opts) > 0 {
		return a.Plot(x2, y2, opts[0])
	}
	return a.Plot(x2, y2)
}

// Scatter3D projects x/y/z values and draws markers through projected points.
func (a *Axes3D) Scatter3D(x, y, z []float64, opts ...ScatterOptions) *Scatter2D {
	projected := a.projectedData(x, y, z)
	if len(projected) == 0 {
		return nil
	}

	x2, y2 := make([]float64, len(projected)), make([]float64, len(projected))
	for i, p := range projected {
		x2[i] = p.X
		y2[i] = p.Y
	}

	if len(opts) > 0 {
		return a.Scatter(x2, y2, opts[0])
	}
	return a.Scatter(x2, y2)
}

// PlotSurface creates a placeholder mesh-like line-strip by plotting the input
// sequence as connected edges.
func (a *Axes3D) PlotSurface(x, y, z []float64, opts ...PlotOptions) *Line2D {
	return a.Plot3D(x, y, z, opts...)
}

// Text3D projects a point and renders arbitrary text at the projected location.
func (a *Axes3D) Text3D(x, y, z float64, text string, opts ...TextOptions) *Text {
	if a == nil || a.Axes == nil {
		return nil
	}
	p := a.ProjectPoint(x, y, z)
	return a.Text(p.X, p.Y, text, opts...)
}

// PlotSurface creates wireframe line segments from a structured z grid.
func (a *Axes3D) PlotSurfaceGrid(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	return a.Wireframe(x, y, z, opts...)
}

// Wireframe draws a structured wireframe as line segments.
func (a *Axes3D) Wireframe(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	segments := a.projectWireframeSegments(x, y, z)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  alpha,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	return collection
}

// Contour projects a structured z grid and emits a placeholder wireframe contour.
func (a *Axes3D) Contour(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	segments := a.projectedContourSegments(x, y, z, 7)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  alpha,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	return collection
}

// Contourf projects a structured z grid and emits filled contour bands.
func (a *Axes3D) Contourf(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	color := a.NextColor()
	alpha := 1.0
	label := ""
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}
	color.A *= alpha

	contourOptions := ContourOptions{
		Color: &color,
	}

	polygons, colors := a.projectedContourFillPolygons(x, y, z, contourOptions, 7)
	if len(polygons) == 0 {
		return nil
	}

	collection := &PolyCollection{
		Polygons: polygons,
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords: Coords(CoordData),
				Label:  label,
				Alpha:  1,
			},
			FaceColors: colors,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	return collection
}

// Surface is an explicit alias for structured wireframe plotting.
func (a *Axes3D) Surface(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	return a.Wireframe(x, y, z, opts...)
}

func (a *Axes3D) projectedContourSegments(x, y []float64, z [][]float64, levelCount int) [][]geom.Pt {
	tri, values, ok := a.projectedContourTriangulation(x, y, z)
	if !ok {
		return nil
	}
	if levelCount <= 0 {
		levelCount = 7
	}

	levels := contourLevels(values, nil, levelCount, false)
	if len(levels) == 0 {
		return nil
	}

	segments := make([][]geom.Pt, 0)
	for _, level := range levels {
		for _, polyline := range stitchContourSegments(contourSegmentsForLevel(tri, values, level)) {
			if len(polyline) >= 2 {
				segments = append(segments, polyline)
			}
		}
	}
	return segments
}

func (a *Axes3D) projectedContourFillPolygons(x, y []float64, z [][]float64, opt ContourOptions, levelCount int) ([][]geom.Pt, []render.Color) {
	tri, values, ok := a.projectedContourTriangulation(x, y, z)
	if !ok {
		return nil, nil
	}
	if levelCount <= 0 {
		levelCount = 7
	}

	levels := contourLevels(values, nil, levelCount, true)
	if len(levels) < 2 {
		return nil, nil
	}

	mapping := resolveScalarMapValues(values, "", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	polygons, polygonColors := contourBandPolygons(tri, values, levels, opt, mapping, 1.0)
	if len(polygons) == 0 {
		return nil, nil
	}

	filteredPolygons := make([][]geom.Pt, 0, len(polygons))
	filteredColors := make([]render.Color, 0, len(polygonColors))
	for i, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		if i < len(polygonColors) {
			filteredColors = append(filteredColors, polygonColors[i])
			filteredPolygons = append(filteredPolygons, polygon)
		}
	}
	return filteredPolygons, filteredColors
}

func (a *Axes3D) projectedContourTriangulation(x, y []float64, z [][]float64) (Triangulation, []float64, bool) {
	if a == nil || len(z) == 0 {
		return Triangulation{}, nil, false
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return Triangulation{}, nil, false
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return Triangulation{}, nil, false
		}
	}

	points := make([]geom.Pt, 0, rows*cols)
	values := make([]float64, 0, rows*cols)
	triangles := make([][3]int, 0, (rows-1)*(cols-1)*2)
	mask := make([]bool, 0, (rows-1)*(cols-1)*2)
	index := func(row, col int) int { return row*cols + col }

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			p := a.ProjectPoint(x[col], y[row], z[row][col])
			points = append(points, p)
			values = append(values, z[row][col])
		}
	}
	for row := 0; row+1 < rows; row++ {
		for col := 0; col+1 < cols; col++ {
			p00 := index(row, col)
			p10 := index(row, col+1)
			p01 := index(row+1, col)
			p11 := index(row+1, col+1)
			triangles = append(triangles, [3]int{p00, p10, p11})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p10, p11}))
			triangles = append(triangles, [3]int{p00, p11, p01})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p11, p01}))
		}
	}

	tri := Triangulation{
		X:         make([]float64, len(points)),
		Y:         make([]float64, len(points)),
		Triangles: triangles,
		Mask:      mask,
	}
	for i, pt := range points {
		tri.X[i] = pt.X
		tri.Y[i] = pt.Y
	}
	return tri, values, true
}

// Trisurf projects a triangulated unstructured surface mesh as wireframe edges.
func (a *Axes3D) Trisurf(tri Triangulation, z []float64, opts ...PlotOptions) *LineCollection {
	if a == nil || len(tri.X) == 0 {
		return nil
	}
	if err := tri.Validate(); err != nil || len(z) != len(tri.X) {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
		}
		if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}

	edgeSet := map[[2]int]struct{}{}
	segments := make([][]geom.Pt, 0, len(tri.Triangles)*3)
	for triIdx, t := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		edges := [][2]int{
			{t[0], t[1]},
			{t[1], t[2]},
			{t[2], t[0]},
		}
		for _, edge := range edges {
			u := edge[0]
			v := edge[1]
			if u > v {
				u, v = v, u
			}
			key := [2]int{u, v}
			if _, ok := edgeSet[key]; ok {
				continue
			}
			edgeSet[key] = struct{}{}

			p0 := a.ProjectPoint(tri.X[t[0]], tri.Y[t[0]], z[t[0]])
			p1 := a.ProjectPoint(tri.X[t[1]], tri.Y[t[1]], z[t[1]])
			p2 := a.ProjectPoint(tri.X[t[2]], tri.Y[t[2]], z[t[2]])
			switch key {
			case sortedPair(t[0], t[1]):
				segments = append(segments, []geom.Pt{p0, p1})
			case sortedPair(t[1], t[2]):
				segments = append(segments, []geom.Pt{p1, p2})
			case sortedPair(t[0], t[2]):
				segments = append(segments, []geom.Pt{p0, p2})
			}
		}
	}
	if len(segments) == 0 {
		return nil
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  alpha,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	return collection
}

func sortedPair(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}

// Bar3DOptions configures projected wireframe bars.
type Bar3DOptions struct {
	Color     *render.Color
	LineWidth *float64
	Alpha     *float64
	Label     string
}

// Bar3D draws a simple projected wireframe column for each x/y/z sample.
func (a *Axes3D) Bar3D(x, y, z, dx, dy, dz []float64, opts ...Bar3DOptions) *LineCollection {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	if len(dx) < n {
		n = len(dx)
	}
	if len(dy) < n {
		n = len(dy)
	}
	if len(dz) < n {
		n = len(dz)
	}
	if n <= 0 || a == nil {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	if len(opts) > 0 {
		o := opts[0]
		if o.Color != nil {
			color = *o.Color
		}
		if o.LineWidth != nil {
			lineWidth = *o.LineWidth
		}
		if o.Alpha != nil && *o.Alpha >= 0 && *o.Alpha <= 1 {
			alpha = *o.Alpha
		}
		label = o.Label
	}

	segments := make([][]geom.Pt, 0, n*8)
	for i := 0; i < n; i++ {
		w := dx[i]
		d := dy[i]
		xc := x[i]
		yc := y[i]
		bottom := z[i]
		top := z[i] + dz[i]
		if top < bottom {
			bottom, top = top, bottom
		}

		x0 := xc - w/2
		x1 := xc + w/2
		y0 := yc - d/2
		y1 := yc + d/2

		p00 := a.ProjectPoint(x0, y0, top)
		p10 := a.ProjectPoint(x1, y0, top)
		p11 := a.ProjectPoint(x1, y1, top)
		p01 := a.ProjectPoint(x0, y1, top)
		q00 := a.ProjectPoint(x0, y0, bottom)
		q10 := a.ProjectPoint(x1, y0, bottom)
		q11 := a.ProjectPoint(x1, y1, bottom)
		q01 := a.ProjectPoint(x0, y1, bottom)

		segments = append(segments,
			[]geom.Pt{p00, p10},
			[]geom.Pt{p10, p11},
			[]geom.Pt{p11, p01},
			[]geom.Pt{p01, p00},
		)
		segments = append(segments,
			[]geom.Pt{p00, q00},
			[]geom.Pt{p10, q10},
			[]geom.Pt{p11, q11},
			[]geom.Pt{p01, q01},
		)
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  alpha,
		},
		Segments:  segments,
		Color:     color,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	return collection
}

// Voxel projects unstructured rectangular prisms as wireframe voxels.
func (a *Axes3D) Voxel(x, y, z, dx, dy, dz []float64, opts ...PlotOptions) *LineCollection {
	if len(opts) == 0 {
		return a.Bar3D(x, y, z, dx, dy, dz)
	}

	o := opts[0]
	voxelOpts := make([]Bar3DOptions, 1)
	if o.Color != nil {
		voxelOpts[0].Color = o.Color
	}
	if o.LineWidth != nil {
		voxelOpts[0].LineWidth = o.LineWidth
	}
	if o.Alpha != nil {
		voxelOpts[0].Alpha = o.Alpha
	}
	voxelOpts[0].Label = o.Label
	return a.Bar3D(x, y, z, dx, dy, dz, voxelOpts...)
}

// SetView updates the 3D viewing angles in degrees.
func (a *Axes3D) SetView(elevationDeg, azimuthDeg float64) {
	if a == nil {
		return
	}
	a.elevationDeg = elevationDeg
	a.azimuthDeg = azimuthDeg
}

// SetDistance sets the perspective distance used by the 3D projection.
// Non-positive values disable perspective scaling.
func (a *Axes3D) SetDistance(distance float64) {
	if a == nil {
		return
	}
	a.distance = distance
}

// SetDefaults sets standard Matplotlib-like defaults for elevation, azimuth,
// and perspective distance.
func (a *Axes3D) SetDefaults() {
	if a == nil {
		return
	}
	a.elevationDeg = default3DElevationDeg
	a.azimuthDeg = default3DAzimuthDeg
	a.distance = default3DDistance
}

// View reports the current 3D orientation state.
func (a *Axes3D) View() (elevationDeg, azimuthDeg, distance float64) {
	if a == nil {
		return 0, 0, 0
	}
	return a.elevationDeg, a.azimuthDeg, a.distance
}

// ProjectPoint projects a single 3D point into this Axes3D data space.
func (a *Axes3D) ProjectPoint(x, y, z float64) geom.Pt {
	if a == nil {
		return geom.Pt{}
	}
	return project3DPoint(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance)
}

func (a *Axes3D) projectedData(x, y, z []float64) []geom.Pt {
	if a == nil || a.Axes == nil {
		return nil
	}

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	if n <= 0 {
		return nil
	}

	pts := make([]geom.Pt, n)
	for i := 0; i < n; i++ {
		pts[i] = a.ProjectPoint(x[i], y[i], z[i])
	}
	return pts
}

func (a *Axes3D) projectWireframeSegments(x, y []float64, z [][]float64) [][]geom.Pt {
	if a == nil || len(z) == 0 {
		return nil
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < rows || len(y) < cols {
		return nil
	}
	for i := 1; i < rows; i++ {
		if len(z[i]) != cols {
			return nil
		}
	}

	segments := make([][]geom.Pt, 0, 2*(rows*(cols-1)+cols*(rows-1)))
	for row := 0; row < rows; row++ {
		for col := 0; col < cols-1; col++ {
			p0 := a.ProjectPoint(x[row], y[col], z[row][col])
			p1 := a.ProjectPoint(x[row], y[col+1], z[row][col+1])
			segments = append(segments, []geom.Pt{p0, p1})
		}
	}
	for col := 0; col < cols; col++ {
		for row := 0; row < rows-1; row++ {
			p0 := a.ProjectPoint(x[row], y[col], z[row][col])
			p1 := a.ProjectPoint(x[row+1], y[col], z[row+1][col])
			segments = append(segments, []geom.Pt{p0, p1})
		}
	}
	return segments
}

func project3DPoint(x, y, z, elevationDeg, azimuthDeg, distance float64) geom.Pt {
	az := azimuthDeg * math.Pi / 180
	v := elevationDeg * math.Pi / 180

	// Rotate around Z, then tilt around X.
	x2 := x*math.Cos(az) - y*math.Sin(az)
	y2 := x*math.Sin(az) + y*math.Cos(az)
	z2 := y2*math.Sin(v) + z*math.Cos(v)
	y2 = y2*math.Cos(v) - z*math.Sin(v)

	if distance <= 0 {
		return geom.Pt{X: x2, Y: y2}
	}

	s := distance / (distance - z2)
	return geom.Pt{X: x2 * s, Y: y2 * s}
}
