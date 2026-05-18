package pdf

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	// defaultFontHeight matches the SVG backend default. It is the fallback
	// ascent the MeasureText path returns when no text drawer is wired up.
	defaultFontHeight = 13.0
)

// state captures the parts of renderer state that change with Save/Restore.
type state struct {
	// inContent reports whether the saved state lies inside a content stream.
	// Save before any draw call is permitted and behaves like the initial
	// identity transform.
	inContent bool
}

type pdfImage struct {
	name     string
	width    int
	height   int
	rgb      []byte
	alpha    []byte
	hasAlpha bool
	filter   string
}

type pdfImageObject struct {
	pdfImage
	objectID int
	smaskID  int
}

type pdfHatchPattern struct {
	name      string
	hatch     string
	faceColor render.Color
	lineColor render.Color
	lineWidth float64
	spacing   float64
}

type pdfHatchPatternObject struct {
	pdfHatchPattern
	objectID int
}

type pdfFormXObject struct {
	name     string
	path     geom.Path
	paintOp  string
	bbox     geom.Rect
	lineJoin render.LineJoin
	lineCap  render.LineCap
}

type pdfFormXObjectObject struct {
	pdfFormXObject
	objectID int
}

type pdfAlphaState struct {
	name        string
	strokeAlpha float64
	fillAlpha   float64
}

// Renderer implements render.Renderer by emitting a PDF document.
//
// The renderer buffers a single content stream. Calling End() finalizes the
// document into memory; SavePDF then flushes the in-memory PDF bytes to disk.
// The buffer is reusable: callers can Begin/End again to overwrite the
// previous document.
type Renderer struct {
	width      int
	height     int
	viewport   geom.Rect
	background render.Color
	resolution uint

	began bool
	stack []state

	// content is the page content stream under construction.
	content       bytes.Buffer
	images        []pdfImage
	imageIDs      map[string]string
	hatchPatterns []pdfHatchPattern
	hatchIDs      map[string]string
	forms         []pdfFormXObject
	formIDs       map[string]string
	alphaStates   []pdfAlphaState
	alphaIDs      map[string]string
	// document is the fully serialized PDF bytes ready for write.
	document []byte

	// pdfOpts carries setter-supplied options. SavePDFWithOptions overrides
	// fields directly for that single call.
	pdfOpts render.PDFOptions

	lastFontKey string
}

// Compile-time interface assertions.
var (
	_ render.Renderer               = (*Renderer)(nil)
	_ render.PNGExporter            = nil // explicitly not implemented
	_ render.PDFExporter            = (*Renderer)(nil)
	_ render.DPIAware               = (*Renderer)(nil)
	_ render.ImageTransformer       = (*Renderer)(nil)
	_ render.TextPather             = (*Renderer)(nil)
	_ render.FontTextDrawer         = (*Renderer)(nil)
	_ render.FontRotatedTextDrawer  = (*Renderer)(nil)
	_ render.FontVerticalTextDrawer = (*Renderer)(nil)
	_ render.NativeHatcher          = (*Renderer)(nil)
	_ render.MarkerDrawer           = (*Renderer)(nil)
	_ render.PathCollectionDrawer   = (*Renderer)(nil)
	_ render.PDFOptionExporter      = (*Renderer)(nil)
	_ render.PDFOptionSetter        = (*Renderer)(nil)
)

// New constructs a PDF renderer that produces a single-page document of the
// given width and height in points (1 point = 1/72 inch).
func New(width, height int, background render.Color) (*Renderer, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("pdf: invalid size %dx%d", width, height)
	}
	r := &Renderer{
		width:      width,
		height:     height,
		background: background,
		resolution: 72,
		pdfOpts:    render.DefaultPDFOptions(),
	}
	return r, nil
}

// SetResolution implements render.DPIAware. PDF coordinates are always in
// points, so the renderer keeps DPI as informational state for callers that
// want to scale rasterized output later.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi == 0 {
		dpi = 72
	}
	r.resolution = dpi
}

// SetPDFOptions implements render.PDFOptionSetter.
func (r *Renderer) SetPDFOptions(opts render.PDFOptions) {
	r.pdfOpts = opts
}

// Begin starts a drawing session for the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("pdf: Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.content.Reset()
	r.document = nil
	r.stack = r.stack[:0]
	r.images = r.images[:0]
	r.imageIDs = map[string]string{}
	r.hatchPatterns = r.hatchPatterns[:0]
	r.hatchIDs = map[string]string{}
	r.forms = r.forms[:0]
	r.formIDs = map[string]string{}
	r.alphaStates = r.alphaStates[:0]
	r.alphaIDs = map[string]string{}
	r.lastFontKey = ""

	// PDF's coordinate origin is bottom-left with +Y up. matplotlib-go uses a
	// top-left origin with +Y down. Flip Y once at the top of the content
	// stream so all subsequent draws can use the matplotlib coordinate frame
	// unchanged.
	fmt.Fprintf(&r.content, "1 0 0 -1 0 %s cm\n", shortFloat(float64(r.height)))

	if r.background.A > 0 {
		// Paint the page background as a filled rectangle in the now-flipped
		// frame.
		writeFillColor(&r.content, r.background)
		fmt.Fprintf(&r.content, "0 0 %s %s re f\n",
			shortFloat(float64(r.width)),
			shortFloat(float64(r.height)),
		)
	}
	return nil
}

// End finalizes the current drawing session.
func (r *Renderer) End() error {
	if !r.began {
		return errors.New("pdf: End called before Begin")
	}
	r.began = false
	doc, err := buildDocument(r.width, r.height, r.content.Bytes(), r.images, r.hatchPatterns, r.forms, r.alphaStates, r.pdfOpts)
	if err != nil {
		return err
	}
	r.document = doc
	return nil
}

