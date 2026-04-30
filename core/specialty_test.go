package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestAxesEventplotBuildsSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	events := ax.Eventplot([][]float64{
		{0.5, 1.5},
		{2.0},
	}, EventPlotOptions{
		LineOffsets: []float64{1, 3},
		LineLengths: []float64{0.4, 1.2},
		Label:       "events",
	})
	if events == nil {
		t.Fatal("expected event collection")
	}
	if len(events.Segments) != 3 {
		t.Fatalf("len(events.Segments) = %d, want 3", len(events.Segments))
	}
	if got := events.Segments[0][0]; got != (geom.Pt{X: 0.5, Y: 0.8}) {
		t.Fatalf("first segment start = %+v, want {0.5 0.8}", got)
	}
	if got := events.Segments[2][1]; got != (geom.Pt{X: 2.0, Y: 3.6}) {
		t.Fatalf("third segment end = %+v, want {2.0 3.6}", got)
	}
}

func TestAxesHexbinAggregatesValues(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	hex := ax.Hexbin(
		[]float64{0.1, 0.2, 0.85},
		[]float64{0.1, 0.2, 0.8},
		HexbinOptions{
			GridSizeX: 2,
			GridSizeY: 2,
			C:         []float64{2, 4, 9},
			Reduce:    "mean",
			Extent: &geom.Rect{
				Min: geom.Pt{X: 0, Y: 0},
				Max: geom.Pt{X: 1, Y: 1},
			},
			Label: "hex",
		},
	)
	if hex == nil {
		t.Fatal("expected hexbin collection")
	}
	if len(hex.Values) != 3 {
		t.Fatalf("len(hex.Values) = %d, want 3", len(hex.Values))
	}
	if hex.Values[0] != 2 || hex.Values[1] != 4 || hex.Values[2] != 9 {
		t.Fatalf("unexpected values %v", hex.Values)
	}
	if hex.Counts[0] != 1 || hex.Counts[1] != 1 || hex.Counts[2] != 1 {
		t.Fatalf("unexpected counts %v", hex.Counts)
	}
	if !floatApprox(hex.BinCenters[1].X, 0.25, 1e-6) || !floatApprox(hex.BinCenters[1].Y, 0.25, 1e-6) {
		t.Fatalf("second center = %+v, want near {0.25 0.25}", hex.BinCenters[1])
	}
	if len(hex.EdgeColors) != len(hex.FaceColors) {
		t.Fatalf("hex edge colors len = %d, want face-colored edges for %d faces", len(hex.EdgeColors), len(hex.FaceColors))
	}
	mapping := hex.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 2 || mapping.VMax != 9 {
		t.Fatalf("unexpected scalar map %+v", mapping)
	}
}

func TestAxesPieCreatesWedgesAndLabels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	pie := ax.Pie([]float64{2, 3}, PieOptions{
		Labels:  []string{"A", "B"},
		AutoPct: "%.0f%%",
	})
	if pie == nil {
		t.Fatal("expected pie container")
	}
	if len(pie.Wedges) != 2 || len(pie.Labels) != 2 || len(pie.AutoText) != 2 {
		t.Fatalf("unexpected pie counts wedges=%d labels=%d auto=%d", len(pie.Wedges), len(pie.Labels), len(pie.AutoText))
	}
	if pie.Labels[0].ClipOn || pie.AutoText[0].ClipOn {
		t.Fatal("expected pie label text to draw outside the axes clip")
	}
	if pie.AutoText[0].Color != (render.Color{}) {
		t.Fatalf("autopct color = %+v, want default text color", pie.AutoText[0].Color)
	}
	if pie.Wedges[0].Theta1 != 0 || pie.Wedges[0].Theta2 != 144 {
		t.Fatalf("unexpected first wedge angles %.1f -> %.1f", pie.Wedges[0].Theta1, pie.Wedges[0].Theta2)
	}
	if bounds := pie.Wedges[0].Bounds(nil); bounds == (geom.Rect{}) {
		t.Fatal("expected wedge bounds")
	}
}

