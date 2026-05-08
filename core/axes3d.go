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
	default3DDataMargin   = 0.05
	default3DComputedZ    = 2.5
	default3DSurfaceCount = 50
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
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	gridLineWidth := 0.8
	axisLineWidth := 0.8
	gridColor := render.Color{R: 0.70, G: 0.70, B: 0.70, A: 1}
	axisColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	if ctx.RC.GridLineWidth > 0 {
		gridLineWidth = ctx.RC.GridLineWidth
	}
	if ctx.RC.AxisLineWidth > 0 {
		axisLineWidth = ctx.RC.AxisLineWidth
	}
	if ctx.RC.GridColor.A > 0 {
		gridColor = ctx.RC.GridColor
	}
	if ctx.RC.AxesEdgeColor.A > 0 {
		axisColor = ctx.RC.AxesEdgeColor
	}

	panes := f.axes.activePanePolygonsProjected(frameMins, frameMaxs, mins, maxs)
	(&PolyCollection{
		Polygons: panes,
		PatchCollection: PatchCollection{
			Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
			FaceColors: axes3DPaneFaceColors(),
			EdgeColor:  render.Color{A: 0},
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}).Draw(r, ctx)

	segments := f.axes.frameSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   segments,
		Color:      gridColor,
		LineWidth:  gridLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	axisLines := f.axes.axisLineSegmentsProjected(frameMins, frameMaxs, mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   axisLines,
		Color:      axisColor,
		LineWidth:  axisLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	tickSegments := f.axes.axisTickSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	(&LineCollection{
		Collection: Collection{Coords: Coords(CoordData), Alpha: 1},
		Segments:   tickSegments,
		Color:      axisColor,
		LineWidth:  axisLineWidth,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	}).Draw(r, ctx)

	if textRen, ok := r.(render.TextDrawer); ok {
		f.axes.draw3DTickLabels(textRen, r, ctx, frameMins, frameMaxs, mins, maxs)
		f.axes.draw3DAxisLabels(textRen, r, ctx, frameMins, frameMaxs)
	}
}

func (f *axes3DFrame) Z() float64 { return -1000 }

func (f *axes3DFrame) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func axes3DPaneFaceColors() []render.Color {
	return []render.Color{
		{R: 0.95, G: 0.95, B: 0.95, A: 0.5},
		{R: 0.90, G: 0.90, B: 0.90, A: 0.5},
		{R: 0.925, G: 0.925, B: 0.925, A: 0.5},
	}
}

func (f *axes3DFrame) DrawOverlay(r render.Renderer, ctx *DrawContext) {
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
		scatter.z = a.points3DCollectionZ(x, y, z)
		a.add3DReprojector(func() {
			reprojectScatter3D(scatter, a.projectedData(x, y, z))
			scatter.z = a.points3DCollectionZ(x, y, z)
		}, limitsChanged)
		return scatter
	}
	scatter := a.Scatter(x2, y2)
	scatter.z = a.points3DCollectionZ(x, y, z)
	a.add3DReprojector(func() {
		reprojectScatter3D(scatter, a.projectedData(x, y, z))
		scatter.z = a.points3DCollectionZ(x, y, z)
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
			z:      a.grid3DCollectionZ(x, y, z),
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
			collection.z = a.grid3DCollectionZ(x, y, z)
		}
	}, limitsChanged)
	return collection
}

// Contour projects a structured z grid and emits a placeholder wireframe contour.
func (a *Axes3D) Contour(x, y []float64, z [][]float64, opts ...PlotOptions) *LineCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	levelCount := 7
	var explicitLevels []float64
	if len(opts) > 0 && opts[0].LevelCount > 0 {
		levelCount = opts[0].LevelCount
	}
	if len(opts) > 0 && len(opts[0].Levels) > 0 {
		explicitLevels = opts[0].Levels
	}
	segments, segmentLevels, _, values, zorder := a.projectedContourLineData(x, y, z, levelCount, explicitLevels)
	if len(segments) == 0 {
		return nil
	}

	color := a.NextColor()
	lineWidth := 1.0
	alpha := 1.0
	label := ""
	colorOverride := false
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			color = *opt.Color
			colorOverride = true
		}
		if opt.LineWidth != nil {
			lineWidth = *opt.LineWidth
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		label = opt.Label
	}

	colors := []render.Color(nil)
	collectionAlpha := alpha
	if !colorOverride {
		mapping := resolveScalarMapValues(values, "viridis", nil, nil)
		colors = make([]render.Color, len(segmentLevels))
		for i, level := range segmentLevels {
			colors[i] = mapping.Color(level, alpha)
		}
		collectionAlpha = 1
	}

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  collectionAlpha,
			z:      zorder,
		},
		Segments:  segments,
		Color:     color,
		Colors:    colors,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			segments, segmentLevels, _, values, zorder := a.projectedContourLineData(x, y, z, levelCount, explicitLevels)
			collection.Segments = segments
			if !colorOverride {
				mapping := resolveScalarMapValues(values, "viridis", nil, nil)
				colors := make([]render.Color, len(segmentLevels))
				for i, level := range segmentLevels {
					colors[i] = mapping.Color(level, alpha)
				}
				collection.Colors = colors
			}
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