// Save pushes graphics state.
func (r *Renderer) Save() {
	r.stack = append(r.stack, state{inContent: r.began})
	if r.began {
		r.content.WriteString("q\n")
	}
}

// Restore pops graphics state.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return
	}
	top := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	if top.inContent && r.began {
		r.content.WriteString("Q\n")
	}
}

// ClipRect installs a rectangular clip.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if !r.began {
		return
	}
	fmt.Fprintf(&r.content, "%s %s %s %s re W n\n",
		shortFloat(rect.Min.X),
		shortFloat(rect.Min.Y),
		shortFloat(rect.W()),
		shortFloat(rect.H()),
	)
}

// ClipPath installs an arbitrary path clip.
func (r *Renderer) ClipPath(p geom.Path) {
	if !r.began {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}
	r.content.WriteString("W n\n")
}

// SupportsNativeHatch reports that the PDF backend emits hatch fills as
// native PDF tiling pattern resources.
func (r *Renderer) SupportsNativeHatch() bool { return true }

// Path draws a path using the provided paint.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !r.began || paint == nil {
		return
	}
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	hasFill := paint.Fill.A > 0 || hasHatch
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}

	if hasFill {
		r.writeAlphaState(paint)
		if hasHatch {
			writePatternFill(&r.content, r.registerHatchPattern(*paint))
		} else {
			writeFillColor(&r.content, paint.Fill)
		}
	}
	if hasStroke {
		if !hasFill {
			r.writeAlphaState(paint)
		}
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
	}

	switch {
	case hasFill && hasStroke:
		r.content.WriteString("B\n")
	case hasFill:
		r.content.WriteString("f\n")
	case hasStroke:
		r.content.WriteString("S\n")
	}
}

func (r *Renderer) writeAlphaState(paint *render.Paint) {
	if paint == nil {
		return
	}
	strokeAlpha := 1.0
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		strokeAlpha = clamp01(paint.Stroke.A)
	}
	fillAlpha := 1.0
	if paint.Fill.A > 0 || (paint.Hatch != "" && paint.HatchColor.A > 0) {
		fillAlpha = clamp01(paint.Fill.A)
		if paint.Hatch != "" && paint.HatchColor.A > 0 {
			fillAlpha = clamp01(paint.HatchColor.A)
		}
	}
	if strokeAlpha >= 1 && fillAlpha >= 1 {
		return
	}
	name := r.registerAlphaState(strokeAlpha, fillAlpha)
	fmt.Fprintf(&r.content, "/%s gs\n", name)
}

func (r *Renderer) registerAlphaState(strokeAlpha, fillAlpha float64) string {
	if r.alphaIDs == nil {
		r.alphaIDs = map[string]string{}
	}
	strokeAlpha = clamp01(strokeAlpha)
	fillAlpha = clamp01(fillAlpha)
	key := shortFloat(strokeAlpha) + "\x00" + shortFloat(fillAlpha)
	if id, ok := r.alphaIDs[key]; ok {
		return id
	}
	id := fmt.Sprintf("A%d", len(r.alphaStates)+1)
	r.alphaIDs[key] = id
	r.alphaStates = append(r.alphaStates, pdfAlphaState{
		name:        id,
		strokeAlpha: strokeAlpha,
		fillAlpha:   fillAlpha,
	})
	return id
}

func (r *Renderer) registerHatchPattern(paint render.Paint) string {
	if r.hatchIDs == nil {
		r.hatchIDs = map[string]string{}
	}
	lineWidth := paint.HatchLineWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	spacing := paint.HatchSpacing
	if spacing <= 0 {
		spacing = 8
	}
	key := hatchPatternKey(paint.Hatch, paint.Fill, paint.HatchColor, lineWidth, spacing)
	if id, ok := r.hatchIDs[key]; ok {
		return id
	}
	id := fmt.Sprintf("Pa%d", len(r.hatchPatterns)+1)
	r.hatchIDs[key] = id
	r.hatchPatterns = append(r.hatchPatterns, pdfHatchPattern{
		name:      id,
		hatch:     paint.Hatch,
		faceColor: paint.Fill,
		lineColor: paint.HatchColor,
		lineWidth: lineWidth,
		spacing:   spacing,
	})
	return id
}

// DrawMarkers renders one marker path at many display-space offsets using a
// reusable Form XObject for the marker geometry.
func (r *Renderer) DrawMarkers(batch render.MarkerBatch) bool {
	if !r.began || len(batch.Marker.C) == 0 || len(batch.Items) == 0 || !batch.Marker.Validate() {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		marker := affinePath(batch.Marker, item.Transform)
		if !marker.Validate() || len(marker.C) == 0 {
			continue
		}
		paint := item.Paint
		paintOp := paintOperator(&paint)
		if paintOp == "" {
			continue
		}
		name := r.registerFormXObject("M", marker, paintOp, &paint)
		r.writePaintState(&paint)
		fmt.Fprintf(&r.content, "q\n1 0 0 1 %s %s cm\n/%s Do\nQ\n",
			shortFloat(item.Offset.X),
			shortFloat(item.Offset.Y),
			name,
		)
		emitted = true
	}
	return emitted
}

