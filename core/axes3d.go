package core

import (
	"math"
	"sort"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	default3DAzimuthDeg   = -60
	default3DElevationDeg = 30
	default3DDistance     = 10
	default3DFocalLength  = 1
	default3DDataMargin   = 1.0 / 48.0
	default3DViewMin      = -0.095
	default3DViewMax      = 0.09
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
	hasData      bool
	dataMin      vec3
	dataMax      vec3
	reprojectors []func()
}

type axes3DFrame struct {
	axes *Axes3D
}

func (f *axes3DFrame) Draw(r render.Renderer, ctx *DrawContext) {
	if f == nil || f.axes == nil || r == nil || ctx == nil {
		return
	}
	mins, maxs := f.axes.projectionLimits()
	if !f.axes.hasData {
		mins, maxs = vec3{0, 0, 0}, vec3{1, 1, 1}
	}

	panes := [][]geom.Pt{
		{
			f.axes.ProjectPoint(mins[0], mins[1], mins[2]),
			f.axes.ProjectPoint(maxs[0], mins[1], mins[2]),
			f.axes.ProjectPoint(maxs[0], maxs[1], mins[2]),
			f.axes.ProjectPoint(mins[0], maxs[1], mins[2]),
		},
		{
			f.axes.ProjectPoint(mins[0], maxs[1], mins[2]),
			f.axes.ProjectPoint(maxs[0], maxs[1], mins[2]),
			f.axes.ProjectPoint(maxs[0], maxs[1], maxs[2]),
			f.axes.ProjectPoint(mins[0], maxs[1], maxs[2]),
		},
		{
			f.axes.ProjectPoint(mins[0], mins[1], mins[2]),
			f.axes.ProjectPoint(mins[0], maxs[1], mins[2]),
			f.axes.ProjectPoint(mins[0], maxs[1], maxs[2]),
			f.axes.ProjectPoint(mins[0], mins[1], maxs[2]),
		},
	}
	(&PolyCollection{
		Polygons: panes,
		PatchCollection: PatchCollection{
			Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
			FaceColor:  render.Color{R: 0.96, G: 0.96, B: 0.96, A: 0.55},
			EdgeColor:  render.Color{A: 0},
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}).Draw(r, ctx)

	segments := f.axes.frameSegments(mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   segments,
		Color:      render.Color{R: 0.70, G: 0.70, B: 0.70, A: 1},
		LineWidth:  0.6,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)
}

func (f *axes3DFrame) Z() float64 { return -1000 }

func (f *axes3DFrame) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func (f *axes3DFrame) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if f == nil || f.axes == nil || r == nil || ctx == nil {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}
	mins, maxs := f.axes.projectionLimits()
	if !f.axes.hasData {
		mins, maxs = vec3{0, 0, 0}, vec3{1, 1, 1}
	}
	f.axes.draw3DTickLabels(textRen, r, ctx, mins, maxs)
	f.axes.draw3DAxisLabels(textRen, r, ctx, mins, maxs)
}

// NewAxes3D wraps an existing axes and configures 3D default view settings.
func NewAxes3D(ax *Axes) *Axes3D {
	if ax == nil {
		return nil
	}
	axes := &Axes3D{
		Axes:         ax,
		azimuthDeg:   default3DAzimuthDeg,
		elevationDeg: default3DElevationDeg,
		distance:     default3DDistance,
	}
	ax.Add(&axes3DFrame{axes: axes})
	return axes
}

// Plot3D projects x/y/z values and draws a line through projected points.
func (a *Axes3D) Plot3D(x, y, z []float64, opts ...PlotOptions) *Line2D {
	limitsChanged := a.observe3DData(x, y, z)
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
		line := a.Plot(x2, y2, opts[0])
		a.add3DReprojector(func() {
			reprojectLine3D(line, a.projectedData(x, y, z))
		}, limitsChanged)
		return line
	}
	line := a.Plot(x2, y2)
	a.add3DReprojector(func() {
		reprojectLine3D(line, a.projectedData(x, y, z))
	}, limitsChanged)
	return line
}

// Scatter3D projects x/y/z values and draws markers through projected points.
func (a *Axes3D) Scatter3D(x, y, z []float64, opts ...ScatterOptions) *Scatter2D {
	limitsChanged := a.observe3DData(x, y, z)
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
		scatter := a.Scatter(x2, y2, opts[0])
		a.add3DReprojector(func() {
			reprojectScatter3D(scatter, a.projectedData(x, y, z))
		}, limitsChanged)
		return scatter
	}
	scatter := a.Scatter(x2, y2)
	a.add3DReprojector(func() {
		reprojectScatter3D(scatter, a.projectedData(x, y, z))
	}, limitsChanged)
	return scatter
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
	limitsChanged := a.observe3DPoint(x, y, z)
	p := a.ProjectPoint(x, y, z)
	txt := a.Text(p.X, p.Y, text, opts...)
	a.add3DReprojector(func() {
		if txt != nil {
			txt.Position = a.ProjectPoint(x, y, z)
		}
	}, limitsChanged)
	return txt
}