// Contourf projects a structured z grid and emits filled contour bands.
func (a *Axes3D) Contourf(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	alpha := 0.45
	label := ""
	levelCount := 7
	var explicitLevels []float64
	colorOverride := (*render.Color)(nil)
	offset := (*float64)(nil)
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Color != nil {
			colorOverride = opt.Color
		}
		if opt.Alpha != nil && *opt.Alpha >= 0 && *opt.Alpha <= 1 {
			alpha = *opt.Alpha
		}
		if opt.LevelCount > 0 {
			levelCount = opt.LevelCount
		}
		if len(opt.Levels) > 0 {
			explicitLevels = opt.Levels
		}
		offset = opt.Offset
		label = opt.Label
	}

	polygons, colors, zorder := a.projectedContourFloorPolygons(x, y, z, alpha, colorOverride, levelCount, explicitLevels, offset)
	if len(polygons) == 0 {
		return nil
	}
	paths, pathColors := compoundContourPaths(polygons, colors)

	collection := &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords: Coords(CoordData),
				Label:  label,
				Alpha:  1,
				z:      zorder,
			},
			Paths:      paths,
			FaceColors: pathColors,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			polygons, colors, zorder := a.projectedContourFloorPolygons(x, y, z, alpha, colorOverride, levelCount, explicitLevels, offset)
			paths, pathColors := compoundContourPaths(polygons, colors)
			collection.Polygons = nil
			collection.Paths = paths
			collection.FaceColors = pathColors
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

// Surface draws a structured surface as projected, z-sorted quadrilateral faces.
func (a *Axes3D) Surface(x, y []float64, z [][]float64, opts ...PlotOptions) *PolyCollection {
	limitsChanged := a.observe3DGrid(x, y, z)
	polygons, faceColors, zorder := a.projectSurfacePolygons(x, y, z, opts...)
	if len(polygons) == 0 {
		return nil
	}

	alpha := 0.85
	label := ""
	edgeWidth := 1.0
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
				z:      zorder,
			},
			FaceColors: faceColors,
			EdgeColor:  render.Color{A: 0},
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			polygons, faceColors, zorder := a.projectSurfacePolygons(x, y, z, opts...)
			for i := range faceColors {
				faceColors[i].A *= alpha
			}
			collection.Polygons = polygons
			collection.FaceColors = faceColors
			collection.z = zorder
		}
	}, limitsChanged)
	return collection
}

func (a *Axes3D) projectedContourSegments(x, y []float64, z [][]float64, levelCount int) [][]geom.Pt {
	segments, _, _, _, _ := a.projectedContourLineData(x, y, z, levelCount, nil)
	return segments
}

func (a *Axes3D) projectedContourLineData(x, y []float64, z [][]float64, levelCount int, explicitLevels []float64) ([][]geom.Pt, []float64, []float64, []float64, float64) {
	if a == nil || len(z) == 0 {
		return nil, nil, nil, nil, defaultPatchZ
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil, nil, nil, defaultPatchZ
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil, nil, nil, defaultPatchZ
		}
	}

	values := flattenGridValues(z)
	if levelCount <= 0 {
		levelCount = 7
	}

	levels := contourLevels(values, explicitLevels, levelCount, false)
	if len(levels) == 0 {
		return nil, nil, nil, nil, defaultPatchZ
	}

	rawLines, rawLevels := contourGridPolylines(x[:cols], y[:rows], z, levels)
	segments := make([][]geom.Pt, 0, len(rawLines))
	segmentLevels := make([]float64, 0, len(rawLines))
	depth := math.Inf(1)
	for i, polyline := range rawLines {
		if len(polyline) < 2 {
			continue
		}
		level := rawLevels[i]
		projected := make([]geom.Pt, len(polyline))
		for j, pt := range polyline {
			var zDepth float64
			projected[j], zDepth = a.projectPointDepth(pt.X, pt.Y, level)
			if zDepth < depth {
				depth = zDepth
			}
		}
		segments = append(segments, projected)
		segmentLevels = append(segmentLevels, level)
	}
	return segments, segmentLevels, levels, values, computed3DCollectionZ(depth)
}

