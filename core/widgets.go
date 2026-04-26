package core

import (
	"fmt"
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// ButtonOptions configures a Button widget artist.
type ButtonOptions struct {
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	Pressed   *bool
	FontSize  float64
}

// SliderOptions configures a Slider widget artist.
type SliderOptions struct {
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	FontSize    float64
}

// CheckButtonsOptions configures a CheckButtons widget artist.
type CheckButtonsOptions struct {
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	CheckColor render.Color
	FontSize   float64
}

// RadioButtonsOptions configures a RadioButtons widget artist.
type RadioButtonsOptions struct {
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	DotColor   render.Color
	FontSize   float64
}

// TextBoxOptions configures a TextBox widget artist.
type TextBoxOptions struct {
	FaceColor   render.Color
	EdgeColor   render.Color
	TextColor   render.Color
	Placeholder string
	FontSize    float64
	Active      *bool
}

// Button draws a static button-style control inside its owning axes.
type Button struct {
	Label     string
	FaceColor render.Color
	EdgeColor render.Color
	TextColor render.Color
	Pressed   bool
	FontSize  float64
	z         float64
}

// Slider draws a static slider control inside its owning axes.
type Slider struct {
	Label       string
	Min         float64
	Max         float64
	Value       float64
	FaceColor   render.Color
	TrackColor  render.Color
	FillColor   render.Color
	HandleColor render.Color
	TextColor   render.Color
	FontSize    float64
	z           float64
}

// CheckButtons draws a static checklist-style control.
type CheckButtons struct {
	Labels     []string
	Values     []bool
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	CheckColor render.Color
	FontSize   float64
	z          float64
}

// RadioButtons draws a static radio-button control.
type RadioButtons struct {
	Labels     []string
	Active     int
	FaceColor  render.Color
	EdgeColor  render.Color
	TextColor  render.Color
	DotColor   render.Color
	FontSize   float64
	z          float64
}

// TextBox draws a static text-entry control.
type TextBox struct {
	Label       string
	Value       string
	Placeholder string
	FaceColor   render.Color
	EdgeColor   render.Color
	TextColor   render.Color
	FontSize    float64
	Active      bool
	z           float64
}

// Button adds a button widget artist to the axes.
func (a *Axes) Button(label string, opts ...ButtonOptions) *Button {
	if a == nil {
		return nil
	}
	cfg := ButtonOptions{
		FaceColor: render.Color{R: 0.94, G: 0.95, B: 0.97, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeButtonOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &Button{
		Label:     label,
		FaceColor: cfg.FaceColor,
		EdgeColor: cfg.EdgeColor,
		TextColor: cfg.TextColor,
		Pressed:   boolValue(cfg.Pressed, false),
		FontSize:  cfg.FontSize,
		z:         1200,
	}
	a.Add(w)
	return w
}

// Slider adds a slider widget artist to the axes.
func (a *Axes) Slider(label string, min, max, value float64, opts ...SliderOptions) *Slider {
	if a == nil {
		return nil
	}
	cfg := SliderOptions{
		FaceColor:   render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		TrackColor:  render.Color{R: 0.83, G: 0.85, B: 0.89, A: 1},
		FillColor:   render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
		HandleColor: render.Color{R: 0.09, G: 0.18, B: 0.34, A: 1},
		TextColor:   render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeSliderOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &Slider{
		Label:       label,
		Min:         min,
		Max:         max,
		Value:       clampSliderValue(min, max, value),
		FaceColor:   cfg.FaceColor,
		TrackColor:  cfg.TrackColor,
		FillColor:   cfg.FillColor,
		HandleColor: cfg.HandleColor,
		TextColor:   cfg.TextColor,
		FontSize:    cfg.FontSize,
		z:           1200,
	}
	a.Add(w)
	return w
}

// CheckButtons adds a check-button widget artist to the axes.
func (a *Axes) CheckButtons(labels []string, active []bool, opts ...CheckButtonsOptions) *CheckButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	cfg := CheckButtonsOptions{
		FaceColor:  render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		EdgeColor:  render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor:  render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
		CheckColor: render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeCheckButtonsOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	values := make([]bool, len(labels))
	copy(values, active)
	w := &CheckButtons{
		Labels:     append([]string(nil), labels...),
		Values:     values,
		FaceColor:  cfg.FaceColor,
		EdgeColor:  cfg.EdgeColor,
		TextColor:  cfg.TextColor,
		CheckColor: cfg.CheckColor,
		FontSize:   cfg.FontSize,
		z:          1200,
	}
	a.Add(w)
	return w
}

// RadioButtons adds a radio-button widget artist to the axes.
func (a *Axes) RadioButtons(labels []string, active int, opts ...RadioButtonsOptions) *RadioButtons {
	if a == nil || len(labels) == 0 {
		return nil
	}
	cfg := RadioButtonsOptions{
		FaceColor: render.Color{R: 0.96, G: 0.97, B: 0.98, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
		DotColor:  render.Color{R: 0.85, G: 0.32, B: 0.17, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeRadioButtonsOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &RadioButtons{
		Labels:     append([]string(nil), labels...),
		Active:     clampInt(active, 0, len(labels)-1),
		FaceColor:  cfg.FaceColor,
		EdgeColor:  cfg.EdgeColor,
		TextColor:  cfg.TextColor,
		DotColor:   cfg.DotColor,
		FontSize:   cfg.FontSize,
		z:          1200,
	}
	a.Add(w)
	return w
}

// TextBox adds a text-box widget artist to the axes.
func (a *Axes) TextBox(label, value string, opts ...TextBoxOptions) *TextBox {
	if a == nil {
		return nil
	}
	cfg := TextBoxOptions{
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0.74, G: 0.76, B: 0.80, A: 1},
		TextColor: render.Color{R: 0.12, G: 0.13, B: 0.16, A: 1},
	}
	if len(opts) > 0 {
		cfg = mergeTextBoxOptions(cfg, opts[0])
	}
	prepareWidgetAxes(a)
	w := &TextBox{
		Label:       label,
		Value:       value,
		Placeholder: cfg.Placeholder,
		FaceColor:   cfg.FaceColor,
		EdgeColor:   cfg.EdgeColor,
		TextColor:   cfg.TextColor,
		FontSize:    cfg.FontSize,
		Active:      boolValue(cfg.Active, false),
		z:           1200,
	}
	a.Add(w)
	return w
}

func (b *Button) Draw(r render.Renderer, ctx *DrawContext) {
	if b == nil || r == nil || ctx == nil {
		return
	}
	bounds := insetRect(ctx.Clip, 6)
	fill := b.FaceColor
	if b.Pressed {
		fill = mixColor(fill, render.Color{R: 0, G: 0, B: 0, A: 1}, 0.12)
	}
	drawWidgetPanel(r, bounds, fill, b.EdgeColor, 1.25, 10)
	drawCenteredWidgetText(r, ctx, geom.Pt{
		X: bounds.Min.X + bounds.W()/2,
		Y: bounds.Min.Y + bounds.H()/2,
	}, b.Label, b.FontSize, b.TextColor)
}

func (b *Button) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
func (b *Button) Z() float64                    { return b.z }

func (s *Slider) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, s.FaceColor, render.Color{A: 0}, 0, 12)
	textColor := s.TextColor
	fontSize := resolvedFontSize(s.FontSize, ctx)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Min.X + 14, Y: panel.Min.Y + 22}, s.Label, fontSize, textColor, TextAlignLeft, textLayoutVAlignTop)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Max.X - 14, Y: panel.Min.Y + 22}, fmt.Sprintf("%.2f", s.Value), fontSize, textColor, TextAlignRight, textLayoutVAlignTop)

	track := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 14, Y: panel.Max.Y - 26},
		Max: geom.Pt{X: panel.Max.X - 14, Y: panel.Max.Y - 14},
	}
	drawWidgetPanel(r, track, s.TrackColor, render.Color{A: 0}, 0, track.H()/2)
	fraction := sliderFraction(s.Min, s.Max, s.Value)
	fill := track
	fill.Max.X = fill.Min.X + track.W()*fraction
	drawWidgetPanel(r, fill, s.FillColor, render.Color{A: 0}, 0, fill.H()/2)
	handleX := track.Min.X + track.W()*fraction
	handle := ellipsePath(track.H()*1.9, track.H()*1.9)
	handlePath := applyAffinePath(handle, patchAffine(geom.Pt{X: handleX, Y: track.Min.Y + track.H()/2}, 0))
	r.Path(handlePath, &render.Paint{
		Fill:      s.HandleColor,
		Stroke:    mixColor(s.HandleColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.2),
		LineWidth: 1,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	})
}

