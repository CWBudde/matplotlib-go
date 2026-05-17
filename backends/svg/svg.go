package svg

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	quantizationGrid  = 1e-6
	defaultFontHeight = 13.0
)

type state struct {
	clipRect      *geom.Rect
	clipPathDepth int
}

type svgNode struct {
	content string
	// clipIDs lists active clip-path defs in outer-to-inner order. An empty
	// slice means the node is unclipped. The serializer wraps the content in
	// nested <g clip-path="url(#…)"> groups, opening outer-most first.
	clipIDs []string
}

// clipDef describes one <clipPath> entry in the SVG <defs> block. Exactly one
// of rect or path is populated. Storing both in a single ordered slice keeps
// emission in registration order so def IDs and document order stay aligned.
type clipDef struct {
	id   string
	rect *geom.Rect
	path string
}

type fontFaceDef struct {
	family string
	data   string
	mime   string
	format string
}

// Renderer implements render.Renderer using SVG path/text recording.
type Renderer struct {
	width      int
	height     int
	viewport   geom.Rect
	began      bool
	stack      []state
	clipRect   *geom.Rect
	resolution uint
	background render.Color

	nodes         []svgNode
	clipDefs      map[string]string // rect-key → clipDef ID
	clipPathDefs  map[string]string // path data → clipDef ID
	clipOrder     []clipDef         // registration-order defs (rects and paths interleaved)
	clipPathStack []string          // active path-clip IDs in outer-to-inner order
	clipIDCounter int

	fontFaces     map[string]fontFaceDef
	fontFaceOrder []fontFaceDef

	lastFontKey string
	texManager  *tex.Manager
	texErr      error
}

var (
	_ render.Renderer           = (*Renderer)(nil)
	_ render.DPIAware           = (*Renderer)(nil)
	_ render.TextDrawer         = (*Renderer)(nil)
	_ render.RotatedTextDrawer  = (*Renderer)(nil)
	_ render.VerticalTextDrawer = (*Renderer)(nil)
	_ render.TextPather         = (*Renderer)(nil)
	_ render.TeXMetricer        = (*Renderer)(nil)
	_ render.TeXDrawer          = (*Renderer)(nil)
	_ render.RotatedTeXDrawer   = (*Renderer)(nil)
	_ render.SVGExporter        = (*Renderer)(nil)
)

// New creates a new SVG renderer with the specified dimensions and background color.
func New(w, h int, bg render.Color) (*Renderer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("svg: width and height must be positive")
	}

	return &Renderer{
		width:        w,
		height:       h,
		background:   bg,
		resolution:   72,
		clipDefs:     map[string]string{},
		clipPathDefs: map[string]string{},
		fontFaces:    map[string]fontFaceDef{},
		texManager:   tex.NewManager(tex.ManagerConfig{}),
	}, nil
}

// Begin starts a drawing session with the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}

	r.began = true
	r.viewport = viewport
	r.nodes = nil
	r.stack = r.stack[:0]
	r.clipRect = nil
	r.clipDefs = map[string]string{}
	r.clipPathDefs = map[string]string{}
	r.clipOrder = nil
	r.clipPathStack = nil
	r.clipIDCounter = 0
	r.fontFaces = map[string]fontFaceDef{}
	r.fontFaceOrder = nil
	r.lastFontKey = ""
	return nil
}

// End finishes the drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("End called before Begin")
	}

	r.began = false
	r.stack = r.stack[:0]
	r.clipRect = nil
	return nil
}

// Save pushes the current graphics state onto the stack.
func (r *Renderer) Save() {
	var clipCopy *geom.Rect
	if r.clipRect != nil {
		copyRect := *r.clipRect
		clipCopy = &copyRect
	}
	r.stack = append(r.stack, state{
		clipRect:      clipCopy,
		clipPathDepth: len(r.clipPathStack),
	})
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return
	}

	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = s.clipRect
	if s.clipPathDepth < len(r.clipPathStack) {
		r.clipPathStack = r.clipPathStack[:s.clipPathDepth]
	}
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	clipRect := normalizeRect(rect)
	if r.clipRect == nil {
		r.clipRect = &clipRect
		return
	}

	intersected := r.clipRect.Intersect(clipRect)
	r.clipRect = &intersected
}