type surfaceFace struct {
	polygon []geom.Pt
	value   float64
	depth   float64
}

func (a *Axes3D) projectSurfacePolygons(x, y []float64, z [][]float64, opts ...PlotOptions) ([][]geom.Pt, []render.Color, float64) {
	if a == nil || len(z) == 0 {
		return nil, nil, 0
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil, 0
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil, 0
		}
	}

	faces := make([]surfaceFace, 0, (rows-1)*(cols-1))
	values := make([]float64, 0, (rows-1)*(cols-1))
	collectionDepth := math.Inf(1)
	rowIndices := surfaceGridSampleIndices(rows, default3DSurfaceCount)
	colIndices := surfaceGridSampleIndices(cols, default3DSurfaceCount)
	for rowIdx := 0; rowIdx+1 < len(rowIndices); rowIdx++ {
		row0 := rowIndices[rowIdx]
		row1 := rowIndices[rowIdx+1]
		for colIdx := 0; colIdx+1 < len(colIndices); colIdx++ {
			col0 := colIndices[colIdx]
			col1 := colIndices[colIdx+1]
			polygon := make([]geom.Pt, 0, 2*(row1-row0)+2*(col1-col0))
			value := 0.0
			depth := 0.0
			valid := true
			count := 0
			surfacePatchPerimeter(row0, row1, col0, col1, func(row, col int) {
				if !valid {
					return
				}
				zVal := z[row][col]
				if !isFinite3D(x[col], y[row], zVal) {
					valid = false
					return
				}
				pt, zDepth := a.projectPointDepth(x[col], y[row], zVal)
				polygon = append(polygon, pt)
				value += zVal
				depth += zDepth
				if zDepth < collectionDepth {
					collectionDepth = zDepth
				}
				count++
			})
			if !valid {
				continue
			}
			value /= float64(count)
			depth /= float64(count)
			faces = append(faces, surfaceFace{polygon: polygon, value: value, depth: depth})
			values = append(values, value)
		}
	}
	if len(faces) == 0 {
		return nil, nil, 0
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
	return polygons, colors, computed3DCollectionZ(collectionDepth)
}

func surfaceGridSampleIndices(length, count int) []int {
	if length <= 0 {
		return nil
	}
	if count <= 0 {
		count = default3DSurfaceCount
	}
	stride := int(math.Ceil(float64(length) / float64(count)))
	if stride < 1 {
		stride = 1
	}

	indices := make([]int, 0, (length+stride-1)/stride+1)
	if (length-1)%stride == 0 {
		for i := 0; i < length; i += stride {
			indices = append(indices, i)
		}
		return indices
	}

	for i := 0; i < length-1; i += stride {
		indices = append(indices, i)
	}
	return append(indices, length-1)
}

func surfacePatchPerimeter(row0, row1, col0, col1 int, emit func(row, col int)) {
	for col := col0; col < col1; col++ {
		emit(row0, col)
	}
	for row := row0; row < row1; row++ {
		emit(row, col1)
	}
	for col := col1; col > col0; col-- {
		emit(row1, col)
	}
	for row := row1; row > row0; row-- {
		emit(row, col0)
	}
}

