package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAddAxes3DConfiguresProjection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	if got, want := ax.ProjectionName(), "3d"; got != want {
		t.Fatalf("projection name = %q, want %q", got, want)
	}
	xMin, xMax := ax.effectiveXScale().Domain()
	yMin, yMax := ax.effectiveYScale().Domain()
	if !approx(xMin, default3DViewMin, 1e-12) || !approx(xMax, default3DViewMax, 1e-12) ||
		!approx(yMin, default3DViewMin, 1e-12) || !approx(yMax, default3DViewMax, 1e-12) {
		t.Fatalf("3D view domain = x(%v,%v) y(%v,%v), want (%v,%v)", xMin, xMax, yMin, yMax, default3DViewMin, default3DViewMax)
	}
	layout := ax.adjustedLayout(fig)
	if !approx(layout.W(), layout.H(), 1e-12) {
		t.Fatalf("3D axes layout = %+v, want square active box", layout)
	}

	elev, azim, distance := ax.View()
	if !approx(elev, default3DElevationDeg, 1e-12) ||
		!approx(azim, default3DAzimuthDeg, 1e-12) ||
		distance != default3DDistance {
		t.Fatalf("View = (%v, %v, %v), want (%v, %v, %v)", elev, azim, distance, default3DElevationDeg, default3DAzimuthDeg, default3DDistance)
	}
}

func TestAxes3DProjectionPointDefaults(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	got := ax.ProjectPoint(1, 2, 3)
	if !approx(got.X, 1, 1e-12) || !approx(got.Y, 2, 1e-12) {
		t.Fatalf("ProjectPoint(1,2,3) = %+v, want {1 2}", got)
	}
}

func TestAxes3DProjectPointMatchesMatplotlibDefaultProjection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	got := ax.ProjectPoint(1, 1, 1)
	if !approx(got.X, 0.0783182204915425, 1e-12) ||
		!approx(got.Y, 0.04773147362601089, 1e-12) {
		t.Fatalf("ProjectPoint(1,1,1) = %+v, want Matplotlib default projection {0.0783182204915425 0.04773147362601089}", got)
	}
}

func TestAxes3DProjectPointMatchesMatplotlibBasicDataLimits(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe(x, y, z)
	ax.Surface(x, y, z)
	ax.Contour(x, y, z)
	ax.Bar3D([]float64{0.2}, []float64{0.3}, []float64{0.4}, []float64{0.2}, []float64{0.2}, []float64{0.3})
	ax.Text3D(0.2, 0.8, 0.6, "demo point")

	got := ax.ProjectPoint(1, 1, 2)
	if !approx(got.X, 0.0711768607286225, 1e-12) ||
		!approx(got.Y, 0.043379132331248196, 1e-12) {
		t.Fatalf("ProjectPoint(1,1,2) with mplot3d_basic limits = %+v, want Matplotlib projection {0.0711768607286225 0.043379132331248196}", got)
	}
}

func TestAxes3DScatterDefaultColorUsesShapeCycle(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	palette := fig.RC.Palette()

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	scatter := ax.Scatter3D([]float64{0.5}, []float64{0.2}, []float64{0.1})
	nextLine := ax.Plot3D([]float64{0, 1}, []float64{1, 0}, []float64{0, 1})

	if got, want := line.Col, palette[0]; got != want {
		t.Fatalf("first 3D line color = %+v, want %+v", got, want)
	}
	if got, want := scatter.Color, palette[0]; got != want {
		t.Fatalf("3D scatter color = %+v, want independent shape cycle first color %+v", got, want)
	}
	if got, want := nextLine.Col, palette[1]; got != want {
		t.Fatalf("second 3D line color = %+v, want line cycle second color %+v", got, want)
	}
}

func TestAxes3DPlot3DUsesProjectedCoordinates(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	line := ax.Plot3D([]float64{0, 1}, []float64{0, 0}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}
	if got, want := len(line.XY), 2; got != want {
		t.Fatalf("projected points = %d, want %d", got, want)
	}
}

