// Package agg implements the render.Renderer interface using the AGG (Anti-Grain Geometry)
// rendering library via github.com/cwbudde/agg_go. AGG provides high-quality
// anti-aliased 2D rendering with sub-pixel accuracy.
package agg

import (
	"errors"
	"image"
	"image/png"
	"math"
	"os"
	"sync"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	agglib "github.com/cwbudde/agg_go"
)

// dejaVuFontPath holds the path to the extracted DejaVu Sans TTF, initialised once.
// DejaVu Sans is the same font matplotlib uses by default.
var (
	fontOnce sync.Once
	fontPath string
	fontErr  error
)

func loadFontPath() (string, error) {
	fontOnce.Do(func() {
		f, err := os.CreateTemp("", "matplotlib-go-*.ttf")
		if err != nil {
			fontErr = err
			return
		}
		_, err = f.Write(dejavusans.TTF)
		f.Close()
		if err != nil {
			os.Remove(f.Name())
			fontErr = err
			return
		}
		fontPath = f.Name()
	})
	return fontPath, fontErr
}

// Renderer implements render.Renderer using the AGG rendering backend.
type Renderer struct {
	ctx         *aggSurface
	width       int
	height      int
	resolution  uint
	began       bool
	viewport    geom.Rect
	stack       []state
	clipRect    *geom.Rect
	fontPath    string // path to TrueType font; empty means use GSV fallback
	fallback    bool   // true if any text path had to fall back to GSV
	lastFontKey string
	outlineText *agglib.FreeTypeOutlineText
}

// state represents a saved graphics state.
type state struct {
	clipRect *geom.Rect
}

var _ render.Renderer = (*Renderer)(nil)
var _ render.DPIAware = (*Renderer)(nil)
var _ render.TextDrawer = (*Renderer)(nil)
var _ render.RotatedTextDrawer = (*Renderer)(nil)
var _ render.VerticalTextDrawer = (*Renderer)(nil)
var _ render.TextBounder = (*Renderer)(nil)
var _ render.TextFontMetricer = (*Renderer)(nil)
var _ render.TextPather = (*Renderer)(nil)
var _ render.ImageTransformer = (*Renderer)(nil)
var _ render.PNGExporter = (*Renderer)(nil)

// New creates a new AGG renderer with the specified dimensions and background color.
// Returns an error if width or height are not positive.
func New(w, h int, bg render.Color) (*Renderer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("agg: width and height must be positive")
	}

	ctx := newAggSurface(w, h)

	// Clear with background color
	bgColor := renderColorToAGG(bg)
	ctx.Clear(bgColor)

	r := &Renderer{
		ctx:        ctx,
		width:      w,
		height:     h,
		resolution: 72,
	}

	// Prefer DejaVu Sans (the same default font Matplotlib ships with) via AGG's
	// raster FreeType text path. Fall back to the legacy GSV vector font only
	// when the FreeType-backed path is unavailable in the current build.
	fp, err := loadFontPath()
	if err == nil {
		r.fontPath = fp
	}
	if err := r.ctx.ConfigureTextFont(r.fontPath, 12, r.resolution); err != nil {
		r.fallback = true
	}

	return r, nil
}

// SetResolution sets the font rendering resolution used for text metrics and glyph sizing.
func (r *Renderer) SetResolution(dpi uint) {
	if dpi > 0 {
		r.resolution = dpi
	}
	r.ctx.SetResolution(r.resolution)
}

// Begin starts a drawing session with the given viewport.
func (r *Renderer) Begin(viewport geom.Rect) error {
	if r.began {
		return errors.New("Begin called twice")
	}
	r.began = true
	r.viewport = viewport
	r.stack = r.stack[:0]
	r.clipRect = nil
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
		rectCopy := *r.clipRect
		clipCopy = &rectCopy
	}
	r.stack = append(r.stack, state{clipRect: clipCopy})
	r.ctx.PushTransform()
}

// Restore pops the graphics state from the stack.
func (r *Renderer) Restore() {
	if len(r.stack) == 0 {
		return
	}
	s := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.clipRect = s.clipRect
	r.ctx.PopTransform()
	r.applyClipRect()
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if r.clipRect == nil {
		r.clipRect = &rect
	} else {
		intersected := r.clipRect.Intersect(rect)
		r.clipRect = &intersected
	}
	r.applyClipRect()
}

