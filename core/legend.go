package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
)

// LegendLocation controls where the legend box is anchored inside the axes.
type LegendLocation uint8

const (
	LegendUpperRight LegendLocation = iota
	LegendUpperLeft
	LegendLowerRight
	LegendLowerLeft
)

type legendEntryKind uint8

const (
	legendEntryLine legendEntryKind = iota
	legendEntryMarker
	legendEntryPatch
)

type legendEntry struct {
	Label string

	kind legendEntryKind

	lineColor render.Color
	lineWidth float64
	dashes    []float64

	marker          MarkerType
	markerFill      render.Color
	markerEdge      render.Color
	markerEdgeWidth float64

	patchFill      render.Color
	patchEdge      render.Color
	patchEdgeWidth float64
}

// Legend renders a styled legend box inside an axes.
// If no explicit internal entries are present, labeled artists on the owning axes are collected automatically.
type Legend struct {
	Axes *Axes

	entries []legendEntry

	Location        LegendLocation
	Padding         float64
	Inset           float64
	RowGap          float64
	SampleWidth     float64
	SampleTextGap   float64
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	z               float64
}

// NewLegend creates a legend bound to the provided axes.
func NewLegend(ax *Axes) *Legend {
	rc := style.Default
	if ax != nil {
		rc = ax.resolvedRC()
	}
	return &Legend{
		Axes:            ax,
		Location:        LegendUpperRight,
		Padding:         10,
		Inset:           10,
		RowGap:          6,
		SampleWidth:     24,
		SampleTextGap:   8,
		CornerRadius:    0,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     1,
		z:               1_000,
	}
}

// AddLegend appends a legend to the axes.
func (a *Axes) AddLegend() *Legend {
	legend := NewLegend(a)
	a.Add(legend)
	return legend
}

// Draw renders the legend box and entries.
func (l *Legend) Draw(r render.Renderer, ctx *DrawContext) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	entries := l.entries
	if len(entries) == 0 {
		entries = l.collectEntries()
	}
	if len(entries) == 0 {
		return
	}

	fontSize := l.FontSize
	if fontSize <= 0 {
		fontSize = ctx.RC.FontSize * 0.92
	}
	if fontSize < 8 {
		fontSize = 8
	}

	maxLabelWidth := 0.0
	rowHeights := make([]float64, len(entries))
	labelMetrics := make([]render.TextMetrics, len(entries))
	for i, entry := range entries {
		metrics := r.MeasureText(entry.Label, fontSize, ctx.RC.FontKey)
		labelMetrics[i] = metrics
		if metrics.W > maxLabelWidth {
			maxLabelWidth = metrics.W
		}
		rowHeight := math.Max(metrics.H, fontSize)
		if rowHeight < 12 {
			rowHeight = 12
		}
		rowHeights[i] = rowHeight
	}

	contentHeight := 0.0
	for _, h := range rowHeights {
		contentHeight += h
	}
	if len(rowHeights) > 1 {
		contentHeight += l.RowGap * float64(len(rowHeights)-1)
	}

	boxWidth := l.Padding*2 + l.SampleWidth + l.SampleTextGap + maxLabelWidth
	boxHeight := l.Padding*2 + contentHeight
	box := l.legendBoxRect(ctx.Clip, boxWidth, boxHeight)

	r.Path(pixelRectPath(box), &render.Paint{
		Fill:      l.BackgroundColor,
		Stroke:    l.BorderColor,
		LineWidth: l.BorderWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})

	y := box.Min.Y + l.Padding
	for i, entry := range entries {
		rowHeight := rowHeights[i]
		centerY := y + rowHeight/2
		labelMetric := labelMetrics[i]

		l.drawSample(r, entry, geom.Rect{
			Min: geom.Pt{X: box.Min.X + l.Padding, Y: centerY - rowHeight/2},
			Max: geom.Pt{X: box.Min.X + l.Padding + l.SampleWidth, Y: centerY + rowHeight/2},
		})

		baselineY := centerY - labelMetric.H/2 + labelMetric.Ascent
		textRen.DrawText(
			entry.Label,
			geom.Pt{X: box.Min.X + l.Padding + l.SampleWidth + l.SampleTextGap, Y: baselineY},
			fontSize,
			l.TextColor,
		)

		y += rowHeight + l.RowGap
	}
}

// Z returns the legend z-order.
func (l *Legend) Z() float64 {
	return l.z
}

// Bounds returns an empty rect because legends do not contribute to data bounds.
func (l *Legend) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