// PlotSurfaceGrid creates a filled surface from a structured z grid.
func (a *Axes3D) PlotSurfaceGrid(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	return a.Surface(x, y, z, opts...)
}

// Wireframe draws a structured wireframe as line segments.
func (a *Axes3D) Wireframe(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
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
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectWireframeSegments(x, y, z)
		}
	}, limitsChanged)
	return collection
}

// Contour projects a structured z grid and emits a placeholder wireframe contour.
func (a *Axes3D) Contour(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
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
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectedContourSegments(x, y, z, 7)
		}
	}, limitsChanged)
	return collection
}

// Contourf projects a structured z grid and emits filled contour bands.
func (a *Axes3D) Contourf(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	alpha := 0.45
	label := ""
	colorOverride := (*render.Color)(nil)
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			colorOverride = opt.Color
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}

	polygons, colors := a.projectedContourFloorPolygons(x, y, z, alpha, colorOverride)
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
	a.add3DReprojector(func() {
		if collection != nil {
			polygons, colors := a.projectedContourFloorPolygons(x, y, z, alpha, colorOverride)
			collection.Polygons = polygons
			collection.FaceColors = colors
		}
	}, limitsChanged)
	return collection
}

// Surface draws a structured surface as projected, z-sorted quadrilateral faces.
func (a *Axes3D) Surface(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	polygons, faceColors := a.projectSurfacePolygons(x, y, z, opts...)
	if len(polygons) == 0 {
		return nil
	}

	alpha := 0.85
	label := ""
	edgeWidth := 0.35
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		if opt.LineWidth != nil && *opt.LineWidth >= 0 {
			edgeWidth = *opt.LineWidth
		}
		label = opt.Label
	}
	for i := range faceColors {
		faceColors[i].A *= alpha
	}

	collection := &PolyCollection{
		Polygons: polygons,
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords: Coords(CoordData),
				Label:  label,
				Alpha:  1,
			},
			FaceColors: faceColors,
			EdgeColor:  render.Color{R: 0.20, G: 0.20, B: 0.20, A: 0.35},
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			polygons, faceColors := a.projectSurfacePolygons(x, y, z, opts...)
			for i := range faceColors {
				faceColors[i].A *= alpha
			}
			collection.Polygons = polygons
			collection.FaceColors = faceColors
		}
	}, limitsChanged)
	return collection
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

type surfaceFace struct {
	polygon []geom.Pt
	value   float64
	depth   float64
}

func (a *Axes3D) projectSurfacePolygons(x, y []float64, z [][]float64, opts ...PlotOptions) ([][]geom.Pt, []render.Color) {
	if a == nil || len(z) == 0 {
		return nil, nil
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil
		}
	}

	faces := make([]surfaceFace, 0, (rows-1)*(cols-1))
	values := make([]float64, 0, (rows-1)*(cols-1))
	for row := 0; row+1 < rows; row++ {
		for col := 0; col+1 < cols; col++ {
			corners := [4][3]float64{
				{x[col], y[row], z[row][col]},
				{x[col+1], y[row], z[row][col+1]},
				{x[col+1], y[row+1], z[row+1][col+1]},
				{x[col], y[row+1], z[row+1][col]},
			}
			polygon := make([]geom.Pt, 0, len(corners))
			value := 0.0
			depth := 0.0
			valid := true
			for _, c := range corners {
				if !isFinite3D(c[0], c[1], c[2]) {
					valid = false
					break
				}
				pt, zDepth := a.projectPointDepth(c[0], c[1], c[2])
				polygon = append(polygon, pt)
				value += c[2]
				depth += zDepth
			}
			if !valid {
				continue
			}
			value /= float64(len(corners))
			depth /= float64(len(corners))
			faces = append(faces, surfaceFace{polygon: polygon, value: value, depth: depth})
			values = append(values, value)
		}
	}
	if len(faces) == 0 {
		return nil, nil
	}

	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})

	colorOverride := (*render.Color)(nil)
	if len(opts) > 0 && opts[0].Color != nil {
		colorOverride = opts[0].Color
	}
	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	cmap := matcolor.GetColormap(mapping.Colormap)
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	for i, face := range faces {
		polygons[i] = face.polygon
		if colorOverride != nil {
			colors[i] = *colorOverride
			continue
		}
		colors[i] = cmap.At(mapping.Normalize(face.value))
	}
	return polygons, colors
}