func (s *Slider) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
func (s *Slider) Z() float64                    { return s.z }

func (c *CheckButtons) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, c.FaceColor, c.EdgeColor, 1.1, 12)
	if len(c.Labels) == 0 {
		return
	}
	rowHeight := panel.H() / float64(len(c.Labels))
	fontSize := resolvedFontSize(c.FontSize, ctx)
	for i, label := range c.Labels {
		rowMinY := panel.Min.Y + rowHeight*float64(i)
		rowMaxY := rowMinY + rowHeight
		boxSize := math.Min(16, rowHeight*0.42)
		box := geom.Rect{
			Min: geom.Pt{X: panel.Min.X + 14, Y: rowMinY + (rowHeight-boxSize)/2},
			Max: geom.Pt{X: panel.Min.X + 14 + boxSize, Y: rowMaxY - (rowHeight-boxSize)/2},
		}
		drawWidgetPanel(r, box, render.Color{R: 1, G: 1, B: 1, A: 1}, c.EdgeColor, 1, 3)
		if i < len(c.Values) && c.Values[i] {
			path := geom.Path{}
			path.MoveTo(geom.Pt{X: box.Min.X + box.W()*0.18, Y: box.Min.Y + box.H()*0.56})
			path.LineTo(geom.Pt{X: box.Min.X + box.W()*0.42, Y: box.Max.Y - box.H()*0.20})
			path.LineTo(geom.Pt{X: box.Max.X - box.W()*0.16, Y: box.Min.Y + box.H()*0.22})
			r.Path(path, &render.Paint{
				Stroke:    c.CheckColor,
				LineWidth: 2,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapRound,
			})
		}
		drawWidgetText(r, ctx, geom.Pt{X: box.Max.X + 10, Y: rowMinY + rowHeight/2}, label, fontSize, c.TextColor, TextAlignLeft, textLayoutVAlignCenter)
	}
}

