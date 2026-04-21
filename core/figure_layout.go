package core

import (
	"math"
	"sort"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

type axisAlignmentKey struct {
	side AxisSide
	line int
}

type figureTextAlignment struct {
	titleExtents  map[axisAlignmentKey]float64
	xLabelExtents map[axisAlignmentKey]float64
	yLabelExtents map[axisAlignmentKey]float64
}

func newAxesDrawContext(ax *Axes, fig *Figure, figureRect, clip geom.Rect) *DrawContext {
	proj := cloneProjection(ax.projection)
	if proj == nil {
		proj, _ = lookupProjection("rectilinear")
	}
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      ax.effectiveXScale(),
			YScale:      ax.effectiveYScale(),
			DataToAxes:  proj.DataToAxes(ax),
			AxesToPixel: transform.NewDisplayRectTransform(clip),
		},
		Axes:       ax,
		Projection: proj,
		RC:         ax.effectiveRC(fig),
		Clip:       clip,
		FigureRect: figureRect,
	}
}

func newFigureDrawContext(fig *Figure, figureRect geom.Rect) *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			DataToAxes:  transform.NewScaleTransform(transform.NewLinear(0, 1), transform.NewLinear(0, 1)),
			AxesToPixel: transform.NewDisplayRectTransform(figureRect),
		},
		Axes:       nil,
		Projection: cloneProjection(nil),
		RC:         fig.RC,
		Clip:       figureRect,
		FigureRect: figureRect,
	}
}

func computeFigureTextAlignment(fig *Figure, r render.Renderer, figureRect geom.Rect) figureTextAlignment {
	alignment := figureTextAlignment{
		titleExtents:  map[axisAlignmentKey]float64{},
		xLabelExtents: map[axisAlignmentKey]float64{},
		yLabelExtents: map[axisAlignmentKey]float64{},
	}
	if fig == nil {
		return alignment
	}

	for _, ax := range fig.Children {
		px := ax.adjustedLayout(fig)
		ctx := newAxesDrawContext(ax, fig, figureRect, px)

		if ax.Title != "" {
			key := alignmentKey(AxisTop, spinePixelY(AxisTop, px))
			extent := titleTopExtent(ax, r, ctx, px)
			if current, ok := alignment.titleExtents[key]; !ok || extent < current {
				alignment.titleExtents[key] = extent
			}
		}

		if ax.XLabel != "" {
			side := ax.effectiveXLabelSide()
			key := alignmentKey(side, spinePixelY(side, px))
			extent := xLabelExtent(ax, r, ctx, px, side)
			if side == AxisTop {
				if current, ok := alignment.xLabelExtents[key]; !ok || extent < current {
					alignment.xLabelExtents[key] = extent
				}
			} else if current, ok := alignment.xLabelExtents[key]; !ok || extent > current {
				alignment.xLabelExtents[key] = extent
			}
		}

		if ax.YLabel != "" {
			side := ax.effectiveYLabelSide()
			key := alignmentKey(side, spinePixelX(side, px))
			extent := yLabelExtent(ax, r, ctx, px, side)
			if side == AxisRight {
				if current, ok := alignment.yLabelExtents[key]; !ok || extent > current {
					alignment.yLabelExtents[key] = extent
				}
			} else if current, ok := alignment.yLabelExtents[key]; !ok || extent < current {
				alignment.yLabelExtents[key] = extent
			}
		}
	}

	return alignment
}

func alignmentKey(side AxisSide, coord float64) axisAlignmentKey {
	return axisAlignmentKey{
		side: side,
		line: int(math.Round(coord)),
	}
}

func titleTopExtent(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) float64 {
	extent := spinePixelY(AxisTop, px)
	for _, candidate := range []*Axis{ax.effectiveXAxis(), ax.effectiveTopAxis()} {
		if candidate == nil || candidate.Side != AxisTop || !candidate.ShowLabels {
			continue
		}
		if tickBounds, ok := axisTickLabelBounds(candidate, r, ctx); ok {
			extent = math.Min(extent, tickBounds.Min.Y)
		}
	}
	return extent
}