func (a *Axes3D) projectedContourFloorPolygons(x, y []float64, z [][]float64, alpha float64, colorOverride *render.Color, levelCount int, explicitLevels []float64, offset *float64) ([][]geom.Pt, []render.Color, float64) {
	if a == nil || len(z) == 0 {
		return nil, nil, defaultPatchZ
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return nil, nil, defaultPatchZ
	}
	for row := 1; row < rows; row++ {
		if len(z[row]) != cols {
			return nil, nil, defaultPatchZ
		}
	}

	zMin, zMax := zGridRange(z)
	if zMin == zMax {
		zMin -= 0.5
		zMax += 0.5
	}
	floorZ := zMin - 0.2*(zMax-zMin)
	if offset != nil && isFinite(*offset) {
		floorZ = *offset
	}

	values := flattenGridValues(z)
	levels := contourLevels(values, explicitLevels, levelCount, true)
	if len(levels) < 2 {
		return nil, nil, defaultPatchZ
	}

	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	opt := ContourOptions{}
	if colorOverride != nil {
		opt.Color = colorOverride
	}
	rawPolygons, colors := contourGridBandPolygons(x[:cols], y[:rows], z, levels, opt, mapping, alpha)
	if len(rawPolygons) == 0 {
		return nil, nil, defaultPatchZ
	}

	polygons := make([][]geom.Pt, 0, len(rawPolygons))
	projectedColors := make([]render.Color, 0, len(colors))
	collectionDepth := math.Inf(1)
	for i, polygon := range rawPolygons {
		if len(polygon) < 3 {
			continue
		}
		projected := make([]geom.Pt, len(polygon))
		for j, pt := range polygon {
			projectedPt, zDepth := a.projectPointDepth(pt.X, pt.Y, floorZ)
			projected[j] = projectedPt
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		polygons = append(polygons, projected)
		if i < len(colors) {
			projectedColors = append(projectedColors, colors[i])
		} else if colorOverride != nil {
			color := *colorOverride
			color.A *= alpha
			projectedColors = append(projectedColors, color)
		} else {
			projectedColors = append(projectedColors, mapping.Color(0, 1))
		}
	}
	return polygons, projectedColors, computed3DCollectionZ(collectionDepth)
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

func compoundContourPaths(polygons [][]geom.Pt, colors []render.Color) ([]geom.Path, []render.Color) {
	paths := make([]geom.Path, 0)
	pathColors := make([]render.Color, 0)
	var current geom.Path
	var currentColor render.Color
	haveCurrent := false
	flush := func() {
		if !haveCurrent || len(current.C) == 0 {
			return
		}
		paths = append(paths, current)
		pathColors = append(pathColors, currentColor)
		current = geom.Path{}
		haveCurrent = false
	}
	for i, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		color := render.Color{}
		if i < len(colors) {
			color = colors[i]
		}
		if haveCurrent && color != currentColor {
			flush()
		}
		if !haveCurrent {
			currentColor = color
			haveCurrent = true
		}
		for j, pt := range polygon {
			if j == 0 {
				current.MoveTo(pt)
			} else {
				current.LineTo(pt)
			}
		}
		current.Close()
	}
	flush()
	return paths, pathColors
}

func flattenGridValues(z [][]float64) []float64 {
	values := make([]float64, 0)
	for _, row := range z {
		values = append(values, row...)
	}
	return values
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

// Trisurf projects a triangulated unstructured surface mesh as filled polygons.
func (a *Axes3D) Trisurf(tri Triangulation, z []float64, opts ...PlotOptions) *PolyCollection {
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

	faceColor := color
	faceColor.A *= alpha
	faces, faceColors, faceZ := a.projectTriangulationFaces(tri, z, faceColor)
	if len(faces) == 0 {
		return nil
	}
	collection := &PolyCollection{
		Polygons: faces,
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords: Coords(CoordData),
				Label:  label,
				Alpha:  1,
				z:      faceZ,
			},
			FaceColors: faceColors,
			EdgeColor:  render.Color{A: 0},
			EdgeWidth:  lineWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
	}
	a.Add(collection)
	a.add3DReprojector(func() {
		if collection != nil {
			faces, faceColors, faceZ := a.projectTriangulationFaces(tri, z, faceColor)
			collection.Polygons = faces
			collection.FaceColors = faceColors
			collection.z = faceZ
		}
	}, limitsChanged)
	return collection
}

func reprojectScatter3D(scatter *Scatter2D, points []geom.Pt) {
	if scatter == nil {
		return
	}
	scatter.XY = append(scatter.XY[:0], points...)
}

func reprojectLine3D(line *Line2D, points []geom.Pt) {
	if line == nil {
		return
	}
	line.XY = append(line.XY[:0], points...)
}

func (a *Axes3D) projectTriangulationFaces(tri Triangulation, z []float64, baseColor render.Color) ([][]geom.Pt, []render.Color, float64) {
	type triFace struct {
		polygon []geom.Pt
		color   render.Color
		depth   float64
	}
	faces := make([]triFace, 0, len(tri.Triangles))
	collectionDepth := math.Inf(1)
	for triIdx, t := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		polygon := make([]geom.Pt, 0, 3)
		points := [3]vec3{}
		depth := 0.0
		valid := true
		for i, idx := range t {
			if idx < 0 || idx >= len(tri.X) || idx >= len(tri.Y) || idx >= len(z) || !isFinite3D(tri.X[idx], tri.Y[idx], z[idx]) {
				valid = false
				break
			}
			points[i] = vec3{tri.X[idx], tri.Y[idx], z[idx]}
			pt, zDepth := a.projectPointDepth(tri.X[idx], tri.Y[idx], z[idx])
			polygon = append(polygon, pt)
			depth += zDepth
			if zDepth < collectionDepth {
				collectionDepth = zDepth
			}
		}
		if valid {
			normal := points[0].sub(points[1]).cross(points[1].sub(points[2]))
			faces = append(faces, triFace{
				polygon: polygon,
				color:   shade3DFaceColor(baseColor, normal),
				depth:   depth / 3,
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
	return polygons, colors, computed3DCollectionZ(collectionDepth)
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
	return a.frameSegmentsProjected(mins, maxs, mins, maxs, mins, maxs)
}

func (a *Axes3D) frameSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3) [][]geom.Pt {
	project := func(x, y, z float64) geom.Pt {
		return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, projMins, projMaxs)
	}
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
		return project(x, y, z)
	}

	edges := [][2][3]int{
		{{0, 0, 0}, {1, 0, 0}}, {{1, 0, 0}, {1, 1, 0}}, {{1, 1, 0}, {0, 1, 0}}, {{0, 1, 0}, {0, 0, 0}},
		{{0, 0, 1}, {1, 0, 1}}, {{1, 0, 1}, {1, 1, 1}}, {{1, 1, 1}, {0, 1, 1}}, {{0, 1, 1}, {0, 0, 1}},
		{{0, 0, 0}, {0, 0, 1}}, {{1, 0, 0}, {1, 0, 1}}, {{1, 1, 0}, {1, 1, 1}}, {{0, 1, 0}, {0, 1, 1}},
	}
	segments := make([][]geom.Pt, 0, len(edges)+18)
	for _, edge := range edges {
		p0 := corner(edge[0][0], edge[0][1], edge[0][2])
		p1 := corner(edge[1][0], edge[1][1], edge[1][2])
		segments = append(segments, []geom.Pt{p0, p1})
	}
	segments = append(segments, a.frameGridSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs)...)
	return segments
}

