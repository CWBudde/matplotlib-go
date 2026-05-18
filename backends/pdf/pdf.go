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
	content bytes.Buffer
	// document is the fully serialized PDF bytes ready for write.
	document []byte

	// pdfOpts carries setter-supplied options. SavePDFWithOptions overrides
	// fields directly for that single call.
	pdfOpts render.PDFOptions
}

// Compile-time interface assertions.
var (
	_ render.Renderer     = (*Renderer)(nil)
	_ render.PNGExporter  = nil // explicitly not implemented
	_ render.PDFExporter  = (*Renderer)(nil)
	_ render.DPIAware     = (*Renderer)(nil)
	_ render.PDFOptionExporter = (*Renderer)(nil)
	_ render.PDFOptionSetter   = (*Renderer)(nil)
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
	doc, err := buildDocument(r.width, r.height, r.content.Bytes(), r.pdfOpts)
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

// Path draws a path using the provided paint.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !r.began || paint == nil {
		return
	}
	hasFill := paint.Fill.A > 0
	hasStroke := paint.Stroke.A > 0 && paint.LineWidth > 0
	if !hasFill && !hasStroke {
		return
	}
	if !writePathOps(&r.content, p) {
		return
	}

	if hasFill {
		writeFillColor(&r.content, paint.Fill)
	}
	if hasStroke {
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

// Image draws a raster image into the destination rectangle. Image XObject
// support is not yet implemented; this method is a no-op so callers fall back
// to renderer-neutral expansion (typically a path-based placeholder).
func (r *Renderer) Image(_ render.Image, _ geom.Rect) {
	// TODO(phase1.1): emit /XObject /Image with FlateDecode pixel data.
}

// GlyphRun draws shaped glyphs. The initial backend has no font subsetting,
// so glyph drawing is a no-op. Callers can wrap the renderer in a TextPather
// shim to render text-as-path.
func (r *Renderer) GlyphRun(_ render.GlyphRun, _ render.Color) {
	// TODO(phase1.1): route through the text-as-path fallback.
}

// MeasureText returns rough metrics so layout code does not divide by zero.
// A future revision will plumb in the shared font manager.
func (r *Renderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	if text == "" {
		return render.TextMetrics{}
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
	doc, err := buildDocument(r.width, r.height, r.content.Bytes(), opts)
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

// --- PDF document assembly ---------------------------------------------------

// buildDocument assembles the PDF bytes for one page given the encoded
// content stream.
func buildDocument(width, height int, contentStream []byte, opts render.PDFOptions) ([]byte, error) {
	// We emit five indirect objects:
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
	fmt.Fprintf(&w.buf, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] /Contents 4 0 R /Resources << >> >>",
		width, height)
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

// Image alpha helper used by the TODO image path; kept here so the eventual
// XObject path can pre-multiply alpha consistently with other backends.
var _ = image.NewRGBA
