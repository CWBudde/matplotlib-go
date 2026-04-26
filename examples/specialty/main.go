package main

import (
	"fmt"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(980, 720)

	// Use explicit normalized axes rectangles to match test/matplotlib_ref/plots/specialty_artists.py.
	eventAx := fig.AddAxes(axisRect(0.07, 0.57, 0.34, 0.94))
	eventAx.SetTitle("Eventplot")
	eventAx.SetXLim(0, 10)
	eventAx.SetYLim(0.4, 3.6)
	eventAx.AddXGrid()
	eventAx.Eventplot([][]float64{
		{0.8, 1.4, 3.1, 4.6, 7.3},
		{1.2, 2.9, 4.0, 6.4, 8.6},
		{0.5, 2.2, 5.4, 6.8, 9.1},
	}, core.EventPlotOptions{
		LineOffsets: []float64{1, 2, 3},
		LineLengths: []float64{0.6, 0.7, 0.8},
		Colors: []render.Color{
			{R: 0.18, G: 0.44, B: 0.74, A: 1},
			{R: 0.84, G: 0.38, B: 0.16, A: 1},
			{R: 0.20, G: 0.63, B: 0.42, A: 1},
		},
	})

	hexAx := fig.AddAxes(axisRect(0.39, 0.57, 0.66, 0.94))
	hexAx.SetTitle("Hexbin")
	hexAx.SetXLim(0, 1)
	hexAx.SetYLim(0, 1)
	hexAx.Hexbin(
		[]float64{0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88},
		[]float64{0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58},
		core.HexbinOptions{
			GridSizeX: 7,
			C:         []float64{1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1},
			Reduce:    "mean",
			Label:     "clusters",
		},
	)

	pieAx := fig.AddAxes(axisRect(0.73, 0.57, 0.98, 0.94))
	pieAx.SetTitle("Pie")
	pieAx.Pie([]float64{28, 22, 18, 32}, core.PieOptions{
		Labels:        []string{"Core", "I/O", "Render", "Docs"},
		AutoPct:       "%.0f%%",
		StartAngle:    90,
		LabelDistance: 1.08,
		Explode:       []float64{0, 0.04, 0, 0.02},
	})

	violinAx := fig.AddAxes(axisRect(0.07, 0.08, 0.34, 0.45))
	violinAx.SetTitle("Violin")
	violinAx.SetXLim(0.4, 3.6)
	violinAx.SetYLim(0.8, 5.2)
	violinAx.AddYGrid()
	violinAx.Violinplot([][]float64{
		{1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2},
		{1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0},
		{2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6},
	}, core.ViolinOptions{
		ShowMeans: boolPtr(true),
		Label:     "distribution",
	})

	tableAx := fig.AddAxes(axisRect(0.39, 0.08, 0.66, 0.45))
	tableAx.SetTitle("Table")
	tableAx.ShowFrame = false
	if tableAx.XAxis != nil {
		tableAx.XAxis.ShowTicks = false
		tableAx.XAxis.ShowLabels = false
	}
	if tableAx.YAxis != nil {
		tableAx.YAxis.ShowTicks = false
		tableAx.YAxis.ShowLabels = false
	}
	tableAx.Table(core.TableOptions{
		ColLabels: []string{"Metric", "Q1", "Q2"},
		RowLabels: []string{"A", "B"},
		CellText: [][]string{
			{"Latency", "18ms", "14ms"},
			{"Throughput", "220/s", "265/s"},
		},
		BBox: geom.Rect{
			Min: geom.Pt{X: 0.04, Y: 0.18},
			Max: geom.Pt{X: 0.96, Y: 0.82},
		},
	})

	sankeyAx := fig.AddAxes(axisRect(0.73, 0.08, 0.98, 0.45))
	sankeyAx.SetTitle("Sankey")
	sankeyAx.ShowFrame = false
	if sankeyAx.XAxis != nil {
		sankeyAx.XAxis.ShowTicks = false
		sankeyAx.XAxis.ShowLabels = false
	}
	if sankeyAx.YAxis != nil {
		sankeyAx.YAxis.ShowTicks = false
		sankeyAx.YAxis.ShowLabels = false
	}
	builder := core.NewSankey(sankeyAx, core.SankeyOptions{
		Center: geom.Pt{X: 0.18, Y: 0.5},
	})
	builder.Add([]float64{-2, 3, 1.5}, core.SankeyAddOptions{
		Labels:       []string{"Waste", "CPU", "Cache"},
		Orientations: []int{-1, 1, 1},
	})

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      980,
		Height:     720,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        96,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}
	if err := core.SavePNG(fig, r, "specialty.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved specialty.png")
}

func boolPtr(v bool) *bool {
	return &v
}

func axisRect(x0, y0, x1, y1 float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y0},
		Max: geom.Pt{X: x1, Y: y1},
	}
}
