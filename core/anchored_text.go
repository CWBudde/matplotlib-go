package core

import (
	"strings"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/style"
)

// AnchoredTextOptions configures an anchored annotation box.
type AnchoredTextOptions struct {
	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
}

// AnchoredTextBox draws a boxed block of text anchored to an axes or figure corner.
type AnchoredTextBox struct {
	Content string

	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	z               float64
}

// AddAnchoredText appends an anchored text box inside an axes.
func (a *Axes) AddAnchoredText(text string, opts ...AnchoredTextOptions) *AnchoredTextBox {
	box := newAnchoredTextBox(text, a.resolvedRC(), opts...)
	a.Add(box)
	return box
}

// AddAnchoredText appends a figure-level anchored text box.
func (f *Figure) AddAnchoredText(text string, opts ...AnchoredTextOptions) *AnchoredTextBox {
	rc := style.CurrentDefaults()
	if f != nil {
		rc = f.RC
	}
	box := newAnchoredTextBox(text, rc, opts...)
	if f != nil {
		f.Add(box)
	}
	return box
}

func newAnchoredTextBox(text string, rc style.RC, opts ...AnchoredTextOptions) *AnchoredTextBox {
	cfg := AnchoredTextOptions{
		Location:        LegendUpperLeft,
		Padding:         8,
		Inset:           8,
		RowGap:          4,
		CornerRadius:    0,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     1,
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Padding <= 0 {
			cfg.Padding = 8
		}
		if cfg.Inset <= 0 {
			cfg.Inset = 8
		}
		if cfg.RowGap < 0 {
			cfg.RowGap = 0
		}
		if cfg.BackgroundColor == (render.Color{}) {
			cfg.BackgroundColor = rc.LegendBackground
		}
		if cfg.BorderColor == (render.Color{}) {
			cfg.BorderColor = rc.LegendBorderColor
		}
		if cfg.TextColor == (render.Color{}) {
			cfg.TextColor = rc.LegendTextColor
		}
		if cfg.BorderWidth <= 0 {
			cfg.BorderWidth = 1
		}
	}

	return &AnchoredTextBox{
		Content:         text,
		Location:        cfg.Location,
		Locator:         cfg.Locator,
		Padding:         cfg.Padding,
		Inset:           cfg.Inset,
		RowGap:          cfg.RowGap,
		CornerRadius:    cfg.CornerRadius,
		BackgroundColor: cfg.BackgroundColor,
		BorderColor:     cfg.BorderColor,
		TextColor:       cfg.TextColor,
		BorderWidth:     cfg.BorderWidth,
		FontSize:        cfg.FontSize,
		z:               950,
	}
}

// Draw renders the anchored text box.
func (a *AnchoredTextBox) Draw(r render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	lines := strings.Split(a.Content, "\n")
	if len(lines) == 0 {
		return
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	maxWidth := 0.0
	lineHeights := make([]float64, len(lines))
	layouts := make([]singleLineTextLayout, len(lines))
	for i, line := range lines {
		layout := measureSingleLineTextLayout(r, line, fontSize, ctx.RC.FontKey)
		layouts[i] = layout
		if layout.Width > maxWidth {
			maxWidth = layout.Width
		}
		lineHeight := layout.Height
		if lineHeight < fontSize {
			lineHeight = fontSize
		}
		lineHeights[i] = lineHeight
	}

	contentHeight := 0.0
	for _, h := range lineHeights {
		contentHeight += h
	}
	if len(lineHeights) > 1 {
		contentHeight += a.RowGap * float64(len(lineHeights)-1)
	}

	box := resolveAnchoredBoxRect(a.Locator, ctx.Clip, maxWidth+a.Padding*2, contentHeight+a.Padding*2, a.Location, a.Inset)
	r.Path(pixelRectPath(box), &render.Paint{
		Fill:      a.BackgroundColor,
		Stroke:    a.BorderColor,
		LineWidth: a.BorderWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})

	y := box.Min.Y + a.Padding
	for i, line := range lines {
		layout := layouts[i]
		if line == "" {
			y += lineHeights[i] + a.RowGap
			continue
		}
		anchor := geom.Pt{
			X: box.Min.X + a.Padding,
			Y: y,
		}
		drawDisplayText(
			textRen,
			line,
			alignedSingleLineOrigin(anchor, layout, TextAlignLeft, textLayoutVAlignTop),
			fontSize,
			resolvedTextColor(a.TextColor, ctx),
			ctx.RC.FontKey,
		)
		y += lineHeights[i] + a.RowGap
	}
}

// Bounds returns an empty rect so anchored text boxes do not affect autoscaling.
func (a *AnchoredTextBox) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the anchored text box z-order.
func (a *AnchoredTextBox) Z() float64 { return a.z }

// SetLocator overrides the anchored-box placement strategy for this text box.
func (a *AnchoredTextBox) SetLocator(locator AnchoredBoxLocator) {
	if a == nil {
		return
	}
	a.Locator = locator
}

func (a *AnchoredTextBox) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if a == nil || r == nil || ctx == nil {
		return geom.Rect{}, false
	}

	lines := strings.Split(a.Content, "\n")
	if len(lines) == 0 {
		return geom.Rect{}, false
	}

	fontSize := resolvedFontSize(a.FontSize, ctx)
	maxWidth := 0.0
	contentHeight := 0.0
	for _, line := range lines {
		layout := measureSingleLineTextLayout(r, line, fontSize, ctx.RC.FontKey)
		if layout.Width > maxWidth {
			maxWidth = layout.Width
		}
		lineHeight := layout.Height
		if lineHeight < fontSize {
			lineHeight = fontSize
		}
		contentHeight += lineHeight
	}
	if len(lines) > 1 {
		contentHeight += a.RowGap * float64(len(lines)-1)
	}

	return resolveAnchoredBoxRect(a.Locator, ctx.Clip, maxWidth+a.Padding*2, contentHeight+a.Padding*2, a.Location, a.Inset), true
}

func anchoredBoxRect(clip geom.Rect, width, height float64, location LegendLocation, inset float64) geom.Rect {
	var minX, minY float64
	switch location {
	case LegendUpperLeft:
		minX = clip.Min.X + inset
		minY = clip.Min.Y + inset
	case LegendLowerRight:
		minX = clip.Max.X - inset - width
		minY = clip.Max.Y - inset - height
	case LegendLowerLeft:
		minX = clip.Min.X + inset
		minY = clip.Max.Y - inset - height
	default:
		minX = clip.Max.X - inset - width
		minY = clip.Min.Y + inset
	}
	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: minX + width, Y: minY + height},
	}
}
