package render

import "math"

// TextLineLayout captures renderer-measured single-line text geometry.
type TextLineLayout struct {
	Width         float64
	InkBounds     TextBounds
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

// MeasureTextLineLayout combines run metrics, ink bounds, and font-wide
// vertical metrics into one stable line layout record.
func MeasureTextLineLayout(r Renderer, text string, size float64, fontKey string) TextLineLayout {
	if r == nil || text == "" || size <= 0 {
		return TextLineLayout{}
	}

	metrics := r.MeasureText(text, size, fontKey)
	bounds, haveBounds := measureTextBounds(r, text, size, fontKey)

	fontHeights := FontHeightMetrics{
		Ascent:  max0(metrics.Ascent),
		Descent: max0(metrics.Descent),
	}
	if metricer, ok := r.(TextFontMetricer); ok {
		if measured, ok := metricer.MeasureFontHeights(size, fontKey); ok {
			fontHeights = measured
		}
	}

	layout := TextLineLayout{
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

	layout.RunAscent, layout.RunDescent = ActualTextVerticalExtents(metrics, bounds, layout.HaveInkBounds)
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

// ActualTextVerticalExtents returns the run-specific ascent/descent, preferring
// ink bounds when the renderer can provide them.
func ActualTextVerticalExtents(metrics TextMetrics, bounds TextBounds, haveBounds bool) (ascent, descent float64) {
	if metrics.Ascent > 0 || metrics.Descent > 0 {
		return max0(metrics.Ascent), max0(metrics.Descent)
	}
	if haveBounds {
		ascent = max0(-bounds.Y)
		descent = max0(bounds.Y + bounds.H)
		return ascent, descent
	}
	return max0(metrics.Ascent), max0(metrics.Descent)
}

func measureTextBounds(r Renderer, text string, size float64, fontKey string) (TextBounds, bool) {
	if bounder, ok := r.(TextBounder); ok {
		return bounder.MeasureTextBounds(text, size, fontKey)
	}
	return TextBounds{}, false
}

func max0(v float64) float64 {
	return math.Max(v, 0)
}
