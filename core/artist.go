package core

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"matplotlib-go/color"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
	"matplotlib-go/transform"
)

// Artist is anything that can draw itself with a z-order and optional bounds.
type Artist interface {
	Draw(r render.Renderer, ctx *DrawContext)
	Z() float64
	Bounds(ctx *DrawContext) geom.Rect
}

// OverlayArtist is an optional extension for artists that need to render
// outside the axes clip, such as annotations with labels offset from the plot.
type OverlayArtist interface {
	DrawOverlay(r render.Renderer, ctx *DrawContext)
}

// ArtistFunc adapts a function to an Artist.
type ArtistFunc func(r render.Renderer, ctx *DrawContext)

func (f ArtistFunc) Draw(r render.Renderer, ctx *DrawContext) { f(r, ctx) }
func (f ArtistFunc) Z() float64                               { return 0 }
func (f ArtistFunc) Bounds(_ *DrawContext) geom.Rect          { return geom.Rect{} }

// DrawContext carries per-draw state like transforms and style.
type DrawContext struct {
	// DataToPixel maps data coordinates to pixels.
	DataToPixel Transform2D
	// Styling configuration in effect.
	RC style.RC
	// Clip is the axes pixel rectangle.
	Clip geom.Rect
	// FigureRect is the figure display rectangle in pixels.
	FigureRect geom.Rect
}

// Transform2D wires x/y scales with an axes->pixel affine transform.
type Transform2D struct {
	XScale      transform.Scale
	YScale      transform.Scale
	DataToAxes  transform.T
	AxesToPixel transform.T
}

// Apply transforms a data-space point to pixel coordinates.
func (t *Transform2D) Apply(p geom.Pt) geom.Pt {
	tr := t.transData()
	if tr == nil {
		return p
	}
	return tr.Apply(p)
}

// Invert transforms a pixel-space point back into data coordinates.
func (t *Transform2D) Invert(p geom.Pt) (geom.Pt, bool) {
	tr := t.transData()
	if tr == nil {
		return p, true
	}
	return tr.Invert(p)
}

func (t *Transform2D) transData() transform.T {
	if t == nil {
		return nil
	}

	dataToAxes := t.DataToAxes
	if dataToAxes == nil {
		dataToAxes = transform.NewScaleTransform(t.XScale, t.YScale)
	}

	switch {
	case dataToAxes == nil:
		return t.AxesToPixel
	case t.AxesToPixel == nil:
		return dataToAxes
	default:
		return transform.Chain{A: dataToAxes, B: t.AxesToPixel}
	}
}

// CoordinateSpace identifies a Matplotlib-style coordinate system.
type CoordinateSpace uint8

const (
	CoordData CoordinateSpace = iota
	CoordAxes
	CoordFigure
)

// CoordinateSpec identifies the x/y coordinate spaces used by a point.
type CoordinateSpec struct {
	X CoordinateSpace
	Y CoordinateSpace
}

// Coords uses the same coordinate space for x and y.
func Coords(space CoordinateSpace) CoordinateSpec {
	return CoordinateSpec{X: space, Y: space}
}

// BlendCoords uses separate coordinate spaces for x and y.
func BlendCoords(xSpace, ySpace CoordinateSpace) CoordinateSpec {
	return CoordinateSpec{X: xSpace, Y: ySpace}
}

// TransData returns the Matplotlib-style data->display transform.
func (ctx *DrawContext) TransData() transform.T {
	if ctx == nil {
		return nil
	}
	return ctx.DataToPixel.transData()
}

// TransAxes returns the Matplotlib-style axes-fraction->display transform.
func (ctx *DrawContext) TransAxes() transform.T {
	if ctx == nil {
		return nil
	}
	if ctx.DataToPixel.AxesToPixel != nil {
		return ctx.DataToPixel.AxesToPixel
	}
	if rect, ok := unitSquareBounds(nil, ctx.Clip); ok {
		return transform.NewDisplayRectTransform(rect)
	}
	return nil
}

// TransFigure returns the Matplotlib-style figure-fraction->display transform.
func (ctx *DrawContext) TransFigure() transform.T {
	if ctx == nil {
		return nil
	}
	rect := ctx.FigureRect
	if rect == (geom.Rect{}) {
		rect = ctx.Clip
	}
	if rect == (geom.Rect{}) {
		return nil
	}
	return transform.NewDisplayRectTransform(rect)
}

// TransformFor resolves a coordinate specification into a display transform.
func (ctx *DrawContext) TransformFor(spec CoordinateSpec) transform.T {
	if spec.X == spec.Y {
		return ctx.transformForSpace(spec.X)
	}

	xTrans, okX := ctx.separableTransformForSpace(spec.X)
	yTrans, okY := ctx.separableTransformForSpace(spec.Y)
	if !okX || !okY {
		return nil
	}
	return transform.Blend(xTrans, yTrans)
}

func (ctx *DrawContext) transformForSpace(space CoordinateSpace) transform.T {
	switch space {
	case CoordAxes:
		return ctx.TransAxes()
	case CoordFigure:
		return ctx.TransFigure()
	default:
		return ctx.TransData()
	}
}

func (ctx *DrawContext) separableTransformForSpace(space CoordinateSpace) (transform.Separable, bool) {
	switch space {
	case CoordAxes:
		return ctx.separableAxesTransform()
	case CoordFigure:
		tr := ctx.TransFigure()
		sep, ok := tr.(transform.Separable)
		return sep, ok
	default:
		return ctx.separableDataTransform()
	}
}

func (ctx *DrawContext) separableAxesTransform() (transform.Separable, bool) {
	if ctx == nil {
		return transform.SeparableT{}, false
	}
	if sep, ok := ctx.DataToPixel.AxesToPixel.(transform.Separable); ok {
		return sep, true
	}
	rect, ok := unitSquareBounds(ctx.DataToPixel.AxesToPixel, ctx.Clip)
	if !ok {
		return transform.SeparableT{}, false
	}
	return transform.NewDisplayRectTransform(rect), true
}

func (ctx *DrawContext) separableDataTransform() (transform.Separable, bool) {
	axesTrans, ok := ctx.separableAxesTransform()
	if !ok {
		return transform.SeparableT{}, false
	}

	if ctx != nil && ctx.DataToPixel.DataToAxes != nil {
		sep, ok := ctx.DataToPixel.DataToAxes.(transform.Separable)
		if !ok {
			return transform.SeparableT{}, false
		}
		return transform.ChainSeparable(sep, axesTrans), true
	}

	return transform.ChainSeparable(
		transform.NewScaleTransform(ctx.DataToPixel.XScale, ctx.DataToPixel.YScale),
		axesTrans,
	), true
}