// DrawPathCollection renders display-space paths through Form XObject
// templates with per-item paint state applied at invocation time.
func (r *Renderer) DrawPathCollection(batch render.PathCollectionBatch) bool {
	if !r.began || len(batch.Items) == 0 {
		return false
	}
	emitted := false
	for i := range batch.Items {
		item := batch.Items[i]
		if !item.Path.Validate() || len(item.Path.C) == 0 {
			continue
		}
		paint := item.Paint
		if item.Hatch != "" {
			paint.Hatch = item.Hatch
			paint.HatchColor = item.HatchColor
			paint.HatchLineWidth = item.HatchWidth
			paint.HatchSpacing = item.HatchSpacing
		}
		paintOp := paintOperator(&paint)
		if paintOp == "" {
			continue
		}
		name := r.registerFormXObject("P", item.Path, paintOp, &paint)
		r.writePaintState(&paint)
		fmt.Fprintf(&r.content, "/%s Do\n", name)
		emitted = true
	}
	return emitted
}

func (r *Renderer) registerFormXObject(prefix string, path geom.Path, paintOp string, paint *render.Paint) string {
	if r.formIDs == nil {
		r.formIDs = map[string]string{}
	}
	key := formXObjectKey(prefix, path, paintOp, paint)
	if id, ok := r.formIDs[key]; ok {
		return id
	}
	name := fmt.Sprintf("%s%d", prefix, len(r.forms)+1)
	bbox, ok := pathBounds(path)
	if !ok {
		bbox = geom.Rect{}
	}
	padding := formPadding(paint)
	bbox = bbox.Inflate(padding, padding)
	r.formIDs[key] = name
	r.forms = append(r.forms, pdfFormXObject{
		name:     name,
		path:     clonePath(path),
		paintOp:  paintOp,
		bbox:     bbox,
		lineJoin: paint.LineJoin,
		lineCap:  paint.LineCap,
	})
	return name
}

// Image draws a raster image into the destination rectangle as a PDF image
// XObject. RGBA images with alpha get a grayscale soft mask.
func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	matrix := geom.Affine{A: dst.W(), D: dst.H(), E: dst.Min.X, F: dst.Min.Y}
	r.drawImageWithMatrix(img, matrix)
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates; PDF image
// XObjects paint a unit square, so the current transformation matrix includes
// the source image dimensions.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if !r.began || img == nil {
		return
	}
	w, h := img.Size()
	if w <= 0 || h <= 0 {
		return
	}
	matrix := geom.Affine{
		A: transform.A * float64(w),
		B: transform.B * float64(w),
		C: transform.C * float64(h),
		D: transform.D * float64(h),
		E: transform.E,
		F: transform.F,
	}
	r.drawImageWithMatrix(img, matrix)
}

func (r *Renderer) drawImageWithMatrix(img render.Image, matrix geom.Affine) {
	if jpegSource, ok := img.(render.JPEGImage); ok {
		pdfImg, ok := encodePDFJPEGImage("", jpegSource)
		if !ok {
			return
		}
		name := r.registerImage(pdfImg)
		r.writeImageInvocation(matrix, name)
		return
	}
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return
	}
	pdfImg, ok := encodePDFImage(
		"",
		rgbaSource.RGBA(),
		imageAlphaMultiplier(img),
	)
	if !ok {
		return
	}
	name := r.registerImage(pdfImg)
	r.writeImageInvocation(matrix, name)
}

func (r *Renderer) writeImageInvocation(matrix geom.Affine, name string) {
	fmt.Fprintf(&r.content, "q\n%s %s %s %s %s %s cm\n/%s Do\nQ\n",
		shortFloat(matrix.A),
		shortFloat(matrix.B),
		shortFloat(matrix.C),
		shortFloat(matrix.D),
		shortFloat(matrix.E),
		shortFloat(matrix.F),
		name,
	)
}

func (r *Renderer) registerImage(img pdfImage) string {
	if r.imageIDs == nil {
		r.imageIDs = map[string]string{}
	}
	key := imageKey(img)
	if name, ok := r.imageIDs[key]; ok {
		return name
	}
	img.name = fmt.Sprintf("Im%d", len(r.images)+1)
	r.imageIDs[key] = img.name
	r.images = append(r.images, img)
	return img.name
}

// GlyphRun draws shaped glyphs as filled outlines. GlyphRun only carries glyph
// IDs, so this remains a practical fallback for simple code-point-shaped runs
// until embedded font subsetting lands.
func (r *Renderer) GlyphRun(run render.GlyphRun, textColor render.Color) {
	if len(run.Glyphs) == 0 {
		return
	}
	if run.FontKey != "" {
		r.lastFontKey = run.FontKey
	}
	size := run.Size
	if size <= 0 {
		size = 12
	}
	penX := run.Origin.X
	penY := run.Origin.Y
	for _, glyph := range run.Glyphs {
		if glyph.ID != 0 {
			text := string(rune(glyph.ID))
			origin := geom.Pt{X: penX + glyph.Offset.X, Y: penY + glyph.Offset.Y}
			r.DrawTextWithFont(text, origin, size, textColor, r.lastFontKey)
		}
		advance := glyph.Advance
		if advance <= 0 && glyph.ID != 0 {
			advance = r.MeasureText(string(rune(glyph.ID)), size, r.lastFontKey).W
		}
		penX += advance
	}
}

// MeasureText returns rough metrics so layout code does not divide by zero.
// A future revision will plumb in the shared font manager.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" {
		return render.TextMetrics{}
	}
	if fontKey != "" {
		r.lastFontKey = fontKey
	}
	if size <= 0 {
		size = defaultFontHeight
	}
	// Crude width estimate; consistent across backends that lack a real font
	// shaper. Refined once the shared font pipeline is wired up.
	width := size * 0.5 * float64(len(text))
	return render.TextMetrics{
		W:       width,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

// TextPath converts text to vector glyph outlines through the shared font
// manager. This mirrors Matplotlib's text-as-path PDF mode until Type 0 font
// subsetting is implemented.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}
	return render.TextPath(text, origin, size, fontKey)
}