// ClipPath pushes a path-based clip region onto the active clip stack. Each
// active path clip becomes its own <clipPath> def and the rendered content is
// wrapped in nested <g clip-path="…"> groups (outer-most first) so SVG's
// natural clip-on-clip composition applies.
func (r *Renderer) ClipPath(p geom.Path) {
	if !p.Validate() {
		return
	}
	d := buildPathData(p)
	if d == "" {
		return
	}
	id := r.registerPathClip(d)
	r.clipPathStack = append(r.clipPathStack, id)
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !p.Validate() || paint == nil {
		return
	}

	d := buildPathData(p)
	if d == "" {
		return
	}

	hasFill := paint.Fill.A > 0
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}

	var b strings.Builder
	b.WriteString(`<path`)
	writeAttr(&b, "d", d)

	if hasFill {
		fillColor, fillAlpha := colorToStyle(paint.Fill)
		writeAttr(&b, "fill", fillColor)
		if fillAlpha < 1 {
			writeFloatAttr(&b, "fill-opacity", fillAlpha)
		}
	} else {
		writeAttr(&b, "fill", "none")
	}

	if hasStroke {
		strokeColor, strokeAlpha := colorToStyle(paint.Stroke)
		writeAttr(&b, "stroke", strokeColor)
		if strokeAlpha < 1 {
			writeFloatAttr(&b, "stroke-opacity", strokeAlpha)
		}
		writeFloatAttr(&b, "stroke-width", paint.LineWidth)
		writeAttr(&b, "stroke-linejoin", mapLineJoin(paint.LineJoin))
		writeAttr(&b, "stroke-linecap", mapLineCap(paint.LineCap))
		if paint.MiterLimit > 0 {
			writeFloatAttr(&b, "stroke-miterlimit", paint.MiterLimit)
		}

		if len(paint.Dashes) >= 2 {
			writeAttr(&b, "stroke-dasharray", dashedArray(paint.Dashes))
		}
	} else {
		writeAttr(&b, "stroke", "none")
	}

	b.WriteString(" />")

	r.nodes = append(r.nodes, svgNode{
		content: b.String(),
		clipIDs: r.currentClipIDs(),
	})
}

// Image draws an image within the destination rectangle.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	rgba := asRGBAImage(img)
	if rgba == nil {
		return
	}
	r.renderImageNode(rgba, dst, "")
}

func (r *Renderer) renderImageNode(rgba *image.RGBA, dst geom.Rect, transform string) {
	x := dst.Min.X
	y := dst.Min.Y
	w := dst.W()
	h := dst.H()
	if w < 0 {
		x += w
		w = -w
	}
	if h < 0 {
		y += h
		h = -h
	}
	if w <= 0 || h <= 0 {
		return
	}

	encoded, err := encodeImage(rgba)
	if err != nil {
		return
	}

	uri := "data:image/png;base64," + encoded

	var b strings.Builder
	b.WriteString(`<image x="`)
	b.WriteString(formatFloat(x))
	b.WriteString(`" y="`)
	b.WriteString(formatFloat(y))
	b.WriteString(`" width="`)
	b.WriteString(formatFloat(w))
	b.WriteString(`" height="`)
	b.WriteString(formatFloat(h))
	b.WriteString(`" preserveAspectRatio="none"`)
	b.WriteString(` href="`)
	b.WriteString(uri)
	b.WriteString(`" xlink:href="`)
	b.WriteString(uri)
	b.WriteString(`"`)
	if transform != "" {
		writeAttr(&b, "transform", transform)
	}
	b.WriteString(` />`)

	r.nodes = append(r.nodes, svgNode{
		content: b.String(),
		clipIDs: r.currentClipIDs(),
	})
}

// GlyphRun draws a run of glyph IDs as characters where available.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}

	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}

	penX := run.Origin.X
	penY := run.Origin.Y

	size := run.Size
	if size <= 0 {
		size = 12
	}

	for _, glyph := range run.Glyphs {
		if glyph.ID == 0 {
			if glyph.Advance > 0 {
				penX += glyph.Advance
			}
			continue
		}

		r.DrawText(string(rune(glyph.ID)), geom.Pt{X: penX + glyph.Offset.X, Y: penY + glyph.Offset.Y}, size, textColor)

		advance := glyph.Advance
		if advance <= 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), size, run.FontKey).W
		}
		penX += advance
	}
}