func TestAxesViolinplotAddsCollections(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	violins := ax.Violinplot([][]float64{
		{1, 2, 2.5, 3, 4},
		{2, 2.1, 2.2, 3.4, 3.6},
	}, ViolinOptions{
		ShowMeans: specialtyBoolPtr(true),
		Label:     "spread",
	})
	if violins == nil {
		t.Fatal("expected violin container")
	}
	if violins.Bodies == nil || len(violins.Bodies.Polygons) != 2 {
		t.Fatal("expected two violin bodies")
	}
	if violins.Medians == nil || len(violins.Medians.Segments) != 2 {
		t.Fatal("expected median segments")
	}
	if violins.Means == nil || len(violins.Means.Segments) != 2 {
		t.Fatal("expected mean segments")
	}
	if violins.Extrema == nil || len(violins.Extrema.Segments) != 6 {
		t.Fatalf("expected extrema segments, got %d", len(violins.Extrema.Segments))
	}
}

func TestAxesTableDrawsCellsAndText(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	table := ax.Table(TableOptions{
		CellText:  [][]string{{"1.0", "2.0"}, {"3.0", "4.0"}},
		RowLabels: []string{"R1", "R2"},
		ColLabels: []string{"C1", "C2"},
		BBox:      geom.Rect{Min: geom.Pt{X: 0.15, Y: 0.15}, Max: geom.Pt{X: 0.85, Y: 0.55}},
	})
	if table == nil {
		t.Fatal("expected table artist")
	}
	if len(table.Cells) != 3 || len(table.Cells[0]) != 3 {
		t.Fatalf("unexpected table grid %dx%d", len(table.Cells), len(table.Cells[0]))
	}
	if table.Cells[0][0].Rect != (geom.Rect{}) {
		t.Fatalf("top-left row-label/header intersection rect = %+v, want empty", table.Cells[0][0].Rect)
	}
	if got, want := table.Cells[1][0].Rect.Max.X, table.BBox.Min.X; !floatApprox(got, want, 1e-12) {
		t.Fatalf("row label right edge = %v, want bbox left %v", got, want)
	}
	if table.HeaderTextColor != (render.Color{A: 1}) || table.EdgeColor != (render.Color{A: 1}) {
		t.Fatalf("table defaults headerText=%+v edge=%+v, want opaque black", table.HeaderTextColor, table.EdgeColor)
	}

	var renderer specialtyRecordingRenderer
	DrawFigure(fig, &renderer)
	if renderer.pathCount < 8 {
		t.Fatalf("expected at least 8 cell paths, got %d", renderer.pathCount)
	}
	if len(renderer.texts) < 8 {
		t.Fatalf("expected cell/header text draws, got %v", renderer.texts)
	}
}

func TestAxesTableHonorsAlignmentPadding(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	table := ax.Table(TableOptions{
		CellText:  [][]string{{"L", "R"}},
		RowLabels: []string{"row"},
		ColLabels: []string{"C1", "C2"},
		BBox:      geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}},
		FontSize:  10,
		CellLoc:   "left",
		RowLoc:    "right",
		ColLoc:    "center",
	})
	if table == nil {
		t.Fatal("expected table artist")
	}
	if got := table.Cells[1][1].HAlign; got != TextAlignLeft {
		t.Fatalf("data cell align = %v, want left", got)
	}
	if got := table.Cells[1][0].HAlign; got != TextAlignRight {
		t.Fatalf("row label align = %v, want right", got)
	}
	if got := table.Cells[0][1].HAlign; got != TextAlignCenter {
		t.Fatalf("column label align = %v, want center", got)
	}

	var renderer specialtyRecordingRenderer
	table.Draw(&renderer, &DrawContext{
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}},
	})
	origin, ok := renderer.textOrigins["L"]
	if !ok {
		t.Fatalf("expected data text draw, got %v", renderer.texts)
	}
	if !floatApprox(origin.X, 5, 1e-12) {
		t.Fatalf("left-aligned data text origin x = %v, want 10%% cell padding at 5px", origin.X)
	}
}

