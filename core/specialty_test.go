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
	if len(hex.Values) != 2 {
		t.Fatalf("len(hex.Values) = %d, want 2", len(hex.Values))
	}
	if hex.Values[0] != 3 {
		t.Fatalf("hex.Values[0] = %v, want 3", hex.Values[0])
	}
	if hex.Counts[0] != 2 || hex.Counts[1] != 1 {
		t.Fatalf("unexpected counts %v", hex.Counts)
	}
	mapping := hex.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 3 || mapping.VMax != 9 {
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

	var renderer specialtyRecordingRenderer
	DrawFigure(fig, &renderer)
	if renderer.pathCount < 9 {
		t.Fatalf("expected at least 9 cell paths, got %d", renderer.pathCount)
	}
	if len(renderer.texts) < 8 {
		t.Fatalf("expected cell/header text draws, got %v", renderer.texts)
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
	if diagram.Trunk == nil || len(diagram.Ribbons) != 2 || len(diagram.Labels) != 2 {
		t.Fatalf("unexpected sankey content %+v", diagram)
	}
	if finished := builder.Finish(); len(finished) != 1 {
		t.Fatalf("Finish len = %d, want 1", len(finished))
	}
	if len(ax.Artists) < 5 {
		t.Fatalf("expected artists to be added to axes, got %d", len(ax.Artists))
	}
}

type specialtyRecordingRenderer struct {
	render.NullRenderer
	pathCount int
	texts     []string
}

func (r *specialtyRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *specialtyRecordingRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
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