// MeasureText returns text metrics based on a built-in monospace-compatible font.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}

	scale := size / defaultFontHeight
	if scale <= 0 {
		return render.TextMetrics{}
	}

	face := basicfont.Face7x13
	width := float64(font.MeasureString(face, text).Ceil())
	height := float64(face.Metrics().Height.Ceil())
	ascent := float64(face.Metrics().Ascent.Ceil())
	desc := float64(face.Metrics().Descent.Ceil())

	if width <= 0 || height <= 0 {
		return render.TextMetrics{}
	}

	return render.TextMetrics{
		W:       quantize(width * scale),
		H:       quantize(height * scale),
		Ascent:  quantize(ascent * scale),
		Descent: quantize(desc * scale),
	}
}

// DrawText renders text using an SVG <text> element.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}

	r.renderTextNode(text, origin.X, origin.Y, size, textColor, "")
}

// DrawTextRotated renders text using Matplotlib-like anchor rotation. The
// anchor is the bottom-center of the unrotated text box.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if text == "" || size <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}

	metrics := r.MeasureText(text, size, "")
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}

	origin := geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
	transform := rotateTransform(-angle*180/math.Pi, anchor.X, anchor.Y)
	r.renderTextNode(text, origin.X, origin.Y, size, textColor, transform)
}

// DrawTextVertical renders one character per line.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}

	runes := []rune(text)
	lineMetrics := r.MeasureText("M", size, "")
	lineH := lineMetrics.H
	if lineH <= 0 {
		lineH = size
	}

	totalH := lineH * float64(len(runes))
	y := center.Y - totalH/2 + lineMetrics.Ascent

	for _, ch := range runes {
		s := string(ch)
		chMetrics := r.MeasureText(s, size, "")
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			continue
		}

		x := center.X - chMetrics.W/2
		r.renderTextNode(s, x, y, size, textColor, "")
		y += lineH
	}
}

// TextPath converts text to a vector path using the shared font manager.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// LastTeXError returns the most recent TeX pipeline error recorded by MeasureTeX
// or DrawTeX. A nil value means the last TeX operation succeeded.
func (r *Renderer) LastTeXError() error {
	if r == nil {
		return nil
	}
	return r.texErr
}

// MeasureTeX measures a TeX string by rendering it through the external
// latex+dvipng cache and using the resulting tight PNG dimensions.
func (r *Renderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}
	return result.Metrics, true
}

// DrawTeX embeds a TeX-rendered PNG as an SVG image element.
func (r *Renderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - result.Metrics.Ascent}
	r.renderImageNode(img, geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	}, "")
	return true
}

// DrawTeXRotated embeds a TeX-rendered PNG and rotates it around the
// Matplotlib-style text rotation anchor.
func (r *Renderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}

	metrics := result.Metrics
	origin := geom.Pt{X: anchor.X - metrics.W/2, Y: anchor.Y - metrics.Descent}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	transform := rotateTransform(-angle*180/math.Pi, anchor.X, anchor.Y)
	r.renderImageNode(img, geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	}, transform)
	return true
}

func (r *Renderer) renderTeX(text string, size float64, fontKey string) (tex.RenderResult, bool) {
	if r == nil || text == "" || size <= 0 {
		return tex.RenderResult{}, false
	}
	if r.texManager == nil {
		r.texManager = tex.NewManager(tex.ManagerConfig{})
	}
	result, err := r.texManager.Render(text, size, r.resolution, fontKey)
	if err != nil {
		r.texErr = err
		return tex.RenderResult{}, false
	}
	r.texErr = nil
	return result, true
}

// SetResolution sets raster-free text metric scale basis.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
	}
}