func unitSquareBounds(tr transform.T, fallback geom.Rect) (geom.Rect, bool) {
	if tr == nil {
		if fallback == (geom.Rect{}) {
			return geom.Rect{}, false
		}
		return fallback, true
	}

	corners := []geom.Pt{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 0, Y: 1},
		{X: 1, Y: 1},
	}

	rect := geom.Rect{Min: tr.Apply(corners[0]), Max: tr.Apply(corners[0])}
	for _, corner := range corners[1:] {
		pt := tr.Apply(corner)
		if pt.X < rect.Min.X {
			rect.Min.X = pt.X
		}
		if pt.Y < rect.Min.Y {
			rect.Min.Y = pt.Y
		}
		if pt.X > rect.Max.X {
			rect.Max.X = pt.X
		}
		if pt.Y > rect.Max.Y {
			rect.Max.Y = pt.Y
		}
	}
	return rect, true
}

// Figure is the root of the Artist tree. It contains Axes children.
type Figure struct {
	SizePx   geom.Pt
	RC       style.RC
	Children []*Axes
}

// NewFigure creates a new figure with pixel dimensions and optional style overrides.
func NewFigure(w, h int, opts ...style.Option) *Figure {
	rc := style.Apply(style.Default, opts...)
	return &Figure{
		SizePx:   geom.Pt{X: float64(w), Y: float64(h)},
		RC:       rc,
		Children: nil,
	}
}

// Axes represents an axes region inside a figure.
type Axes struct {
	RectFraction geom.Rect // [0..1] fraction in figure coords
	RC           *style.RC // nil => inherit figure RC
	XScale       transform.Scale
	YScale       transform.Scale
	Artists      []Artist
	zsorted      bool

	// Axis control
	XAxis      *Axis // bottom x-axis
	YAxis      *Axis // left y-axis
	XAxisTop   *Axis // optional top x-axis
	YAxisRight *Axis // optional right y-axis
	ShowFrame  bool  // draw top and right border lines when no explicit top/right axis exists

	// Text labels
	Title  string // title above the plot
	XLabel string // x-axis label below ticks
	YLabel string // y-axis label left of ticks

	// Color cycling for multiple series
	ColorCycle *color.ColorCycle

	aspectMode  string
	aspectValue float64
	boxAspect   float64

	shareX *Axes
	shareY *Axes
	figure *Figure
}

// TickParams controls axis tick visibility and styling.
type TickParams struct {
	Axis          string
	Which         string
	Color         *render.Color
	Length        *float64
	Width         *float64
	ShowTicks     *bool
	ShowLabels    *bool
	LabelRotation *float64
	LabelPad      *float64
	LabelHAlign   *TextAlign
	LabelVAlign   *TextVerticalAlign
}

// LocatorParams controls the target tick density for automatic locators.
type LocatorParams struct {
	Axis       string
	MajorCount int
	MinorCount int
}

// AddAxes appends an Axes to the Figure. If opts are provided, the Axes gets its
// own RC copy; otherwise it inherits from the Figure.
func (f *Figure) AddAxes(r geom.Rect, opts ...style.Option) *Axes {
	var rc *style.RC
	effective := f.RC
	if len(opts) > 0 {
		v := style.Apply(f.RC, opts...)
		rc = &v
		effective = v
	}
	ax := &Axes{
		RectFraction: r,
		RC:           rc,
		XScale:       transform.NewLinear(0, 1),
		YScale:       transform.NewLinear(0, 1),
		XAxis:        NewXAxis(),
		YAxis:        NewYAxis(),
		ShowFrame:    true,
		ColorCycle:   color.NewColorCycle(effective.Palette()),
		aspectMode:   "auto",
		aspectValue:  1,
		figure:       f,
	}
	ax.applyStyleDefaults(effective)
	f.Children = append(f.Children, ax)
	return ax
}

// Add registers an Artist with the Axes.
func (a *Axes) Add(art Artist) { a.Artists = append(a.Artists, art); a.zsorted = false }

// SetXLim sets the x-axis limits.
func (a *Axes) SetXLim(minVal, maxVal float64) {
	target := a.xScaleRoot()
	target.XScale = replaceScaleDomain(target.XScale, minVal, maxVal)
}

// SetYLim sets the y-axis limits.
func (a *Axes) SetYLim(minVal, maxVal float64) {
	target := a.yScaleRoot()
	target.YScale = replaceScaleDomain(target.YScale, minVal, maxVal)
}

// SetXScale replaces the x-axis scale while preserving the current view limits.
func (a *Axes) SetXScale(name string, opts ...transform.ScaleOption) error {
	return a.setScale(true, name, opts...)
}

// SetYScale replaces the y-axis scale while preserving the current view limits.
func (a *Axes) SetYScale(name string, opts ...transform.ScaleOption) error {
	return a.setScale(false, name, opts...)
}

// SetXLimLog sets the x-axis to logarithmic scale with given limits.
func (a *Axes) SetXLimLog(minVal, maxVal, base float64) {
	target := a.xScaleRoot()
	target.XScale = transform.NewLog(minVal, maxVal, base)
	configureScaleAxes(target.XAxis, target.XAxisTop, "log", transform.ResolveScaleOptions(
		transform.WithScaleDomain(minVal, maxVal),
		transform.WithScaleBase(base),
	))
}

// SetYLimLog sets the y-axis to logarithmic scale with given limits.
func (a *Axes) SetYLimLog(minVal, maxVal, base float64) {
	target := a.yScaleRoot()
	target.YScale = transform.NewLog(minVal, maxVal, base)
	configureScaleAxes(target.YAxis, target.YAxisRight, "log", transform.ResolveScaleOptions(
		transform.WithScaleDomain(minVal, maxVal),
		transform.WithScaleBase(base),
	))
}

// InvertX reverses the x-axis direction while preserving the underlying scale type.
func (a *Axes) InvertX() {
	target := a.xScaleRoot()
	if target == nil || target.XScale == nil {
		return
	}
	target.XScale = toggleInvertedScale(target.XScale)
}

// InvertY reverses the y-axis direction while preserving the underlying scale type.
func (a *Axes) InvertY() {
	target := a.yScaleRoot()
	if target == nil || target.YScale == nil {
		return
	}
	target.YScale = toggleInvertedScale(target.YScale)
}