func (a *Axes3D) activePanePolygons(mins, maxs vec3) [][]geom.Pt {
	return a.activePanePolygonsProjected(mins, maxs, mins, maxs)
}

func (a *Axes3D) activePanePolygonsProjected(mins, maxs, projMins, projMaxs vec3) [][]geom.Pt {
	planes := [6][4][3]int{
		{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	project := func(corner [3]int) geom.Pt {
		x := mins[0]
		if corner[0] == 1 {
			x = maxs[0]
		}
		y := mins[1]
		if corner[1] == 1 {
			y = maxs[1]
		}
		z := mins[2]
		if corner[2] == 1 {
			z = maxs[2]
		}
		return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, projMins, projMaxs)
	}

	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	panes := make([][]geom.Pt, 0, 3)
	for axis := range 3 {
		planeIndex := 2 * axis
		if highs[axis] {
			planeIndex++
		}
		plane := planes[planeIndex]
		polygon := make([]geom.Pt, len(plane))
		for i, corner := range plane {
			polygon[i] = project(corner)
		}
		panes = append(panes, polygon)
	}
	return panes
}

func (a *Axes3D) frameGridSegments(mins, maxs vec3) [][]geom.Pt {
	return a.frameGridSegmentsProjected(mins, maxs, mins, maxs, mins, maxs)
}

func (a *Axes3D) axisLineSegmentsProjected(mins, maxs, projMins, projMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return project3DPointWithLimits(p[0], p[1], p[2], a.elevationDeg, a.azimuthDeg, a.distance, projMins, projMaxs)
	}
	pairs := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	segments := make([][]geom.Pt, 0, len(pairs))
	for _, pair := range pairs {
		segments = append(segments, []geom.Pt{project(pair[0]), project(pair[1])})
	}
	return segments
}

func (a *Axes3D) axisTickSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return project3DPointWithLimits(p[0], p[1], p[2], a.elevationDeg, a.azimuthDeg, a.distance, projMins, projMaxs)
	}
	pairs := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	tickDirs := [3]int{1, 0, 0}
	segments := make([][]geom.Pt, 0, 24)
	for axis, pair := range pairs {
		tickDir := tickDirs[axis]
		tickDelta := (tickMaxs[tickDir] - tickMins[tickDir]) / 12
		if !highs[tickDir] {
			tickDelta = -tickDelta
		}
		outward := pair[0][tickDir] + 0.1*tickDelta
		inward := pair[0][tickDir] - 0.2*tickDelta
		for _, tick := range frameAxisTicks(tickMins[axis], tickMaxs[axis]) {
			p0 := pair[0]
			p1 := pair[0]
			p0[axis] = tick
			p1[axis] = tick
			p0[tickDir] = outward
			p1[tickDir] = inward
			segments = append(segments, []geom.Pt{project(p0), project(p1)})
		}
	}
	return segments
}

func (a *Axes3D) axisLineEdgePointPairs(mins, maxs, projMins, projMaxs vec3) [][2]vec3 {
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	juggled := [3][3]int{
		{1, 0, 2},
		{0, 1, 2},
		{0, 2, 1},
	}
	pairs := make([][2]vec3, 0, 3)
	for axis := range 3 {
		p0 := minmax
		p0[juggled[axis][0]] = maxmin[juggled[axis][0]]
		p1 := p0
		p1[juggled[axis][1]] = maxmin[juggled[axis][1]]
		pairs = append(pairs, [2]vec3{p0, p1})
	}
	return pairs
}

