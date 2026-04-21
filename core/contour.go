package core

import (
	"math"
	"sort"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// ContourOptions configures contour, contourf, and tricontour rendering.
type ContourOptions struct {
	X              []float64
	Y              []float64
	XEdges         []float64
	YEdges         []float64
	Levels         []float64
	LevelCount     int
	Colormap       *string
	Colors         []render.Color
	Color          *render.Color
	LineWidth      *float64
	Alpha          *float64
	LabelLines     bool
	LabelFormatter Formatter
	LabelFontSize  *float64
	LabelColor     *render.Color
	Label          string
}

type contourLabel struct {
	Text     string
	Position geom.Pt
	Angle    float64
	Color    render.Color
}

// ContourSet stores the artists created by contour/contourf calls.
type ContourSet struct {
	Levels         []float64
	Lines          *LineCollection
	Fills          *PolyCollection
	LabelFormatter Formatter
	LabelFontSize  float64
	LabelColor     render.Color
	labels         []contourLabel
	z              float64
}

// Contour draws isolines over a rectilinear scalar grid.
func (a *Axes) Contour(data [][]float64, opts ...ContourOptions) *ContourSet {
	tri, values, ok := contourGridTriangulation(data, opts)
	if !ok {
		return nil
	}
	return a.buildContourSet(tri, values, false, opts...)
}

// Contourf draws filled contour bands over a rectilinear scalar grid.
func (a *Axes) Contourf(data [][]float64, opts ...ContourOptions) *ContourSet {
	tri, values, ok := contourGridTriangulation(data, opts)
	if !ok {
		return nil
	}
	return a.buildContourSet(tri, values, true, opts...)
}

// TriContour draws isolines over an explicit triangulation.
func (a *Axes) TriContour(tri Triangulation, values []float64, opts ...ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}
	return a.buildContourSet(tri, values, false, opts...)
}

// TriContourf draws filled contour bands over an explicit triangulation.
func (a *Axes) TriContourf(tri Triangulation, values []float64, opts ...ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}
	return a.buildContourSet(tri, values, true, opts...)
}

// Draw renders the contour set's filled bands and/or line collection.
func (c *ContourSet) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil {
		return
	}
	if c.Fills != nil {
		c.Fills.Draw(r, ctx)
	}
	if c.Lines != nil {
		c.Lines.Draw(r, ctx)
	}
}

// DrawOverlay renders contour labels outside the axes clip.
func (c *ContourSet) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil || len(c.labels) == 0 {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	fontSize := resolvedFontSize(c.LabelFontSize, ctx)
	for _, label := range c.labels {
		text := normalizeDisplayText(label.Text)
		if text == "" {
			continue
		}
		displayPt := ctx.DataToPixel.Apply(label.Position)
		color := label.Color
		if color == (render.Color{}) {
			color = resolvedTextColor(c.LabelColor, ctx)
		}

		if rotated, ok := r.(render.RotatedTextDrawer); ok {
			layout := measureSingleLineTextLayout(r, text, fontSize, ctx.RC.FontKey)
			anchor := geom.Pt{X: displayPt.X, Y: displayPt.Y + layout.Height*0.5}
			rotated.DrawTextRotated(text, anchor, fontSize, label.Angle, color)
			continue
		}

		layout := measureSingleLineTextLayout(r, text, fontSize, ctx.RC.FontKey)
		origin := alignedSingleLineOrigin(displayPt, layout, TextAlignCenter, textLayoutVAlignCenter)
		textRen.DrawText(text, origin, fontSize, color)
	}
}

// Bounds returns the union of the contour set's line and fill geometry.
func (c *ContourSet) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil {
		return geom.Rect{}
	}
	if c.Lines == nil {
		if c.Fills == nil {
			return geom.Rect{}
		}
		return c.Fills.Bounds(ctx)
	}
	if c.Fills == nil {
		return c.Lines.Bounds(ctx)
	}
	return unionRect(c.Lines.Bounds(ctx), c.Fills.Bounds(ctx))
}

// Z returns the contour set's draw order.
func (c *ContourSet) Z() float64 {
	if c == nil {
		return 0
	}
	return c.z
}

