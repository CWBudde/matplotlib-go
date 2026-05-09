package test

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestLargeScatter_Golden(t *testing.T) {
	runGoldenTest(t, "large_scatter", renderLargeScatter)
}

func TestMixedCollection_Golden(t *testing.T) {
	runGoldenTest(t, "mixed_collection", renderMixedCollection)
}

func TestQuadMesh_Golden(t *testing.T) {
	runGoldenTest(t, "quad_mesh", renderQuadMesh)
}

func TestGouraudTriangles_Golden(t *testing.T) {
	runGoldenTest(t, "gouraud_triangles", renderGouraudTriangles)
}

func TestClipPathBatch_Golden(t *testing.T) {
	runGoldenTest(t, "clip_path_batch", renderClipPathBatch)
}

func renderLargeScatter() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.09, Y: 0.13}, Max: geom.Pt{X: 0.95, Y: 0.88}})
	ax.SetTitle("RendererAgg marker batch")
	ax.SetXLim(-0.5, 14.5)
	ax.SetYLim(-1.5, 11.5)
	ax.AddYGrid()

	points := make([]geom.Pt, 0, 180)
	sizes := make([]float64, 0, 180)
	colors := make([]render.Color, 0, 180)
	edges := make([]render.Color, 0, 180)
	for i := 0; i < 180; i++ {
		x := float64(i%15) + 0.24*math.Sin(float64(i)*0.73)
		y := float64((i*7)%12) + 0.28*math.Cos(float64(i)*0.41)
		points = append(points, geom.Pt{X: x, Y: y})
		radius := 4.0 + float64((i*11)%9)
		sizes = append(sizes, core.ScatterAreaFromRadius(radius, fig.RC.DPI))
		t := float64(i%30) / 29.0
		colors = append(colors, render.Color{R: 0.12 + 0.70*t, G: 0.58 - 0.25*t, B: 0.88 - 0.56*t, A: 0.72})
		edges = append(edges, render.Color{R: 0.08, G: 0.10 + 0.28*t, B: 0.18, A: 0.95})
	}
	ax.Add(&core.Scatter2D{
		XY:         points,
		Sizes:      sizes,
		Colors:     colors,
		EdgeColors: edges,
		EdgeWidth:  0.75,
		Marker:     core.MarkerCircle,
		Label:      "batched markers",
	})
	ax.AddLegend()

	return renderFixtureFigure(fig, 980, 620)
}

func renderMixedCollection() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg mixed path collection")
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 6)
	ax.XAxis.Locator = core.MultipleLocator{Base: 2}
	ax.YAxis.Locator = core.MultipleLocator{Base: 1}

	paths := []geom.Path{
		fixtureRectPath(0.8, 0.7, 1.4, 1.2),
		fixtureTrianglePath(2.9, 1.0, 0.9),
		fixtureDiamondPath(4.6, 1.2, 0.7),
		fixtureStarPath(6.4, 1.0, 0.75),
		fixtureRectPath(7.8, 0.7, 1.1, 1.7),
		fixtureTrianglePath(1.7, 4.0, 1.1),
		fixtureDiamondPath(3.8, 4.1, 0.8),
		fixtureStarPath(5.8, 4.0, 0.85),
		fixtureRectPath(7.5, 3.3, 1.5, 1.1),
	}
	faces := []render.Color{
		{R: 0.13, G: 0.47, B: 0.70, A: 0.65},
		{R: 1.00, G: 0.50, B: 0.05, A: 0.72},
		{R: 0.17, G: 0.63, B: 0.17, A: 0.70},
		{R: 0.84, G: 0.15, B: 0.16, A: 0.62},
		{R: 0.58, G: 0.40, B: 0.74, A: 0.70},
		{R: 0.55, G: 0.34, B: 0.29, A: 0.66},
		{R: 0.89, G: 0.47, B: 0.76, A: 0.66},
		{R: 0.50, G: 0.50, B: 0.50, A: 0.70},
		{R: 0.74, G: 0.74, B: 0.13, A: 0.70},
	}
	edges := []render.Color{
		{R: 0.02, G: 0.14, B: 0.23, A: 1},
		{R: 0.46, G: 0.21, B: 0.02, A: 1},
		{R: 0.02, G: 0.30, B: 0.06, A: 1},
		{R: 0.45, G: 0.04, B: 0.05, A: 1},
		{R: 0.28, G: 0.17, B: 0.42, A: 1},
		{R: 0.31, G: 0.17, B: 0.14, A: 1},
		{R: 0.44, G: 0.19, B: 0.37, A: 1},
		{R: 0.20, G: 0.20, B: 0.20, A: 1},
		{R: 0.36, G: 0.36, B: 0.04, A: 1},
	}
	widths := []float64{1.1, 1.6, 1.0, 1.8, 1.2, 1.4, 1.0, 1.6, 1.2}
	ax.AddCollection(&core.PatchCollection{
		Collection: core.Collection{Label: "mixed collection"},
		Paths:      paths,
		FaceColors: faces,
		EdgeColors: edges,
		EdgeWidths: widths,
		LineJoin:   render.JoinMiter,
		LineCap:    render.CapButt,
	})

	return renderFixtureFigure(fig, 980, 620)
}