func (a *Axes3D) frameGridSegmentsProjected(mins, maxs, projMins, projMaxs, tickMins, tickMaxs vec3) [][]geom.Pt {
	project := func(p vec3) geom.Pt {
		return project3DPointWithLimits(p[0], p[1], p[2], a.elevationDeg, a.azimuthDeg, a.distance, projMins, projMaxs)
	}
	highs := a.activePaneHighsProjected(mins, maxs, projMins, projMaxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	segments := make([][]geom.Pt, 0, 18)
	limits := [][2]float64{
		{tickMins[0], tickMaxs[0]},
		{tickMins[1], tickMaxs[1]},
		{tickMins[2], tickMaxs[2]},
	}
	for index, lim := range limits {
		for _, tick := range frameAxisTicks(lim[0], lim[1]) {
			p0 := minmax
			p1 := minmax
			p2 := minmax
			p0[index] = tick
			p1[index] = tick
			p2[index] = tick
			first := (index + 1) % 3
			second := (index + 2) % 3
			p0[first] = maxmin[first]
			p2[second] = maxmin[second]
			segments = append(segments, []geom.Pt{project(p0), project(p1), project(p2)})
		}
	}
	return segments
}

func (a *Axes3D) activePaneHighs(mins, maxs vec3) [3]bool {
	return a.activePaneHighsProjected(mins, maxs, mins, maxs)
}

func (a *Axes3D) activePaneHighsProjected(mins, maxs, projMins, projMaxs vec3) [3]bool {
	planes := [6][4]int{
		{0, 3, 7, 4}, {1, 2, 6, 5},
		{0, 1, 5, 4}, {3, 2, 6, 7},
		{0, 1, 2, 3}, {4, 5, 6, 7},
	}
	corners := [8]vec3{
		{mins[0], mins[1], mins[2]},
		{maxs[0], mins[1], mins[2]},
		{maxs[0], maxs[1], mins[2]},
		{mins[0], maxs[1], mins[2]},
		{mins[0], mins[1], maxs[2]},
		{maxs[0], mins[1], maxs[2]},
		{maxs[0], maxs[1], maxs[2]},
		{mins[0], maxs[1], maxs[2]},
	}
	depths := [8]float64{}
	for i, corner := range corners {
		depths[i] = a.projectPointDepthWithProjectionLimits(corner[0], corner[1], corner[2], projMins, projMaxs)
	}

	means0 := [3]float64{}
	means1 := [3]float64{}
	highs := [3]bool{}
	equals := [3]bool{}
	equalCount := 0
	for axis := range 3 {
		means0[axis] = meanPlaneDepth(depths, planes[2*axis])
		means1[axis] = meanPlaneDepth(depths, planes[2*axis+1])
		highs[axis] = means0[axis] < means1[axis]
		if math.Abs(means0[axis]-means1[axis]) <= math.Nextafter(1, 2)-1 {
			equals[axis] = true
			equalCount++
		}
	}
	if equalCount == 2 {
		vertical := -1
		for i := range equals {
			if !equals[i] {
				vertical = i
				break
			}
		}
		switch vertical {
		case 2:
			highs[0], highs[1] = true, true
		case 1:
			highs[0], highs[2] = true, false
		case 0:
			highs[1], highs[2] = false, false
		}
	}
	return highs
}

func meanPlaneDepth(depths [8]float64, plane [4]int) float64 {
	return (depths[plane[0]] + depths[plane[1]] + depths[plane[2]] + depths[plane[3]]) / 4
}

func computed3DCollectionZ(projectedDepth float64) float64 {
	if math.IsNaN(projectedDepth) || math.IsInf(projectedDepth, 0) {
		return defaultPatchZ
	}
	return default3DComputedZ - projectedDepth
}

func (a *Axes3D) points3DCollectionZ(x, y, z []float64) float64 {
	n := minLen(x, y, z)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		if !isFinite3D(x[i], y[i], z[i]) {
			continue
		}
		_, zDepth := a.projectPointDepth(x[i], y[i], z[i])
		if zDepth < depth {
			depth = zDepth
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) grid3DCollectionZ(x, y []float64, z [][]float64) float64 {
	if a == nil || len(z) == 0 {
		return defaultPatchZ
	}
	rows := len(z)
	cols := len(z[0])
	if cols == 0 || len(x) < cols || len(y) < rows {
		return defaultPatchZ
	}
	depth := math.Inf(1)
	for row := 0; row < rows; row++ {
		if len(z[row]) != cols {
			return computed3DCollectionZ(depth)
		}
		for col := 0; col < cols; col++ {
			if !isFinite3D(x[col], y[row], z[row][col]) {
				continue
			}
			_, zDepth := a.projectPointDepth(x[col], y[row], z[row][col])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) triangulation3DCollectionZ(tri Triangulation, z []float64) float64 {
	if a == nil {
		return defaultPatchZ
	}
	depth := math.Inf(1)
	for triIdx, t := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		for _, idx := range t {
			if idx < 0 || idx >= len(tri.X) || idx >= len(tri.Y) || idx >= len(z) || !isFinite3D(tri.X[idx], tri.Y[idx], z[idx]) {
				continue
			}
			_, zDepth := a.projectPointDepth(tri.X[idx], tri.Y[idx], z[idx])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) bar3DCollectionZ(x, y, z, dx, dy, dz []float64) float64 {
	n := minLen(x, y, z, dx, dy, dz)
	depth := math.Inf(1)
	for i := 0; i < n; i++ {
		x0, x1 := x[i], x[i]+dx[i]
		y0, y1 := y[i], y[i]+dy[i]
		z0, z1 := z[i], z[i]+dz[i]
		corners := [8][3]float64{
			{x0, y0, z0}, {x1, y0, z0}, {x1, y1, z0}, {x0, y1, z0},
			{x0, y0, z1}, {x1, y0, z1}, {x1, y1, z1}, {x0, y1, z1},
		}
		for _, corner := range corners {
			if !isFinite3D(corner[0], corner[1], corner[2]) {
				continue
			}
			_, zDepth := a.projectPointDepth(corner[0], corner[1], corner[2])
			if zDepth < depth {
				depth = zDepth
			}
		}
	}
	return computed3DCollectionZ(depth)
}

func (a *Axes3D) projectPointDepthWithLimits(x, y, z float64, mins, maxs vec3) (geom.Pt, float64) {
	if a == nil {
		return geom.Pt{}, 0
	}
	if a.distance <= 0 {
		return project3DPointWithLimits(x, y, z, a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs), z
	}
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs)
	tx, ty, tz := transform3DPoint(m, x, y, z)
	return geom.Pt{X: tx, Y: ty}, tz
}

func (a *Axes3D) projectPointDepthWithProjectionLimits(x, y, z float64, mins, maxs vec3) float64 {
	if a.distance <= 0 {
		return z
	}
	m := default3DProjectionMatrix(a.elevationDeg, a.azimuthDeg, a.distance, mins, maxs)
	_, _, tz := transform3DPoint(m, x, y, z)
	return tz
}

func (a *Axes3D) draw3DTickLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, mins, maxs, tickMins, tickMaxs vec3) {
	fontSize := ctx.RC.TickLabelSize("x")
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	centers, deltas := axes3DLabelCentersDeltas(ctx, tickMins, tickMaxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (defaultTickPadPt + 8) * deltas[i]
	}
	axisLines := a.axisLineEdgePointPairs(mins, maxs, tickMins, tickMaxs)
	tickDirs := [3]int{1, 0, 0}

	xTicks := frameAxisTicks(tickMins[0], tickMaxs[0])
	for i, tick := range xTicks {
		pos := axisLines[0][0]
		pos[0] = tick
		pos[tickDirs[0]] = axisLines[0][0][tickDirs[0]]
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 0), tickMins, tickMaxs)
		draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, xTicks), anchor, fontSize, textColor, textLayoutVAlignTop)
	}
	yTicks := frameAxisTicks(tickMins[1], tickMaxs[1])
	for i, tick := range yTicks {
		pos := axisLines[1][0]
		pos[1] = tick
		pos[tickDirs[1]] = axisLines[1][0][tickDirs[1]]
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 1), tickMins, tickMaxs)
		draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, yTicks), anchor, fontSize, textColor, textLayoutVAlignTop)
	}
	zTicks := frameAxisTicks(tickMins[2], tickMaxs[2])
	for i, tick := range zTicks {
		pos := axisLines[2][0]
		pos[2] = tick
		pos[tickDirs[2]] = axisLines[2][0][tickDirs[2]]
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 2), tickMins, tickMaxs)
		draw3DTextAtAnchorAligned(textRen, r, ctx, format3DTick(tick, i, zTicks), anchor, fontSize, textColor, textLayoutVAlignTop)
	}
}