func TestAxes3DReprojectsExistingArtistsWhenDataLimitsExpand(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}

	ax.Wireframe(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{
			{0, 2},
			{0, 2},
		},
	)

	got := line.XY[1]
	if !approx(got.X, 0.06981276096054631, 1e-12) ||
		!approx(got.Y, 0.009353136460382655, 1e-12) {
		t.Fatalf("reprojected line endpoint = %+v, want Matplotlib projection with autoscale margins", got)
	}
}

func TestAxes3DProjectionLimitsUseMatplotlibAutoscaleMargins(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[0], -0.05, 1e-12) || !approx(maxs[0], 1.05, 1e-12) ||
		!approx(mins[1], -0.05, 1e-12) || !approx(maxs[1], 1.05, 1e-12) ||
		!approx(mins[2], -0.1, 1e-12) || !approx(maxs[2], 2.1, 1e-12) {
		t.Fatalf("projection limits = %v..%v, want Matplotlib autoscale margins [-0.05 -0.05 -0.1]..[1.05 1.05 2.1]", mins, maxs)
	}
}

func TestAxes3DFrameLimitsAddMatplotlibViewMargin(t *testing.T) {
	mins, maxs := axes3DFrameLimits(vec3{-0.05, -0.05, -0.1}, vec3{1.05, 1.05, 2.1})
	if !approx(mins[0], -0.07291666666666667, 1e-12) ||
		!approx(maxs[0], 1.0729166666666667, 1e-12) ||
		!approx(mins[2], -0.14583333333333334, 1e-12) ||
		!approx(maxs[2], 2.1458333333333335, 1e-12) {
		t.Fatalf("frame limits = %v..%v, want Matplotlib axis3d _get_coord_info view margin", mins, maxs)
	}
}

func TestAxes3DCollectionsUseComputedDepthZOrder(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	low := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 0}, {0, 0}},
	)
	high := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{1, 1}, {1, 1}},
	)
	if line == nil || low == nil || high == nil {
		t.Fatalf("expected line and surface artists, got line=%v low=%v high=%v", line, low, high)
	}
	if !(low.Z() > line.Z() && high.Z() > line.Z()) {
		t.Fatalf("3D surface zorders = low %.6g high %.6g, want both above 3D line %.6g like Matplotlib computed_zorder", low.Z(), high.Z(), line.Z())
	}
	if !(high.Z() > low.Z()) {
		t.Fatalf("3D surface zorders = low %.6g high %.6g, want higher projected plane drawn after lower plane", low.Z(), high.Z())
	}
}

func TestAxes3DWireframeGeneratesLineCollection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 4; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
}

func TestAxes3DWireframeTreatsZRowsAsYAndColumnsAsX(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	x := []float64{10, 20, 30}
	y := []float64{1, 2}
	z := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
	}
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := collection.Segments[0][0], (Pt{X: 10, Y: 1}); got != want {
		t.Fatalf("first wireframe point = %+v, want %+v", got, want)
	}
	if got, want := collection.Segments[0][1], (Pt{X: 20, Y: 1}); got != want {
		t.Fatalf("first wireframe row segment end = %+v, want %+v", got, want)
	}
}

func TestAxes3DFrameSegmentsUseMatplotlibActiveGridPlanes(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	segments := ax.frameSegments(vec3{0, 0, 0}, vec3{1, 1, 1})
	want := []Pt{
		ax.ProjectPoint(0.2, 0, 0),
		ax.ProjectPoint(0.2, 1, 0),
		ax.ProjectPoint(0.2, 1, 1),
	}
	if !contains3DSegment(segments, want, 1e-12) {
		t.Fatalf("missing Matplotlib-style x gridline through active panes; want %+v in %+v", want, segments)
	}
}