func TestSankeyBuilderCreatesDiagram(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	builder := NewSankey(ax, SankeyOptions{
		Center: geom.Pt{X: 0.2, Y: 0.5},
		Scale:  0.1,
	})
	if builder == nil {
		t.Fatal("expected sankey builder")
	}
	diagram := builder.Add([]float64{-2, 3}, SankeyAddOptions{
		Labels:       []string{"Loss", "Gain"},
		Orientations: []int{-1, 1},
	})
	if diagram == nil {
		t.Fatal("expected sankey diagram")
	}
	if diagram.Trunk == nil || len(diagram.Ribbons) != 1 || len(diagram.Labels) != 2 || len(diagram.Values) != 2 {
		t.Fatalf("unexpected sankey content %+v", diagram)
	}
	if got, want := diagram.Trunk.Height, 0.3; !floatApprox(got, want, 1e-12) {
		t.Fatalf("trunk height = %v, want max flow sum scaled to %v", got, want)
	}
	if diagram.Values[0].Content != "2" || diagram.Values[1].Content != "3" {
		t.Fatalf("unexpected flow value labels %q %q", diagram.Values[0].Content, diagram.Values[1].Content)
	}
	if finished := builder.Finish(); len(finished) != 1 {
		t.Fatalf("Finish len = %d, want 1", len(finished))
	}
	if len(ax.Artists) < 5 {
		t.Fatalf("expected artists to be added to axes, got %d", len(ax.Artists))
	}
}

func TestSankeyMatchesMatplotlibSingleDiagramGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	builder := NewSankey(ax, SankeyOptions{Scale: 0.16, Offset: 0.2})
	diagram := builder.Add([]float64{-2, 3, 1.5}, SankeyAddOptions{
		Labels:       []string{"Waste", "CPU", "Cache"},
		Orientations: []int{-1, 1, 1},
	})
	if diagram == nil || diagram.Patch == nil {
		t.Fatal("expected sankey diagram patch")
	}
	builder.Finish()

	wantTips := []geom.Pt{
		{X: 0.66, Y: -0.5694289299236832},
		{X: -0.74, Y: 0.4086160885174527},
		{X: -1.35, Y: 0.5093080442587265},
	}
	for i, want := range wantTips {
		if !floatApprox(diagram.Tips[i].X, want.X, 1e-12) || !floatApprox(diagram.Tips[i].Y, want.Y, 1e-12) {
			t.Fatalf("tip %d = %+v, want %+v", i, diagram.Tips[i], want)
		}
		if diagram.Angles[i] != sankeyDown {
			t.Fatalf("angle %d = %d, want DOWN", i, diagram.Angles[i])
		}
	}

	bounds, ok := pathBounds(diagram.Patch.Path)
	if !ok {
		t.Fatal("expected path bounds")
	}
	if !floatApprox(bounds.Min.X, -1.47, 1e-12) ||
		!floatApprox(bounds.Max.X, 0.85, 1e-12) ||
		!floatApprox(bounds.Min.Y, -0.5694289299236832, 1e-12) ||
		!floatApprox(bounds.Max.Y, 0.61, 1e-12) {
		t.Fatalf("path bounds = %+v", bounds)
	}

	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if !floatApprox(xMin, -1.87, 1e-12) || !floatApprox(xMax, 1.25, 1e-12) ||
		!floatApprox(yMin, -1.1694289299236833, 1e-12) || !floatApprox(yMax, 1.1093080442587264, 1e-12) {
		t.Fatalf("finished limits = x(%v, %v) y(%v, %v)", xMin, xMax, yMin, yMax)
	}
}

type specialtyRecordingRenderer struct {
	render.NullRenderer
	pathCount   int
	texts       []string
	textOrigins map[string]geom.Pt
}

func (r *specialtyRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *specialtyRecordingRenderer) DrawText(text string, pt geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	if r.textOrigins == nil {
		r.textOrigins = map[string]geom.Pt{}
	}
	r.textOrigins[text] = pt
}

func (r *specialtyRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.55,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func specialtyBoolPtr(v bool) *bool {
	return &v
}