func (a *Axes3D) draw3DAxisLabels(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, mins, maxs vec3) {
	fontSize := axisLabelFontSize(ctx)
	textColor := ctx.RC.DefaultAxesLabelColor()
	projMins, projMaxs := a.projectionLimits()
	centers, deltas := axes3DLabelCentersDeltas(ctx, projMins, projMaxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (4 + 21) * deltas[i]
	}
	axisLines := a.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)
	if a.XLabel != "" {
		pos := midpoint3D(axisLines[0][0], axisLines[0][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 0), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.XLabel, anchor, fontSize, textColor)
	}
	if a.YLabel != "" {
		pos := midpoint3D(axisLines[1][0], axisLines[1][1])
		anchor := a.project3DLabelAnchor(ctx, move3DLabelFromCenter(pos, centers, labelDeltas, 1), projMins, projMaxs)
		draw3DTextAtAnchor(textRen, r, ctx, a.YLabel, anchor, fontSize, textColor)
	}
}

func axes3DLabelCentersDeltas(ctx *DrawContext, mins, maxs vec3) (vec3, vec3) {
	centers := vec3{}
	deltas := vec3{}
	dpi := 100.0
	clipWidth := 1.0
	clipHeight := 1.0
	if ctx != nil {
		if ctx.RC.DPI > 0 {
			dpi = ctx.RC.DPI
		}
		if ctx.Clip.W() > 0 {
			clipWidth = ctx.Clip.W()
		}
		if ctx.Clip.H() > 0 {
			clipHeight = ctx.Clip.H()
		}
	}
	deltasPerPoint := 48 / (72 * (clipWidth + clipHeight) / dpi)
	for i := range 3 {
		centers[i] = (mins[i] + maxs[i]) / 2
		deltas[i] = (maxs[i] - mins[i]) / 12 * deltasPerPoint
	}
	return centers, deltas
}