func (c *CheckButtons) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
func (c *CheckButtons) Z() float64                    { return c.z }

func (rdo *RadioButtons) Draw(r render.Renderer, ctx *DrawContext) {
	if rdo == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, rdo.FaceColor, rdo.EdgeColor, 1.1, 12)
	if len(rdo.Labels) == 0 {
		return
	}
	rowHeight := panel.H() / float64(len(rdo.Labels))
	fontSize := resolvedFontSize(rdo.FontSize, ctx)
	for i, label := range rdo.Labels {
		center := geom.Pt{X: panel.Min.X + 24, Y: panel.Min.Y + rowHeight*float64(i) + rowHeight/2}
		outer := ellipsePath(16, 16)
		outerPath := applyAffinePath(outer, patchAffine(center, 0))
		r.Path(outerPath, &render.Paint{
			Fill:      render.Color{R: 1, G: 1, B: 1, A: 1},
			Stroke:    rdo.EdgeColor,
			LineWidth: 1,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
		if i == clampInt(rdo.Active, 0, len(rdo.Labels)-1) {
			inner := ellipsePath(8, 8)
			innerPath := applyAffinePath(inner, patchAffine(center, 0))
			r.Path(innerPath, &render.Paint{Fill: rdo.DotColor})
		}
		drawWidgetText(r, ctx, geom.Pt{X: center.X + 16, Y: center.Y}, label, fontSize, rdo.TextColor, TextAlignLeft, textLayoutVAlignCenter)
	}
}

func (rdo *RadioButtons) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
func (rdo *RadioButtons) Z() float64                    { return rdo.z }

func (t *TextBox) Draw(r render.Renderer, ctx *DrawContext) {
	if t == nil || r == nil || ctx == nil {
		return
	}
	panel := insetRect(ctx.Clip, 4)
	drawWidgetPanel(r, panel, render.Color{A: 0}, render.Color{A: 0}, 0, 0)
	fontSize := resolvedFontSize(t.FontSize, ctx)
	drawWidgetText(r, ctx, geom.Pt{X: panel.Min.X + 4, Y: panel.Min.Y + 20}, t.Label, fontSize, t.TextColor, TextAlignLeft, textLayoutVAlignTop)

	input := geom.Rect{
		Min: geom.Pt{X: panel.Min.X + 4, Y: panel.Min.Y + 30},
		Max: geom.Pt{X: panel.Max.X - 4, Y: panel.Max.Y - 8},
	}
	edge := t.EdgeColor
	if t.Active {
		edge = mixColor(edge, render.Color{R: 0.16, G: 0.42, B: 0.76, A: 1}, 0.65)
	}
	drawWidgetPanel(r, input, t.FaceColor, edge, 1.2, 8)

	display := t.Value
	displayColor := t.TextColor
	if display == "" {
		display = t.Placeholder
		displayColor = mixColor(t.TextColor, render.Color{R: 1, G: 1, B: 1, A: 1}, 0.45)
	}
	drawWidgetText(r, ctx, geom.Pt{X: input.Min.X + 12, Y: input.Min.Y + input.H()/2}, display, fontSize, displayColor, TextAlignLeft, textLayoutVAlignCenter)
	if t.Active {
		caretX := input.Min.X + 16 + math.Min(input.W()-24, float64(len(display))*fontSize*0.42)
		r.Path(pixelLinePath(
			geom.Pt{X: caretX, Y: input.Min.Y + 8},
			geom.Pt{X: caretX, Y: input.Max.Y - 8},
		), &render.Paint{
			Stroke:    edge,
			LineWidth: 1.2,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
		})
	}
}

func (t *TextBox) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }
func (t *TextBox) Z() float64                    { return t.z }