// ScalarMap exposes the contour fill's scalar mapping for colorbars.
func (c *ContourSet) ScalarMap() ScalarMapInfo {
	if c == nil || c.Fills == nil {
		return ScalarMapInfo{}
	}
	return c.Fills.ScalarMap()
}

func (c *ContourSet) legendEntry() (legendEntry, bool) {
	if c == nil {
		return legendEntry{}, false
	}
	if c.Fills != nil {
		return c.Fills.legendEntry()
	}
	if c.Lines != nil {
		return c.Lines.legendEntry()
	}
	return legendEntry{}, false
}

func (a *Axes) buildContourSet(tri Triangulation, values []float64, filled bool, opts ...ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}

	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	levels := contourLevels(values, opt.Levels, opt.LevelCount, filled)
	if (!filled && len(levels) == 0) || (filled && len(levels) < 2) {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	lineWidth := 1.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	colorFallback := a.NextColor()
	cmapName := ""
	if opt.Colormap != nil {
		cmapName = *opt.Colormap
	}
	mapping := resolveScalarMapValues(values, cmapName, nil, nil)
	if filled {
		mapping.VMin = levels[0]
		mapping.VMax = levels[len(levels)-1]
	}

	set := &ContourSet{
		Levels:         append([]float64(nil), levels...),
		LabelFormatter: contourFormatter(opt.LabelFormatter),
		LabelFontSize:  valueOrDefaultFloat(opt.LabelFontSize, 10),
		z:              0,
	}
	if opt.LabelColor != nil {
		set.LabelColor = *opt.LabelColor
	}

	if filled {
		polygons, faceColors := contourBandPolygons(tri, values, levels, opt, mapping, alpha)
		if len(polygons) > 0 {
			cmap := ""
			vmin := 0.0
			vmax := 0.0
			if opt.Color == nil && len(opt.Colors) == 0 {
				cmap = mapping.Colormap
				vmin = mapping.VMin
				vmax = mapping.VMax
			}
			set.Fills = &PolyCollection{
				PatchCollection: PatchCollection{
					Collection: Collection{
						Coords:   Coords(CoordData),
						Label:    opt.Label,
						Alpha:    1,
						Colormap: cmap,
						VMin:     vmin,
						VMax:     vmax,
					},
					FaceColors: faceColors,
					LineJoin:   render.JoinMiter,
					LineCap:    render.CapButt,
				},
				Polygons: polygons,
			}
		}
	} else {
		polylines, polylineLevels := contourPolylines(tri, values, levels)
		if len(polylines) > 0 {
			colors := make([]render.Color, len(polylines))
			for i, level := range polylineLevels {
				colors[i] = contourLineColor(level, levels, opt, mapping, alpha, colorFallback)
			}
			set.Lines = &LineCollection{
				Collection: Collection{
					Coords: Coords(CoordData),
					Label:  opt.Label,
					Alpha:  1,
				},
				Segments:  polylines,
				Colors:    colors,
				LineWidth: lineWidth,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapRound,
			}
			if opt.LabelLines {
				set.labels = contourLabels(polylines, polylineLevels, colors, set.LabelFormatter)
			}
		}
	}

	if set.Fills == nil && set.Lines == nil {
		return nil
	}
	a.Add(set)
	return set
}

func contourGridTriangulation(data [][]float64, opts []ContourOptions) (Triangulation, []float64, bool) {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok || rows < 2 || cols < 2 {
		return Triangulation{}, nil, false
	}

	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	xCoords := resolvedContourCoords(cols, opt.X, opt.XEdges)
	yCoords := resolvedContourCoords(rows, opt.Y, opt.YEdges)
	if len(xCoords) != cols || len(yCoords) != rows {
		return Triangulation{}, nil, false
	}

	values := make([]float64, 0, rows*cols)
	xPoints := make([]float64, 0, rows*cols)
	yPoints := make([]float64, 0, rows*cols)
	for yi := 0; yi < rows; yi++ {
		for xi := 0; xi < cols; xi++ {
			xPoints = append(xPoints, xCoords[xi])
			yPoints = append(yPoints, yCoords[yi])
			values = append(values, data[yi][xi])
		}
	}

	triangles := make([][3]int, 0, (rows-1)*(cols-1)*2)
	mask := make([]bool, 0, (rows-1)*(cols-1)*2)
	index := func(row, col int) int { return row*cols + col }
	for yi := 0; yi+1 < rows; yi++ {
		for xi := 0; xi+1 < cols; xi++ {
			p00 := index(yi, xi)
			p10 := index(yi, xi+1)
			p01 := index(yi+1, xi)
			p11 := index(yi+1, xi+1)

			triangles = append(triangles, [3]int{p00, p10, p11})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p10, p11}))
			triangles = append(triangles, [3]int{p00, p11, p01})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p11, p01}))
		}
	}

	return Triangulation{
		X:         xPoints,
		Y:         yPoints,
		Triangles: triangles,
		Mask:      mask,
	}, values, true
}

