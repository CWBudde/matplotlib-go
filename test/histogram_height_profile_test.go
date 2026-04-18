package test

import (
	"image"
	"image/color"
	"math"
	"path/filepath"
	"testing"

	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/test/imagecmp"
	"matplotlib-go/transform"
)

const (
	histogramFigureWidth  = 640
	histogramFigureHeight = 360

	// Histogram axes rectangle in figure-fraction coordinates, shared by all histogram tests.
	histogramAxesMinX = 0.12
	histogramAxesMinY = 0.12
	histogramAxesMaxX = 0.95
	histogramAxesMaxY = 0.90
)

type histogramHeightSeries struct {
	edges   []float64
	heights []float64
}

type histogramHeightCase struct {
	name    string
	render  func() image.Image
	series  []histogramHeightSeries
	maxDiff int
}

func TestHistogramTopHeights_ParsedDataParity(t *testing.T) {
	payload := loadMatplotlibRNGDebugPayload(t)
	getSeries := func(name string) histogramHeightSeries {
		data, ok := payload.HistogramData[name]
		if !ok {
			t.Fatalf("missing payload series %s", name)
		}
		if len(data.Edges) < 2 || len(data.Heights) == 0 {
			t.Fatalf("invalid payload series %s: edges=%d heights=%d", name, len(data.Edges), len(data.Heights))
		}
		return histogramHeightSeries{
			edges:   data.Edges,
			heights: data.Heights,
		}
	}

	basic := getSeries("hist_basic")
	density := getSeries("hist_density")
	strat1 := getSeries("hist_strategies_data1")
	strat2 := getSeries("hist_strategies_data2")

	axes := geom.Rect{
		Min: geom.Pt{X: histogramAxesMinX, Y: histogramAxesMinY},
		Max: geom.Pt{X: histogramAxesMaxX, Y: histogramAxesMaxY},
	}

	tests := []histogramHeightCase{
		{
			name:    "hist_basic",
			render:  renderHistBasic,
			series:  []histogramHeightSeries{basic},
			maxDiff: 1,
		},
		{
			name:    "hist_density",
			render:  renderHistDensity,
			series:  []histogramHeightSeries{density},
			maxDiff: 1,
		},
		{
			name:    "hist_strategies",
			render:  renderHistStrategies,
			series:  []histogramHeightSeries{strat1, strat2},
			maxDiff: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			expected := expectedHistogramTopProfile(t, axes, tc.series)

			got := tc.render()
			assertHistogramTopProfile(t, "got", got, expected, tc.maxDiff)

			goldenPath := filepath.Join("..", "testdata", "golden", tc.name+".png")
			golden, err := imagecmp.LoadPNG(goldenPath)
			if err != nil {
				t.Fatalf("load golden image %s: %v", goldenPath, err)
			}
			assertHistogramTopProfile(t, "golden", golden, expected, tc.maxDiff)

			refPath := filepath.Join("..", "testdata", "matplotlib_ref", tc.name+".png")
			ref, err := imagecmp.LoadPNG(refPath)
			if err != nil {
				t.Fatalf("load matplotlib reference image %s: %v", refPath, err)
			}
			assertHistogramTopProfile(t, "matplotlib_ref", ref, expected, tc.maxDiff)
		})
	}
}

func expectedHistogramTopProfile(t *testing.T, axes geom.Rect, series []histogramHeightSeries) []int {
	t.Helper()

	if len(series) == 0 {
		t.Fatal("expected at least one series")
	}

	xMin := math.Inf(1)
	xMax := math.Inf(-1)
	yMax := 0.0
	for _, s := range series {
		if len(s.edges) < 2 {
			continue
		}
		if s.edges[0] < xMin {
			xMin = s.edges[0]
		}
		if s.edges[len(s.edges)-1] > xMax {
			xMax = s.edges[len(s.edges)-1]
		}
		for _, h := range s.heights {
			if h > yMax {
				yMax = h
			}
		}
	}
	if xMin >= xMax || yMax <= 0 {
		t.Fatal("invalid histogram geometry for profile expectations")
	}
	axMargin := 0.05

	ax := geom.Rect{
		Min: geom.Pt{X: float64(histogramFigureWidth) * axes.Min.X, Y: float64(histogramFigureHeight) * axes.Min.Y},
		Max: geom.Pt{X: float64(histogramFigureWidth) * axes.Max.X, Y: float64(histogramFigureHeight) * axes.Max.Y},
	}
	xSpan := xMax - xMin
	xMin -= xSpan * axMargin
	xMax += xSpan * axMargin
	yScaleMax := yMax
	yScaleMax *= (1.0 + axMargin)

	xScale := transform.NewLinear(xMin, xMax)
	yScale := transform.NewLinear(0, yScaleMax)
	transform2D := core.Transform2D{
		XScale:      xScale,
		YScale:      yScale,
		AxesToPixel: transform.NewAffine(histogramAxesToPixel(ax)),
	}

	xStart := int(math.Ceil(ax.Min.X))
	if xStart < 0 {
		xStart = 0
	}
	xEnd := int(math.Floor(ax.Max.X))
	if xEnd > histogramFigureWidth {
		xEnd = histogramFigureWidth
	}
	height := make([]int, xEnd-xStart)
	for i := range height {
		height[i] = -1
	}

	for _, s := range series {
		if len(s.edges) < 2 || len(s.heights) == 0 {
			continue
		}

		for binIdx := 0; binIdx < len(s.heights) && binIdx+1 < len(s.edges); binIdx++ {
			h := s.heights[binIdx]
			if h <= 0 {
				continue
			}

			leftPt := transform2D.Apply(geom.Pt{X: s.edges[binIdx], Y: 0})
			rightPt := transform2D.Apply(geom.Pt{X: s.edges[binIdx+1], Y: 0})
			topPt := transform2D.Apply(geom.Pt{X: s.edges[binIdx], Y: h})

			left := math.Min(leftPt.X, rightPt.X)
			right := math.Max(leftPt.X, rightPt.X)
			top := int(math.Round(topPt.Y))

			for x := xStart; x < xEnd; x++ {
				colCenter := float64(x) + 0.5
				if colCenter < left || colCenter >= right {
					continue
				}
				idx := x - xStart
				if height[idx] < 0 || top < height[idx] {
					height[idx] = top
				}
			}
		}
	}

	return height
}

