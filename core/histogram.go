package core

import (
	"math"
	"sort"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// BinStrategy specifies how to automatically determine histogram bin count.
type BinStrategy uint8

const (
	BinStrategyAuto    BinStrategy = iota // Sturges for n<1000, Scott otherwise
	BinStrategySturges                    // ceil(log2(n)) + 1
	BinStrategyScott                      // 3.49 * std * n^(-1/3)
	BinStrategyFD                         // 2 * IQR * n^(-1/3) (Freedman-Diaconis)
	BinStrategySqrt                       // ceil(sqrt(n))
)

// HistNorm specifies how to normalize histogram bin heights.
type HistNorm uint8

const (
	HistNormCount       HistNorm = iota // raw counts (default)
	HistNormProbability                 // count/total — each bar is fraction of total
	HistNormDensity                     // count/(total*width) — area integrates to 1
)

// Hist2D renders histogram plots computed from raw data.
// Bars span from edge[i] to edge[i+1] with no gap between adjacent bins.
type Hist2D struct {
	Data      []float64    // raw data values
	Bins      int          // number of bins (0 = auto)
	BinEdges  []float64    // explicit bin edges; overrides Bins when len > 1
	BinStrat  BinStrategy  // automatic binning strategy (used when Bins==0 and BinEdges==nil)
	Norm      HistNorm     // normalization mode
	Color     render.Color // bar fill color
	EdgeColor render.Color // bar outline color
	EdgeWidth float64      // bar outline width in pixels (0 = no outline)
	Alpha     float64      // alpha transparency (0-1, 0 means 1.0)
	Label     string       // series label for legend
	z         float64      // z-order

	// Computed lazily on first Draw/Bounds call.
	computed bool
	counts   []float64 // normalized heights per bin
	edges    []float64 // bin edge values (len = len(counts)+1)
}

// compute calculates bin edges and counts from Data.
func (h *Hist2D) compute() {
	if h.computed {
		return
	}
	h.computed = true

	if len(h.Data) == 0 {
		return
	}

	// Determine bin edges.
	if len(h.BinEdges) > 1 {
		h.edges = h.BinEdges
	} else {
		h.edges = computeBinEdges(h.Data, h.Bins, h.BinStrat)
	}

	nBins := len(h.edges) - 1
	if nBins <= 0 {
		return
	}

	// Count data in each bin. The last bin includes the right edge.
	raw := make([]float64, nBins)
	for _, v := range h.Data {
		idx := findBin(v, h.edges)
		if idx >= 0 && idx < nBins {
			raw[idx]++
		}
	}

	// Apply normalization.
	total := float64(len(h.Data))
	h.counts = make([]float64, nBins)
	for i, c := range raw {
		switch h.Norm {
		case HistNormProbability:
			h.counts[i] = c / total
		case HistNormDensity:
			width := h.edges[i+1] - h.edges[i]
			if width > 0 {
				h.counts[i] = c / (total * width)
			}
		default: // HistNormCount
			h.counts[i] = c
		}
	}
}

// findBin returns the bin index for value v using the given edges.
// The last bin is closed on both sides [edges[n-1], edges[n]].
func findBin(v float64, edges []float64) int {
	n := len(edges) - 1
	if v < edges[0] || v > edges[n] {
		return -1
	}
	if v == edges[n] {
		return n - 1 // include right edge in last bin
	}
	// Binary search.
	lo, hi := 0, n-1
	for lo < hi {
		mid := (lo + hi) / 2
		if v < edges[mid+1] {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// computeBinEdges computes evenly-spaced bin edges from data.
func computeBinEdges(data []float64, nBins int, strat BinStrategy) []float64 {
	// Find data range.
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	// If all values identical, create a single bin around that value.
	if minV == maxV {
		half := math.Max(math.Abs(minV)*0.5, 0.5)
		minV -= half
		maxV += half
		nBins = 1
	}

	if nBins <= 0 {
		nBins = autoBinCount(data, strat)
	}
	if nBins < 1 {
		nBins = 1
	}

	edges := make([]float64, nBins+1)
	width := (maxV - minV) / float64(nBins)
	for i := range edges {
		edges[i] = minV + float64(i)*width
	}
	// Ensure last edge is exactly maxV to avoid floating-point drift.
	edges[nBins] = maxV
	return edges
}

// autoBinCount chooses bin count based on strategy.
func autoBinCount(data []float64, strat BinStrategy) int {
	n := len(data)
	if n == 0 {
		return 1
	}

	switch strat {
	case BinStrategySturges:
		return sturgesBins(n)
	case BinStrategyScott:
		return scottBins(data)
	case BinStrategyFD:
		return fdBins(data)
	case BinStrategySqrt:
		return int(math.Ceil(math.Sqrt(float64(n))))
	default: // BinStrategyAuto
		if n < 1000 {
			return sturgesBins(n)
		}
		return scottBins(data)
	}
}

func sturgesBins(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Ceil(math.Log2(float64(n)))) + 1
}

func scottBins(data []float64) int {
	n := len(data)
	sigma := stddev(data)
	if sigma == 0 {
		return sturgesBins(n)
	}
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	h := 3.49 * sigma * math.Pow(float64(n), -1.0/3.0)
	if h <= 0 {
		return sturgesBins(n)
	}
	k := int(math.Ceil((maxV - minV) / h))
	if k < 1 {
		return 1
	}
	return k
}

func fdBins(data []float64) int {
	n := len(data)
	iqr := computeIQR(data)
	if iqr == 0 {
		return sturgesBins(n)
	}
	minV, maxV := data[0], data[0]
	for _, v := range data[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	h := 2.0 * iqr * math.Pow(float64(n), -1.0/3.0)
	if h <= 0 {
		return sturgesBins(n)
	}
	k := int(math.Ceil((maxV - minV) / h))
	if k < 1 {
		return 1
	}
	return k
}

func stddev(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(len(data))
	variance := 0.0
	for _, v := range data {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(data) - 1)
	return math.Sqrt(variance)
}

func computeIQR(data []float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	n := len(sorted)
	q1 := sorted[n/4]
	q3 := sorted[3*n/4]
	return q3 - q1
}

// Draw renders the histogram bars.
func (h *Hist2D) Draw(r render.Renderer, ctx *DrawContext) {
	h.compute()
	if len(h.counts) == 0 {
		return
	}

	alpha := h.Alpha
	if alpha <= 0 {
		alpha = 1.0
	}
	if alpha > 1 {
		alpha = 1.0
	}

	fillColor := h.Color
	fillColor.A *= alpha

	edgeColor := h.EdgeColor
	edgeColor.A *= alpha

	for i, count := range h.counts {
		left := h.edges[i]
		right := h.edges[i+1]

		// Transform corners to pixel space.
		bl := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: 0})
		br := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: 0})
		tr := ctx.DataToPixel.Apply(geom.Pt{X: right, Y: count})
		tl := ctx.DataToPixel.Apply(geom.Pt{X: left, Y: count})

		path := geom.Path{}
		path.C = append(path.C, geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath)
		path.V = append(path.V, bl, br, tr, tl)

		paint := render.Paint{Fill: fillColor}
		if h.EdgeWidth > 0 && edgeColor.A > 0 {
			paint.Stroke = edgeColor
			paint.LineWidth = h.EdgeWidth
			paint.LineJoin = render.JoinMiter
			paint.LineCap = render.CapSquare
		}

		r.Path(path, &paint)
	}
}

// Z returns the z-order for sorting.
func (h *Hist2D) Z() float64 {
	return h.z
}

// Bounds returns the bounding box of the histogram in data coordinates.
func (h *Hist2D) Bounds(*DrawContext) geom.Rect {
	h.compute()
	if len(h.counts) == 0 || len(h.edges) < 2 {
		return geom.Rect{}
	}

	maxCount := h.counts[0]
	for _, c := range h.counts[1:] {
		if c > maxCount {
			maxCount = c
		}
	}

	return geom.Rect{
		Min: geom.Pt{X: h.edges[0], Y: 0},
		Max: geom.Pt{X: h.edges[len(h.edges)-1], Y: maxCount},
	}
}

// BinCounts returns the computed bin edges and counts.
// Useful for inspecting histogram results without drawing.
func (h *Hist2D) BinCounts() (edges, counts []float64) {
	h.compute()
	return h.edges, h.counts
}