func triangleFinite(values []float64, tri [3]int) bool {
	return isFinite(values[tri[0]]) && isFinite(values[tri[1]]) && isFinite(values[tri[2]])
}

func resolvedContourCoords(size int, coords, edges []float64) []float64 {
	switch {
	case len(coords) == size:
		return append([]float64(nil), coords...)
	case len(edges) == size:
		return append([]float64(nil), edges...)
	case len(edges) == size+1:
		out := make([]float64, size)
		for i := 0; i < size; i++ {
			out[i] = (edges[i] + edges[i+1]) * 0.5
		}
		return out
	default:
		out := make([]float64, size)
		for i := range out {
			out[i] = float64(i)
		}
		return out
	}
}

func contourLevels(values, explicit []float64, levelCount int, filled bool) []float64 {
	if len(explicit) > 0 {
		levels := make([]float64, 0, len(explicit))
		for _, level := range explicit {
			if isFinite(level) {
				levels = append(levels, level)
			}
		}
		sort.Float64s(levels)
		return dedupeFloat64(levels)
	}

	if levelCount <= 0 {
		levelCount = 7
	}
	if filled && levelCount < 2 {
		levelCount = 2
	}

	minValue, maxValue := finiteRange(values)
	if !isFinite(minValue) || !isFinite(maxValue) {
		return nil
	}
	if minValue == maxValue {
		if filled {
			return []float64{minValue, minValue + 1}
		}
		return []float64{minValue}
	}

	levels := make([]float64, levelCount)
	step := (maxValue - minValue) / float64(levelCount-1)
	for i := range levels {
		levels[i] = minValue + float64(i)*step
	}
	return levels
}

func contourPolylines(tri Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
	var polylines [][]geom.Pt
	var polylineLevels []float64
	for _, level := range levels {
		segments := contourSegmentsForLevel(tri, values, level)
		for _, polyline := range stitchContourSegments(segments) {
			if len(polyline) < 2 {
				continue
			}
			polylines = append(polylines, polyline)
			polylineLevels = append(polylineLevels, level)
		}
	}
	return polylines, polylineLevels
}

func contourBandPolygons(tri Triangulation, values, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color) {
	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		for triIdx, triangle := range tri.Triangles {
			if tri.masked(triIdx) {
				continue
			}
			polygon := triangleBandPolygon(
				[3]geom.Pt{tri.point(triangle[0]), tri.point(triangle[1]), tri.point(triangle[2])},
				[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
				low,
				high,
			)
			if len(polygon) < 3 {
				continue
			}
			polygons = append(polygons, polygon)
			colors = append(colors, color)
		}
	}
	return polygons, colors
}

func contourSegmentsForLevel(tri Triangulation, values []float64, level float64) [][]geom.Pt {
	segments := make([][]geom.Pt, 0, len(tri.Triangles))
	for triIdx, triangle := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		segment, ok := triangleContourSegment(
			[3]geom.Pt{tri.point(triangle[0]), tri.point(triangle[1]), tri.point(triangle[2])},
			[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
			level,
		)
		if ok {
			segments = append(segments, segment)
		}
	}
	return segments
}