// XInverted reports whether the effective x-axis direction is reversed.
func (a *Axes) XInverted() bool {
	if a == nil {
		return false
	}
	return scaleDomainDescending(a.effectiveXScale())
}

// YInverted reports whether the effective y-axis direction is reversed.
func (a *Axes) YInverted() bool {
	if a == nil {
		return false
	}
	return scaleDomainDescending(a.effectiveYScale())
}

// AutoScale computes axis limits from the data bounds of all artists,
// adding the specified margin fraction on each side (e.g. 0.05 = 5%).
// A margin of 0 fits exactly to the data. If no artists have non-zero bounds,
// limits remain unchanged.
func (a *Axes) AutoScale(margin float64) {
	targetX := a.xScaleRoot()
	targetY := a.yScaleRoot()

	var xMin, xMax, yMin, yMax float64
	first := true

	for _, art := range a.Artists {
		b := art.Bounds(nil)
		if b.W() == 0 && b.H() == 0 && b.Min.X == 0 && b.Min.Y == 0 {
			continue // skip zero-bounds artists (grids, etc.)
		}
		if first {
			xMin, xMax = b.Min.X, b.Max.X
			yMin, yMax = b.Min.Y, b.Max.Y
			first = false
		} else {
			if b.Min.X < xMin {
				xMin = b.Min.X
			}
			if b.Max.X > xMax {
				xMax = b.Max.X
			}
			if b.Min.Y < yMin {
				yMin = b.Min.Y
			}
			if b.Max.Y > yMax {
				yMax = b.Max.Y
			}
		}
	}
	if first {
		return // no data artists
	}

	// Apply margin
	xSpan := xMax - xMin
	ySpan := yMax - yMin
	if xSpan == 0 {
		xSpan = 1 // avoid zero-span
	}
	if ySpan == 0 {
		ySpan = 1
	}
	xMin -= xSpan * margin
	xMax += xSpan * margin
	yMin -= ySpan * margin
	yMax += ySpan * margin

	targetX.XScale = replaceScaleDomain(targetX.XScale, xMin, xMax)
	targetY.YScale = replaceScaleDomain(targetY.YScale, yMin, yMax)
}

// AddGrid adds grid lines for the specified axis.
func (a *Axes) AddGrid(axis AxisSide) *Grid {
	grid := NewGrid(axis)
	rc := a.resolvedRC()
	grid.Color = rc.GridColor
	grid.LineWidth = rc.GridLineWidth
	grid.MinorColor = rc.MinorGridColor
	grid.MinorLineWidth = rc.MinorGridLineWidth
	a.Add(grid)
	return grid
}

// AddXGrid adds vertical grid lines based on x-axis ticks.
func (a *Axes) AddXGrid() *Grid {
	return a.AddGrid(AxisBottom)
}

// AddYGrid adds horizontal grid lines based on y-axis ticks.
func (a *Axes) AddYGrid() *Grid {
	return a.AddGrid(AxisLeft)
}

// SetTitle sets the title displayed above the plot.
func (a *Axes) SetTitle(title string) { a.Title = title }

// SetXLabel sets the label displayed below the x-axis.
func (a *Axes) SetXLabel(label string) { a.XLabel = label }

// SetYLabel sets the label displayed left of the y-axis.
func (a *Axes) SetYLabel(label string) { a.YLabel = label }

// SetAspect configures the data aspect ratio for the axes box.
// Supported values are "auto", "equal", and "ratio" (which requires one numeric value).
func (a *Axes) SetAspect(mode string, value ...float64) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "auto":
		a.aspectMode = "auto"
		a.aspectValue = 1
	case "equal":
		a.aspectMode = "equal"
		a.aspectValue = 1
	case "ratio":
		if len(value) == 0 || value[0] <= 0 || math.IsNaN(value[0]) || math.IsInf(value[0], 0) {
			return fmt.Errorf("ratio aspect requires a positive finite value")
		}
		a.aspectMode = "ratio"
		a.aspectValue = value[0]
	default:
		return fmt.Errorf("unsupported aspect mode %q", mode)
	}
	return nil
}

// SetAxisEqual is a convenience helper for a 1:1 data aspect ratio.
func (a *Axes) SetAxisEqual() {
	_ = a.SetAspect("equal")
}

// SetBoxAspect constrains the physical height/width ratio of the axes box.
func (a *Axes) SetBoxAspect(aspect float64) error {
	if a == nil {
		return nil
	}
	if aspect <= 0 || math.IsNaN(aspect) || math.IsInf(aspect, 0) {
		return fmt.Errorf("box aspect must be a positive finite value")
	}
	a.boxAspect = aspect
	return nil
}

// ClearBoxAspect removes any physical box-aspect constraint.
func (a *Axes) ClearBoxAspect() {
	if a == nil {
		return
	}
	a.boxAspect = 0
}

// NextColor returns the next color in the axes color cycle.
func (a *Axes) NextColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.ColorCycle.Next()
}

// PeekColor returns the current color without advancing the cycle.
func (a *Axes) PeekColor() render.Color {
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(a.resolvedRC().Palette())
	}
	return a.ColorCycle.Peek()
}

// ResetColorCycle resets the color cycle to the first color.
func (a *Axes) ResetColorCycle() {
	if a.ColorCycle != nil {
		a.ColorCycle.Reset()
	}
}

// TopAxis returns the explicit top x-axis, creating it on first use.
func (a *Axes) TopAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.XAxisTop == nil {
		a.XAxisTop = cloneAxisForSide(a.XAxis, AxisTop)
	}
	return a.XAxisTop
}

// RightAxis returns the explicit right y-axis, creating it on first use.
func (a *Axes) RightAxis() *Axis {
	if a == nil {
		return nil
	}
	if a.YAxisRight == nil {
		a.YAxisRight = cloneAxisForSide(a.YAxis, AxisRight)
	}
	return a.YAxisRight
}

// HideTopAxis removes the explicit top x-axis.
func (a *Axes) HideTopAxis() {
	if a == nil {
		return
	}
	a.XAxisTop = nil
}

// HideRightAxis removes the explicit right y-axis.
func (a *Axes) HideRightAxis() {
	if a == nil {
		return
	}
	a.YAxisRight = nil
}