func renderQuadMesh() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg quad mesh")
	ax.SetXLim(0, 9)
	ax.SetYLim(0, 6)
	ax.XAxis.Locator = core.MultipleLocator{Base: 1}
	ax.YAxis.Locator = core.MultipleLocator{Base: 1}

	data := make([][]float64, 6)
	for y := range data {
		data[y] = make([]float64, 9)
		for x := range data[y] {
			data[y][x] = 0.45 + 0.38*math.Sin(float64(x)*0.7) + 0.22*math.Cos(float64(y)*1.1)
		}
	}
	cmap := "viridis"
	vmin, vmax := -0.15, 1.1
	edgeColor := render.Color{R: 0.96, G: 0.96, B: 0.96, A: 1}
	edgeWidth := 0.65
	ax.PColorMesh(data, core.MeshOptions{
		XEdges:    []float64{0, 1.1, 1.9, 3.0, 3.7, 4.9, 5.8, 6.7, 7.9, 9.0},
		YEdges:    []float64{0, 0.8, 1.7, 2.9, 3.6, 4.8, 6.0},
		Colormap:  &cmap,
		VMin:      &vmin,
		VMax:      &vmax,
		EdgeColor: &edgeColor,
		EdgeWidth: &edgeWidth,
		Label:     "quad mesh",
	})

	return renderFixtureFigure(fig, 980, 620)
}

func renderGouraudTriangles() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg Gouraud triangles")
	ax.SetXLim(0, 4)
	ax.SetYLim(0, 3.2)
	ax.XAxis.Locator = core.MultipleLocator{Base: 0.5}
	ax.YAxis.Locator = core.MultipleLocator{Base: 0.5}

	ax.Add(&gouraudFixtureArtist{
		points: []geom.Pt{
			{X: 0.35, Y: 0.35},
			{X: 1.80, Y: 0.30},
			{X: 3.40, Y: 0.55},
			{X: 0.80, Y: 1.70},
			{X: 2.20, Y: 2.70},
			{X: 3.55, Y: 1.75},
		},
		triangles: [][3]int{{0, 1, 3}, {1, 4, 3}, {1, 2, 4}, {2, 5, 4}},
		values:    []float64{0.05, 0.38, 0.82, 0.62, 1.00, 0.28},
	})

	return renderFixtureFigure(fig, 980, 620)
}

func renderClipPathBatch() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg clip path batch")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5.4)
	ax.XAxis.Locator = core.MultipleLocator{Base: 1}
	ax.YAxis.Locator = core.MultipleLocator{Base: 1}
	ax.AddYGrid()
	ax.Add(&clipPathBatchFixtureArtist{})

	return renderFixtureFigure(fig, 980, 620)
}

type gouraudFixtureArtist struct {
	points    []geom.Pt
	triangles [][3]int
	values    []float64
}

func (a *gouraudFixtureArtist) Draw(r render.Renderer, ctx *core.DrawContext) {
	drawer, ok := r.(render.GouraudTriangleDrawer)
	if !ok || ctx == nil || len(a.points) == 0 {
		return
	}
	tr := ctx.TransformFor(core.Coords(core.CoordData))
	mapping := core.ScalarMapInfo{Colormap: "viridis", VMin: 0, VMax: 1}.Resolved()
	batch := render.GouraudTriangleBatch{Antialiased: true}
	for _, idx := range a.triangles {
		var tri render.GouraudTriangle
		for i, pointIndex := range idx {
			pt := a.points[pointIndex]
			if tr != nil {
				pt = tr.Apply(pt)
			}
			tri.P[i] = pt
			tri.Color[i] = mapping.Color(a.values[pointIndex], 1)
		}
		batch.Triangles = append(batch.Triangles, tri)
	}
	drawer.DrawGouraudTriangles(batch)
}

func (a *gouraudFixtureArtist) Z() float64 { return 0 }

func (a *gouraudFixtureArtist) Bounds(*core.DrawContext) geom.Rect {
	if len(a.points) == 0 {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: a.points[0], Max: a.points[0]}
	for _, pt := range a.points[1:] {
		bounds.Min.X = math.Min(bounds.Min.X, pt.X)
		bounds.Min.Y = math.Min(bounds.Min.Y, pt.Y)
		bounds.Max.X = math.Max(bounds.Max.X, pt.X)
		bounds.Max.Y = math.Max(bounds.Max.Y, pt.Y)
	}
	return bounds
}

type clipPathBatchFixtureArtist struct{}

