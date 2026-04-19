package core

import (
	"math"
	"strings"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

// TextAlign controls horizontal text anchoring.
type TextAlign uint8

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// TextVerticalAlign controls vertical text anchoring.
type TextVerticalAlign uint8

const (
	TextVAlignBaseline TextVerticalAlign = iota
	TextVAlignBottom
	TextVAlignMiddle
	TextVAlignTop
)

// TextOptions configures a Text artist.
type TextOptions struct {
	FontSize float64
	Color    render.Color
	HAlign   TextAlign
	VAlign   TextVerticalAlign
	Coords   CoordinateSpec
	OffsetX  float64
	OffsetY  float64
}

// AnnotationOptions configures an Annotation artist.
type AnnotationOptions struct {
	Coords        CoordinateSpec
	OffsetX       float64
	OffsetY       float64
	FontSize      float64
	Color         render.Color
	ArrowColor    render.Color
	ArrowWidth    float64
	ArrowHeadSize float64
	HAlign        TextAlign
	VAlign        TextVerticalAlign
}

// Text renders arbitrary text at a data-space position.
type Text struct {
	Position geom.Pt
	Content  string

	FontSize float64
	Color    render.Color
	HAlign   TextAlign
	VAlign   TextVerticalAlign
	Coords   CoordinateSpec
	OffsetX  float64
	OffsetY  float64
	z        float64
}

// Annotation renders text offset from a data point with an arrow.
type Annotation struct {
	Point   geom.Pt
	Content string

	OffsetX       float64
	OffsetY       float64
	FontSize      float64
	Color         render.Color
	ArrowColor    render.Color
	ArrowWidth    float64
	ArrowHeadSize float64
	HAlign        TextAlign
	VAlign        TextVerticalAlign
	Coords        CoordinateSpec
	z             float64
}

var basicSymbolReplacer = strings.NewReplacer(
	`\\alpha`, "α",
	`\\beta`, "β",
	`\\gamma`, "γ",
	`\\delta`, "δ",
	`\\epsilon`, "ε",
	`\\theta`, "θ",
	`\\lambda`, "λ",
	`\\mu`, "μ",
	`\\pi`, "π",
	`\\sigma`, "σ",
	`\\omega`, "ω",
	`\\Delta`, "Δ",
	`\\Theta`, "Θ",
	`\\Lambda`, "Λ",
	`\\Sigma`, "Σ",
	`\\Omega`, "Ω",
	`\\pm`, "±",
	`\\times`, "×",
	`\\deg`, "°",
	`\\le`, "≤",
	`\\ge`, "≥",
	`\\neq`, "≠",
	`\\rightarrow`, "→",
)

// Text adds arbitrary text positioned in data coordinates.
func (a *Axes) Text(x, y float64, text string, opts ...TextOptions) *Text {
	opt := TextOptions{
		HAlign: TextAlignLeft,
		VAlign: TextVAlignBaseline,
	}
	if len(opts) > 0 {
		opt = opts[0]
	}

	artist := &Text{
		Position: geom.Pt{X: x, Y: y},
		Content:  text,
		FontSize: opt.FontSize,
		Color:    opt.Color,
		HAlign:   opt.HAlign,
		VAlign:   opt.VAlign,
		Coords:   opt.Coords,
		OffsetX:  opt.OffsetX,
		OffsetY:  opt.OffsetY,
		z:        500,
	}
	a.Add(artist)
	return artist
}

// Annotate adds an arrow annotation pointing to a data-space point.
func (a *Axes) Annotate(text string, x, y float64, opts ...AnnotationOptions) *Annotation {
	opt := AnnotationOptions{
		OffsetX:       28,
		OffsetY:       -20,
		ArrowWidth:    1.25,
		ArrowHeadSize: 8,
	}
	if len(opts) > 0 {
		opt = opts[0]
		if opt.OffsetX == 0 && opt.OffsetY == 0 {
			opt.OffsetX = 28
			opt.OffsetY = -20
		}
		if opt.ArrowWidth <= 0 {
			opt.ArrowWidth = 1.25
		}
		if opt.ArrowHeadSize <= 0 {
			opt.ArrowHeadSize = 8
		}
	}

	artist := &Annotation{
		Point:         geom.Pt{X: x, Y: y},
		Content:       text,
		OffsetX:       opt.OffsetX,
		OffsetY:       opt.OffsetY,
		FontSize:      opt.FontSize,
		Color:         opt.Color,
		ArrowColor:    opt.ArrowColor,
		ArrowWidth:    opt.ArrowWidth,
		ArrowHeadSize: opt.ArrowHeadSize,
		HAlign:        annotationHAlign(opt),
		VAlign:        annotationVAlign(opt),
		Coords:        opt.Coords,
		z:             900,
	}
	a.Add(artist)
	return artist
}

// Draw renders text inside the axes clip.
func (t *Text) Draw(r render.Renderer, ctx *DrawContext) {
	if t == nil || ctx == nil {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	content := normalizeDisplayText(t.Content)
	if content == "" {
		return
	}

	fontSize := resolvedFontSize(t.FontSize, ctx)
	anchor := transformedPoint(ctx, t.Coords, t.Position, t.OffsetX, t.OffsetY)
	metrics := r.MeasureText(content, fontSize, ctx.RC.FontKey)
	origin := alignedTextOrigin(anchor, metrics, t.HAlign, t.VAlign)
	textRen.DrawText(content, origin, fontSize, resolvedTextColor(t.Color, ctx))
}

// Bounds returns an empty rect so labels do not affect autoscaling.
func (t *Text) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the text z-order.
func (t *Text) Z() float64 { return t.z }

// Draw is a no-op because annotations render outside the axes clip via DrawOverlay.
func (a *Annotation) Draw(render.Renderer, *DrawContext) {}

// DrawOverlay renders the full annotation without the axes clip applied.
func (a *Annotation) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	content := normalizeDisplayText(a.Content)
	if content == "" {
		return
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	target := transformedPoint(ctx, a.Coords, a.Point, 0, 0)
	anchor := transformedPoint(ctx, a.Coords, a.Point, a.OffsetX, a.OffsetY)
	metrics := r.MeasureText(content, fontSize, ctx.RC.FontKey)
	origin := alignedTextOrigin(anchor, metrics, a.HAlign, a.VAlign)
	box := textBounds(origin, metrics)
	start := nearestPointOnRect(box, target)

	linePaint := render.Paint{
		Stroke:    defaultArrowColor(a.ArrowColor, a.Color),
		LineWidth: a.ArrowWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	}
	r.Path(pixelLinePath(start, target), &linePaint)

	head := annotationHeadPath(start, target, a.ArrowHeadSize)
	if len(head.C) > 0 {
		r.Path(head, &render.Paint{
			Fill:      defaultArrowColor(a.ArrowColor, a.Color),
			Stroke:    resolvedArrowColor(a.ArrowColor, a.Color, ctx),
			LineWidth: a.ArrowWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
	}

	textRen.DrawText(content, origin, fontSize, resolvedTextColor(a.Color, ctx))
}

// Bounds returns an empty rect so annotations do not affect autoscaling.
func (a *Annotation) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the annotation z-order.
func (a *Annotation) Z() float64 { return a.z }

func normalizeDisplayText(text string) string {
	return basicSymbolReplacer.Replace(text)
}

func defaultTextColor(c render.Color) render.Color {
	return resolvedTextColor(c, nil)
}

func resolvedTextColor(c render.Color, ctx *DrawContext) render.Color {
	if c == (render.Color{}) {
		if ctx != nil {
			return ctx.RC.DefaultTextColor()
		}
		return render.Color{R: 0, G: 0, B: 0, A: 1}
	}
	if c.A == 0 && (c.R != 0 || c.G != 0 || c.B != 0) {
		c.A = 1
	}
	return c
}

func defaultArrowColor(arrow, text render.Color) render.Color {
	return resolvedArrowColor(arrow, text, nil)
}

func resolvedArrowColor(arrow, text render.Color, ctx *DrawContext) render.Color {
	if arrow == (render.Color{}) {
		return resolvedTextColor(text, ctx)
	}
	if arrow.A == 0 && (arrow.R != 0 || arrow.G != 0 || arrow.B != 0) {
		arrow.A = 1
	}
	return arrow
}

func resolvedFontSize(size float64, ctx *DrawContext) float64 {
	if size > 0 {
		return size
	}
	if ctx != nil && ctx.RC.FontSize > 0 {
		return ctx.RC.FontSize
	}
	return 12
}

func annotationHAlign(opt AnnotationOptions) TextAlign {
	if opt.HAlign != TextAlignLeft {
		return opt.HAlign
	}
	if opt.OffsetX < 0 {
		return TextAlignRight
	}
	return TextAlignLeft
}

func annotationVAlign(opt AnnotationOptions) TextVerticalAlign {
	if opt.VAlign != TextVAlignBaseline {
		return opt.VAlign
	}
	if opt.OffsetY > 0 {
		return TextVAlignTop
	}
	if opt.OffsetY < 0 {
		return TextVAlignBottom
	}
	return TextVAlignMiddle
}

func alignedTextOrigin(anchor geom.Pt, metrics render.TextMetrics, hAlign TextAlign, vAlign TextVerticalAlign) geom.Pt {
	origin := geom.Pt{X: anchor.X, Y: anchor.Y}

	switch hAlign {
	case TextAlignCenter:
		origin.X -= metrics.W / 2
	case TextAlignRight:
		origin.X -= metrics.W
	}

	switch vAlign {
	case TextVAlignTop:
		origin.Y += metrics.Ascent
	case TextVAlignMiddle:
		origin.Y += (metrics.Ascent - metrics.Descent) / 2
	case TextVAlignBottom:
		origin.Y -= metrics.Descent
	}

	return origin
}

func textBounds(origin geom.Pt, metrics render.TextMetrics) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent},
		Max: geom.Pt{X: origin.X + metrics.W, Y: origin.Y + metrics.Descent},
	}
}

func nearestPointOnRect(rect geom.Rect, pt geom.Pt) geom.Pt {
	return geom.Pt{
		X: clampFloat(pt.X, rect.Min.X, rect.Max.X),
		Y: clampFloat(pt.Y, rect.Min.Y, rect.Max.Y),
	}
}

func annotationHeadPath(start, tip geom.Pt, size float64) geom.Path {
	dx := tip.X - start.X
	dy := tip.Y - start.Y
	length := math.Hypot(dx, dy)
	if length == 0 || size <= 0 {
		return geom.Path{}
	}

	ux := dx / length
	uy := dy / length
	base := geom.Pt{X: tip.X - ux*size, Y: tip.Y - uy*size}
	perpX := -uy * size * 0.45
	perpY := ux * size * 0.45

	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{
			tip,
			{X: base.X + perpX, Y: base.Y + perpY},
			{X: base.X - perpX, Y: base.Y - perpY},
		},
	}
}

func pixelLinePath(p1, p2 geom.Pt) geom.Path {
	return geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{p1, p2},
	}
}

func clampFloat(v, minVal, maxVal float64) float64 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func transformedPoint(ctx *DrawContext, spec CoordinateSpec, p geom.Pt, dxPx, dyPx float64) geom.Pt {
	if ctx == nil {
		return p
	}

	base := ctx.TransformFor(spec)
	if base == nil {
		return p
	}
	if dxPx != 0 || dyPx != 0 {
		return transform.NewOffset(base, geom.Pt{X: dxPx, Y: dyPx}).Apply(p)
	}
	return base.Apply(p)
}