// ClipPath sets a path-based clip region (not yet supported, no-op).
func (r *Renderer) ClipPath(_ geom.Path) {
	// AGG supports path clipping at the rasterizer level, but not through Agg2D's simple API.
	// For now this is a no-op, same as gobasic.
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !p.Validate() || paint == nil {
		return
	}

	// Fill first if requested
	if paint.Fill.A > 0 {
		r.buildPath(p)
		r.ctx.SetFillColor(renderColorToAGG(paint.Fill))
		r.ctx.Fill()
	}

	// Then stroke if requested
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		r.buildPath(p)
		r.ctx.SetStrokeColor(renderColorToAGG(paint.Stroke))
		r.ctx.SetStrokeWidth(paint.LineWidth)

		// Map line join
		switch paint.LineJoin {
		case render.JoinMiter:
			r.ctx.SetLineJoin(agglib.JoinMiter)
		case render.JoinRound:
			r.ctx.SetLineJoin(agglib.JoinRound)
		case render.JoinBevel:
			r.ctx.SetLineJoin(agglib.JoinBevel)
		}

		// Map line cap
		switch paint.LineCap {
		case render.CapButt:
			r.ctx.SetLineCap(agglib.CapButt)
		case render.CapRound:
			r.ctx.SetLineCap(agglib.CapRound)
		case render.CapSquare:
			r.ctx.SetLineCap(agglib.CapSquare)
		}

		// Set miter limit
		if paint.MiterLimit > 0 {
			r.ctx.SetMiterLimit(paint.MiterLimit)
		}

		// Handle dashes
		r.ctx.ClearDashes()
		if len(paint.Dashes) >= 2 {
			r.ctx.SetDashPattern(paint.Dashes)
		}

		r.ctx.Stroke()

		// Clean up dashes
		if len(paint.Dashes) >= 2 {
			r.ctx.ClearDashes()
		}
	}
}

// buildPath converts a geom.Path into AGG path commands on the current context.
// Coordinates are quantized to ensure deterministic rendering across platforms.
func (r *Renderer) buildPath(p geom.Path) {
	r.ctx.BeginPath()

	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return
			}
			pt := quantizePt(p.V[vi])
			r.ctx.MoveTo(pt.X, pt.Y)
			vi++
		case geom.LineTo:
			if vi >= len(p.V) {
				return
			}
			pt := quantizePt(p.V[vi])
			r.ctx.LineTo(pt.X, pt.Y)
			vi++
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return
			}
			ctrl := quantizePt(p.V[vi])
			to := quantizePt(p.V[vi+1])
			r.ctx.QuadricCurveTo(ctrl.X, ctrl.Y, to.X, to.Y)
			vi += 2
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return
			}
			c1 := quantizePt(p.V[vi])
			c2 := quantizePt(p.V[vi+1])
			to := quantizePt(p.V[vi+2])
			r.ctx.CubicCurveTo(c1.X, c1.Y, c2.X, c2.Y, to.X, to.Y)
			vi += 3
		case geom.ClosePath:
			r.ctx.ClosePath()
		}
	}
}

func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	aggImg, ok := renderImageToAGG(img)
	if !ok {
		return
	}

	agg := r.ctx
	prevFilter := agg.GetImageFilter()
	prevResample := agg.GetImageResample()
	agg.SetImageFilter(agglib.NoFilter)
	agg.SetImageResample(agglib.NoResample)
	defer func() {
		agg.SetImageFilter(prevFilter)
		agg.SetImageResample(prevResample)
	}()

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

	_ = agg.DrawImageScaled(aggImg, x, y, w, h)
}

// ImageTransformed draws an image using the provided affine transformation.
// Used by core.Image2D when rotation is requested.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, affine geom.Affine) {
	aggImg, ok := renderImageToAGG(img)
	if !ok {
		return
	}

	agg := r.ctx
	prevFilter := agg.GetImageFilter()
	prevResample := agg.GetImageResample()
	agg.SetImageFilter(agglib.NoFilter)
	agg.SetImageResample(agglib.NoResample)
	defer func() {
		agg.SetImageFilter(prevFilter)
		agg.SetImageResample(prevResample)
	}()

	transform := agglib.NewTransformationsFromValues(
		affine.A,
		affine.B,
		affine.C,
		affine.D,
		affine.E,
		affine.F,
	)
	_ = agg.DrawImageTransformed(aggImg, transform)
}