func TestAxes3DFrameGridSegmentsUseAxisTickLocations(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-3.45, -3.45, -0.88}
	maxs := vec3{3.45, 3.45, 0.78}
	segments := ax.frameSegments(mins, maxs)
	highs := ax.activePaneHighs(mins, maxs)
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

	p0 := minmax
	p1 := minmax
	p2 := minmax
	p0[0], p1[0], p2[0] = -3, -3, -3
	p0[1] = maxmin[1]
	p2[2] = maxmin[2]
	want := []Pt{
		project3DPointWithLimits(p0[0], p0[1], p0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p1[0], p1[1], p1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p2[0], p2[1], p2[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !contains3DSegment(segments, want, 1e-12) {
		t.Fatalf("missing gridline at Matplotlib AutoLocator tick -3; want %+v in %+v", want, segments)
	}
}

func TestAxes3DFrameAxisTicksMatchMatplotlibDensity(t *testing.T) {
	ticks := frameAxisTicks(-0.1, 2.1)
	if !containsFloat64Approx(ticks, 0.25, 1e-12) {
		t.Fatalf("3D frame ticks = %v, want Matplotlib-like 0.25 z tick", ticks)
	}
}

func TestAxes3DAxisLineSegmentsUseMatplotlibCameraFacingEdges(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	segments := ax.axisLineSegmentsProjected(frameMins, frameMaxs, mins, maxs)
	if got, want := len(segments), 3; got != want {
		t.Fatalf("axis line count = %d, want %d", got, want)
	}

	highs := ax.activePaneHighsProjected(frameMins, frameMaxs, mins, maxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = frameMaxs[i]
			maxmin[i] = frameMins[i]
		} else {
			minmax[i] = frameMins[i]
			maxmin[i] = frameMaxs[i]
		}
	}
	x0 := minmax
	x0[1] = maxmin[1]
	x1 := x0
	x1[0] = maxmin[0]
	wantX := []Pt{
		project3DPointWithLimits(x0[0], x0[1], x0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(x1[0], x1[1], x1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !pointsEqual(segments[0], wantX, 1e-12) {
		t.Fatalf("x axis line = %+v, want Matplotlib camera-facing edge %+v", segments[0], wantX)
	}
}

func TestAxes3DTickSegmentsUseMatplotlibInwardOutwardFactors(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	segments := ax.axisTickSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	if len(segments) == 0 {
		t.Fatal("axis tick segments are empty")
	}

	pair := ax.axisLineEdgePointPairs(frameMins, frameMaxs, mins, maxs)[0]
	tick := frameAxisTicks(mins[0], maxs[0])[0]
	tickDir := 1
	tickDelta := (maxs[tickDir] - mins[tickDir]) / 12
	if !ax.activePaneHighsProjected(frameMins, frameMaxs, mins, maxs)[tickDir] {
		tickDelta = -tickDelta
	}
	p0 := pair[0]
	p1 := pair[0]
	p0[0] = tick
	p1[0] = tick
	p0[tickDir] = pair[0][tickDir] + 0.1*tickDelta
	p1[tickDir] = pair[0][tickDir] - 0.2*tickDelta
	want := []Pt{
		project3DPointWithLimits(p0[0], p0[1], p0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p1[0], p1[1], p1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !pointsEqual(segments[0], want, 1e-12) {
		t.Fatalf("first x tick segment = %+v, want Matplotlib inward/outward segment %+v", segments[0], want)
	}
}

func TestAxes3DPanePolygonsUseMatplotlibActivePanes(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{0, 0, 0}
	maxs := vec3{1, 1, 1}
	panes := ax.activePanePolygons(mins, maxs)
	if got, want := len(panes), 3; got != want {
		t.Fatalf("pane count = %d, want %d", got, want)
	}
	highs := ax.activePaneHighs(mins, maxs)
	expectedPlanes := [6][4][3]int{
		{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	for axis := range 3 {
		planeIndex := 2 * axis
		if highs[axis] {
			planeIndex++
		}
		want := projectPlaneCorners(ax, expectedPlanes[planeIndex], mins, maxs)
		if !pointsEqual(panes[axis], want, 1e-12) {
			t.Fatalf("pane %d = %+v, want active Matplotlib pane %+v", axis, panes[axis], want)
		}
	}
}

func TestAxes3DPaneFaceColorsMatchMatplotlibDefaults(t *testing.T) {
	got := axes3DPaneFaceColors()
	want := []render.Color{
		{R: 0.95, G: 0.95, B: 0.95, A: 0.5},
		{R: 0.90, G: 0.90, B: 0.90, A: 0.5},
		{R: 0.925, G: 0.925, B: 0.925, A: 0.5},
	}
	if len(got) != len(want) {
		t.Fatalf("pane face colors = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pane face color %d = %+v, want Matplotlib default %+v", i, got[i], want[i])
		}
	}
}

func TestAxes3DSurfaceCreatesProjectedPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	collection := ax.Surface(x, y, z)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 1; got != want {
		t.Fatalf("surface polygon count = %d, want %d", got, want)
	}
	if got, want := len(collection.FaceColors), 1; got != want {
		t.Fatalf("surface face color count = %d, want %d", got, want)
	}
}

func TestAxes3DSurfaceUsesMatplotlibDefaultSampleCounts(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := make([]float64, 90)
	for i := range x {
		x[i] = float64(i)
	}
	y := make([]float64, 70)
	for i := range y {
		y[i] = float64(i)
	}
	z := make([][]float64, len(y))
	for row := range z {
		z[row] = make([]float64, len(x))
		for col := range z[row] {
			z[row][col] = float64(row + col)
		}
	}

	collection := ax.Surface(x, y, z)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 35*45; got != want {
		t.Fatalf("surface polygon count = %d, want Matplotlib default rcount/ccount sampled count %d", got, want)
	}
}

func projectPlaneCorners(ax *Axes3D, plane [4][3]int, mins, maxs vec3) []Pt {
	points := make([]Pt, len(plane))
	for i, corner := range plane {
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
		points[i] = ax.ProjectPoint(x, y, z)
	}
	return points
}

func pointsEqual(got, want []Pt, tol float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if !approx(got[i].X, want[i].X, tol) || !approx(got[i].Y, want[i].Y, tol) {
			return false
		}
	}
	return true
}

func contains3DSegment(segments [][]Pt, want []Pt, tol float64) bool {
	for _, segment := range segments {
		if len(segment) != len(want) {
			continue
		}
		matches := true
		for i := range want {
			if !approx(segment[i].X, want[i].X, tol) || !approx(segment[i].Y, want[i].Y, tol) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func containsFloat64Approx(values []float64, want, tol float64) bool {
	for _, got := range values {
		if approx(got, want, tol) {
			return true
		}
	}
	return false
}

func TestAxes3DContourAndContourfCreateCollections(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	contour := ax.Contour(x, y, z)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if contourf := ax.Contourf(x, y, z); contourf == nil {
		t.Fatal("Contourf returned nil")
	}
}

func TestAxes3DContourfProjectsFilledContourBands(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	fill := ax.Contourf(x, y, z)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, cellCount := len(fill.Polygons), 1; got <= cellCount {
		t.Fatalf("Contourf polygon count = %d, want filled contour band polygons rather than %d grid cell", got, cellCount)
	}
}

func TestAxes3DContourfUsesExplicitZOffset(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	offset := -3.0
	fill := ax.Contourf(x, y, z, PlotOptions{LevelCount: 3, Offset: &offset})
	if fill == nil || len(fill.Polygons) == 0 || len(fill.Polygons[0]) == 0 {
		t.Fatalf("Contourf returned no polygons: %+v", fill)
	}

	tri, values, ok := contourGridTriangulation(z, []ContourOptions{{X: x, Y: y}})
	if !ok {
		t.Fatal("contourGridTriangulation failed")
	}
	levels := contourLevels(values, nil, 3, true)
	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	rawPolygons, _ := contourBandPolygons(tri, values, levels, ContourOptions{}, mapping, 0.45)
	if len(rawPolygons) == 0 || len(rawPolygons[0]) == 0 {
		t.Fatal("expected raw contour band polygons")
	}
	want := ax.ProjectPoint(rawPolygons[0][0].X, rawPolygons[0][0].Y, offset)
	if got := fill.Polygons[0][0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("Contourf first point = %+v, want projection at explicit offset %+v", got, want)
	}
}

func TestAxes3DContourfUsesStructuredGridBandPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	offset := -1.0
	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0.5, 1.5}, Offset: &offset},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, want := len(fill.Polygons), 1; got != want {
		t.Fatalf("Contourf polygons = %d, want one structured quad band polygon", got)
	}
}

func TestAxes3DContourProjectsLinesAtContourLevels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	ax.observe3DGrid(x, y, z)

	levelCount := 3
	got := ax.projectedContourSegments(x, y, z, levelCount)
	tri, values, ok := contourGridTriangulation(z, []ContourOptions{{X: x, Y: y}})
	if !ok {
		t.Fatal("contourGridTriangulation failed")
	}
	levels := contourLevels(values, nil, levelCount, false)
	rawLines, rawLevels := contourPolylines(tri, values, levels)
	want := make([][]Pt, len(rawLines))
	for i, line := range rawLines {
		want[i] = make([]Pt, len(line))
		for j, pt := range line {
			want[i][j] = ax.ProjectPoint(pt.X, pt.Y, rawLevels[i])
		}
	}
	if len(got) != len(want) {
		t.Fatalf("contour segment count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !pointsEqual(got[i], want[i], 1e-12) {
			t.Fatalf("contour segment %d = %+v, want x/y contour projected at level z %+v", i, got[i], want[i])
		}
	}
}

func TestAxes3DContourUsesLevelColorsByDefault(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	contour := ax.Contour(x, y, z, PlotOptions{LevelCount: 4})
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if len(contour.Colors) != len(contour.Segments) {
		t.Fatalf("contour colors = %d, segments = %d; want per-level colormapped colors by default", len(contour.Colors), len(contour.Segments))
	}
	if len(contour.Colors) > 1 && contour.Colors[0] == contour.Colors[len(contour.Colors)-1] {
		t.Fatalf("first and last contour colors are both %+v, want level-dependent colors", contour.Colors[0])
	}
}

func TestAxes3DContourZOrderUsesContourGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{-1, 0, 1}
	y := []float64{-1, 0, 1}
	z := [][]float64{
		{0.2, 0.6, 0.2},
		{0.6, 1.0, 0.6},
		{0.2, 0.6, 0.2},
	}
	surface := ax.Surface(x, y, z)
	contour := ax.Contour(x, y, z, PlotOptions{LevelCount: 4})
	if surface == nil || contour == nil {
		t.Fatalf("expected surface and contour collections, got surface=%v contour=%v", surface, contour)
	}
	if !(surface.Z() > contour.Z()) {
		t.Fatalf("surface zorder %.6g, contour zorder %.6g; want surface drawn over 3D contour lines like Matplotlib computed_zorder", surface.Z(), contour.Z())
	}
}

func TestAxes3DDrawsYAxisEndpointTickLabels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	mins, maxs := ax.projectionLimits()
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, mins, maxs, mins, maxs)

	xTicks := frameAxisTicks(mins[0], maxs[0])
	yTicks := frameAxisTicks(mins[1], maxs[1])
	zTicks := frameAxisTicks(mins[2], maxs[2])
	if got, want := len(r.texts), len(xTicks)+len(yTicks)+len(zTicks); got != want {
		t.Fatalf("3D tick label count = %d, want x+y+z endpoint labels included (%d)", got, want)
	}
}

func TestAxes3DTickLabelsUseMatplotlibDataSpaceOffset(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe([]float64{0, 1}, []float64{0, 1}, z)
	ax.Surface([]float64{0, 1}, []float64{0, 1}, z)

	mins, maxs := ax.projectionLimits()
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, frameMins, frameMaxs, mins, maxs)

	if len(r.texts) == 0 {
		t.Fatal("expected 3D tick labels to be drawn")
	}
	xTicks := frameAxisTicks(mins[0], maxs[0])
	label := format3DTick(xTicks[0], 0, xTicks)
	if got := r.texts[0]; got != label {
		t.Fatalf("first tick label = %q, want first x tick %q", got, label)
	}
	fontSize := ctx.RC.TickLabelSize("x")
	expectedAnchor := expectedMatplotlib3DTickLabelAnchor(ax, ctx, 0, xTicks[0], frameMins, frameMaxs, mins, maxs)
	layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
	want := alignedSingleLineOrigin(expectedAnchor, layout, TextAlignCenter, textLayoutVAlignCenter)
	if !approx(r.positions[0].X, want.X, 1e-9) || !approx(r.positions[0].Y, want.Y, 1e-9) {
		t.Fatalf("first x tick label origin = %+v, want Matplotlib data-space offset origin %+v", r.positions[0], want)
	}
}