// SetXAxisSide repositions the primary x-axis to the requested side.
func (a *Axes) SetXAxisSide(side AxisSide) error {
	if a == nil {
		return nil
	}
	if side != AxisBottom && side != AxisTop {
		return fmt.Errorf("x-axis side must be bottom or top")
	}
	if a.XAxis == nil {
		a.XAxis = cloneAxisForSide(nil, side)
	} else {
		a.XAxis.Side = side
	}
	if side == AxisTop {
		a.XAxisTop = nil
	}
	return nil
}

// SetYAxisSide repositions the primary y-axis to the requested side.
func (a *Axes) SetYAxisSide(side AxisSide) error {
	if a == nil {
		return nil
	}
	if side != AxisLeft && side != AxisRight {
		return fmt.Errorf("y-axis side must be left or right")
	}
	if a.YAxis == nil {
		a.YAxis = cloneAxisForSide(nil, side)
	} else {
		a.YAxis.Side = side
	}
	if side == AxisRight {
		a.YAxisRight = nil
	}
	return nil
}

// MoveXAxisToTop is a convenience helper that repositions the primary x-axis.
func (a *Axes) MoveXAxisToTop() error { return a.SetXAxisSide(AxisTop) }

// MoveXAxisToBottom is a convenience helper that repositions the primary x-axis.
func (a *Axes) MoveXAxisToBottom() error { return a.SetXAxisSide(AxisBottom) }

// MoveYAxisToLeft is a convenience helper that repositions the primary y-axis.
func (a *Axes) MoveYAxisToLeft() error { return a.SetYAxisSide(AxisLeft) }

// MoveYAxisToRight is a convenience helper that repositions the primary y-axis.
func (a *Axes) MoveYAxisToRight() error { return a.SetYAxisSide(AxisRight) }

// SetXTickLabelPosition controls whether x tick labels appear on the bottom,
// top, both, or neither side.
func (a *Axes) SetXTickLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "bottom":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = true
		}
		if a.XAxisTop != nil {
			a.XAxisTop.ShowLabels = false
		}
	case "top":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = false
		}
		a.TopAxis().ShowLabels = true
	case "both":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = true
		}
		a.TopAxis().ShowLabels = true
	case "none":
		if a.XAxis != nil {
			a.XAxis.ShowLabels = false
		}
		if a.XAxisTop != nil {
			a.XAxisTop.ShowLabels = false
		}
	default:
		return fmt.Errorf("unsupported x tick label position %q", position)
	}
	return nil
}

// SetYTickLabelPosition controls whether y tick labels appear on the left,
// right, both, or neither side.
func (a *Axes) SetYTickLabelPosition(position string) error {
	if a == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "", "left":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = true
		}
		if a.YAxisRight != nil {
			a.YAxisRight.ShowLabels = false
		}
	case "right":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = false
		}
		a.RightAxis().ShowLabels = true
	case "both":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = true
		}
		a.RightAxis().ShowLabels = true
	case "none":
		if a.YAxis != nil {
			a.YAxis.ShowLabels = false
		}
		if a.YAxisRight != nil {
			a.YAxisRight.ShowLabels = false
		}
	default:
		return fmt.Errorf("unsupported y tick label position %q", position)
	}
	return nil
}

// MinorticksOn enables default minor locators on the requested axis selection.
func (a *Axes) MinorticksOn(axis string) error {
	if err := validateAxisSpec(axis); err != nil {
		return err
	}
	for _, ax := range a.axesForSpec(axis) {
		enableMinorTicks(ax)
	}
	return nil
}

// MinorticksOff disables minor locators on the requested axis selection.
func (a *Axes) MinorticksOff(axis string) error {
	if err := validateAxisSpec(axis); err != nil {
		return err
	}
	for _, ax := range a.axesForSpec(axis) {
		if ax != nil {
			ax.MinorLocator = nil
		}
	}
	return nil
}

// LocatorParams updates the target major/minor tick density for the selected axes.
func (a *Axes) LocatorParams(params LocatorParams) error {
	if err := validateAxisSpec(params.Axis); err != nil {
		return err
	}
	for _, axis := range a.axesForSpec(params.Axis) {
		if axis == nil {
			continue
		}
		if params.MajorCount > 0 {
			axis.MajorTickCount = params.MajorCount
		}
		if params.MinorCount > 0 {
			axis.MinorTickCount = params.MinorCount
		}
	}
	return nil
}

// TickParams applies visibility/styling updates to the selected axis ticks.
func (a *Axes) TickParams(params TickParams) error {
	if err := validateAxisSpec(params.Axis); err != nil {
		return err
	}
	which := normalizeTickWhich(params.Which)
	if which == "" {
		return fmt.Errorf("unsupported tick selection %q", params.Which)
	}

	for _, axis := range a.axesForSpec(params.Axis) {
		if axis == nil {
			continue
		}
		if params.Color != nil {
			axis.Color = *params.Color
		}
		if params.Width != nil {
			axis.LineWidth = *params.Width
		}
		if params.ShowTicks != nil {
			axis.ShowTicks = *params.ShowTicks
		}
		if params.ShowLabels != nil {
			switch which {
			case "major":
				axis.ShowLabels = *params.ShowLabels
			case "minor":
				axis.ShowMinorLabels = *params.ShowLabels
			case "both":
				axis.ShowLabels = *params.ShowLabels
				axis.ShowMinorLabels = *params.ShowLabels
			}
		}
		if params.Length != nil {
			switch which {
			case "major":
				axis.TickSize = *params.Length
			case "minor":
				axis.MinorTickSize = *params.Length
			case "both":
				axis.TickSize = *params.Length
				axis.MinorTickSize = *params.Length
			}
		}
		switch which {
		case "major":
			applyTickLabelParams(&axis.MajorLabelStyle, params)
		case "minor":
			applyTickLabelParams(&axis.MinorLabelStyle, params)
		case "both":
			applyTickLabelParams(&axis.MajorLabelStyle, params)
			applyTickLabelParams(&axis.MinorLabelStyle, params)
		}
	}
	return nil
}

// TwinX creates an overlay axes sharing the x-scale with an independent y-scale on the right.
func (a *Axes) TwinX() *Axes {
	twin := a.newOverlayAxes()
	if twin == nil {
		return nil
	}
	twin.shareX = a.xScaleRoot()
	if twin.XAxis != nil {
		twin.XAxis.ShowSpine = false
		twin.XAxis.ShowTicks = false
		twin.XAxis.ShowLabels = false
	}
	if twin.YAxis != nil {
		twin.YAxis.ShowSpine = false
		twin.YAxis.ShowTicks = false
		twin.YAxis.ShowLabels = false
	}
	twin.ShowFrame = false
	twin.RightAxis()
	return twin
}

