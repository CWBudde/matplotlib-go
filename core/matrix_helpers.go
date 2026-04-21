package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// MatShowOptions configures Axes.MatShow.
type MatShowOptions struct {
	Colormap     *string
	VMin         *float64
	VMax         *float64
	Alpha        *float64
	Aspect       string
	IntegerTicks *bool
	Label        string
}

// SpyOptions configures Axes.Spy.
type SpyOptions struct {
	Precision  float64
	UseImage   *bool
	Marker     *MarkerType
	MarkerSize float64
	Color      *render.Color
	Alpha      *float64
	Aspect     string
	Label      string
}

// AnnotatedHeatmapOptions configures Axes.AnnotatedHeatmap.
type AnnotatedHeatmapOptions struct {
	MatShowOptions
	Format        string
	FontSize      float64
	TextColor     render.Color
	TextColorHigh render.Color
	Threshold     *float64
	SkipNaN       bool
	NaNText       string
}

// SpyResult groups the artists and coordinates produced by Axes.Spy.
type SpyResult struct {
	Image   *Image2D
	Markers *PathCollection
	Indices []geom.Pt
}

// AnnotatedHeatmapResult stores the image plus per-cell annotations.
type AnnotatedHeatmapResult struct {
	Image  *Image2D
	Labels []*Text
}

// MatShow renders a matrix with centered integer ticks and an equal aspect.
func (a *Axes) MatShow(data [][]float64, opts ...MatShowOptions) *Image2D {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	cfg := MatShowOptions{
		Aspect: "equal",
	}
	if len(opts) > 0 {
		opt := opts[0]
		if opt.Colormap != nil {
			cfg.Colormap = opt.Colormap
		}
		if opt.VMin != nil {
			cfg.VMin = opt.VMin
		}
		if opt.VMax != nil {
			cfg.VMax = opt.VMax
		}
		if opt.Alpha != nil {
			cfg.Alpha = opt.Alpha
		}
		if opt.Aspect != "" {
			cfg.Aspect = opt.Aspect
		}
		if opt.IntegerTicks != nil {
			cfg.IntegerTicks = opt.IntegerTicks
		}
		if opt.Label != "" {
			cfg.Label = opt.Label
		}
	}

	xMin := -0.5
	xMax := float64(cols) - 0.5
	yMin := -0.5
	yMax := float64(rows) - 0.5
	img := a.Image(data, ImageOptions{
		Colormap: cfg.Colormap,
		VMin:     cfg.VMin,
		VMax:     cfg.VMax,
		Alpha:    cfg.Alpha,
		XMin:     &xMin,
		XMax:     &xMax,
		YMin:     &yMin,
		YMax:     &yMax,
		Origin:   ImageOriginUpper,
		Label:    cfg.Label,
	})
	if img == nil {
		return nil
	}

	if cfg.Aspect != "" {
		_ = a.SetAspect(cfg.Aspect)
	}
	a.SetXLim(xMin, xMax)
	a.SetYLim(yMin, yMax)
	if !a.YInverted() {
		a.InvertY()
	}
	if boolValue(cfg.IntegerTicks, true) {
		applyMatrixTicks(a, rows, cols)
	}
	return img
}

// Spy visualizes the sparsity pattern of a matrix.
func (a *Axes) Spy(data [][]float64, opts ...SpyOptions) *SpyResult {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	cfg := SpyOptions{
		MarkerSize: 7,
		Aspect:     "equal",
	}
	if len(opts) > 0 {
		opt := opts[0]
		cfg.Precision = opt.Precision
		cfg.UseImage = opt.UseImage
		cfg.Marker = opt.Marker
		if opt.MarkerSize > 0 {
			cfg.MarkerSize = opt.MarkerSize
		}
		if opt.Color != nil {
			cfg.Color = opt.Color
		}
		if opt.Alpha != nil {
			cfg.Alpha = opt.Alpha
		}
		if opt.Aspect != "" {
			cfg.Aspect = opt.Aspect
		}
		if opt.Label != "" {
			cfg.Label = opt.Label
		}
	}
	if cfg.MarkerSize <= 0 {
		cfg.MarkerSize = 7
	}
	marker := markerValue(cfg.Marker, MarkerSquare)

	indices := make([]geom.Pt, 0)
	mask := make([][]float64, rows)
	for row := 0; row < rows; row++ {
		mask[row] = make([]float64, cols)
		for col := 0; col < cols; col++ {
			value := data[row][col]
			if !isFinite(value) || math.Abs(value) <= cfg.Precision {
				continue
			}
			mask[row][col] = 1
			indices = append(indices, geom.Pt{X: float64(col), Y: float64(row)})
		}
	}

	result := &SpyResult{Indices: indices}
	if boolValue(cfg.UseImage, false) {
		cmap := "gray"
		vMin := 0.0
		vMax := 1.0
		result.Image = a.MatShow(mask, MatShowOptions{
			Colormap:     &cmap,
			VMin:         &vMin,
			VMax:         &vMax,
			Alpha:        cfg.Alpha,
			Aspect:       cfg.Aspect,
			IntegerTicks: boolPtr(true),
			Label:        cfg.Label,
		})
		return result
	}

	color := render.Color{A: 1}
	if cfg.Color != nil {
		color = *cfg.Color
	}
	alpha := 1.0
	if cfg.Alpha != nil {
		alpha = clampOneToOne(*cfg.Alpha)
	}
	path := (&Scatter2D{Marker: marker}).markerPrototypePath()
	lineOnly := marker == MarkerPlus || marker == MarkerCross
	pc := &PathCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  cfg.Label,
			Alpha:  alpha,
		},
		Path:          path,
		Offsets:       append([]geom.Pt(nil), indices...),
		Size:          cfg.MarkerSize * stemMarkerScale,
		PathInDisplay: true,
		FaceColor:     color,
		EdgeColor:     color,
		EdgeWidth:     1,
		LineOnly:      lineOnly,
	}
	a.AddCollection(pc)
	result.Markers = pc

	xMin := -0.5
	xMax := float64(cols) - 0.5
	yMin := -0.5
	yMax := float64(rows) - 0.5
	if cfg.Aspect != "" {
		_ = a.SetAspect(cfg.Aspect)
	}
	a.SetXLim(xMin, xMax)
	a.SetYLim(yMin, yMax)
	if !a.YInverted() {
		a.InvertY()
	}
	applyMatrixTicks(a, rows, cols)
	return result
}

