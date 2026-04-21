package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

type textLayoutVerticalAlign uint8

const (
	textLayoutVAlignTop textLayoutVerticalAlign = iota
	textLayoutVAlignBottom
	textLayoutVAlignCenter
	textLayoutVAlignBaseline
	textLayoutVAlignCenterBaseline
)

type singleLineTextLayout struct {
	Width         float64
	InkBounds     render.TextBounds
	HaveInkBounds bool
	RunAscent     float64
	RunDescent    float64
	MinAscent     float64
	MinDescent    float64
	LineGap       float64
	Ascent        float64
	Descent       float64
	Height        float64
}

func measureSingleLineTextLayout(r render.Renderer, text string, size float64, fontKey string) singleLineTextLayout {
	display := normalizeDisplayText(text)
	metrics := r.MeasureText(display, size, fontKey)
	bounds, haveBounds := measureTextBounds(r, display, size, fontKey)

	fontHeights := render.FontHeightMetrics{
		Ascent:  max0(metrics.Ascent),
		Descent: max0(metrics.Descent),
	}
	if metricer, ok := r.(render.TextFontMetricer); ok {
		if measured, ok := metricer.MeasureFontHeights(size, fontKey); ok {
			fontHeights = measured
		}
	}

	layout := singleLineTextLayout{
		Width:         metrics.W,
		InkBounds:     bounds,
		HaveInkBounds: haveBounds && bounds.W > 0 && bounds.H > 0,
		LineGap:       max0(fontHeights.LineGap),
		MinAscent:     max0(fontHeights.Ascent),
		MinDescent:    max0(fontHeights.Descent),
	}
	if layout.Width <= 0 && layout.HaveInkBounds {
		layout.Width = bounds.W
	}

	layout.RunAscent, layout.RunDescent = actualTextVerticalExtents(metrics, bounds, layout.HaveInkBounds)
	layout.Ascent = math.Max(layout.RunAscent, layout.MinAscent)
	layout.Descent = math.Max(layout.RunDescent, layout.MinDescent)
	layout.Height = layout.Ascent + layout.Descent

	if layout.Height <= 0 {
		layout.Ascent = max0(metrics.Ascent)
		layout.Descent = max0(metrics.Descent)
		layout.Height = layout.Ascent + layout.Descent
	}
	if layout.Height <= 0 && layout.HaveInkBounds {
		layout.Ascent = max0(-bounds.Y)
		layout.Descent = max0(bounds.Y + bounds.H)
		layout.Height = layout.Ascent + layout.Descent
	}

	return layout
}

func actualTextVerticalExtents(metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) (ascent, descent float64) {
	if haveBounds {
		ascent = max0(-bounds.Y)
		descent = max0(bounds.Y + bounds.H)
		return ascent, descent
	}
	return max0(metrics.Ascent), max0(metrics.Descent)
}

func textBaselineOffset(layout singleLineTextLayout, align textLayoutVerticalAlign) float64 {
	switch align {
	case textLayoutVAlignTop:
		return layout.Ascent
	case textLayoutVAlignBottom:
		return -layout.Descent
	case textLayoutVAlignCenter:
		return (layout.Ascent - layout.Descent) / 2
	case textLayoutVAlignCenterBaseline:
		return layout.Ascent / 2
	default:
		return 0
	}
}

func textHorizontalOriginOffset(layout singleLineTextLayout, align TextAlign) float64 {
	switch align {
	case TextAlignLeft:
		if layout.HaveInkBounds {
			return -layout.InkBounds.X
		}
		return 0
	case TextAlignRight:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W
		}
		return layout.Width
	default:
		if layout.HaveInkBounds {
			return layout.InkBounds.X + layout.InkBounds.W/2
		}
		return layout.Width / 2
	}
}

func alignedSingleLineOrigin(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign, vAlign textLayoutVerticalAlign) geom.Pt {
	return geom.Pt{
		X: anchor.X - textHorizontalOriginOffset(layout, hAlign),
		Y: anchor.Y + textBaselineOffset(layout, vAlign),
	}
}

func layoutVerticalAlign(vAlign TextVerticalAlign, preferCenterBaseline bool) textLayoutVerticalAlign {
	switch vAlign {
	case TextVAlignTop:
		return textLayoutVAlignTop
	case TextVAlignBottom:
		return textLayoutVAlignBottom
	case TextVAlignBaseline:
		return textLayoutVAlignBaseline
	case TextVAlignMiddle:
		if preferCenterBaseline {
			return textLayoutVAlignCenterBaseline
		}
		return textLayoutVAlignCenter
	default:
		return textLayoutVAlignBaseline
	}
}

func max0(v float64) float64 {
	return math.Max(v, 0)
}