func (a *clipPathBatchFixtureArtist) Draw(r render.Renderer, ctx *core.DrawContext) {
	if ctx == nil {
		return
	}
	drawer, ok := r.(render.QuadMeshDrawer)
	if !ok {
		return
	}
	tr := ctx.TransformFor(core.Coords(core.CoordData))
	if tr == nil {
		return
	}

	clip := transformFixturePath(clipBatchPath(), tr)
	r.Save()
	r.ClipPath(clip)
	drawer.DrawQuadMesh(clipBatchQuadMesh(ctx))
	r.Restore()

	r.Path(clip, &render.Paint{
		Stroke:    render.Color{R: 0.05, G: 0.08, B: 0.12, A: 1},
		LineWidth: 2.0,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func (a *clipPathBatchFixtureArtist) Z() float64 { return 0 }

func (a *clipPathBatchFixtureArtist) Bounds(*core.DrawContext) geom.Rect {
	return geom.Rect{Min: geom.Pt{X: 0.45, Y: 0.45}, Max: geom.Pt{X: 5.55, Y: 5.05}}
}

func clipBatchQuadMesh(ctx *core.DrawContext) render.QuadMeshBatch {
	xEdges := []float64{0, 0.75, 1.5, 2.35, 3.1, 4.0, 4.85, 5.45, 6.0}
	yEdges := []float64{0, 0.7, 1.55, 2.4, 3.2, 4.15, 5.4}
	mapping := core.ScalarMapInfo{Colormap: "viridis", VMin: -0.35, VMax: 1.15}.Resolved()
	tr := ctx.TransformFor(core.Coords(core.CoordData))
	batch := render.QuadMeshBatch{Cells: make([]render.QuadMeshCell, 0, (len(xEdges)-1)*(len(yEdges)-1))}
	for yi := 0; yi+1 < len(yEdges); yi++ {
		for xi := 0; xi+1 < len(xEdges); xi++ {
			cx := (xEdges[xi] + xEdges[xi+1]) * 0.5
			cy := (yEdges[yi] + yEdges[yi+1]) * 0.5
			value := 0.45 + 0.42*math.Sin(cx*1.15) + 0.33*math.Cos(cy*1.35) + 0.06*float64((xi+yi)%3)
			local := [4]geom.Pt{
				{X: xEdges[xi], Y: yEdges[yi]},
				{X: xEdges[xi+1], Y: yEdges[yi]},
				{X: xEdges[xi+1], Y: yEdges[yi+1]},
				{X: xEdges[xi], Y: yEdges[yi+1]},
			}
			var quad [4]geom.Pt
			for i, pt := range local {
				quad[i] = tr.Apply(pt)
			}
			batch.Cells = append(batch.Cells, render.QuadMeshCell{
				Quad:        quad,
				Face:        mapping.Color(value, 0.84),
				Edge:        render.Color{R: 0.97, G: 0.97, B: 0.97, A: 0.72},
				LineWidth:   0.55,
				Antialiased: true,
			})
		}
	}
	return batch
}

func clipBatchPath() geom.Path {
	points := []geom.Pt{
		{X: 0.55, Y: 1.10},
		{X: 2.05, Y: 0.50},
		{X: 3.10, Y: 1.05},
		{X: 5.35, Y: 0.80},
		{X: 4.70, Y: 2.45},
		{X: 5.50, Y: 4.05},
		{X: 3.70, Y: 3.80},
		{X: 2.55, Y: 5.05},
		{X: 1.75, Y: 3.55},
		{X: 0.55, Y: 3.85},
		{X: 1.20, Y: 2.35},
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func transformFixturePath(path geom.Path, tr interface{ Apply(geom.Pt) geom.Pt }) geom.Path {
	out := geom.Path{C: append([]geom.Cmd(nil), path.C...), V: make([]geom.Pt, len(path.V))}
	for i, pt := range path.V {
		out.V[i] = tr.Apply(pt)
	}
	return out
}

func renderFixtureFigure(fig *core.Figure, w, h int) image.Image {
	r, err := agg.New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func fixtureRectPath(x, y, w, h float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y + h})
	path.LineTo(geom.Pt{X: x, Y: y + h})
	path.Close()
	return path
}

func fixtureTrianglePath(cx, cy, r float64) geom.Path {
	path := geom.Path{}
	for i := 0; i < 3; i++ {
		angle := -math.Pi/2 + float64(i)*2*math.Pi/3
		pt := geom.Pt{X: cx + r*math.Cos(angle), Y: cy + r*math.Sin(angle)}
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func fixtureDiamondPath(cx, cy, r float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: cx, Y: cy + r})
	path.LineTo(geom.Pt{X: cx + r, Y: cy})
	path.LineTo(geom.Pt{X: cx, Y: cy - r})
	path.LineTo(geom.Pt{X: cx - r, Y: cy})
	path.Close()
	return path
}

func fixtureStarPath(cx, cy, r float64) geom.Path {
	path := geom.Path{}
	for i := 0; i < 10; i++ {
		radius := r
		if i%2 == 1 {
			radius = r * 0.45
		}
		angle := -math.Pi/2 + float64(i)*math.Pi/5
		pt := geom.Pt{X: cx + radius*math.Cos(angle), Y: cy + radius*math.Sin(angle)}
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}