type axes3DTextRecorder struct {
	render.NullRenderer
	texts     []string
	positions []geom.Pt
}

func (r *axes3DTextRecorder) DrawText(text string, pos geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.positions = append(r.positions, pos)
}

func expectedMatplotlib3DTickLabelAnchor(ax *Axes3D, ctx *DrawContext, axis int, tick float64, mins, maxs, projMins, projMaxs vec3) geom.Pt {
	pair := ax.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)[axis]
	pos := pair[0]
	pos[axis] = tick
	tickDirs := [3]int{1, 0, 0}
	pos[tickDirs[axis]] = pair[0][tickDirs[axis]]

	centers, deltas := testAxes3DLabelCentersDeltas(ctx, projMins, projMaxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (defaultTickPadPt + 8) * deltas[i]
	}
	pos = testMove3DLabelFromCenter(pos, centers, labelDeltas, axis)
	projected := project3DPointWithLimits(pos[0], pos[1], pos[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, projMins, projMaxs)
	return ctx.TransformFor(Coords(CoordData)).Apply(projected)
}

func testAxes3DLabelCentersDeltas(ctx *DrawContext, mins, maxs vec3) (vec3, vec3) {
	centers := vec3{}
	deltas := vec3{}
	dpi := 100.0
	if ctx != nil && ctx.RC.DPI > 0 {
		dpi = ctx.RC.DPI
	}
	deltasPerPoint := 48 / (72 * (ctx.Clip.W() + ctx.Clip.H()) / dpi)
	for i := range 3 {
		centers[i] = (mins[i] + maxs[i]) / 2
		deltas[i] = (maxs[i] - mins[i]) / 12 * deltasPerPoint
	}
	return centers, deltas
}