func assertHistogramTopProfile(
	t *testing.T,
	name string,
	img image.Image,
	expected []int,
	maxDiff int,
) {
	t.Helper()

	ax := geom.Rect{
		Min: geom.Pt{X: float64(histogramFigureWidth) * histogramAxesMinX, Y: float64(histogramFigureHeight) * histogramAxesMinY},
		Max: geom.Pt{X: float64(histogramFigureWidth) * histogramAxesMaxX, Y: float64(histogramFigureHeight) * histogramAxesMaxY},
	}
	observed := extractHistogramTopProfile(img, ax)

	if len(observed) != len(expected) {
		t.Fatalf("%s histogram profile width mismatch: got %d, expected %d", name, len(observed), len(expected))
	}

	mismatches := 0
	for i := range expected {
		x := i + int(math.Ceil(histogramAxesMinX*float64(histogramFigureWidth)))
		e := expected[i]
		o := observed[i]
		if e < 0 {
			if o >= 0 {
				mismatches++
				if mismatches <= 8 {
					t.Logf("%s unexpected top at x=%d: got=%d expected=%d", name, x, o, e)
				}
			}
			continue
		}
		if o < 0 {
			mismatches++
			if mismatches <= 8 {
				t.Logf("%s profile missing top at x=%d (expected=%d)", name, x, e)
			}
			continue
		}
		diff := int(math.Abs(float64(o - e)))
		if diff > maxDiff {
			mismatches++
			if mismatches <= 8 {
				t.Logf("%s profile mismatch at x=%d: got=%d expected=%d diff=%d", name, x, o, e, diff)
			}
		}
	}

	if mismatches > 0 {
		t.Fatalf("%s histogram height profile mismatch count=%d", name, mismatches)
	}
}

func extractHistogramTopProfile(img image.Image, axes geom.Rect) []int {
	bounds := img.Bounds()
	xStart := int(math.Ceil(axes.Min.X))
	yStart := int(math.Ceil(axes.Min.Y))
	xEnd := int(math.Floor(axes.Max.X))
	yEnd := int(math.Floor(axes.Max.Y)) - 1

	if xStart < bounds.Min.X {
		xStart = bounds.Min.X
	}
	if xEnd > bounds.Max.X {
		xEnd = bounds.Max.X
	}
	if yStart < bounds.Min.Y {
		yStart = bounds.Min.Y
	}
	if yEnd > bounds.Max.Y {
		yEnd = bounds.Max.Y
	}

	profile := make([]int, xEnd-xStart)
	for i := range profile {
		profile[i] = -1
	}

	for x := xStart; x < xEnd; x++ {
		for y := yStart; y < yEnd; y++ {
			rgba := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			if isHistogramFillPixel(rgba) {
				profile[x-xStart] = y
				break
			}
		}
	}

	return profile
}

func isHistogramFillPixel(p color.RGBA) bool {
	if p.A == 0 {
		return false
	}
	// Discard pure/near-white background.
	if p.R > 248 && p.G > 248 && p.B > 248 {
		return false
	}

	maxCh := p.R
	if p.G > maxCh {
		maxCh = p.G
	}
	if p.B > maxCh {
		maxCh = p.B
	}
	minCh := p.R
	if p.G < minCh {
		minCh = p.G
	}
	if p.B < minCh {
		minCh = p.B
	}
	if maxCh-minCh <= 2 {
		return false
	}

	return true
}

func histogramAxesToPixel(rect geom.Rect) geom.Affine {
	sx := rect.W()
	sy := -rect.H()
	tx := rect.Min.X
	ty := rect.Max.Y
	return geom.Affine{A: sx, D: sy, E: tx, F: ty}
}