// TwinY creates an overlay axes sharing the y-scale with an independent x-scale on the top.
func (a *Axes) TwinY() *Axes {
	twin := a.newOverlayAxes()
	if twin == nil {
		return nil
	}
	twin.shareY = a.yScaleRoot()
	if twin.YAxis != nil {
		twin.YAxis.ShowSpine = false
		twin.YAxis.ShowTicks = false
		twin.YAxis.ShowLabels = false
	}
	if twin.XAxis != nil {
		twin.XAxis.ShowSpine = false
		twin.XAxis.ShowTicks = false
		twin.XAxis.ShowLabels = false
	}
	twin.ShowFrame = false
	twin.TopAxis()
	return twin
}

// SecondaryXAxis creates an overlay axes that displays transformed x ticks on the requested side.
func (a *Axes) SecondaryXAxis(side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if side != AxisTop && side != AxisBottom {
		return nil, fmt.Errorf("secondary x-axis side must be top or bottom")
	}
	return a.newSecondaryAxes(true, side, forward, inverse)
}

// SecondaryYAxis creates an overlay axes that displays transformed y ticks on the requested side.
func (a *Axes) SecondaryYAxis(side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if side != AxisLeft && side != AxisRight {
		return nil, fmt.Errorf("secondary y-axis side must be left or right")
	}
	return a.newSecondaryAxes(false, side, forward, inverse)
}

// layout computes the pixel rectangle for this Axes inside the Figure.
func (a *Axes) layout(f *Figure) (pixelRect geom.Rect) {
	// Map fraction [0..1] to pixel coordinates
	minPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Min.X, Y: f.SizePx.Y * a.RectFraction.Min.Y}
	maxPt := geom.Pt{X: f.SizePx.X * a.RectFraction.Max.X, Y: f.SizePx.Y * a.RectFraction.Max.Y}
	return geom.Rect{Min: minPt, Max: maxPt}
}

func (a *Axes) xScaleRoot() *Axes {
	if a == nil {
		return nil
	}
	cur := a
	for cur.shareX != nil {
		cur = cur.shareX
		if cur == nil {
			return a
		}
	}
	return cur
}

func (a *Axes) yScaleRoot() *Axes {
	if a == nil {
		return nil
	}
	cur := a
	for cur.shareY != nil {
		cur = cur.shareY
		if cur == nil {
			return a
		}
	}
	return cur
}

func (a *Axes) effectiveXAxis() *Axis {
	return a.XAxis
}

func (a *Axes) effectiveYAxis() *Axis {
	return a.YAxis
}

func (a *Axes) effectiveTopAxis() *Axis {
	if a == nil {
		return nil
	}
	return a.XAxisTop
}

func (a *Axes) effectiveRightAxis() *Axis {
	if a == nil {
		return nil
	}
	return a.YAxisRight
}

func (a *Axes) effectiveXScale() transform.Scale {
	if a.shareX != nil {
		return a.shareX.effectiveXScale()
	}
	return a.XScale
}

func (a *Axes) effectiveYScale() transform.Scale {
	if a.shareY != nil {
		return a.shareY.effectiveYScale()
	}
	return a.YScale
}

func (a *Axes) setScale(isX bool, name string, opts ...transform.ScaleOption) error {
	var target *Axes
	var current transform.Scale
	var primary, secondary *Axis

	if isX {
		target = a.xScaleRoot()
		current = target.XScale
		primary = target.XAxis
		secondary = target.XAxisTop
	} else {
		target = a.yScaleRoot()
		current = target.YScale
		primary = target.YAxis
		secondary = target.YAxisRight
	}

	minVal, maxVal := currentScaleDomain(current)
	cfg := transform.ResolveScaleOptions(append([]transform.ScaleOption{
		transform.WithScaleDomain(minVal, maxVal),
	}, opts...)...)
	scale, err := transform.NewScaleWithOptions(name, cfg)
	if err != nil {
		return err
	}

	if isX {
		target.XScale = scale
	} else {
		target.YScale = scale
	}
	configureScaleAxes(primary, secondary, name, cfg)
	return nil
}

// effectiveRC resolves the RC for this axes, inheriting from the Figure if needed.
func (a *Axes) effectiveRC(f *Figure) style.RC {
	if a.RC != nil {
		return *a.RC
	}
	if a.figure != nil {
		return a.figure.RC
	}
	if f == nil {
		return style.Default
	}
	return f.RC
}

func (a *Axes) resolvedRC() style.RC {
	if a == nil {
		return style.Default
	}
	return a.effectiveRC(a.figure)
}

func currentScaleDomain(s transform.Scale) (float64, float64) {
	if s == nil {
		return 0, 1
	}
	return s.Domain()
}

func replaceScaleDomain(s transform.Scale, minVal, maxVal float64) transform.Scale {
	switch v := s.(type) {
	case nil:
		return transform.NewLinear(minVal, maxVal)
	case transform.DomainSetter:
		return v.WithDomain(minVal, maxVal)
	case invertedScale:
		return replaceScaleDomain(v.base, minVal, maxVal)
	default:
		return transform.NewLinear(minVal, maxVal)
	}
}

func configureScaleAxes(primary, secondary *Axis, scaleName string, cfg transform.ScaleOptions) {
	configureScaleAxis(primary, scaleName, cfg)
	configureScaleAxis(secondary, scaleName, cfg)
}

func configureScaleAxis(axis *Axis, scaleName string, cfg transform.ScaleOptions) {
	if axis == nil {
		return
	}

	switch strings.ToLower(scaleName) {
	case "log":
		axis.Locator = LogLocator{Base: cfg.Base, Minor: false}
		axis.Formatter = LogFormatter{Base: cfg.Base}
		if len(cfg.Subs) > 0 {
			axis.MinorLocator = LogLocator{Base: cfg.Base, Minor: true, Subs: cfg.Subs}
		} else {
			axis.MinorLocator = nil
		}
	default:
		axis.Locator = LinearLocator{}
		axis.Formatter = ScalarFormatter{Prec: 3}
		axis.MinorLocator = nil
	}
}