func (a *Axes3D) projectedContourFloorPolygons(x, y []float64, z [][]float64, alpha float64, colorOverride *render.Color) ([][]geom.Pt, []render.Color) {
	if a == nil || len(z) == 0 {
		return nil, nil
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil
		}
	}

	zMin, zMax := zGridRange(z)
	if zMin == zMax {
		zMin -= 0.5
		zMax += 0.5
	}
	floorZ := zMin - 0.2*(zMax-zMin)

	type floorCell struct {
		polygon []geom.Pt
		value   float64
	}
	cells := make([]floorCell, 0, (rows-1)*(cols-1))
	values := make([]float64, 0, (rows-1)*(cols-1))
	for row := 0; row+1 < rows; row++ {
		for col := 0; col+1 < cols; col++ {
			value := (z[row][col] + z[row][col+1] + z[row+1][col+1] + z[row+1][col]) / 4
			polygon := []geom.Pt{
				a.ProjectPoint(x[col], y[row], floorZ),
				a.ProjectPoint(x[col+1], y[row], floorZ),
				a.ProjectPoint(x[col+1], y[row+1], floorZ),
				a.ProjectPoint(x[col], y[row+1], floorZ),
			}
			cells = append(cells, floorCell{polygon: polygon, value: value})
			values = append(values, value)
		}
	}
	if len(cells) == 0 {
		return nil, nil
	}

	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	cmap := matcolor.GetColormap(mapping.Colormap)
	polygons := make([][]geom.Pt, len(cells))
	colors := make([]render.Color, len(cells))
	for i, cell := range cells {
		polygons[i] = cell.polygon
		if colorOverride != nil {
			colors[i] = *colorOverride
		} else {
			colors[i] = cmap.At(mapping.Normalize(cell.value))
		}
		colors[i].A *= alpha
	}
	return polygons, colors
}

func zGridRange(z [][]float64) (float64, float64) {
	minVal, maxVal := 0.0, 0.0
	first := true
	for _, row := range z {
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				continue
			}
			if first {
				minVal, maxVal = v, v
				first = false
				continue
			}
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}
	return minVal, maxVal
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
	limitsChanged := a.observe3DTriangulation(tri, z)

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

	segments := a.projectTriangulationEdges(tri, z)
	if len(segments) == 0 {
		return nil
	}
	faces := a.projectTriangulationFaces(tri, z)
	var faceCollection *PolyCollection
	if len(faces) > 0 {
		faceColor := color
		faceColor.A *= alpha
		faceCollection = &PolyCollection{
			Polygons: faces,
			PatchCollection: PatchCollection{
				Collection: Collection{
					Coords: Coords(CoordData),
					Label:  label,
					Alpha:  1,
				},
				FaceColor: faceColor,
				EdgeColor: render.Color{R: color.R, G: color.G, B: color.B, A: 0.35 * alpha},
				EdgeWidth: lineWidth,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
			},
		}
		a.Add(faceCollection)
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
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectTriangulationEdges(tri, z)
		}
		if faceCollection != nil {
			faceCollection.Polygons = a.projectTriangulationFaces(tri, z)
		}
	}, limitsChanged)
	return collection
}

func sortedPair(a, b int) [2]int {
	if a < b {
		return [2]int{a, b}
	}
	return [2]int{b, a}
}

func reprojectLine3D(line *Line2D, points []geom.Pt) {
	if line == nil {
		return
	}
	line.XY = append(line.XY[:0], points...)
}

func reprojectScatter3D(scatter *Scatter2D, points []geom.Pt) {
	if scatter == nil {
		return
	}
	scatter.XY = append(scatter.XY[:0], points...)
}

func (a *Axes3D) projectTriangulationEdges(tri Triangulation, z []float64) [][]geom.Pt {
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
	return segments
}

func (a *Axes3D) projectTriangulationFaces(tri Triangulation, z []float64) [][]geom.Pt {
	type triFace struct {
		polygon []geom.Pt
		depth   float64
	}
	faces := make([]triFace, 0, len(tri.Triangles))
	for triIdx, t := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		polygon := make([]geom.Pt, 0, 3)
		depth := 0.0
		valid := true
		for _, idx := range t {
			if idx < 0 || idx >= len(tri.X) || idx >= len(tri.Y) || idx >= len(z) || !isFinite3D(tri.X[idx], tri.Y[idx], z[idx]) {
				valid = false
				break
			}
			pt, zDepth := a.projectPointDepth(tri.X[idx], tri.Y[idx], z[idx])
			polygon = append(polygon, pt)
			depth += zDepth
		}
		if valid {
			faces = append(faces, triFace{polygon: polygon, depth: depth / 3})
		}
	}
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})
	polygons := make([][]geom.Pt, len(faces))
	for i, face := range faces {
		polygons[i] = face.polygon
	}
	return polygons
}