func move3DLabelFromCenter(pos, centers, deltas vec3, axis int) vec3 {
	for i := range 3 {
		if i == axis {
			continue
		}
		if pos[i] < centers[i] {
			pos[i] -= deltas[i]
		} else {
			pos[i] += deltas[i]
		}
	}
	return pos
}

func midpoint3D(a, b vec3) vec3 {
	return vec3{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2, (a[2] + b[2]) / 2}
}

func (a *Axes3D) project3DLabelAnchor(ctx *DrawContext, pos, projMins, projMaxs vec3) geom.Pt {
	projected := project3DPointWithLimits(pos[0], pos[1], pos[2], a.elevationDeg, a.azimuthDeg, a.distance, projMins, projMaxs)
	return ctx.TransformFor(Coords(CoordData)).Apply(projected)
}

func draw3DTextAtAnchor(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, label string, anchor geom.Pt, fontSize float64, textColor render.Color) {
	draw3DTextAtAnchorAligned(textRen, r, ctx, label, anchor, fontSize, textColor, textLayoutVAlignCenter)
}

func draw3DTextAtAnchorAligned(textRen render.TextDrawer, r render.Renderer, ctx *DrawContext, label string, anchor geom.Pt, fontSize float64, textColor render.Color, vAlign textLayoutVerticalAlign) {
	if label == "" {
		return
	}
	layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
	origin := alignedSingleLineOrigin(anchor, layout, TextAlignCenter, vAlign)
	drawDisplayText(textRen, label, origin, fontSize, textColor, ctx.RC.FontKey)
}

func frameAxisTicks(minVal, maxVal float64) []float64 {
	ticks := AutoLocator{}.Ticks(minVal, maxVal, 9)
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
	barZ := a.bar3DCollectionZ(x, y, z, dx, dy, dz)
	if len(faces) > 0 {
		faceCollection := &PolyCollection{
			Polygons: faces,
			PatchCollection: PatchCollection{
				Collection: Collection{Coords: Coords(CoordData), Alpha: 1, z: barZ},
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
				faceCollection.z = a.bar3DCollectionZ(x, y, z, dx, dy, dz)
			}
		}, limitsChanged)
	}

	segments := a.projectBar3DSegments(x, y, z, dx, dy, dz)

	collection := &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  label,
			Alpha:  alpha,
			z:      barZ,
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
			collection.z = a.bar3DCollectionZ(x, y, z, dx, dy, dz)
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
	a.reproject3DArtists()
}

// SetDistance sets the perspective distance used by the 3D projection.
// Non-positive values disable perspective scaling.
func (a *Axes3D) SetDistance(distance float64) {
	if a == nil {
		return
	}
	a.distance = distance
	a.reproject3DArtists()
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
	a.reproject3DArtists()
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

func axes3DFrameLimits(mins, maxs vec3) (vec3, vec3) {
	for i := range 3 {
		delta := (maxs[i] - mins[i]) / 12
		mins[i] -= 0.25 * delta
		maxs[i] += 0.25 * delta
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
	scale := 1.8294640721620434 / aspect.norm()
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