// AnnotatedHeatmap renders a matrix display plus a centered value label in each cell.
func (a *Axes) AnnotatedHeatmap(data [][]float64, opts ...AnnotatedHeatmapOptions) *AnnotatedHeatmapResult {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	cfg := AnnotatedHeatmapOptions{
		MatShowOptions: MatShowOptions{
			Aspect:       "equal",
			IntegerTicks: boolPtr(true),
		},
		Format:        "%.3g",
		FontSize:      11,
		TextColor:     render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1},
		TextColorHigh: render.Color{R: 1, G: 1, B: 1, A: 1},
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Aspect == "" {
			cfg.Aspect = "equal"
		}
		if cfg.Format == "" {
			cfg.Format = "%.3g"
		}
		if cfg.FontSize <= 0 {
			cfg.FontSize = 11
		}
		if cfg.TextColor.A == 0 {
			cfg.TextColor = render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}
		}
		if cfg.TextColorHigh.A == 0 {
			cfg.TextColorHigh = render.Color{R: 1, G: 1, B: 1, A: 1}
		}
	}

	img := a.MatShow(data, cfg.MatShowOptions)
	if img == nil {
		return nil
	}

	mapping := img.ScalarMap().Resolved()
	threshold := mapping.VMin + 0.5*(mapping.VMax-mapping.VMin)
	if cfg.Threshold != nil {
		threshold = *cfg.Threshold
	}

	labels := make([]*Text, 0, rows*cols)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			value := data[row][col]
			if !isFinite(value) && cfg.SkipNaN {
				continue
			}
			text := formatAnnotatedHeatmapValue(value, cfg.Format, cfg.NaNText)
			color := cfg.TextColor
			if isFinite(value) && value >= threshold {
				color = cfg.TextColorHigh
			}
			label := a.Text(float64(col), float64(row), text, TextOptions{
				FontSize: cfg.FontSize,
				Color:    color,
				HAlign:   TextAlignCenter,
				VAlign:   TextVAlignMiddle,
			})
			labels = append(labels, label)
		}
	}

	return &AnnotatedHeatmapResult{
		Image:  img,
		Labels: labels,
	}
}

func applyMatrixTicks(a *Axes, rows, cols int) {
	if a == nil {
		return
	}
	xLocator := integerMatrixLocator(cols)
	yLocator := integerMatrixLocator(rows)
	for _, axis := range []*Axis{a.XAxis, a.XAxisTop} {
		if axis == nil {
			continue
		}
		axis.Locator = xLocator
		axis.Formatter = ScalarFormatter{Prec: 0}
	}
	for _, axis := range []*Axis{a.YAxis, a.YAxisRight} {
		if axis == nil {
			continue
		}
		axis.Locator = yLocator
		axis.Formatter = ScalarFormatter{Prec: 0}
	}
}

func integerMatrixLocator(count int) Locator {
	if count <= 0 {
		return NullLocator{}
	}
	if count <= 16 {
		ticks := make([]float64, count)
		for i := range ticks {
			ticks[i] = float64(i)
		}
		return FixedLocator{TicksList: ticks}
	}
	return MaxNLocator{N: 8, Integer: true}
}

func formatAnnotatedHeatmapValue(value float64, pattern, nanText string) string {
	if !isFinite(value) {
		if nanText != "" {
			return nanText
		}
		return "NaN"
	}
	if pattern == "" {
		return ScalarFormatter{Prec: 3}.Format(value)
	}
	return FormatStrFormatter{Pattern: pattern}.Format(value)
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func boolPtr(value bool) *bool {
	return &value
}

func markerValue(value *MarkerType, fallback MarkerType) MarkerType {
	if value == nil {
		return fallback
	}
	return *value
}