// GlyphRun draws a run of glyphs.
func (r *Renderer) GlyphRun(_ render.GlyphRun, _ render.Color) {
	// GlyphRun requires glyph-ID-to-character mapping.
	// Text rendering is done through DrawText helper instead.
}

// MeasureText measures text dimensions using the active font engine.
func (r *Renderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	if text == "" || size <= 0 {
		return render.TextMetrics{}
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)

	var (
		w       float64
		ascent  float64
		descent float64
	)
	switch font.backend {
	case textBackendRaster:
		if metrics, ok := r.measureRasterText(text, font.fontPath, font.size); ok {
			return metrics
		}
		if err := r.ctx.ConfigureTextFont(font.fontPath, font.size, r.resolution); err == nil {
			w, ascent, descent = r.ctx.TextMetrics(text)
			break
		}
		r.fallback = true
		sizePx := r.fontPixelSize(font.size)
		w = measureLocalGSVTextWidth(text, sizePx)
		if x, y, bw, h, ok := measureTextPathBounds(text, sizePx, font.fontPath); ok {
			w = math.Max(w, x+bw)
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else if _, y, _, h, ok := measureLocalGSVTextBounds(text, sizePx); ok {
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else {
			ascent = sizePx
			descent = 0
		}
	default:
		sizePx := r.fontPixelSize(font.size)
		w = measureLocalGSVTextWidth(text, sizePx)
		if _, y, _, h, ok := measureLocalGSVTextBounds(text, sizePx); ok {
			ascent = math.Max(0, -y)
			descent = math.Max(0, y+h)
		} else {
			ascent = sizePx
			descent = 0
		}
	}

	h := ascent + descent
	if h <= 0 {
		h = font.size
	}
	if ascent <= 0 {
		ascent = h
	}
	if descent < 0 {
		descent = 0
	}

	return render.TextMetrics{
		W:       w,
		H:       h,
		Ascent:  ascent,
		Descent: descent,
	}
}

// MeasureTextBounds reports the actual ink bounds of text relative to the
// baseline origin used for DrawText.
func (r *Renderer) MeasureTextBounds(text string, size float64, fontKey string) (render.TextBounds, bool) {
	if text == "" || size <= 0 {
		return render.TextBounds{}, false
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)
	if font.backend == textBackendGSV {
		sizePx := r.fontPixelSize(font.size)
		x, y, w, h, ok := measureLocalGSVTextBounds(text, sizePx)
		if !ok {
			return render.TextBounds{}, false
		}
		return render.TextBounds{X: x, Y: y, W: w, H: h}, true
	}
	if font.backend != textBackendRaster {
		return render.TextBounds{}, false
	}
	sizePx := r.fontPixelSize(font.size)
	if layout, ok := render.LayoutTextGlyphs(text, geom.Pt{}, sizePx, font.fontPath); ok {
		return layout.Bounds, true
	}
	if err := r.ctx.ConfigureTextFont(font.fontPath, font.size, r.resolution); err != nil {
		r.fallback = true
		if x, y, w, h, ok := measureTextPathBounds(text, sizePx, font.fontPath); ok {
			return render.TextBounds{X: x, Y: y, W: w, H: h}, true
		}
		x, y, w, h, ok := measureLocalGSVTextBounds(text, sizePx)
		if !ok {
			return render.TextBounds{}, false
		}
		return render.TextBounds{X: x, Y: y, W: w, H: h}, true
	}

	x, y, w, h := r.ctx.TextBounds(text)
	return render.TextBounds{X: x, Y: y, W: w, H: h}, true
}

// MeasureFontHeights reports font-wide ascent, descent, and line-gap values
// for the current raster text face, distinct from a particular string's ink
// bounds.
func (r *Renderer) MeasureFontHeights(size float64, fontKey string) (render.FontHeightMetrics, bool) {
	if size <= 0 {
		return render.FontHeightMetrics{}, false
	}

	if fontKey != "" {
		r.lastFontKey = fontKey
	} else {
		fontKey = r.lastFontKey
	}

	font := r.configureTextFont(size, fontKey)
	if font.backend == textBackendGSV {
		sizePx := r.fontPixelSize(font.size)
		if _, y, _, h, ok := measureLocalGSVTextBounds("lp", sizePx); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		return render.FontHeightMetrics{}, false
	}
	if font.backend != textBackendRaster {
		return render.FontHeightMetrics{}, false
	}
	if face, err := r.openRasterFace(font.fontPath, font.size); err == nil {
		defer func() { _ = face.Close() }()
		metrics := face.Metrics()
		return render.FontHeightMetrics{
			Ascent:  float64(metrics.Ascent.Ceil()),
			Descent: float64(metrics.Descent.Ceil()),
		}, true
	}
	if err := r.ctx.ConfigureTextFont(font.fontPath, font.size, r.resolution); err != nil {
		r.fallback = true
		sizePx := r.fontPixelSize(font.size)
		if _, y, _, h, ok := measureTextPathBounds("lp", sizePx, font.fontPath); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		if _, y, _, h, ok := measureLocalGSVTextBounds("lp", sizePx); ok {
			return render.FontHeightMetrics{
				Ascent:  math.Max(0, -y),
				Descent: math.Max(0, y+h),
			}, true
		}
		return render.FontHeightMetrics{}, false
	}

	metrics, ok := r.ctx.fontHeightMetrics()
	if !ok {
		return render.FontHeightMetrics{}, false
	}
	return render.FontHeightMetrics{
		Ascent:  metrics.ascent,
		Descent: metrics.descent,
		LineGap: metrics.lineGap,
	}, true
}

// TextPath converts text to a vector path using the renderer's resolved font.
func (r *Renderer) TextPath(text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if text == "" || size <= 0 {
		return geom.Path{}, false
	}
	if fontKey == "" {
		fontKey = r.lastFontKey
	}
	font := r.configureTextFont(size, fontKey)
	if font.fontPath == "" {
		return geom.Path{}, false
	}
	return render.TextPath(text, origin, r.fontPixelSize(size), font.fontPath)
}

// DrawText renders text at the given position with the specified size and color.
// This is a helper method (not part of the Renderer interface).
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}

	font := r.configureTextFont(size, r.lastFontKey)

	switch font.backend {
	case textBackendRaster:
		if r.drawRasterText(text, font.fontPath, origin, font.size, textColor) {
			return
		}
		if err := r.ctx.ConfigureTextFont(font.fontPath, font.size, r.resolution); err == nil {
			r.ctx.SetFillColor(renderColorToAGG(textColor))
			r.ctx.SetStrokeColor(renderColorToAGG(textColor))
			r.ctx.DrawText(text, origin.X, origin.Y)
			return
		}
		sizePx := r.fontPixelSize(font.size)
		if r.drawTextPathFallback(text, origin, sizePx, textColor, font.fontPath) {
			return
		}
		r.fallback = true
		fallthrough
	default:
		sizePx := r.fontPixelSize(font.size)
		r.ctx.SetStrokeColor(renderColorToAGG(textColor))
		r.ctx.SetStrokeWidth(math.Max(1, sizePx*0.08))
		r.ctx.SetLineCap(agglib.CapRound)
		r.ctx.SetLineJoin(agglib.JoinRound)
		if appendLocalGSVText(r.ctx, origin.X, origin.Y, sizePx, text) {
			r.ctx.Stroke()
		}
	}
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

	bounds, haveBounds := r.MeasureTextBounds(text, size, "")
	origin := rotatedTextOrigin(anchor, metrics, bounds, haveBounds)
	font := r.configureTextFont(size, r.lastFontKey)

	r.ctx.PushTransform()
	defer r.ctx.PopTransform()

	r.ctx.Translate(-anchor.X, -anchor.Y)
	r.ctx.Rotate(-angle)
	r.ctx.Translate(anchor.X, anchor.Y)

	if font.fontPath != "" {
		sizePx := r.fontPixelSize(font.size)
		if r.drawTextPathFallback(text, origin, sizePx, textColor, font.fontPath) {
			return
		}
		if face, err := r.configureOutlineFont(font.fontPath, sizePx); err == nil {
			r.ctx.SetFillColor(renderColorToAGG(textColor))
			r.ctx.SetStrokeColor(renderColorToAGG(textColor))
			if drawTrueTypeOutlineText(r.ctx, face, origin.X, origin.Y, text) {
				return
			}
		}
		r.fallback = true
	}

	sizePx := r.fontPixelSize(font.size)
	r.ctx.SetStrokeColor(renderColorToAGG(textColor))
	r.ctx.SetStrokeWidth(math.Max(1, sizePx*0.08))
	r.ctx.SetLineCap(agglib.CapRound)
	r.ctx.SetLineJoin(agglib.JoinRound)
	if appendLocalGSVText(r.ctx, origin.X, origin.Y, sizePx, text) {
		r.ctx.Stroke()
	}
}