// DrawText renders text as filled glyph paths using the most recently resolved
// font key when one has been primed by MeasureText or DrawTextWithFont.
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	r.DrawTextWithFont(text, origin, size, textColor, r.lastFontKey)
}

// DrawTextWithFont renders text as filled glyph paths using an explicit font.
func (r *Renderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	path, ok := r.TextPath(text, origin, size, fontKey)
	if !ok {
		return
	}
	r.Path(path, &render.Paint{Fill: textColor})
}

// DrawTextRotated renders text as outlines around Matplotlib's bottom-center
// rotated-text anchor.
func (r *Renderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	r.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, r.lastFontKey)
}

// DrawTextRotatedWithFont renders rotated text as filled glyph paths.
func (r *Renderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 || math.IsNaN(angle) || math.IsInf(angle, 0) {
		return
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 || metrics.H <= 0 {
		return
	}
	origin := geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
	path, ok := r.TextPath(text, origin, size, fontKey)
	if !ok {
		return
	}
	path = affinePath(path, rotationAffine(angle, anchor))
	r.Path(path, &render.Paint{Fill: textColor})
}

// DrawTextVertical renders one character per line as filled glyph paths.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	r.DrawTextVerticalWithFont(text, center, size, textColor, r.lastFontKey)
}

// DrawTextVerticalWithFont renders vertical text with an explicit font key.
func (r *Renderer) DrawTextVerticalWithFont(text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if !r.began || text == "" || size <= 0 || textColor.A <= 0 {
		return
	}
	runes := []rune(text)
	lineMetrics := r.MeasureText("M", size, fontKey)
	lineH := lineMetrics.H
	if lineH <= 0 {
		lineH = size
	}
	totalH := lineH * float64(len(runes))
	y := center.Y - totalH/2 + lineMetrics.Ascent
	for _, ch := range runes {
		s := string(ch)
		chMetrics := r.MeasureText(s, size, fontKey)
		if chMetrics.W <= 0 || chMetrics.H <= 0 {
			y += lineH
			continue
		}
		origin := geom.Pt{X: center.X - chMetrics.W/2, Y: y}
		r.DrawTextWithFont(s, origin, size, textColor, fontKey)
		y += lineH
	}
}

// SavePDF writes the buffered document to path.
func (r *Renderer) SavePDF(path string) error {
	if len(r.document) == 0 {
		return errors.New("pdf: SavePDF called before End")
	}
	return os.WriteFile(path, r.document, 0o644)
}

// SavePDFWithOptions writes the buffered document to path using opts to
// override any setter-supplied options for this single call.
func (r *Renderer) SavePDFWithOptions(path string, opts render.PDFOptions) error {
	if !r.began && len(r.document) == 0 {
		return errors.New("pdf: SavePDFWithOptions called before End")
	}
	doc, err := buildDocument(r.width, r.height, r.content.Bytes(), r.images, r.hatchPatterns, r.forms, r.alphaStates, opts)
	if err != nil {
		return err
	}
	return os.WriteFile(path, doc, 0o644)
}

// Bytes returns the serialized PDF document. Returns an error if End has not
// been called since the last Begin.
func (r *Renderer) Bytes() ([]byte, error) {
	if len(r.document) == 0 {
		return nil, errors.New("pdf: document is empty; call End first")
	}
	return r.document, nil
}

// writePathOps emits PDF path-construction operators for path p. Returns
// false if the path is empty or invalid.
func writePathOps(w *bytes.Buffer, p geom.Path) bool {
	if !p.Validate() || len(p.C) == 0 {
		return false
	}
	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s m\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.LineTo:
			pt := p.V[vi]
			vi++
			fmt.Fprintf(w, "%s %s l\n", shortFloat(pt.X), shortFloat(pt.Y))
		case geom.QuadTo:
			// PDF has no quadratic curve operator; promote to cubic.
			// The previous endpoint must exist; if it does not, skip.
			if vi == 0 {
				vi += 2
				continue
			}
			prev := lastEndpoint(p, vi)
			ctrl := p.V[vi]
			end := p.V[vi+1]
			vi += 2
			c1 := geom.Pt{
				X: prev.X + (2.0/3.0)*(ctrl.X-prev.X),
				Y: prev.Y + (2.0/3.0)*(ctrl.Y-prev.Y),
			}
			c2 := geom.Pt{
				X: end.X + (2.0/3.0)*(ctrl.X-end.X),
				Y: end.Y + (2.0/3.0)*(ctrl.Y-end.Y),
			}
			fmt.Fprintf(w, "%s %s %s %s %s %s c\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.CubicTo:
			c1 := p.V[vi]
			c2 := p.V[vi+1]
			end := p.V[vi+2]
			vi += 3
			fmt.Fprintf(w, "%s %s %s %s %s %s c\n",
				shortFloat(c1.X), shortFloat(c1.Y),
				shortFloat(c2.X), shortFloat(c2.Y),
				shortFloat(end.X), shortFloat(end.Y),
			)
		case geom.ClosePath:
			w.WriteString("h\n")
		}
	}
	return true
}

// lastEndpoint returns the endpoint emitted by the command immediately before
// vi. Used to promote quadratic curves to cubics.
func lastEndpoint(p geom.Path, vi int) geom.Pt {
	// Walk commands counting vertices to find the verb that ends at vi-1.
	consumed := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo, geom.LineTo:
			consumed++
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.QuadTo:
			consumed += 2
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.CubicTo:
			consumed += 3
			if consumed == vi {
				return p.V[consumed-1]
			}
		case geom.ClosePath:
			// No vertices.
		}
	}
	if vi > 0 && vi-1 < len(p.V) {
		return p.V[vi-1]
	}
	return geom.Pt{}
}

