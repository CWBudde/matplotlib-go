package core

import "testing"

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
	if !approx(got.X, 0.0815927872062553, 1e-12) ||
		!approx(got.Y, 0.04972717646245116, 1e-12) {
		t.Fatalf("ProjectPoint(1,1,1) = %+v, want Matplotlib default projection {0.0815927872062553 0.04972717646245116}", got)
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
	if !approx(got.X, 0.07666983110555299, 1e-12) ||
		!approx(got.Y, 0.010271809664281765, 1e-12) {
		t.Fatalf("reprojected line endpoint = %+v, want Matplotlib projection with z margin", got)
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

func TestAxes3DTrisurfCreatesSegments(t *testing.T) {
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
	if got, want := len(collection.Segments), 5; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
	foundFaces := false
	for _, artist := range ax.Artists {
		polys, ok := artist.(*PolyCollection)
		if ok && len(polys.Polygons) == 2 {
			foundFaces = true
			break
		}
	}
	if !foundFaces {
		t.Fatal("Trisurf did not add filled projected triangle faces")
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