// SaveSVG writes all recorded content into an SVG document.
func (r *Renderer) SaveSVG(path string) error {
	if path == "" {
		return errors.New("svg: path is required")
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(r.renderSVG())
	return err
}

func (r *Renderer) renderSVG() string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"`+"\n"+`width="%s" height="%s" viewBox="0 0 %d %d"`+"\n"+`preserveAspectRatio="xMidYMid meet">`+"\n",
		formatFloat(float64(r.width)),
		formatFloat(float64(r.height)),
		r.width,
		r.height))

	if len(r.clipOrder) > 0 || len(r.fontFaceOrder) > 0 {
		b.WriteString("  <defs>\n")
		if len(r.fontFaceOrder) > 0 {
			b.WriteString("    <style type=\"text/css\"><![CDATA[\n")
			for _, face := range r.fontFaceOrder {
				b.WriteString("      @font-face { font-family: \"")
				b.WriteString(face.family)
				b.WriteString("\"; src: url(\"data:")
				b.WriteString(face.mime)
				b.WriteString(";base64,")
				b.WriteString(face.data)
				b.WriteString("\") format(\"")
				b.WriteString(face.format)
				b.WriteString("\"); }\n")
			}
			b.WriteString("    ]]></style>\n")
		}
		for _, clip := range r.clipOrder {
			b.WriteString("    <clipPath id=\"" + clip.id + "\" clipPathUnits=\"userSpaceOnUse\">")
			if clip.rect != nil {
				w := clip.rect.W()
				h := clip.rect.H()
				b.WriteString("<rect x=\"")
				b.WriteString(formatFloat(clip.rect.Min.X))
				b.WriteString(`" y="`)
				b.WriteString(formatFloat(clip.rect.Min.Y))
				b.WriteString(`" width="`)
				b.WriteString(formatFloat(w))
				b.WriteString(`" height="`)
				b.WriteString(formatFloat(h))
				b.WriteString(`" />`)
			} else {
				b.WriteString(`<path d="`)
				b.WriteString(clip.path)
				b.WriteString(`" />`)
			}
			b.WriteString("</clipPath>\n")
		}
		b.WriteString("  </defs>\n")
	}

	bgColor, bgAlpha := colorToStyle(r.background)
	b.WriteString("  <rect x=\"0\" y=\"0\" width=\"100%\" height=\"100%\" ")
	if bgAlpha <= 0 {
		b.WriteString(`fill="none" />`)
		b.WriteString("\n")
	} else {
		b.WriteString(`fill="`)
		b.WriteString(bgColor)
		b.WriteString(`"`)
		if bgAlpha < 1 {
			b.WriteString(` fill-opacity="`)
			b.WriteString(formatFloat(bgAlpha))
			b.WriteString(`"`)
		}
		b.WriteString(" />\n")
	}

	for _, node := range r.nodes {
		if len(node.clipIDs) == 0 {
			b.WriteString("  ")
			b.WriteString(node.content)
			b.WriteString("\n")
			continue
		}
		b.WriteString("  ")
		for _, id := range node.clipIDs {
			b.WriteString("<g clip-path=\"url(#")
			b.WriteString(id)
			b.WriteString(")\">")
		}
		b.WriteString(node.content)
		for range node.clipIDs {
			b.WriteString("</g>")
		}
		b.WriteString("\n")
	}

	b.WriteString("</svg>\n")
	return b.String()
}

func (r *Renderer) renderTextNode(text string, x, y, size float64, textColor render.Color, transform string) {
	if text == "" || size <= 0 {
		return
	}

	var content strings.Builder
	content.WriteString(`<text`)
	writeFloatAttr(&content, "x", x)
	writeFloatAttr(&content, "y", y)
	writeFloatAttr(&content, "font-size", size)
	writeAttr(&content, "font-family", r.svgFontFamily(r.lastFontKey))
	writeAttr(&content, "fill", colorToHex(textColor))
	alpha := clamp01(textColor.A)
	if alpha < 1 {
		writeFloatAttr(&content, "fill-opacity", alpha)
	}
	if transform != "" {
		writeAttr(&content, "transform", transform)
	}
	content.WriteString(">")
	content.WriteString(escapeText(text))
	content.WriteString("</text>")

	r.nodes = append(r.nodes, svgNode{
		content: content.String(),
		clipIDs: r.currentClipIDs(),
	})
}