func writeFillColor(w *bytes.Buffer, c render.Color) {
	// Alpha requires an ExtGState entry; deferred to a later iteration. The
	// document stays opaque until graphics-state dictionaries are wired up.
	fmt.Fprintf(w, "%s %s %s rg\n",
		shortFloat(clamp01(c.R)),
		shortFloat(clamp01(c.G)),
		shortFloat(clamp01(c.B)),
	)
}

func writeStrokeColor(w *bytes.Buffer, c render.Color) {
	fmt.Fprintf(w, "%s %s %s RG\n",
		shortFloat(clamp01(c.R)),
		shortFloat(clamp01(c.G)),
		shortFloat(clamp01(c.B)),
	)
}

func writePatternFill(w *bytes.Buffer, name string) {
	fmt.Fprintf(w, "/Pattern cs\n/%s scn\n", escapeName(name))
}

func (r *Renderer) writePaintState(paint *render.Paint) {
	if paint == nil {
		return
	}
	r.writeAlphaState(paint)
	hasHatch := paint.Hatch != "" && paint.HatchColor.A > 0
	if hasHatch {
		writePatternFill(&r.content, r.registerHatchPattern(*paint))
	} else if paint.Fill.A > 0 {
		writeFillColor(&r.content, paint.Fill)
	}
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		writeStrokeColor(&r.content, paint.Stroke)
		writeLineState(&r.content, paint)
	}
}

func writeLineState(w *bytes.Buffer, paint *render.Paint) {
	if paint.LineWidth > 0 {
		fmt.Fprintf(w, "%s w\n", shortFloat(paint.LineWidth))
	}
	switch paint.LineCap {
	case render.CapButt:
		w.WriteString("0 J\n")
	case render.CapRound:
		w.WriteString("1 J\n")
	case render.CapSquare:
		w.WriteString("2 J\n")
	}
	switch paint.LineJoin {
	case render.JoinMiter:
		w.WriteString("0 j\n")
	case render.JoinRound:
		w.WriteString("1 j\n")
	case render.JoinBevel:
		w.WriteString("2 j\n")
	}
	if paint.MiterLimit > 0 {
		fmt.Fprintf(w, "%s M\n", shortFloat(paint.MiterLimit))
	}
	if len(paint.Dashes) > 0 {
		w.WriteString("[")
		for i, d := range paint.Dashes {
			if i > 0 {
				w.WriteString(" ")
			}
			w.WriteString(shortFloat(d))
		}
		w.WriteString("] 0 d\n")
	} else {
		w.WriteString("[] 0 d\n")
	}
}

// shortFloat mirrors the SVG backend's compact float formatter: up to six
// decimals, trailing zeros stripped, -0 normalized, and NaN/Inf clamped to 0.
func shortFloat(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	if v == 0 {
		return "0"
	}
	s := strconv.FormatFloat(v, 'f', 6, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-0" {
		return "0"
	}
	return s
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

func rotationAffine(angle float64, pivot geom.Pt) geom.Affine {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return translateAffine(pivot).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
}

func translateAffine(p geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: p.X, F: p.Y}
}

func affinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.V) == 0 {
		return path
	}
	out := geom.Path{
		V: make([]geom.Pt, len(path.V)),
		C: append([]geom.Cmd(nil), path.C...),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}

func clonePath(path geom.Path) geom.Path {
	return geom.Path{
		V: append([]geom.Pt(nil), path.V...),
		C: append([]geom.Cmd(nil), path.C...),
	}
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	minX, maxX := path.V[0].X, path.V[0].X
	minY, maxY := path.V[0].Y, path.V[0].Y
	for _, pt := range path.V[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}, true
}

func paintOperator(paint *render.Paint) string {
	if paint == nil {
		return ""
	}
	hasFill := paint.Fill.A > 0 || (paint.Hatch != "" && paint.HatchColor.A > 0)
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	switch {
	case hasFill && hasStroke:
		return "B"
	case hasFill:
		return "f"
	case hasStroke:
		return "S"
	default:
		return ""
	}
}

func formPadding(paint *render.Paint) float64 {
	if paint == nil || paint.Stroke.A <= 0 || paint.LineWidth <= 0 {
		return 0
	}
	padding := paint.LineWidth / 2
	if paint.LineJoin == render.JoinMiter {
		miter := paint.MiterLimit
		if miter <= 0 {
			miter = 10
		}
		padding = math.Max(padding, paint.LineWidth*miter/2)
	}
	return padding
}

func formXObjectKey(prefix string, path geom.Path, paintOp string, paint *render.Paint) string {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('\x00')
	b.WriteString(paintOp)
	b.WriteByte('\x00')
	if paint != nil {
		b.WriteString(strconv.Itoa(int(paint.LineJoin)))
		b.WriteByte('\x00')
		b.WriteString(strconv.Itoa(int(paint.LineCap)))
	}
	for _, cmd := range path.C {
		b.WriteByte(byte(cmd))
	}
	b.WriteByte('\x00')
	for _, pt := range path.V {
		b.WriteString(shortFloat(pt.X))
		b.WriteByte(',')
		b.WriteString(shortFloat(pt.Y))
		b.WriteByte(';')
	}
	return b.String()
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}