func (a *Axes3D) projectBar3DSegments(x, y, z, dx, dy, dz []float64) [][]geom.Pt {
	n := minLen(x, y, z, dx, dy, dz)
	segments := make([][]geom.Pt, 0, n*8)
	for i := 0; i < n; i++ {
		x0 := x[i]
		x1 := x[i] + dx[i]
		y0 := y[i]
		y1 := y[i] + dy[i]
		bottom := z[i]
		top := z[i] + dz[i]
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		if top < bottom {
			bottom, top = top, bottom
		}

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
			[]geom.Pt{p00, q00},
			[]geom.Pt{p10, q10},
			[]geom.Pt{p11, q11},
			[]geom.Pt{p01, q01},
		)
	}
	return segments
}

func (a *Axes3D) projectBar3DFaces(x, y, z, dx, dy, dz []float64) [][]geom.Pt {
	polygons, _ := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, render.Color{R: 1, G: 1, B: 1, A: 1})
	return polygons
}

func (a *Axes3D) projectBar3DShadedFaces(x, y, z, dx, dy, dz []float64, baseColor render.Color) ([][]geom.Pt, []render.Color) {
	type face struct {
		polygon []geom.Pt
		color   render.Color
		depth   float64
	}
	n := minLen(x, y, z, dx, dy, dz)
	faces := make([]face, 0, n*6)
	for i := 0; i < n; i++ {
		x0 := x[i]
		x1 := x[i] + dx[i]
		y0 := y[i]
		y1 := y[i] + dy[i]
		z0 := z[i]
		z1 := z[i] + dz[i]
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		if z1 < z0 {
			z0, z1 = z1, z0
		}
		corners := [8][3]float64{
			{x0, y0, z0}, {x1, y0, z0}, {x1, y1, z0}, {x0, y1, z0},
			{x0, y0, z1}, {x1, y0, z1}, {x1, y1, z1}, {x0, y1, z1},
		}
		faceIndices := [][4]int{
			{0, 1, 2, 3},
			{4, 5, 6, 7},
			{0, 1, 5, 4},
			{1, 2, 6, 5},
			{2, 3, 7, 6},
			{3, 0, 4, 7},
		}
		normals := []vec3{
			{0, 0, -1},
			{0, 0, 1},
			{0, -1, 0},
			{1, 0, 0},
			{0, 1, 0},
			{-1, 0, 0},
		}
		for faceIdx, indices := range faceIndices {
			polygon := make([]geom.Pt, 0, len(indices))
			depth := 0.0
			for _, idx := range indices {
				c := corners[idx]
				pt, zDepth := a.projectPointDepth(c[0], c[1], c[2])
				polygon = append(polygon, pt)
				depth += zDepth
			}
			faces = append(faces, face{
				polygon: polygon,
				color:   shade3DFaceColor(baseColor, normals[faceIdx]),
				depth:   depth / float64(len(indices)),
			})
		}
	}
	sort.SliceStable(faces, func(i, j int) bool {
		return faces[i].depth > faces[j].depth
	})
	polygons := make([][]geom.Pt, len(faces))
	colors := make([]render.Color, len(faces))
	for i, face := range faces {
		polygons[i] = face.polygon
		colors[i] = face.color
	}
	return polygons, colors
}

func shade3DFaceColor(color render.Color, normal vec3) render.Color {
	// Matplotlib art3d._shade_colors maps the light-source dot product from
	// [-1, 1] to [0.3, 1.0] and preserves alpha.
	az := (90.0 - 225.0) * math.Pi / 180
	alt := 19.4712 * math.Pi / 180
	light := vec3{
		math.Cos(az) * math.Cos(alt),
		math.Sin(az) * math.Cos(alt),
		math.Sin(alt),
	}
	shade := 0.65 + 0.35*normal.unit().dot(light)
	if math.IsNaN(shade) {
		shade = 0.65
	}
	shade = math.Max(0.3, math.Min(1, shade))
	color.R *= shade
	color.G *= shade
	color.B *= shade
	return color
}