func (l *Legend) collectEntries() []legendEntry {
	if l == nil || l.Axes == nil {
		return nil
	}

	entries := make([]legendEntry, 0, len(l.Axes.Artists))
	for _, art := range l.Axes.Artists {
		switch v := art.(type) {
		case *Legend:
			continue
		case *Line2D:
			if v.Label == "" {
				continue
			}
			entries = append(entries, legendEntry{
				Label:     v.Label,
				kind:      legendEntryLine,
				lineColor: v.Col,
				lineWidth: v.W,
				dashes:    append([]float64(nil), v.Dashes...),
			})
		case *Scatter2D:
			if v.Label == "" {
				continue
			}
			entry := legendEntry{
				Label:           v.Label,
				kind:            legendEntryMarker,
				marker:          v.Marker,
				markerFill:      v.Color,
				markerEdge:      v.EdgeColor,
				markerEdgeWidth: v.EdgeWidth,
			}
			if len(v.Colors) > 0 {
				entry.markerFill = v.Colors[0]
			}
			if len(v.EdgeColors) > 0 {
				entry.markerEdge = v.EdgeColors[0]
			}
			entries = append(entries, entry)
		case *Bar2D:
			if v.Label == "" {
				continue
			}
			entry := legendEntry{
				Label:          v.Label,
				kind:           legendEntryPatch,
				patchFill:      v.Color,
				patchEdge:      v.EdgeColor,
				patchEdgeWidth: v.EdgeWidth,
			}
			if len(v.Colors) > 0 {
				entry.patchFill = v.Colors[0]
			}
			if len(v.EdgeColors) > 0 {
				entry.patchEdge = v.EdgeColors[0]
			}
			entries = append(entries, entry)
		case *Fill2D:
			if v.Label == "" {
				continue
			}
			entries = append(entries, legendEntry{
				Label:          v.Label,
				kind:           legendEntryPatch,
				patchFill:      v.Color,
				patchEdge:      v.EdgeColor,
				patchEdgeWidth: v.EdgeWidth,
			})
		case *Hist2D:
			if v.Label == "" {
				continue
			}
			entries = append(entries, legendEntry{
				Label:          v.Label,
				kind:           legendEntryPatch,
				patchFill:      v.Color,
				patchEdge:      v.EdgeColor,
				patchEdgeWidth: v.EdgeWidth,
			})
		case *BoxPlot2D:
			if v.Label == "" {
				continue
			}
			entries = append(entries, legendEntry{
				Label:          v.Label,
				kind:           legendEntryPatch,
				patchFill:      v.Color,
				patchEdge:      v.EdgeColor,
				patchEdgeWidth: v.EdgeWidth,
			})
		case *Image2D:
			if v.Label == "" {
				continue
			}
			entries = append(entries, legendEntry{
				Label:          v.Label,
				kind:           legendEntryPatch,
				patchFill:      render.Color{R: 0.45, G: 0.45, B: 0.45, A: 1},
				patchEdge:      render.Color{R: 0.2, G: 0.2, B: 0.2, A: 0.9},
				patchEdgeWidth: 1,
			})
		case *ErrorBar:
			if v.Label == "" {
				continue
			}
			entries = append(entries, legendEntry{
				Label:     v.Label,
				kind:      legendEntryLine,
				lineColor: v.Color,
				lineWidth: v.LineWidth,
			})
		}
	}
	return entries
}

func (l *Legend) legendBoxRect(clip geom.Rect, width, height float64) geom.Rect {
	var minX, minY float64

	switch l.Location {
	case LegendUpperLeft:
		minX = clip.Min.X + l.Inset
		minY = clip.Min.Y + l.Inset
	case LegendLowerRight:
		minX = clip.Max.X - l.Inset - width
		minY = clip.Max.Y - l.Inset - height
	case LegendLowerLeft:
		minX = clip.Min.X + l.Inset
		minY = clip.Max.Y - l.Inset - height
	default:
		minX = clip.Max.X - l.Inset - width
		minY = clip.Min.Y + l.Inset
	}

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: minX + width, Y: minY + height},
	}
}

func (l *Legend) drawSample(r render.Renderer, entry legendEntry, sample geom.Rect) {
	center := geom.Pt{
		X: sample.Min.X + sample.W()/2,
		Y: sample.Min.Y + sample.H()/2,
	}

	switch entry.kind {
	case legendEntryPatch:
		patchRect := geom.Rect{
			Min: geom.Pt{X: sample.Min.X + 2, Y: center.Y - 5},
			Max: geom.Pt{X: sample.Max.X - 2, Y: center.Y + 5},
		}
		r.Path(pixelRectPath(patchRect), &render.Paint{
			Fill:      entry.patchFill,
			Stroke:    entry.patchEdge,
			LineWidth: entry.patchEdgeWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	case legendEntryMarker:
		sampleScatter := Scatter2D{
			Marker:    entry.marker,
			EdgeWidth: entry.markerEdgeWidth,
		}
		r.Path(sampleScatter.createMarkerPath(center, 5), &render.Paint{
			Fill:      entry.markerFill,
			Stroke:    entry.markerEdge,
			LineWidth: entry.markerEdgeWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
	default:
		lineWidth := entry.lineWidth
		if lineWidth <= 0 {
			lineWidth = 1.5
		}
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{
				{X: sample.Min.X + 1, Y: center.Y},
				{X: sample.Max.X - 1, Y: center.Y},
			},
		}
		r.Path(path, &render.Paint{
			Stroke:    entry.lineColor,
			LineWidth: lineWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
			Dashes:    entry.dashes,
		})
	}
}

func pixelRectPath(r geom.Rect) geom.Path {
	path := geom.Path{}
	corners := []geom.Pt{
		r.Min,
		{X: r.Max.X, Y: r.Min.Y},
		r.Max,
		{X: r.Min.X, Y: r.Max.Y},
	}
	for i, corner := range corners {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, corner)
	}
	path.C = append(path.C, geom.ClosePath)
	return path
}