func (a *Axes) applyStyleDefaults(rc style.RC) {
	if a == nil {
		return
	}
	if a.XAxis != nil {
		a.XAxis.Color = rc.AxesEdgeColor
		a.XAxis.LineWidth = rc.AxisLineWidth
	}
	if a.YAxis != nil {
		a.YAxis.Color = rc.AxesEdgeColor
		a.YAxis.LineWidth = rc.AxisLineWidth
	}
	if a.XAxisTop != nil {
		a.XAxisTop.Color = rc.AxesEdgeColor
		a.XAxisTop.LineWidth = rc.AxisLineWidth
	}
	if a.YAxisRight != nil {
		a.YAxisRight.Color = rc.AxesEdgeColor
		a.YAxisRight.LineWidth = rc.AxisLineWidth
	}
	if a.ColorCycle == nil {
		a.ColorCycle = color.NewColorCycle(rc.Palette())
	}
}

// DrawFigure performs a traversal and draws the figure into the renderer.
func DrawFigure(fig *Figure, r render.Renderer) {
	vp := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: fig.SizePx.X, Y: fig.SizePx.Y}}
	_ = r.Begin(vp)
	defer r.End()

	for _, ax := range fig.Children {
		px := ax.adjustedLayout(fig)
		xAxis := ax.effectiveXAxis()
		yAxis := ax.effectiveYAxis()
		topAxis := ax.effectiveTopAxis()
		rightAxis := ax.effectiveRightAxis()

		// Build DrawContext with composed transform
		ctx := &DrawContext{
			DataToPixel: Transform2D{
				XScale:      ax.effectiveXScale(),
				YScale:      ax.effectiveYScale(),
				DataToAxes:  transform.NewScaleTransform(ax.effectiveXScale(), ax.effectiveYScale()),
				AxesToPixel: transform.NewDisplayRectTransform(px),
			},
			RC:         ax.effectiveRC(fig),
			Clip:       px,
			FigureRect: vp,
		}

		if ctx.RC.AxesBackground != fig.RC.FigureBackground() {
			r.Path(pixelRectPath(px), &render.Paint{
				Fill: ctx.RC.AxesBackground,
			})
		}

		// Draw only clipped content (data and grids) while the axes clip is active.
		r.Save()
		r.ClipRect(px)

		if !ax.zsorted {
			sort.SliceStable(ax.Artists, func(i, j int) bool {
				zi, zj := ax.Artists[i].Z(), ax.Artists[j].Z()
				if zi == zj {
					return i < j
				}
				return zi < zj
			})
			ax.zsorted = true
		}
		for _, art := range ax.Artists {
			art.Draw(r, ctx)
		}

		r.Restore()

		setRendererResolution(r, ctx.RC.DPI)
		for _, art := range ax.Artists {
			if overlay, ok := art.(OverlayArtist); ok {
				overlay.DrawOverlay(r, ctx)
			}
		}

		// Draw axes elements outside the clip so spines can straddle the axes edge
		// the same way Matplotlib does.
		if xAxis != nil {
			xAxis.Draw(r, ctx)
		}
		if yAxis != nil {
			yAxis.Draw(r, ctx)
		}
		if topAxis != nil {
			topAxis.Draw(r, ctx)
		}
		if rightAxis != nil {
			rightAxis.Draw(r, ctx)
		}
		if ax.ShowFrame {
			ref := xAxis
			if ref == nil {
				ref = topAxis
			}
			if ref == nil {
				ref = yAxis
			}
			if ref == nil {
				ref = rightAxis
			}
			DrawFrame(r, ctx, ref, topAxis == nil, rightAxis == nil)
		}

		// Draw ticks (outward), tick labels, and text labels outside the clip rect.
		if xAxis != nil {
			xAxis.DrawTicks(r, ctx)
			xAxis.DrawTickLabels(r, ctx)
		}
		if yAxis != nil {
			yAxis.DrawTicks(r, ctx)
			yAxis.DrawTickLabels(r, ctx)
		}
		if topAxis != nil {
			topAxis.DrawTicks(r, ctx)
			topAxis.DrawTickLabels(r, ctx)
		}
		if rightAxis != nil {
			rightAxis.DrawTicks(r, ctx)
			rightAxis.DrawTickLabels(r, ctx)
		}
		drawAxesLabels(ax, r, ctx, px)
	}
}

func setRendererResolution(r render.Renderer, dpi float64) {
	if dpi <= 0 {
		return
	}
	if setter, ok := r.(render.DPIAware); ok {
		setter.SetResolution(uint(math.Round(dpi)))
	}
}

// drawAxesLabels renders title, xlabel, and ylabel outside the clipped axes area.
func drawAxesLabels(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	labelColor := ctx.RC.DefaultTextColor()
	titleSize := titleFontSize(ctx)
	labelSize := ctx.RC.FontSize * 0.97
	if labelSize < 8 {
		labelSize = 8
	}

	// Title: centered above the plot
	if ax.Title != "" {
		metrics := r.MeasureText(ax.Title, titleSize, ctx.RC.FontKey)
		// Matplotlib positions axes titles at y=1.0 with a baseline-aligned
		// transform offset by axes.titlepad (default 6 pt).
		titlePadPx := ctx.RC.DPI * 6.0 / 72.0
		anchor := transform.NewOffset(ctx.TransAxes(), geom.Pt{
			X: titleXAdjustPx(),
			Y: -titlePadPx + titleBaselineAdjustPx(ctx),
		}).Apply(geom.Pt{X: 0.5, Y: 1})
		textRen.DrawText(
			ax.Title,
			alignedTextOrigin(anchor, metrics, TextAlignCenter, TextVAlignBaseline),
			titleSize,
			labelColor,
		)
	}

	// XLabel: centered below the x-axis tick labels
	if ax.XLabel != "" {
		metrics := r.MeasureText(ax.XLabel, labelSize, ctx.RC.FontKey)
		base := geom.Pt{X: 0.5, Y: 0}
		offset := geom.Pt{X: 0, Y: 18 + metrics.Ascent}
		if ax.XAxis != nil && ax.XAxis.Side == AxisTop {
			base.Y = 1
			offset.Y = -18 - metrics.Descent
		}
		anchor := transform.NewOffset(ctx.TransAxes(), offset).Apply(base)
		textRen.DrawText(
			ax.XLabel,
			alignedTextOrigin(anchor, metrics, TextAlignCenter, TextVAlignBaseline),
			labelSize,
			labelColor,
		)
	}

	// YLabel: vertical text if supported, else horizontal fallback
	if ax.YLabel != "" {
		anchor := yLabelAnchorPoint(ax, r, ctx, px)
		angle := math.Pi / 2
		if yAxis := ax.effectiveYAxis(); yAxis != nil && yAxis.Side == AxisRight {
			angle = -math.Pi / 2
		}
		switch ren := r.(type) {
		case render.RotatedTextDrawer:
			ren.DrawTextRotated(ax.YLabel, anchor, labelSize, angle, labelColor)
		case render.VerticalTextDrawer:
			if angle < 0 {
				metrics := r.MeasureText(ax.YLabel, labelSize, ctx.RC.FontKey)
				x := anchor.X - metrics.W/2
				y := px.Min.Y + px.H()/2 + metrics.H/2
				textRen.DrawText(ax.YLabel, geom.Pt{X: x, Y: y}, labelSize, labelColor)
			} else {
				ren.DrawTextVertical(ax.YLabel, geom.Pt{X: anchor.X, Y: anchor.Y}, labelSize, labelColor)
			}
		default:
			metrics := r.MeasureText(ax.YLabel, labelSize, ctx.RC.FontKey)
			x := anchor.X - metrics.W/2
			y := px.Min.Y + px.H()/2 + metrics.H/2
			textRen.DrawText(ax.YLabel, geom.Pt{X: x, Y: y}, labelSize, labelColor)
		}
	}
}