func (r *Renderer) fontPixelSize(size float64) float64 {
	dpi := float64(r.resolution)
	if dpi <= 0 {
		dpi = 72
	}
	return size * dpi / 72.0
}

func (r *Renderer) drawTextPathFallback(text string, origin geom.Pt, size float64, textColor render.Color, fontPath string) bool {
	if fontPath == "" {
		return false
	}
	path, ok := render.TextPath(text, origin, size, fontPath)
	if !ok {
		return false
	}
	r.Path(path, &render.Paint{
		Fill: textColor,
	})
	return true
}

func rotatedTextOrigin(anchor geom.Pt, metrics render.TextMetrics, bounds render.TextBounds, haveBounds bool) geom.Pt {
	if haveBounds && bounds.W > 0 && bounds.H > 0 {
		return geom.Pt{
			X: anchor.X - (bounds.X + bounds.W/2),
			Y: anchor.Y - (bounds.Y + bounds.H),
		}
	}

	return geom.Pt{
		X: anchor.X - metrics.W/2,
		Y: anchor.Y - metrics.Descent,
	}
}

// DrawTextVertical renders text vertically (one character per line, top to bottom).
// This is used for ylabel rendering where true rotation is not available.
func (r *Renderer) DrawTextVertical(text string, center geom.Pt, size float64, textColor render.Color) {
	if text == "" || size <= 0 {
		return
	}

	lineMetrics := r.MeasureText("M", size, "")
	h := lineMetrics.H
	if h <= 0 {
		h = size
	}
	runes := []rune(text)
	totalH := float64(len(runes)) * h
	y := center.Y - totalH/2 + h // start from top, offset by one line height

	for _, ch := range runes {
		s := string(ch)
		w := r.MeasureText(s, size, "").W
		x := center.X - w/2
		r.DrawText(s, geom.Pt{X: x, Y: y}, size, textColor)
		y += h
	}
}