// currentClipIDs returns the active clip-path chain in outer-to-inner order.
// Returns nil when no clip is active.
func (r *Renderer) currentClipIDs() []string {
	count := len(r.clipPathStack)
	if r.clipRect != nil {
		count++
	}
	if count == 0 {
		return nil
	}

	ids := make([]string, 0, count)
	if r.clipRect != nil {
		ids = append(ids, r.registerClip(*r.clipRect))
	}
	ids = append(ids, r.clipPathStack...)
	return ids
}

func (r *Renderer) registerClip(rect geom.Rect) string {
	key := clipKey(rect)
	if id, ok := r.clipDefs[key]; ok {
		return id
	}

	r.clipIDCounter++
	id := "clip" + strconv.Itoa(r.clipIDCounter)
	r.clipDefs[key] = id
	rectCopy := rect
	r.clipOrder = append(r.clipOrder, clipDef{id: id, rect: &rectCopy})
	return id
}

func (r *Renderer) registerPathClip(d string) string {
	if id, ok := r.clipPathDefs[d]; ok {
		return id
	}

	r.clipIDCounter++
	id := "clip" + strconv.Itoa(r.clipIDCounter)
	r.clipPathDefs[d] = id
	r.clipOrder = append(r.clipOrder, clipDef{id: id, path: d})
	return id
}

func clipKey(rect geom.Rect) string {
	q := normalizeRect(rect)
	return fmt.Sprintf("%s,%s,%s,%s",
		formatFloat(q.Min.X),
		formatFloat(q.Min.Y),
		formatFloat(q.Max.X),
		formatFloat(q.Max.Y),
	)
}

func encodeImage(img *image.RGBA) (string, error) {
	if img == nil {
		return "", errors.New("svg: image is nil")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func asRGBAImage(img render.Image) *image.RGBA {
	rgbaImage, ok := img.(interface {
		RGBA() *image.RGBA
	})
	if !ok {
		return nil
	}

	return rgbaImage.RGBA()
}

func colorizeTeXImage(src *image.RGBA, c render.Color) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	r := uint8(clamp01(c.R)*255 + 0.5)
	g := uint8(clamp01(c.G)*255 + 0.5)
	b := uint8(clamp01(c.B)*255 + 0.5)
	alphaScale := clamp01(c.A)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a16 := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			a := uint8(float64(a16>>8)*alphaScale + 0.5)
			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}

func buildPathData(p geom.Path) string {
	if len(p.C) == 0 {
		return ""
	}

	var b strings.Builder
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return ""
			}
			pt := quantizePt(p.V[vi])
			vi++
			b.WriteString("M ")
			b.WriteString(formatFloat(pt.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(pt.Y))
		case geom.LineTo:
			if vi >= len(p.V) {
				return ""
			}
			pt := quantizePt(p.V[vi])
			vi++
			b.WriteString(" L ")
			b.WriteString(formatFloat(pt.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(pt.Y))
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return ""
			}
			ctrl := quantizePt(p.V[vi])
			to := quantizePt(p.V[vi+1])
			vi += 2
			b.WriteString(" Q ")
			b.WriteString(formatFloat(ctrl.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(ctrl.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.Y))
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return ""
			}
			c1 := quantizePt(p.V[vi])
			c2 := quantizePt(p.V[vi+1])
			to := quantizePt(p.V[vi+2])
			vi += 3
			b.WriteString(" C ")
			b.WriteString(formatFloat(c1.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(c1.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(c2.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(c2.Y))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.X))
			b.WriteString(" ")
			b.WriteString(formatFloat(to.Y))
		case geom.ClosePath:
			b.WriteString(" Z")
		default:
			return ""
		}
	}

	d := b.String()
	return strings.TrimSpace(d)
}

func dashedArray(dashes []float64) string {
	if len(dashes) < 2 {
		return ""
	}

	var b strings.Builder
	for i := 0; i < len(dashes)-1; i += 2 {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(formatFloat(dashes[i]))
		b.WriteString(",")
		b.WriteString(formatFloat(dashes[i+1]))
	}

	return b.String()
}

func mapLineJoin(v render.LineJoin) string {
	switch v {
	case render.JoinRound:
		return "round"
	case render.JoinBevel:
		return "bevel"
	default:
		return "miter"
	}
}

func mapLineCap(v render.LineCap) string {
	switch v {
	case render.CapButt:
		return "butt"
	case render.CapRound:
		return "round"
	case render.CapSquare:
		return "square"
	default:
		return "butt"
	}
}