var (
	titleScaleFactor       = 1.0002
	titleXAdjustPxValue    = 0.0
	titleBaselineAdjustPxV = 0.0
)

func titleFontSize(ctx *DrawContext) float64 {
	if ctx == nil || ctx.RC.FontSize <= 0 {
		return 12 * titleScaleFactor
	}
	return ctx.RC.FontSize * titleScaleFactor
}

func titleXAdjustPx() float64 {
	return titleXAdjustPxValue
}

func titleBaselineAdjustPx(ctx *DrawContext) float64 {
	_ = ctx
	return titleBaselineAdjustPxV
}

func yLabelAnchorPoint(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect) geom.Pt {
	anchor := ctx.TransAxes().Apply(geom.Pt{X: 0, Y: 0.5})
	anchor.X = spinePixelX(AxisLeft, px) - axisLabelPadPx(ctx)

	yAxis := ax.effectiveYAxis()
	if yAxis == nil {
		return anchor
	}

	spineX := spinePixelX(yAxis.Side, px)
	if yAxis.Side == AxisRight {
		rightExtent := spineX
		if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
			rightExtent = math.Max(rightExtent, tickBounds.Max.X)
		}
		anchor.X = rightExtent + axisLabelPadPx(ctx)
		return anchor
	}

	leftExtent := spineX
	if tickBounds, ok := axisTickLabelBounds(yAxis, r, ctx); ok {
		leftExtent = math.Min(leftExtent, tickBounds.Min.X)
	}
	anchor.X = leftExtent - axisLabelPadPx(ctx)
	return anchor
}

func axisLabelPadPx(ctx *DrawContext) float64 {
	const labelPadPt = 4.0

	dpi := 96.0
	if ctx != nil && ctx.RC.DPI > 0 {
		dpi = ctx.RC.DPI
	}
	return labelPadPt * dpi / 72.0
}

func spinePixelX(side AxisSide, px geom.Rect) float64 {
	p1, _ := spinePixelEndpoints(side, px)
	return p1.X
}

func (a *Axes) adjustedLayout(f *Figure) geom.Rect {
	px := a.layout(f)
	target := 0.0
	if a.boxAspect > 0 {
		target = a.boxAspect
	} else {
		switch a.aspectMode {
		case "equal":
			target = a.dataAspectTarget(1)
		case "ratio":
			target = a.dataAspectTarget(a.aspectValue)
		}
	}
	if target <= 0 || math.IsNaN(target) || math.IsInf(target, 0) {
		return px
	}
	return rectWithAspect(px, target)
}

func (a *Axes) dataAspectTarget(aspect float64) float64 {
	if a == nil || aspect <= 0 {
		return 0
	}
	xMin, xMax := currentScaleDomain(a.effectiveXScale())
	yMin, yMax := currentScaleDomain(a.effectiveYScale())
	xSpan := math.Abs(xMax - xMin)
	ySpan := math.Abs(yMax - yMin)
	if xSpan == 0 || ySpan == 0 {
		return 0
	}
	return aspect * ySpan / xSpan
}

func rectWithAspect(r geom.Rect, target float64) geom.Rect {
	if target <= 0 {
		return r
	}
	cur := r.H() / r.W()
	switch {
	case cur > target:
		newH := r.W() * target
		pad := (r.H() - newH) / 2
		r.Min.Y += pad
		r.Max.Y -= pad
	case cur < target:
		newW := r.H() / target
		pad := (r.W() - newW) / 2
		r.Min.X += pad
		r.Max.X -= pad
	}
	return r
}

// axesToPixel returns an affine mapping [0..1]^2 (axes space) -> pixel rect.
// This maps axes coordinates to pixel coordinates with Y-flip for mathematical orientation:
// - axes (0,0) -> pixel (px.Min.X, px.Max.Y) [bottom-left]
// - axes (1,1) -> pixel (px.Max.X, px.Min.Y) [top-right]
func axesToPixel(px geom.Rect) geom.Affine {
	sx := px.W()
	sy := -px.H() // Negative to flip Y-axis
	tx := px.Min.X
	ty := px.Max.Y // Start from bottom of pixel rect
	return geom.Affine{A: sx, D: sy, E: tx, F: ty}
}

func cloneAxisForSide(src *Axis, side AxisSide) *Axis {
	var axis Axis
	if src != nil {
		axis = *src
	} else {
		switch side {
		case AxisBottom, AxisTop:
			axis = *NewXAxis()
		case AxisLeft, AxisRight:
			axis = *NewYAxis()
		}
	}
	axis.Side = side
	axis.ShowSpine = true
	axis.ShowTicks = true
	axis.ShowLabels = true
	return &axis
}

func (a *Axes) newOverlayAxes() *Axes {
	if a == nil || a.figure == nil {
		return nil
	}
	overlay := a.figure.AddAxes(a.RectFraction)
	overlay.RC = a.RC
	overlay.aspectMode = a.aspectMode
	overlay.aspectValue = a.aspectValue
	overlay.boxAspect = a.boxAspect
	return overlay
}