func prepareWidgetAxes(a *Axes) {
	a.XAxis.ShowSpine = false
	a.XAxis.ShowTicks = false
	a.XAxis.ShowLabels = false
	a.YAxis.ShowSpine = false
	a.YAxis.ShowTicks = false
	a.YAxis.ShowLabels = false
	a.SetXLim(0, 1)
	a.SetYLim(0, 1)
}

func drawWidgetPanel(r render.Renderer, rect geom.Rect, fill, edge render.Color, width, radius float64) {
	if rect.W() <= 0 || rect.H() <= 0 {
		return
	}
	path := roundedRectPath(rect, math.Min(radius, math.Min(rect.W(), rect.H())/2))
	paint := render.Paint{Fill: fill}
	if edge.A > 0 && width > 0 {
		paint.Stroke = edge
		paint.LineWidth = width
		paint.LineJoin = render.JoinRound
		paint.LineCap = render.CapRound
	}
	r.Path(path, &paint)
}

func drawCenteredWidgetText(r render.Renderer, ctx *DrawContext, center geom.Pt, text string, size float64, color render.Color) {
	drawWidgetText(r, ctx, center, text, size, color, TextAlignCenter, textLayoutVAlignCenter)
}

func drawWidgetText(r render.Renderer, ctx *DrawContext, anchor geom.Pt, text string, size float64, color render.Color, hAlign TextAlign, vAlign textLayoutVerticalAlign) {
	textRen, ok := r.(render.TextDrawer)
	if !ok || displayTextIsEmpty(text) {
		return
	}
	fontSize := resolvedFontSize(size, ctx)
	layout := measureSingleLineTextLayout(r, text, fontSize, ctx.RC.FontKey)
	origin := alignedSingleLineOrigin(anchor, layout, hAlign, vAlign)
	drawDisplayText(textRen, text, origin, fontSize, resolvedTextColor(color, ctx), ctx.RC.FontKey)
}

func insetRect(rect geom.Rect, pad float64) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: rect.Min.X + pad, Y: rect.Min.Y + pad},
		Max: geom.Pt{X: rect.Max.X - pad, Y: rect.Max.Y - pad},
	}
}

func sliderFraction(min, max, value float64) float64 {
	if max <= min {
		return 0
	}
	return clampFloat((value-min)/(max-min), 0, 1)
}

func clampSliderValue(min, max, value float64) float64 {
	if max <= min {
		return min
	}
	return math.Max(min, math.Min(max, value))
}

func clampInt(v, minVal, maxVal int) int {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func mixColor(a, b render.Color, t float64) render.Color {
	t = clampFloat(t, 0, 1)
	return render.Color{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}

func mergeButtonOptions(base, override ButtonOptions) ButtonOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.Pressed != nil {
		base.Pressed = override.Pressed
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}

func mergeSliderOptions(base, override SliderOptions) SliderOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.TrackColor != (render.Color{}) {
		base.TrackColor = override.TrackColor
	}
	if override.FillColor != (render.Color{}) {
		base.FillColor = override.FillColor
	}
	if override.HandleColor != (render.Color{}) {
		base.HandleColor = override.HandleColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}

func mergeCheckButtonsOptions(base, override CheckButtonsOptions) CheckButtonsOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.CheckColor != (render.Color{}) {
		base.CheckColor = override.CheckColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}

func mergeRadioButtonsOptions(base, override RadioButtonsOptions) RadioButtonsOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.DotColor != (render.Color{}) {
		base.DotColor = override.DotColor
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	return base
}

func mergeTextBoxOptions(base, override TextBoxOptions) TextBoxOptions {
	if override.FaceColor != (render.Color{}) {
		base.FaceColor = override.FaceColor
	}
	if override.EdgeColor != (render.Color{}) {
		base.EdgeColor = override.EdgeColor
	}
	if override.TextColor != (render.Color{}) {
		base.TextColor = override.TextColor
	}
	if override.Placeholder != "" {
		base.Placeholder = override.Placeholder
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	if override.Active != nil {
		base.Active = override.Active
	}
	return base
}
