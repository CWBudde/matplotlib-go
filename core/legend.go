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
	markerPath      geom.Path
	markerFill      render.Color
	markerEdge      render.Color
	markerEdgeWidth float64

	patchFill       render.Color
	patchEdge       render.Color
	patchEdgeWidth  float64
	patchHatch      string
	patchHatchColor render.Color
	patchHatchWidth float64
}

type legendEntryProvider interface {
	legendEntry() (legendEntry, bool)
}

// Legend renders a styled legend box inside an axes.
// If no explicit internal entries are present, labeled artists on the owning axes are collected automatically.
type Legend struct {
	Axes   *Axes
	Figure *Figure

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
	rc := style.CurrentDefaults()
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
		FontSize:        rc.LegendSize(),
		z:               1_000,
	}
}

// NewFigureLegend creates a legend bound to the provided figure.
func NewFigureLegend(fig *Figure) *Legend {
	rc := style.CurrentDefaults()
	if fig != nil {
		rc = fig.RC
	}
	return &Legend{
		Figure:          fig,
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
		FontSize:        rc.LegendSize(),
		z:               1_000,
	}
}

// AddLegend appends a legend to the axes.
func (a *Axes) AddLegend() *Legend {
	legend := NewLegend(a)
	a.Add(legend)
	return legend
}

// AddLegend appends a figure-level legend that collects labeled artists from all axes.
func (f *Figure) AddLegend() *Legend {
	legend := NewFigureLegend(f)
	f.Add(legend)
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
		fontSize = ctx.RC.LegendSize()
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
		drawDisplayText(
			textRen,
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
	if l == nil {
		return nil
	}

	switch {
	case l.Axes != nil:
		return collectLegendEntries(l.Axes.Artists)
	case l.Figure != nil:
		var entries []legendEntry
		for _, ax := range l.Figure.Children {
			entries = append(entries, collectLegendEntries(ax.Artists)...)
		}
		return entries
	default:
		return nil
	}
}

func collectLegendEntries(artists []Artist) []legendEntry {
	entries := make([]legendEntry, 0, len(artists))
	for _, art := range artists {
		switch art.(type) {
		case *Legend:
			continue
		default:
			provider, ok := art.(legendEntryProvider)
			if !ok {
				continue
			}
			entry, ok := provider.legendEntry()
			if !ok {
				continue
			}
			entries = append(entries, entry)
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
		patch := Patch{
			FaceColor:  entry.patchFill,
			EdgeColor:  entry.patchEdge,
			EdgeWidth:  entry.patchEdgeWidth,
			Hatch:      entry.patchHatch,
			HatchColor: entry.patchHatchColor,
			HatchWidth: entry.patchHatchWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		}
		patch.drawStyledPath(r, pixelRectPath(patchRect), geom.Path{})
	case legendEntryMarker:
		markerPath := entry.markerPath
		if len(markerPath.C) == 0 {
			sampleScatter := Scatter2D{Marker: entry.marker}
			markerPath = sampleScatter.createMarkerPath(center, 5)
		} else {
			markerPath = scaleAndTranslatePath(markerPath, 5, center)
		}
		r.Path(markerPath, &render.Paint{
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