func (a *Axes3D) frameSegments(mins, maxs vec3) [][]geom.Pt {
	corner := func(xi, yi, zi int) geom.Pt {
		x := mins[0]
		if xi == 1 {
			x = maxs[0]
		}
		y := mins[1]
		if yi == 1 {
			y = maxs[1]
		}
		z := mins[2]
		if zi == 1 {
			z = maxs[2]
		}
		return a.ProjectPoint(x, y, z)
	}

	edges := [][2][3]int{
		{{0, 0, 0}, {1, 0, 0}}, {{1, 0, 0}, {1, 1, 0}}, {{1, 1, 0}, {0, 1, 0}}, {{0, 1, 0}, {0, 0, 0}},
		{{0, 0, 1}, {1, 0, 1}}, {{1, 0, 1}, {1, 1, 1}}, {{1, 1, 1}, {0, 1, 1}}, {{0, 1, 1}, {0, 0, 1}},
		{{0, 0, 0}, {0, 0, 1}}, {{1, 0, 0}, {1, 0, 1}}, {{1, 1, 0}, {1, 1, 1}}, {{0, 1, 0}, {0, 1, 1}},
	}
	segments := make([][]geom.Pt, 0, len(edges)+36)
	for _, edge := range edges {
		p0 := corner(edge[0][0], edge[0][1], edge[0][2])
		p1 := corner(edge[1][0], edge[1][1], edge[1][2])
		segments = append(segments, []geom.Pt{p0, p1})
	}

	for _, x := range frameTicks(mins[0], maxs[0], 6) {
		segments = append(segments,
			[]geom.Pt{a.ProjectPoint(x, mins[1], mins[2]), a.ProjectPoint(x, maxs[1], mins[2])},
			[]geom.Pt{a.ProjectPoint(x, maxs[1], mins[2]), a.ProjectPoint(x, maxs[1], maxs[2])},
		)
	}
	for _, y := range frameTicks(mins[1], maxs[1], 6) {
		segments = append(segments,
			[]geom.Pt{a.ProjectPoint(mins[0], y, mins[2]), a.ProjectPoint(maxs[0], y, mins[2])},
			[]geom.Pt{a.ProjectPoint(mins[0], y, mins[2]), a.ProjectPoint(mins[0], y, maxs[2])},
		)
	}
	for _, z := range frameTicks(mins[2], maxs[2], 6) {
		segments = append(segments,
			[]geom.Pt{a.ProjectPoint(mins[0], mins[1], z), a.ProjectPoint(mins[0], maxs[1], z)},
			[]geom.Pt{a.ProjectPoint(mins[0], maxs[1], z), a.ProjectPoint(maxs[0], maxs[1], z)},
		)
	}
	return segments
}

func frameTicks(minVal, maxVal float64, count int) []float64 {
	if count <= 2 || minVal == maxVal {
		return nil
	}
	ticks := make([]float64, 0, count-2)
	for i := 1; i+1 < count; i++ {
		t := float64(i) / float64(count-1)
		ticks = append(ticks, minVal+(maxVal-minVal)*t)
	}
	return ticks
}

func frameTickValues(minVal, maxVal float64, count int) []float64 {
	if count <= 1 || minVal == maxVal {
		return []float64{minVal}
	}
	ticks := make([]float64, count)
	for i := range count {
		t := float64(i) / float64(count-1)
		ticks[i] = minVal + (maxVal-minVal)*t
	}
	return ticks
}

func (a *Axes3D) draw3DTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, mins, maxs vec3) {
	fontSize := ctx.RC.TickLabelSize("x")
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	center := a.ProjectPoint((mins[0]+maxs[0])/2, (mins[1]+maxs[1])/2, (mins[2]+maxs[2])/2)

	xTicks := frameAxisTicks(mins[0], maxs[0])
	for i, tick := range xTicks {
		a.draw3DLabelAt(textRen, r, ctx, format3DTick(tick, i, xTicks), a.ProjectPoint(tick, mins[1], mins[2]), center, 13, fontSize, textColor)
	}
	yTicks := frameAxisTicks(mins[1], maxs[1])
	for i, tick := range yTicks {
		if i == 0 || i == len(yTicks)-1 {
			continue
		}
		a.draw3DLabelAt(textRen, r, ctx, format3DTick(tick, i, yTicks), a.ProjectPoint(maxs[0], tick, mins[2]), center, 13, fontSize, textColor)
	}
	zTicks := frameAxisTicks(mins[2], maxs[2])
	for i, tick := range zTicks {
		a.draw3DLabelAt(textRen, r, ctx, format3DTick(tick, i, zTicks), a.ProjectPoint(maxs[0], maxs[1], tick), center, 13, fontSize, textColor)
	}
}