func testMove3DLabelFromCenter(pos, centers, deltas vec3, axis int) vec3 {
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

func TestAxes3DBar3DCreatesSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Bar3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
	)
	if collection == nil {
		t.Fatal("Bar3D returned nil")
	}
	if got, want := len(collection.Segments), 16; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
	foundFaces := false
	for _, artist := range ax.Artists {
		polys, ok := artist.(*PolyCollection)
		if ok && len(polys.Polygons) == 6*2 {
			foundFaces = true
			break
		}
	}
	if !foundFaces {
		t.Fatal("Bar3D did not add filled projected cuboid faces")
	}
}

func TestAxes3DTrisurfCreatesSinglePolyCollection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	polyCount := 0
	lineCount := 0
	for _, artist := range ax.Artists {
		switch art := artist.(type) {
		case *PolyCollection:
			if len(art.Polygons) == 2 {
				polyCount++
			}
		case *LineCollection:
			lineCount++
		}
	}
	if polyCount != 1 || lineCount != 0 {
		t.Fatalf("Trisurf artists = %d matching PolyCollection, %d LineCollection; want one Poly3DCollection-equivalent and no separate edge collection", polyCount, lineCount)
	}
}

func TestAxes3DTrisurfShadesFaceColorsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	base := render.Color{R: 1, G: 0.5, B: 0.05, A: 1}
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	collection := ax.Trisurf(tri, []float64{0, 0, 1}, PlotOptions{Color: &base})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := len(collection.FaceColors), 1; got != want {
		t.Fatalf("trisurf face colors = %d, want %d shaded color per face", got, want)
	}
	if collection.FaceColors[0] == base {
		t.Fatalf("trisurf face color = %+v, want Matplotlib-style shaded variant of %+v", collection.FaceColors[0], base)
	}
}

func TestAxes3DVoxelCallsBarLikeSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Voxel(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
	)
	if collection == nil {
		t.Fatal("Voxel returned nil")
	}
	if got, want := len(collection.Segments), 16; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
}

func TestAxes3DText3DProjectsInput(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	text := ax.Text3D(1, 2, 3, "hello")
	if text == nil || text.Content != "hello" {
		t.Fatalf("Text3D returned unexpected value: %#v", text)
	}
	if !approx(text.Position.X, 1, 1e-12) || !approx(text.Position.Y, 2, 1e-12) {
		t.Fatalf("Text position = %+v, want {1 2}", text.Position)
	}
}
