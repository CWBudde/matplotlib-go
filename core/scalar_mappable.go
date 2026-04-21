package core

import (
	"math"

	matcolor "matplotlib-go/color"
	"matplotlib-go/render"
)

// ScalarMappable describes artists that map scalar values through a colormap
// and expose that mapping to helpers such as colorbars.
type ScalarMappable interface {
	ScalarMap() ScalarMapInfo
}

// ScalarMapInfo stores the colormap and value range used by a scalar-mappable
// artist.
type ScalarMapInfo struct {
	Colormap string
	VMin     float64
	VMax     float64
}

// Resolved returns a copy with sane defaults for downstream consumers.
func (m ScalarMapInfo) Resolved() ScalarMapInfo {
	if m.Colormap == "" {
		m.Colormap = "viridis"
	}
	if !isFinite(m.VMin) || !isFinite(m.VMax) {
		m.VMin = 0
		m.VMax = 1
	}
	if m.VMin == m.VMax {
		m.VMax = m.VMin + 1
	}
	return m
}

// Normalize maps a scalar into the colormap domain.
func (m ScalarMapInfo) Normalize(v float64) float64 {
	m = m.Resolved()
	span := m.VMax - m.VMin
	if span == 0 {
		return 0
	}
	return clamp01((v - m.VMin) / span)
}

// Color maps a scalar into a display color using the configured colormap.
func (m ScalarMapInfo) Color(v, alpha float64) render.Color {
	color := matcolor.GetColormap(m.Resolved().Colormap).At(m.Normalize(v))
	color.A *= clampOneToOne(alpha)
	return color
}

func resolveScalarMapGrid(data [][]float64, cmap string, vmin, vmax *float64) ScalarMapInfo {
	minValue, maxValue := dataRange(data)
	if vmin != nil && isFinite(*vmin) {
		minValue = *vmin
	}
	if vmax != nil && isFinite(*vmax) {
		maxValue = *vmax
	}
	return ScalarMapInfo{
		Colormap: resolvedColormapName(cmap),
		VMin:     minValue,
		VMax:     maxValue,
	}.Resolved()
}

func resolveScalarMapValues(values []float64, cmap string, vmin, vmax *float64) ScalarMapInfo {
	minValue, maxValue := finiteRange(values)
	if vmin != nil && isFinite(*vmin) {
		minValue = *vmin
	}
	if vmax != nil && isFinite(*vmax) {
		maxValue = *vmax
	}
	return ScalarMapInfo{
		Colormap: resolvedColormapName(cmap),
		VMin:     minValue,
		VMax:     maxValue,
	}.Resolved()
}

func resolvedColormapName(name string) string {
	if name == "" {
		return "viridis"
	}
	return name
}

func finiteRange(values []float64) (float64, float64) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	if math.IsInf(minValue, 1) || math.IsInf(maxValue, -1) {
		return 0, 1
	}
	if minValue == maxValue {
		return minValue, minValue + 1
	}
	return minValue, maxValue
}

func finiteMatrixSize(data [][]float64) (rows int, cols int, ok bool) {
	rows = len(data)
	if rows == 0 {
		return 0, 0, false
	}
	cols = len(data[0])
	if cols == 0 {
		return 0, 0, false
	}
	for _, row := range data[1:] {
		if len(row) != cols {
			return 0, 0, false
		}
	}
	return rows, cols, true
}