func (a *Axes3D) draw3DAxisLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, mins, maxs vec3) {
	fontSize := axisLabelFontSize(ctx)
	textColor := ctx.RC.DefaultAxesLabelColor()
	center := a.ProjectPoint((mins[0]+maxs[0])/2, (mins[1]+maxs[1])/2, (mins[2]+maxs[2])/2)
	if a.XLabel != "" {
		pos := a.ProjectPoint((mins[0]+maxs[0])/2, mins[1], mins[2])
		a.draw3DLabelAt(textRen, r, ctx, a.XLabel, pos, center, 34, fontSize, textColor)
	}
	if a.YLabel != "" {
		pos := a.ProjectPoint(maxs[0], (mins[1]+maxs[1])/2, mins[2])
		a.draw3DLabelAt(textRen, r, ctx, a.YLabel, pos, center, 34, fontSize, textColor)
	}
}

func (a *Axes3D) draw3DLabelAt(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, label string, projected, projectedCenter geom.Pt, pad, fontSize float64, textColor render.Color) {
	if label == "" {
		return
	}
	tr := ctx.TransformFor(Coords(CoordData))
	anchor := tr.Apply(projected)
	center := tr.Apply(projectedCenter)
	dx := anchor.X - center.X
	dy := anchor.Y - center.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		dy = -1
		length = 1
	}
	anchor.X += pad * dx / length
	anchor.Y += pad * dy / length
	layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
	origin := alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignCenter)
	drawDisplayText(textRen, label, origin, fontSize, textColor, ctx.RC.FontKey)
}

func frameAxisTicks(minVal, maxVal float64) []float64 {
	ticks := AutoLocator{}.Ticks(minVal, maxVal, 8)
	if len(ticks) == 0 {
		return nil
	}
	eps := 1e-12 * math.Max(1, math.Abs(maxVal-minVal))
	filtered := ticks[:0]
	for _, tick := range ticks {
		if tick >= minVal-eps && tick <= maxVal+eps {
			filtered = append(filtered, tick)
		}
	}
	return filtered
}

func format3DTick(v float64, i int, ticks []float64) string {
	step := 0.0
	if len(ticks) > 1 {
		if i+1 < len(ticks) {
			step = math.Abs(ticks[i+1] - ticks[i])
		} else {
			step = math.Abs(ticks[i] - ticks[i-1])
		}
	}
	return formatScalarTickLabel(ScalarFormatter{Prec: 3}, v, step)
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
	n := minLen(x, y, z, dx, dy, dz)
	if n <= 0 || a == nil {
		return nil
	}
	limitsChanged := a.observe3DBarData(x, y, z, dx, dy, dz)

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

	faceColor := color
	if len(opts) > 0 && opts[0].Alpha != nil {
		faceColor.A *= alpha
	} else {
		faceColor.A *= 0.7
	}
	faces, faceColors := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, faceColor)
	if len(faces) > 0 {
		faceCollection := &PolyCollection{
			Polygons: faces,
			PatchCollection: PatchCollection{
				Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
				FaceColors: faceColors,
				EdgeColor:  render.Color{A: 0},
				LineJoin:   render.JoinMiter,
				LineCap:    render.CapButt,
			},
		}
		a.Add(faceCollection)
		a.add3DReprojector(func() {
			if faceCollection != nil {
				faces, faceColors := a.projectBar3DShadedFaces(x, y, z, dx, dy, dz, faceColor)
				faceCollection.Polygons = faces
				faceCollection.FaceColors = faceColors
			}
		}, limitsChanged)
	}

	segments := a.projectBar3DSegments(x, y, z, dx, dy, dz)

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
	a.add3DReprojector(func() {
		if collection != nil {
			collection.Segments = a.projectBar3DSegments(x, y, z, dx, dy, dz)
		}
	}, limitsChanged)
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
	mins, maxs := a.projectionLimits()
	return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs)
}

func (a *Axes3D) projectPointDepth(x, y, z float64) (geom.Pt, float64) {
	if a == nil {
		return geom.Pt{}, 0
	}
	if a.distance <= 0 {
		return a.ProjectPoint(x, y, z), z
	}
	mins, maxs := a.projectionLimits()
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs)
	tx, ty, tz := transform3DPoint(m, x, y, z)
	return geom.Pt{X: tx, Y: ty}, tz
}

func (a *Axes3D) projectionLimits() (vec3, vec3) {
	if a == nil || !a.hasData {
		return vec3{0, 0, 0}, vec3{1, 1, 1}
	}
	mins := a.dataMin
	maxs := a.dataMax
	for i := range 3 {
		if mins[i] == maxs[i] {
			mins[i] -= 0.5
			maxs[i] += 0.5
		}
		margin := (maxs[i] - mins[i]) * default3DDataMargin
		mins[i] -= margin
		maxs[i] += margin
	}
	return mins, maxs
}