func colorToHex(c render.Color) string {
	return fmt.Sprintf("rgb(%d,%d,%d)",
		toByte(c.R),
		toByte(c.G),
		toByte(c.B),
	)
}

func colorToStyle(c render.Color) (string, float64) {
	return colorToHex(c), clamp01(c.A)
}

func toByte(v float64) uint8 {
	v = clamp01(v)
	return uint8(v*255 + 0.5)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func quantize(v float64) float64 {
	return math.Round(v/quantizationGrid) * quantizationGrid
}

func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{X: quantize(p.X), Y: quantize(p.Y)}
}

func normalizeRect(rect geom.Rect) geom.Rect {
	minX := rect.Min.X
	minY := rect.Min.Y
	maxX := rect.Max.X
	maxY := rect.Max.Y

	if maxX < minX {
		minX, maxX = maxX, minX
	}
	if maxY < minY {
		minY, maxY = maxY, minY
	}

	return geom.Rect{
		Min: geom.Pt{X: quantize(minX), Y: quantize(minY)},
		Max: geom.Pt{X: quantize(maxX), Y: quantize(maxY)},
	}
}

func writeAttr(b *strings.Builder, name, value string) {
	b.WriteString(" ")
	b.WriteString(name)
	b.WriteByte('=')
	b.WriteString(strconv.Quote(value))
}

func writeFloatAttr(b *strings.Builder, name string, value float64) {
	b.WriteString(" ")
	b.WriteString(name)
	b.WriteString("=\"")
	b.WriteString(formatFloat(value))
	b.WriteString("\"")
}

func formatFloat(v float64) string {
	return shortFloat(v)
}

// shortFloat formats v with up to 6 decimal digits, mirroring matplotlib's
// _short_float_fmt: trailing zeros (and a trailing decimal point) are stripped,
// negative zero is normalized to "0", and NaN/Inf are clamped to "0". The
// output stays in fixed (non-exponent) notation so SVG number attributes remain
// portable across viewers.
func shortFloat(v float64) string {
	s := strconv.FormatFloat(clampFloat(v), 'f', 6, 64)
	if i := strings.IndexByte(s, '.'); i >= 0 {
		end := len(s)
		for end > i && s[end-1] == '0' {
			end--
		}
		if end > 0 && s[end-1] == '.' {
			end--
		}
		s = s[:end]
	}
	if s == "-0" || s == "" {
		s = "0"
	}
	return s
}

// rotateTransform returns an SVG rotate() transform string with compact float
// formatting. Centralizing the format keeps later phases (matrix transforms)
// free to swap the implementation without touching call sites.
func rotateTransform(angleDeg, cx, cy float64) string {
	return "rotate(" + shortFloat(angleDeg) + " " + shortFloat(cx) + " " + shortFloat(cy) + ")"
}

func clampFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

func fontFamily(key string) string {
	return render.CSSFontFamily(key)
}

func (r *Renderer) svgFontFamily(key string) string {
	if family := r.registerFontFace(key); family != "" {
		return family
	}
	return fontFamily(key)
}

func (r *Renderer) registerFontFace(key string) string {
	path := strings.TrimSpace(key)
	if path == "" || !isFontFile(path) {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return ""
	}
	if r.fontFaces == nil {
		r.fontFaces = map[string]fontFaceDef{}
	}
	if face, ok := r.fontFaces[path]; ok {
		return face.family
	}
	face := fontFaceDef{
		family: "mplgo-font-" + strconv.Itoa(len(r.fontFaces)+1),
		data:   base64.StdEncoding.EncodeToString(data),
		mime:   fontMIME(path),
		format: fontFormat(path),
	}
	r.fontFaces[path] = face
	r.fontFaceOrder = append(r.fontFaceOrder, face)
	return face.family
}

func isFontFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ttf", ".otf", ".ttc", ".dfont":
		return true
	default:
		return false
	}
}

func fontMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf":
		return "font/otf"
	default:
		return "font/ttf"
	}
}

func fontFormat(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".otf":
		return "opentype"
	default:
		return "truetype"
	}
}

func escapeText(text string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(text))
	return b.String()
}