func (a *Axes) newSecondaryAxes(isX bool, side AxisSide, forward func(float64) float64, inverse func(float64) (float64, bool)) (*Axes, error) {
	if a == nil || a.figure == nil {
		return nil, fmt.Errorf("secondary axes require a figure-backed axes")
	}
	if forward == nil || inverse == nil {
		return nil, fmt.Errorf("secondary axes require forward and inverse functions")
	}
	overlay := a.newOverlayAxes()
	if overlay == nil {
		return nil, fmt.Errorf("could not create overlay axes")
	}
	overlay.ShowFrame = false

	if isX {
		overlay.XScale = linkedSecondaryScale{parent: a, isX: true, forward: forward, inverse: inverse}
		overlay.shareY = a.yScaleRoot()
		if overlay.YAxis != nil {
			overlay.YAxis.ShowSpine = false
			overlay.YAxis.ShowTicks = false
			overlay.YAxis.ShowLabels = false
		}
		if overlay.XAxis != nil {
			overlay.XAxis.ShowSpine = false
			overlay.XAxis.ShowTicks = false
			overlay.XAxis.ShowLabels = false
		}
		if side == AxisTop {
			overlay.TopAxis()
		} else {
			overlay.XAxis = cloneAxisForSide(a.XAxis, AxisBottom)
		}
	} else {
		overlay.YScale = linkedSecondaryScale{parent: a, isX: false, forward: forward, inverse: inverse}
		overlay.shareX = a.xScaleRoot()
		if overlay.XAxis != nil {
			overlay.XAxis.ShowSpine = false
			overlay.XAxis.ShowTicks = false
			overlay.XAxis.ShowLabels = false
		}
		if overlay.YAxis != nil {
			overlay.YAxis.ShowSpine = false
			overlay.YAxis.ShowTicks = false
			overlay.YAxis.ShowLabels = false
		}
		if side == AxisRight {
			overlay.RightAxis()
		} else {
			overlay.YAxis = cloneAxisForSide(a.YAxis, AxisLeft)
		}
	}
	return overlay, nil
}

type linkedSecondaryScale struct {
	parent  *Axes
	isX     bool
	forward func(float64) float64
	inverse func(float64) (float64, bool)
}

func (s linkedSecondaryScale) primaryScale() transform.Scale {
	if s.parent == nil {
		return nil
	}
	if s.isX {
		return s.parent.effectiveXScale()
	}
	return s.parent.effectiveYScale()
}

func (s linkedSecondaryScale) Domain() (float64, float64) {
	base := s.primaryScale()
	if base == nil || s.forward == nil {
		return 0, 1
	}
	minVal, maxVal := base.Domain()
	return s.forward(minVal), s.forward(maxVal)
}

func (s linkedSecondaryScale) Fwd(x float64) float64 {
	base := s.primaryScale()
	if base == nil || s.inverse == nil {
		return 0
	}
	primary, ok := s.inverse(x)
	if !ok {
		return math.NaN()
	}
	return base.Fwd(primary)
}

func (s linkedSecondaryScale) Inv(u float64) (float64, bool) {
	base := s.primaryScale()
	if base == nil || s.forward == nil {
		return 0, false
	}
	primary, ok := base.Inv(u)
	if !ok {
		return 0, false
	}
	return s.forward(primary), true
}

func validateAxisSpec(spec string) error {
	switch normalizeAxisSpec(spec) {
	case "", "both", "x", "y", "bottom", "top", "left", "right":
		return nil
	default:
		return fmt.Errorf("unsupported axis selection %q", spec)
	}
}

func normalizeAxisSpec(spec string) string {
	spec = strings.ToLower(strings.TrimSpace(spec))
	if spec == "" {
		return "both"
	}
	return spec
}

func normalizeTickWhich(which string) string {
	switch strings.ToLower(strings.TrimSpace(which)) {
	case "", "both":
		return "both"
	case "major":
		return "major"
	case "minor":
		return "minor"
	default:
		return ""
	}
}

func applyTickLabelParams(style *TickLabelStyle, params TickParams) {
	if style == nil {
		return
	}
	*style = normalizeTickLabelStyle(*style)
	if params.LabelRotation != nil {
		style.Rotation = *params.LabelRotation
	}
	if params.LabelPad != nil {
		style.Pad = *params.LabelPad
	}
	if params.LabelHAlign != nil {
		style.HAlign = *params.LabelHAlign
		style.AutoAlign = false
	}
	if params.LabelVAlign != nil {
		style.VAlign = *params.LabelVAlign
		style.AutoAlign = false
	}
}

func (a *Axes) axesForSpec(spec string) []*Axis {
	switch normalizeAxisSpec(spec) {
	case "x":
		return []*Axis{a.XAxis, a.XAxisTop}
	case "y":
		return []*Axis{a.YAxis, a.YAxisRight}
	case "bottom":
		return []*Axis{a.XAxis}
	case "top":
		return []*Axis{a.XAxisTop}
	case "left":
		return []*Axis{a.YAxis}
	case "right":
		return []*Axis{a.YAxisRight}
	default:
		return []*Axis{a.XAxis, a.XAxisTop, a.YAxis, a.YAxisRight}
	}
}

func enableMinorTicks(axis *Axis) {
	if axis == nil || axis.MinorLocator != nil {
		return
	}
	switch loc := axis.Locator.(type) {
	case LogLocator:
		axis.MinorLocator = LogLocator{Base: loc.Base, Minor: true, Subs: loc.Subs}
	case AutoLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	case MaxNLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	case MultipleLocator:
		axis.MinorLocator = AutoMinorLocator{Major: loc}
	default:
		axis.MinorLocator = MinorLinearLocator{}
	}
}

type invertedScale struct {
	base transform.Scale
}

func (s invertedScale) Fwd(x float64) float64 {
	return 1 - s.base.Fwd(x)
}

func (s invertedScale) Inv(u float64) (float64, bool) {
	return s.base.Inv(1 - u)
}

func (s invertedScale) Domain() (float64, float64) {
	maxVal, minVal := s.base.Domain()
	return minVal, maxVal
}

func toggleInvertedScale(s transform.Scale) transform.Scale {
	if s == nil {
		return nil
	}
	if inv, ok := s.(invertedScale); ok {
		return inv.base
	}
	return invertedScale{base: s}
}

func scaleDomainDescending(s transform.Scale) bool {
	if s == nil {
		return false
	}
	minVal, maxVal := s.Domain()
	return minVal > maxVal
}