func (a *Axes3D) observe3DPoint(x, y, z float64) bool {
	if a == nil || !isFinite3D(x, y, z) {
		return false
	}
	p := vec3{x, y, z}
	if !a.hasData {
		a.dataMin = p
		a.dataMax = p
		a.hasData = true
		return true
	}
	changed := false
	for i := range 3 {
		if p[i] < a.dataMin[i] {
			a.dataMin[i] = p[i]
			changed = true
		}
		if p[i] > a.dataMax[i] {
			a.dataMax[i] = p[i]
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) observe3DData(x, y, z []float64) bool {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if len(z) < n {
		n = len(z)
	}
	changed := false
	for i := 0; i < n; i++ {
		if a.observe3DPoint(x[i], y[i], z[i]) {
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) observe3DGrid(x, y []float64, z [][]float64) bool {
	if a == nil || len(z) == 0 {
		return false
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return false
	}
	changed := false
	for row := 0; row < rows; row++ {
		if len(z[row]) != cols {
			return changed
		}
		for col := 0; col < cols; col++ {
			if a.observe3DPoint(x[col], y[row], z[row][col]) {
				changed = true
			}
		}
	}
	return changed
}

func (a *Axes3D) observe3DTriangulation(tri Triangulation, z []float64) bool {
	n := len(tri.X)
	if len(tri.Y) < n {
		n = len(tri.Y)
	}
	if len(z) < n {
		n = len(z)
	}
	changed := false
	for i := 0; i < n; i++ {
		if a.observe3DPoint(tri.X[i], tri.Y[i], z[i]) {
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) observe3DBarData(x, y, z, dx, dy, dz []float64) bool {
	n := minLen(x, y, z, dx, dy, dz)
	changed := false
	for i := 0; i < n; i++ {
		x0, x1 := x[i], x[i]+dx[i]
		y0, y1 := y[i], y[i]+dy[i]
		z0, z1 := z[i], z[i]+dz[i]
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		if z1 < z0 {
			z0, z1 = z1, z0
		}
		if a.observe3DPoint(x0, y0, z0) {
			changed = true
		}
		if a.observe3DPoint(x1, y1, z1) {
			changed = true
		}
	}
	return changed
}

func (a *Axes3D) add3DReprojector(reproject func(), limitsChanged bool) {
	if a == nil || reproject == nil {
		return
	}
	a.reprojectors = append(a.reprojectors, reproject)
	if limitsChanged {
		a.reproject3DArtists()
	}
}

func (a *Axes3D) reproject3DArtists() {
	if a == nil {
		return
	}
	for _, reproject := range a.reprojectors {
		reproject()
	}
}

func isFinite3D(x, y, z float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsNaN(z) &&
		!math.IsInf(x, 0) && !math.IsInf(y, 0) && !math.IsInf(z, 0)
}

func minLen(slices ...[]float64) int {
	if len(slices) == 0 {
		return 0
	}
	n := len(slices[0])
	for _, s := range slices[1:] {
		if len(s) < n {
			n = len(s)
		}
	}
	return n
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
	if cols == 0 || len(x) < cols || len(y) < rows {
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
			p0 := a.ProjectPoint(x[col], y[row], z[row][col])
			p1 := a.ProjectPoint(x[col+1], y[row], z[row][col+1])
			segments = append(segments, []geom.Pt{p0, p1})
		}
	}
	for col := 0; col < cols; col++ {
		for row := 0; row < rows-1; row++ {
			p0 := a.ProjectPoint(x[col], y[row], z[row][col])
			p1 := a.ProjectPoint(x[col], y[row+1], z[row+1][col])
			segments = append(segments, []geom.Pt{p0, p1})
		}
	}
	return segments
}

func project3DPoint(x, y, z, elevationDeg, azimuthDeg, distance float64) geom.Pt {
	return project3DPointWithLimits(x, y, z, elevationDeg, azimuthDeg, distance, vec3{0, 0, 0}, vec3{1, 1, 1})
}

func project3DPointWithLimits(x, y, z, elevationDeg, azimuthDeg, distance float64, mins, maxs vec3) geom.Pt {
	if distance <= 0 {
		az := azimuthDeg * math.Pi / 180
		v := elevationDeg * math.Pi / 180

		x2 := x*math.Cos(az) - y*math.Sin(az)
		y2 := x*math.Sin(az) + y*math.Cos(az)
		return geom.Pt{X: x2, Y: y2*math.Cos(v) - z*math.Sin(v)}
	}

	m := default3DProjectionMatrix(elevationDeg, azimuthDeg, distance, mins, maxs)
	tx, ty, _ := transform3DPoint(m, x, y, z)
	return geom.Pt{X: tx, Y: ty}
}

func default3DProjectionMatrix(elevationDeg, azimuthDeg, distance float64, mins, maxs vec3) mat4 {
	aspect := default3DBoxAspect()
	world := worldTransformation(mins[0], maxs[0], mins[1], maxs[1], mins[2], maxs[2], aspect)
	center := vec3{0.5 * aspect[0], 0.5 * aspect[1], 0.5 * aspect[2]}

	elev := elevationDeg * math.Pi / 180
	az := azimuthDeg * math.Pi / 180
	viewDir := vec3{
		math.Cos(elev) * math.Cos(az),
		math.Cos(elev) * math.Sin(az),
		math.Sin(elev),
	}
	eye := center.add(viewDir.scale(distance))
	u, v, w := viewAxes(eye, center, vec3{0, 0, 1})
	view := viewTransformation(u, v, w, eye)
	proj := perspectiveTransformation(-distance, distance, default3DFocalLength)
	return proj.mul(view.mul(world))
}

func default3DBoxAspect() vec3 {
	aspect := vec3{4, 4, 3}
	scale := 1.8294640721620434 * 25.0 / 24.0 / aspect.norm()
	return aspect.scale(scale)
}

func worldTransformation(xmin, xmax, ymin, ymax, zmin, zmax float64, aspect vec3) mat4 {
	dx := (xmax - xmin) / aspect[0]
	dy := (ymax - ymin) / aspect[1]
	dz := (zmax - zmin) / aspect[2]
	return mat4{
		{1 / dx, 0, 0, -xmin / dx},
		{0, 1 / dy, 0, -ymin / dy},
		{0, 0, 1 / dz, -zmin / dz},
		{0, 0, 0, 1},
	}
}

func viewAxes(eye, center, vertical vec3) (vec3, vec3, vec3) {
	w := eye.sub(center).unit()
	u := vertical.cross(w).unit()
	v := w.cross(u)
	return u, v, w
}

func viewTransformation(u, v, w, eye vec3) mat4 {
	rot := mat4{
		{u[0], u[1], u[2], 0},
		{v[0], v[1], v[2], 0},
		{w[0], w[1], w[2], 0},
		{0, 0, 0, 1},
	}
	translate := mat4{
		{1, 0, 0, -eye[0]},
		{0, 1, 0, -eye[1]},
		{0, 0, 1, -eye[2]},
		{0, 0, 0, 1},
	}
	return rot.mul(translate)
}

func perspectiveTransformation(zfront, zback, focalLength float64) mat4 {
	b := (zfront + zback) / (zfront - zback)
	c := -2 * (zfront * zback) / (zfront - zback)
	return mat4{
		{focalLength, 0, 0, 0},
		{0, focalLength, 0, 0},
		{0, 0, b, c},
		{0, 0, -1, 0},
	}
}

func transform3DPoint(m mat4, x, y, z float64) (float64, float64, float64) {
	vec := [4]float64{x, y, z, 1}
	var out [4]float64
	for row := range 4 {
		for col := range 4 {
			out[row] += m[row][col] * vec[col]
		}
	}
	if out[3] == 0 {
		return out[0], out[1], out[2]
	}
	return out[0] / out[3], out[1] / out[3], out[2] / out[3]
}

type vec3 [3]float64

func (v vec3) add(other vec3) vec3 {
	return vec3{v[0] + other[0], v[1] + other[1], v[2] + other[2]}
}

func (v vec3) sub(other vec3) vec3 {
	return vec3{v[0] - other[0], v[1] - other[1], v[2] - other[2]}
}

func (v vec3) scale(s float64) vec3 {
	return vec3{v[0] * s, v[1] * s, v[2] * s}
}

func (v vec3) norm() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

func (v vec3) unit() vec3 {
	n := v.norm()
	if n == 0 {
		return vec3{}
	}
	return v.scale(1 / n)
}

func (v vec3) cross(other vec3) vec3 {
	return vec3{
		v[1]*other[2] - v[2]*other[1],
		v[2]*other[0] - v[0]*other[2],
		v[0]*other[1] - v[1]*other[0],
	}
}

func (v vec3) dot(other vec3) float64 {
	return v[0]*other[0] + v[1]*other[1] + v[2]*other[2]
}

type mat4 [4][4]float64

func (m mat4) mul(other mat4) mat4 {
	var out mat4
	for row := range 4 {
		for col := range 4 {
			for k := range 4 {
				out[row][col] += m[row][k] * other[k][col]
			}
		}
	}
	return out
}