func triangleContourSegment(points [3]geom.Pt, values [3]float64, level float64) ([]geom.Pt, bool) {
	intersections := []geom.Pt{}
	edges := [][2]int{{0, 1}, {1, 2}, {2, 0}}
	for _, edge := range edges {
		aIdx, bIdx := edge[0], edge[1]
		aValue := values[aIdx]
		bValue := values[bIdx]
		if !isFinite(aValue) || !isFinite(bValue) {
			return nil, false
		}
		if aValue == bValue {
			continue
		}
		minValue := math.Min(aValue, bValue)
		maxValue := math.Max(aValue, bValue)
		if level < minValue || level > maxValue {
			continue
		}
		t := (level - aValue) / (bValue - aValue)
		if t < 0 || t > 1 {
			continue
		}
		point := interpolatePoint(points[aIdx], points[bIdx], t)
		if !containsPoint(intersections, point) {
			intersections = append(intersections, point)
		}
	}
	if len(intersections) != 2 {
		return nil, false
	}
	return intersections, true
}

type contourVertex struct {
	Point geom.Pt
	Value float64
}

func triangleBandPolygon(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt {
	polygon := []contourVertex{
		{Point: points[0], Value: values[0]},
		{Point: points[1], Value: values[1]},
		{Point: points[2], Value: values[2]},
	}
	polygon = clipContourPolygonMin(polygon, low)
	if len(polygon) < 3 {
		return nil
	}
	polygon = clipContourPolygonMax(polygon, high)
	if len(polygon) < 3 {
		return nil
	}
	out := make([]geom.Pt, len(polygon))
	for i, vertex := range polygon {
		out[i] = vertex.Point
	}
	return out
}

func clipContourPolygonMin(polygon []contourVertex, threshold float64) []contourVertex {
	return clipContourPolygon(polygon, func(value float64) bool {
		return value >= threshold
	}, threshold)
}

func clipContourPolygonMax(polygon []contourVertex, threshold float64) []contourVertex {
	return clipContourPolygon(polygon, func(value float64) bool {
		return value <= threshold
	}, threshold)
}

func clipContourPolygon(polygon []contourVertex, inside func(float64) bool, threshold float64) []contourVertex {
	if len(polygon) == 0 {
		return nil
	}
	out := make([]contourVertex, 0, len(polygon)+2)
	prev := polygon[len(polygon)-1]
	prevInside := inside(prev.Value)
	for _, curr := range polygon {
		currInside := inside(curr.Value)
		if currInside != prevInside && curr.Value != prev.Value {
			t := (threshold - prev.Value) / (curr.Value - prev.Value)
			out = append(out, contourVertex{
				Point: interpolatePoint(prev.Point, curr.Point, t),
				Value: threshold,
			})
		}
		if currInside {
			out = append(out, curr)
		}
		prev = curr
		prevInside = currInside
	}
	return out
}

func contourBandColor(low, high float64, idx int, opt ContourOptions, mapping ScalarMapInfo, alpha float64) render.Color {
	if len(opt.Colors) > 0 {
		color := opt.Colors[idx%len(opt.Colors)]
		color.A *= alpha
		return color
	}
	if opt.Color != nil {
		color := *opt.Color
		color.A *= alpha
		return color
	}
	return mapping.Color((low+high)*0.5, alpha)
}

func contourLineColor(level float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64, fallback render.Color) render.Color {
	if opt.Color != nil {
		color := *opt.Color
		color.A *= alpha
		return color
	}
	if len(opt.Colors) > 0 {
		idx := indexOfLevel(levels, level)
		color := opt.Colors[idx%len(opt.Colors)]
		color.A *= alpha
		return color
	}
	if opt.Colormap != nil {
		return mapping.Color(level, alpha)
	}
	fallback.A *= alpha
	return fallback
}

func contourLabels(polylines [][]geom.Pt, levels []float64, colors []render.Color, formatter Formatter) []contourLabel {
	type candidate struct {
		polyline []geom.Pt
		color    render.Color
	}
	best := map[float64]candidate{}
	bestLen := map[float64]float64{}
	for i, polyline := range polylines {
		length := polylineLength(polyline)
		level := levels[i]
		if length <= bestLen[level] {
			continue
		}
		bestLen[level] = length
		best[level] = candidate{polyline: polyline, color: colors[i]}
	}

	labels := make([]contourLabel, 0, len(best))
	for _, level := range dedupeFloat64(levels) {
		candidate, ok := best[level]
		if !ok {
			continue
		}
		position, angle := polylineLabelPlacement(candidate.polyline)
		labels = append(labels, contourLabel{
			Text:     formatter.Format(level),
			Position: position,
			Angle:    normalizeLabelAngle(angle),
			Color:    candidate.color,
		})
	}
	return labels
}

func contourFormatter(formatter Formatter) Formatter {
	if formatter != nil {
		return formatter
	}
	return ScalarFormatter{Prec: 3}
}

func dedupeFloat64(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := []float64{values[0]}
	for _, value := range values[1:] {
		if math.Abs(value-out[len(out)-1]) <= 1e-12 {
			continue
		}
		out = append(out, value)
	}
	return out
}

func containsPoint(points []geom.Pt, point geom.Pt) bool {
	for _, existing := range points {
		if sameContourPoint(existing, point) {
			return true
		}
	}
	return false
}

func sameContourPoint(a, b geom.Pt) bool {
	return math.Abs(a.X-b.X) <= 1e-9 && math.Abs(a.Y-b.Y) <= 1e-9
}

func interpolatePoint(a, b geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func stitchContourSegments(segments [][]geom.Pt) [][]geom.Pt {
	remaining := make([][]geom.Pt, 0, len(segments))
	for _, segment := range segments {
		if len(segment) >= 2 {
			remaining = append(remaining, append([]geom.Pt(nil), segment...))
		}
	}
	out := [][]geom.Pt{}
	for len(remaining) > 0 {
		polyline := append([]geom.Pt(nil), remaining[0]...)
		remaining = remaining[1:]
		progress := true
		for progress {
			progress = false
			for i := 0; i < len(remaining); i++ {
				segment := remaining[i]
				switch {
				case sameContourPoint(polyline[len(polyline)-1], segment[0]):
					polyline = append(polyline, segment[1:]...)
				case sameContourPoint(polyline[len(polyline)-1], segment[len(segment)-1]):
					polyline = append(polyline, reversePoints(segment[:len(segment)-1])...)
				case sameContourPoint(polyline[0], segment[len(segment)-1]):
					polyline = append(segment[:len(segment)-1], polyline...)
				case sameContourPoint(polyline[0], segment[0]):
					polyline = append(reversePoints(segment[1:]), polyline...)
				default:
					continue
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				progress = true
				break
			}
		}
		out = append(out, polyline)
	}
	return out
}

func reversePoints(points []geom.Pt) []geom.Pt {
	out := append([]geom.Pt(nil), points...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func polylineLength(polyline []geom.Pt) float64 {
	total := 0.0
	for i := 1; i < len(polyline); i++ {
		total += math.Hypot(polyline[i].X-polyline[i-1].X, polyline[i].Y-polyline[i-1].Y)
	}
	return total
}

func polylineLabelPlacement(polyline []geom.Pt) (geom.Pt, float64) {
	total := polylineLength(polyline)
	if total == 0 || len(polyline) < 2 {
		if len(polyline) == 0 {
			return geom.Pt{}, 0
		}
		return polyline[0], 0
	}

	target := total * 0.5
	run := 0.0
	for i := 1; i < len(polyline); i++ {
		segLen := math.Hypot(polyline[i].X-polyline[i-1].X, polyline[i].Y-polyline[i-1].Y)
		if run+segLen >= target {
			t := (target - run) / segLen
			point := interpolatePoint(polyline[i-1], polyline[i], t)
			return point, math.Atan2(polyline[i].Y-polyline[i-1].Y, polyline[i].X-polyline[i-1].X)
		}
		run += segLen
	}

	last := polyline[len(polyline)-1]
	prev := polyline[len(polyline)-2]
	return last, math.Atan2(last.Y-prev.Y, last.X-prev.X)
}

func normalizeLabelAngle(angle float64) float64 {
	for angle > math.Pi/2 {
		angle -= math.Pi
	}
	for angle < -math.Pi/2 {
		angle += math.Pi
	}
	return angle
}

func indexOfLevel(levels []float64, level float64) int {
	for i, candidate := range levels {
		if math.Abs(candidate-level) <= 1e-12 {
			return i
		}
	}
	return 0
}

func valueOrDefaultFloat(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}