func encodePDFImage(name string, src *image.RGBA, alphaMul float64) (pdfImage, bool) {
	if src == nil {
		return pdfImage{}, false
	}
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return pdfImage{}, false
	}
	rgb := make([]byte, 0, width*height*3)
	alpha := make([]byte, 0, width*height)
	hasAlpha := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)
			a := uint8(float64(c.A)*alphaMul + 0.5)
			rgb = append(rgb, c.R, c.G, c.B)
			alpha = append(alpha, a)
			if a != 0xff {
				hasAlpha = true
			}
		}
	}
	return pdfImage{
		name:     name,
		width:    width,
		height:   height,
		rgb:      rgb,
		alpha:    alpha,
		hasAlpha: hasAlpha,
		filter:   "FlateDecode",
	}, true
}

func encodePDFJPEGImage(name string, src render.JPEGImage) (pdfImage, bool) {
	if src == nil {
		return pdfImage{}, false
	}
	width, height := src.Size()
	data := src.JPEGData()
	if width <= 0 || height <= 0 || len(data) == 0 {
		return pdfImage{}, false
	}
	return pdfImage{
		name:   name,
		width:  width,
		height: height,
		rgb:    append([]byte(nil), data...),
		filter: "DCTDecode",
	}, true
}

func imageKey(img pdfImage) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(img.width))
	b.WriteByte('x')
	b.WriteString(strconv.Itoa(img.height))
	b.WriteByte('\x00')
	if img.hasAlpha {
		b.WriteByte('a')
	} else {
		b.WriteByte('o')
	}
	b.WriteByte('\x00')
	b.WriteString(img.filter)
	b.WriteByte('\x00')
	b.WriteString(string(img.rgb))
	b.WriteByte('\x00')
	if img.hasAlpha {
		b.WriteString(string(img.alpha))
	}
	return b.String()
}

func hatchPatternKey(hatch string, face, line render.Color, lineWidth, spacing float64) string {
	return strings.Join([]string{
		hatch,
		shortFloat(face.R),
		shortFloat(face.G),
		shortFloat(face.B),
		shortFloat(face.A),
		shortFloat(line.R),
		shortFloat(line.G),
		shortFloat(line.B),
		shortFloat(line.A),
		shortFloat(lineWidth),
		shortFloat(spacing),
	}, "\x00")
}

func hatchPatternStream(pattern pdfHatchPattern) []byte {
	const side = 72.0
	var buf bytes.Buffer
	if pattern.faceColor.A > 0 {
		writeFillColor(&buf, pattern.faceColor)
		fmt.Fprintf(&buf, "0 0 %s %s re f\n", shortFloat(side), shortFloat(side))
	}
	writeStrokeColor(&buf, pattern.lineColor)
	fmt.Fprintf(&buf, "%s w\n", shortFloat(pattern.lineWidth))
	buf.WriteString("0 J\n")
	for _, line := range hatchPatternLines(pattern.hatch, pattern.spacing) {
		fmt.Fprintf(&buf, "%s %s m\n%s %s l\nS\n",
			shortFloat(line[0].X), shortFloat(line[0].Y),
			shortFloat(line[1].X), shortFloat(line[1].Y),
		)
	}
	return buf.Bytes()
}

