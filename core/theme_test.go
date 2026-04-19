package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/style"
)

func TestAddAxesUsesThemeDefaults(t *testing.T) {
	fig := NewFigure(400, 300, style.WithTheme(style.ThemePublication))
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	if ax.XAxis.Color != fig.RC.AxesEdgeColor {
		t.Fatalf("x axis color = %+v, want %+v", ax.XAxis.Color, fig.RC.AxesEdgeColor)
	}
	if ax.YAxis.LineWidth != fig.RC.AxisLineWidth {
		t.Fatalf("y axis width = %v, want %v", ax.YAxis.LineWidth, fig.RC.AxisLineWidth)
	}
	if got, want := ax.PeekColor(), fig.RC.Palette()[0]; got != want {
		t.Fatalf("palette head = %+v, want %+v", got, want)
	}
}

func TestAddGridAndLegendUseThemeDefaults(t *testing.T) {
	fig := NewFigure(400, 300, style.WithTheme(style.ThemeGGPlot))
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	grid := ax.AddXGrid()
	if grid.Color != fig.RC.GridColor {
		t.Fatalf("grid color = %+v, want %+v", grid.Color, fig.RC.GridColor)
	}
	if grid.MinorColor != fig.RC.MinorGridColor {
		t.Fatalf("minor grid color = %+v, want %+v", grid.MinorColor, fig.RC.MinorGridColor)
	}
	if grid.LineWidth != fig.RC.GridLineWidth {
		t.Fatalf("grid width = %v, want %v", grid.LineWidth, fig.RC.GridLineWidth)
	}

	legend := ax.AddLegend()
	if legend.BackgroundColor != fig.RC.LegendBackground {
		t.Fatalf("legend background = %+v, want %+v", legend.BackgroundColor, fig.RC.LegendBackground)
	}
	if legend.BorderColor != fig.RC.LegendBorderColor {
		t.Fatalf("legend border = %+v, want %+v", legend.BorderColor, fig.RC.LegendBorderColor)
	}
	if legend.TextColor != fig.RC.LegendTextColor {
		t.Fatalf("legend text color = %+v, want %+v", legend.TextColor, fig.RC.LegendTextColor)
	}
}