func xLabelExtent(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide) float64 {
	extent := spinePixelY(side, px)
	xAxis := ax.axisForXLabelSide(side)
	if xAxis == nil {
		return extent
	}
	if tickBounds, ok := axisTickLabelBounds(xAxis, r, ctx); ok {
		if side == AxisTop {
			return math.Min(extent, tickBounds.Min.Y)
		}
		return math.Max(extent, tickBounds.Max.Y)
	}
	return extent
}

func yLabelExtent(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, side AxisSide) float64 {
	extent := spinePixelX(side, px)
	yAxis := ax.axisForYLabelSide(side)
	if yAxis == nil {
		return extent
	}
	if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
		if side == AxisRight {
			return math.Max(extent, tickBounds.Max.X)
		}
		return math.Min(extent, tickBounds.Min.X)
	}
	return extent
}

func drawFigureArtists(fig *Figure, r render.Renderer, figureRect geom.Rect) {
	if fig == nil || len(fig.Artists) == 0 {
		return
	}

	if !fig.zsorted {
		sortArtists(fig.Artists)
		fig.zsorted = true
	}

	ctx := newFigureDrawContext(fig, figureRect)
	for _, art := range fig.Artists {
		art.Draw(r, ctx)
	}
	for _, art := range fig.Artists {
		if overlay, ok := art.(OverlayArtist); ok {
			overlay.DrawOverlay(r, ctx)
		}
	}
}

func drawFigureLabels(fig *Figure, r render.Renderer, figureRect geom.Rect) {
	if fig == nil {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	ctx := newFigureDrawContext(fig, figureRect)
	titleColor := fig.RC.DefaultAxesTitleColor()
	labelColor := fig.RC.DefaultAxesLabelColor()
	titleSize := titleFontSize(ctx)
	labelSize := axisLabelFontSize(ctx)
	centerX := figureRect.Min.X + figureRect.W()/2
	centerY := figureRect.Min.Y + figureRect.H()/2

	if fig.SupTitle != "" {
		layout := measureSingleLineTextLayout(r, fig.SupTitle, titleSize, fig.RC.FontKey)
		anchor := geom.Pt{
			X: centerX,
			Y: figureRect.Min.Y + pointsToPixels(fig.RC, 6),
		}
		drawDisplayText(
			textRen,
			fig.SupTitle,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignTop),
			titleSize,
			titleColor,
		)
	}

	if fig.SupXLabel != "" {
		layout := measureSingleLineTextLayout(r, fig.SupXLabel, labelSize, fig.RC.FontKey)
		anchor := geom.Pt{
			X: centerX,
			Y: figureRect.Max.Y - pointsToPixels(fig.RC, 4),
		}
		drawDisplayText(
			textRen,
			fig.SupXLabel,
			alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignBottom),
			labelSize,
			labelColor,
		)
	}

	if fig.SupYLabel != "" {
		anchor := geom.Pt{
			X: figureRect.Min.X + pointsToPixels(fig.RC, 4),
			Y: centerY,
		}
		switch ren := r.(type) {
		case render.RotatedTextDrawer:
			drawDisplayTextRotated(ren, fig.SupYLabel, anchor, labelSize, math.Pi/2, labelColor)
		case render.VerticalTextDrawer:
			drawDisplayTextVertical(ren, fig.SupYLabel, anchor, labelSize, labelColor)
		default:
			layout := measureSingleLineTextLayout(r, fig.SupYLabel, labelSize, fig.RC.FontKey)
			drawDisplayText(
				textRen,
				fig.SupYLabel,
				alignedSingleLineOrigin(anchor, layout, TextAlignLeft, textLayoutVAlignCenter),
				labelSize,
				labelColor,
			)
		}
	}
}

func pointsToPixels(rc style.RC, points float64) float64 {
	dpi := rc.DPI
	if dpi <= 0 {
		dpi = style.CurrentDefaults().DPI
		if dpi <= 0 {
			dpi = 96
		}
	}
	return points * dpi / 72.0
}

func sortArtists(artists []Artist) {
	if len(artists) < 2 {
		return
	}
	sort.SliceStable(artists, func(i, j int) bool {
		zi, zj := artists[i].Z(), artists[j].Z()
		if zi == zj {
			return i < j
		}
		return zi < zj
	})
}