func hatchPatternLines(hatch string, spacing float64) [][2]geom.Pt {
	if spacing <= 0 {
		spacing = 8
	}
	lines := make([][2]geom.Pt, 0)
	writeHatchLines := func(count int, draw func(float64)) {
		if count <= 0 {
			return
		}
		step := math.Max(2, spacing/float64(count))
		for v := -72.0; v <= 144; v += step {
			draw(v)
		}
	}
	add := func(x1, y1, x2, y2 float64) {
		lines = append(lines, [2]geom.Pt{{X: x1, Y: y1}, {X: x2, Y: y2}})
	}
	verticalCount := strings.Count(hatch, "|") + strings.Count(hatch, "+")
	horizontalCount := strings.Count(hatch, "-") + strings.Count(hatch, "+")
	slashCount := strings.Count(hatch, "/") + strings.Count(hatch, "x") + strings.Count(hatch, "X")
	backslashCount := strings.Count(hatch, `\`) + strings.Count(hatch, "x") + strings.Count(hatch, "X")

	writeHatchLines(verticalCount, func(x float64) { add(x, 0, x, 72) })
	writeHatchLines(horizontalCount, func(y float64) { add(0, y, 72, y) })
	writeHatchLines(slashCount, func(x float64) { add(x, 72, x+72, 0) })
	writeHatchLines(backslashCount, func(x float64) { add(x, 0, x+72, 72) })
	return lines
}

// --- PDF document assembly ---------------------------------------------------

// buildDocument assembles the PDF bytes for one page given the encoded
// content stream.
func buildDocument(width, height int, contentStream []byte, images []pdfImage, hatches []pdfHatchPattern, forms []pdfFormXObject, alphaStates []pdfAlphaState, opts render.PDFOptions) ([]byte, error) {
	imageObjects := assignImageObjects(images, 6)
	hatchObjects := assignHatchObjects(hatches, nextImageObjectID(imageObjects, 6))
	formObjects := assignFormObjects(forms, nextHatchObjectID(hatchObjects, nextImageObjectID(imageObjects, 6)))
	// We emit five fixed indirect objects, followed by image XObjects:
	//   1: /Catalog
	//   2: /Pages
	//   3: /Page
	//   4: content stream
	//   5: /Info (always present so the trailer can reference it)
	w := newPDFWriter()
	w.header()

	w.beginObject(1)
	w.writeString("<< /Type /Catalog /Pages 2 0 R >>")
	w.endObject()

	w.beginObject(2)
	w.writeString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	w.endObject()

	w.beginObject(3)
	fmt.Fprintf(&w.buf, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] /Contents 4 0 R /Resources %s >>",
		width, height, pageResources(imageObjects, hatchObjects, formObjects, alphaStates))
	w.endObject()

	// Compress the content stream with FlateDecode for determinism and size.
	encoded, err := flateEncode(contentStream)
	if err != nil {
		return nil, fmt.Errorf("pdf: flate encode content stream: %w", err)
	}
	w.beginObject(4)
	fmt.Fprintf(&w.buf, "<< /Length %d /Filter /FlateDecode >>\nstream\n", len(encoded))
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()

	w.beginObject(5)
	w.writeInfo(opts)
	w.endObject()

	for _, img := range imageObjects {
		if err := w.writeImageObject(img); err != nil {
			return nil, err
		}
		if img.smaskID != 0 {
			if err := w.writeSoftMaskObject(img); err != nil {
				return nil, err
			}
		}
	}
	for _, hatch := range hatchObjects {
		if err := w.writeHatchPatternObject(hatch); err != nil {
			return nil, err
		}
	}
	for _, form := range formObjects {
		if err := w.writeFormXObject(form); err != nil {
			return nil, err
		}
	}

	xrefOffset := w.buf.Len()
	w.writeXRef()
	w.writeTrailer(xrefOffset)

	return w.buf.Bytes(), nil
}

// pdfWriter helps assemble a PDF document with deterministic xref offsets.
type pdfWriter struct {
	buf     bytes.Buffer
	offsets []int // offsets[i] is the byte offset of object i (1-indexed; offsets[0] is unused)
}

func newPDFWriter() *pdfWriter {
	return &pdfWriter{offsets: []int{0}}
}

func (w *pdfWriter) header() {
	// PDF-1.7 plus a 4-byte binary marker to satisfy PDF readers that look for
	// non-ASCII bytes in the header comment.
	w.buf.WriteString("%PDF-1.7\n")
	w.buf.WriteString("%\xE2\xE3\xCF\xD3\n")
}

func (w *pdfWriter) beginObject(id int) {
	for len(w.offsets) <= id {
		w.offsets = append(w.offsets, 0)
	}
	w.offsets[id] = w.buf.Len()
	fmt.Fprintf(&w.buf, "%d 0 obj\n", id)
}

func (w *pdfWriter) endObject() {
	w.buf.WriteString("\nendobj\n")
}

func (w *pdfWriter) writeString(s string) {
	w.buf.WriteString(s)
}

func (w *pdfWriter) writeInfo(opts render.PDFOptions) {
	w.buf.WriteString("<< /Producer (matplotlib-go)")
	if len(opts.Metadata) > 0 {
		// Sort keys for deterministic order.
		keys := sortedKeys(opts.Metadata)
		for _, k := range keys {
			fmt.Fprintf(&w.buf, " /%s %s", escapeName(k), pdfLiteralString(opts.Metadata[k]))
		}
	}
	if date := resolveCreationDate(opts.CreationDate); !date.IsZero() {
		fmt.Fprintf(&w.buf, " /CreationDate %s", pdfDateString(date))
	}
	w.buf.WriteString(" >>")
}

func (w *pdfWriter) writeImageObject(img pdfImageObject) error {
	filter := img.filter
	if filter == "" {
		filter = "FlateDecode"
	}
	encoded := img.rgb
	if filter == "FlateDecode" {
		var err error
		encoded, err = flateEncode(img.rgb)
		if err != nil {
			return fmt.Errorf("pdf: flate encode image %s: %w", img.name, err)
		}
	}
	w.beginObject(img.objectID)
	fmt.Fprintf(&w.buf,
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8",
		img.width, img.height,
	)
	if img.smaskID != 0 {
		fmt.Fprintf(&w.buf, " /SMask %d 0 R", img.smaskID)
	}
	fmt.Fprintf(&w.buf, " /Length %d /Filter /%s >>\nstream\n", len(encoded), escapeName(filter))
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeSoftMaskObject(img pdfImageObject) error {
	encoded, err := flateEncode(img.alpha)
	if err != nil {
		return fmt.Errorf("pdf: flate encode image soft mask %s: %w", img.name, err)
	}
	w.beginObject(img.smaskID)
	fmt.Fprintf(&w.buf,
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceGray /BitsPerComponent 8 /Length %d /Filter /FlateDecode >>\nstream\n",
		img.width, img.height, len(encoded),
	)
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeHatchPatternObject(hatch pdfHatchPatternObject) error {
	stream := hatchPatternStream(hatch.pdfHatchPattern)
	encoded, err := flateEncode(stream)
	if err != nil {
		return fmt.Errorf("pdf: flate encode hatch pattern %s: %w", hatch.name, err)
	}
	w.beginObject(hatch.objectID)
	fmt.Fprintf(&w.buf,
		"<< /Type /Pattern /PatternType 1 /PaintType 1 /TilingType 1 /BBox [0 0 72 72] /XStep 72 /YStep 72 /Resources << >> /Length %d /Filter /FlateDecode >>\nstream\n",
		len(encoded),
	)
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeFormXObject(form pdfFormXObjectObject) error {
	var stream bytes.Buffer
	formPaint := render.Paint{
		LineJoin: form.lineJoin,
		LineCap:  form.lineCap,
	}
	writeLineState(&stream, &formPaint)
	if !writePathOps(&stream, form.path) {
		return nil
	}
	stream.WriteString(form.paintOp)
	stream.WriteByte('\n')
	encoded, err := flateEncode(stream.Bytes())
	if err != nil {
		return fmt.Errorf("pdf: flate encode form %s: %w", form.name, err)
	}
	w.beginObject(form.objectID)
	fmt.Fprintf(&w.buf,
		"<< /Type /XObject /Subtype /Form /BBox [%s %s %s %s] /Resources << >> /Length %d /Filter /FlateDecode >>\nstream\n",
		shortFloat(form.bbox.Min.X),
		shortFloat(form.bbox.Min.Y),
		shortFloat(form.bbox.Max.X),
		shortFloat(form.bbox.Max.Y),
		len(encoded),
	)
	w.buf.Write(encoded)
	w.writeString("\nendstream")
	w.endObject()
	return nil
}

func (w *pdfWriter) writeXRef() {
	w.buf.WriteString("xref\n")
	fmt.Fprintf(&w.buf, "0 %d\n", len(w.offsets))
	// Object 0 is the head of the free list.
	w.buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(w.offsets); i++ {
		fmt.Fprintf(&w.buf, "%010d 00000 n \n", w.offsets[i])
	}
}

func (w *pdfWriter) writeTrailer(xrefOffset int) {
	fmt.Fprintf(&w.buf,
		"trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(w.offsets), xrefOffset,
	)
}

func flateEncode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, zlib.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// stdlib sort is in the import list of registry but not here; use the
	// minimal in-place insertion sort because the key set is tiny.
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j-1] > keys[j] {
			keys[j-1], keys[j] = keys[j], keys[j-1]
			j--
		}
	}
	return keys
}

func assignImageObjects(images []pdfImage, firstID int) []pdfImageObject {
	if len(images) == 0 {
		return nil
	}
	out := make([]pdfImageObject, len(images))
	nextID := firstID
	for i, img := range images {
		out[i] = pdfImageObject{
			pdfImage: img,
			objectID: nextID,
		}
		nextID++
		if img.hasAlpha {
			out[i].smaskID = nextID
			nextID++
		}
	}
	return out
}

func nextImageObjectID(images []pdfImageObject, firstID int) int {
	nextID := firstID
	for _, img := range images {
		if img.objectID >= nextID {
			nextID = img.objectID + 1
		}
		if img.smaskID >= nextID {
			nextID = img.smaskID + 1
		}
	}
	return nextID
}

func assignHatchObjects(hatches []pdfHatchPattern, firstID int) []pdfHatchPatternObject {
	if len(hatches) == 0 {
		return nil
	}
	out := make([]pdfHatchPatternObject, len(hatches))
	for i, hatch := range hatches {
		out[i] = pdfHatchPatternObject{
			pdfHatchPattern: hatch,
			objectID:        firstID + i,
		}
	}
	return out
}

func nextHatchObjectID(hatches []pdfHatchPatternObject, firstID int) int {
	nextID := firstID
	for _, hatch := range hatches {
		if hatch.objectID >= nextID {
			nextID = hatch.objectID + 1
		}
	}
	return nextID
}

func assignFormObjects(forms []pdfFormXObject, firstID int) []pdfFormXObjectObject {
	if len(forms) == 0 {
		return nil
	}
	out := make([]pdfFormXObjectObject, len(forms))
	for i, form := range forms {
		out[i] = pdfFormXObjectObject{
			pdfFormXObject: form,
			objectID:       firstID + i,
		}
	}
	return out
}

func pageResources(images []pdfImageObject, hatches []pdfHatchPatternObject, forms []pdfFormXObjectObject, alphaStates []pdfAlphaState) string {
	if len(images) == 0 && len(hatches) == 0 && len(forms) == 0 && len(alphaStates) == 0 {
		return "<< >>"
	}
	var b strings.Builder
	b.WriteString("<<")
	if len(images) > 0 || len(forms) > 0 {
		b.WriteString(" /XObject <<")
		for _, img := range images {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(img.name), img.objectID)
		}
		for _, form := range forms {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(form.name), form.objectID)
		}
		b.WriteString(" >>")
	}
	if len(hatches) > 0 {
		b.WriteString(" /Pattern <<")
		for _, hatch := range hatches {
			fmt.Fprintf(&b, " /%s %d 0 R", escapeName(hatch.name), hatch.objectID)
		}
		b.WriteString(" >>")
	}
	if len(alphaStates) > 0 {
		b.WriteString(" /ExtGState <<")
		for _, state := range alphaStates {
			fmt.Fprintf(&b, " /%s << /Type /ExtGState /CA %s /ca %s >>",
				escapeName(state.name),
				shortFloat(state.strokeAlpha),
				shortFloat(state.fillAlpha),
			)
		}
		b.WriteString(" >>")
	}
	b.WriteString(" >>")
	return b.String()
}

// pdfLiteralString encodes s as a PDF literal string. It escapes parentheses
// and backslashes per ISO 32000-1 §7.3.4.2.
func pdfLiteralString(s string) string {
	var b strings.Builder
	b.WriteByte('(')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(', ')', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte(')')
	return b.String()
}

// escapeName escapes a PDF name token per ISO 32000-1 §7.3.5. Only safe-ASCII
// alphanumeric characters and a handful of punctuation are emitted verbatim;
// everything else is hex-escaped.
func escapeName(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			c == '.' || c == '-' || c == '_' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "#%02X", c)
		}
	}
	return b.String()
}

// resolveCreationDate returns the explicit override when set, otherwise the
// SOURCE_DATE_EPOCH environment value, otherwise a zero time (which suppresses
// the /CreationDate entry for full reproducibility).
func resolveCreationDate(explicit time.Time) time.Time {
	if !explicit.IsZero() {
		return explicit
	}
	if v := os.Getenv("SOURCE_DATE_EPOCH"); v != "" {
		if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(secs, 0).UTC()
		}
	}
	return time.Time{}
}

// pdfDateString formats t per ISO 32000-1 §7.9.4 as `(D:YYYYMMDDHHmmSSZ)`.
func pdfDateString(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("(D:%04d%02d%02d%02d%02d%02dZ)",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
}