// GetImage returns the rendered image as a standard Go image.RGBA.
func (r *Renderer) GetImage() *image.RGBA {
	return r.ctx.GetImage().ToGoImage()
}

// SavePNG saves the rendered image to a PNG file.
func (r *Renderer) SavePNG(path string) error {
	img := r.GetImage()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

// renderColorToAGG converts a normalized render.Color to AGG's 8-bit SRGBA
// color type without applying any transfer-curve conversion.
func renderColorToAGG(c render.Color) agglib.Color {
	return agglib.NewColor(
		uint8(math.Round(clamp01(c.R)*255)),
		uint8(math.Round(clamp01(c.G)*255)),
		uint8(math.Round(clamp01(c.B)*255)),
		uint8(math.Round(clamp01(c.A)*255)),
	)
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

// quantize snaps a floating-point value to a fixed grid to ensure
// deterministic rendering across platforms and compiler versions.
const quantizationGrid = 1e-6

func quantize(v float64) float64 {
	return math.Round(v/quantizationGrid) * quantizationGrid
}

func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{X: quantize(p.X), Y: quantize(p.Y)}
}

func (r *Renderer) applyClipRect() {
	if r.clipRect != nil {
		r.ctx.ClipBox(r.clipRect.Min.X, r.clipRect.Min.Y, r.clipRect.Max.X, r.clipRect.Max.Y)
		return
	}
	r.ctx.ClipBox(0, 0, float64(r.width), float64(r.height))
}

// renderImageToAGG converts a renderer image into an AGG image type.
func renderImageToAGG(img render.Image) (*agglib.Image, bool) {
	if img == nil {
		return nil, false
	}

	rgbaImage, ok := img.(render.RGBAImage)
	if !ok {
		return nil, false
	}

	rgba := rgbaImage.RGBA()
	if rgba == nil || rgba.Bounds().Dx() <= 0 || rgba.Bounds().Dy() <= 0 {
		return nil, false
	}

	aggImg, err := agglib.NewImageFromStandardImage(rgba)
	if err != nil {
		return nil, false
	}
	return aggImg, true
}
