package core

import (
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
	render.TextLineLayout
	MathLayout *MathTextLayout
}

func measureSingleLineTextLayout(r render.Renderer, text string, size float64, fontKey string) singleLineTextLayout {
	if layout, ok := layoutDisplayText(r, text, size, fontKey); ok {
		return singleLineTextLayout{
			TextLineLayout: render.TextLineLayout{
				Width:   layout.Width,
				Ascent:  layout.Ascent,
				Descent: layout.Descent,
				Height:  layout.Height,
			},
			MathLayout: &layout,
		}
	}

	display := normalizeDisplayText(text)
	return singleLineTextLayout{
		TextLineLayout: render.MeasureTextLineLayout(r, display, size, fontKey),
	}
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
