// Package agg implements the render.Renderer interface using the AGG (Anti-Grain Geometry)
// rendering library via github.com/MeKo-Christian/agg_go. AGG provides high-quality
// anti-aliased 2D rendering with sub-pixel accuracy.
package agg

import (
	"errors"
	"image"
	"image/png"
	"math"
	"os"
	"sync"

	agglib "github.com/MeKo-Christian/agg_go"
	"golang.org/x/image/font/gofont/goregular"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// goFontPath holds the path to the extracted Go Regular TTF, initialised once.
var (
	goFontOnce sync.Once
	goFontPath string
	goFontErr  error
)

func loadGoFontPath() (string, error) {
	goFontOnce.Do(func() {
		f, err := os.CreateTemp("", "matplotlib-go-*.ttf")
		if err != nil {
			goFontErr = err
			return
		}
		_, err = f.Write(goregular.TTF)
		f.Close()
		if err != nil {
			os.Remove(f.Name())
			goFontErr = err
			return
		}
		goFontPath = f.Name()
	})
	return goFontPath, goFontErr
}

// Renderer implements render.Renderer using the AGG rendering backend.
type Renderer struct {
	ctx      *agglib.Context
	width    int
	height   int
	began    bool
	viewport geom.Rect
	stack    []state
	clipRect *geom.Rect
	fontPath string // path to TrueType font; empty means use GSV fallback
}

// state represents a saved graphics state.
type state struct {
	clipRect *geom.Rect
}

var _ render.Renderer = (*Renderer)(nil)

// New creates a new AGG renderer with the specified dimensions and background color.
// Returns an error if width or height are not positive.
func New(w, h int, bg render.Color) (*Renderer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("agg: width and height must be positive")
	}

	ctx := agglib.NewContext(w, h)

	// Clear with background color
	bgColor := renderColorToAGG(bg)
	ctx.Clear(bgColor)

	r := &Renderer{
		ctx:    ctx,
		width:  w,
		height: h,
	}

	// Prefer TrueType (Go Regular) for crisp text; fall back to built-in GSV.
	if fp, err := loadGoFontPath(); err == nil {
		agg := ctx.GetAgg2D()
		if loadErr := agg.Font(fp, 12.0, false, false, agglib.RasterFontCache, 0); loadErr == nil {
			r.fontPath = fp
		}
	}
	if r.fontPath == "" {
		ctx.GetAgg2D().FontGSV(13.0)
	}

	return r, nil
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

	// Restore clip box
	if r.clipRect != nil {
		r.ctx.GetAgg2D().ClipBox(r.clipRect.Min.X, r.clipRect.Min.Y, r.clipRect.Max.X, r.clipRect.Max.Y)
	} else {
		r.ctx.GetAgg2D().ClipBox(0, 0, float64(r.width), float64(r.height))
	}
}

// ClipRect sets a rectangular clip region.
func (r *Renderer) ClipRect(rect geom.Rect) {
	if r.clipRect == nil {
		r.clipRect = &rect
	} else {
		intersected := r.clipRect.Intersect(rect)
		r.clipRect = &intersected
	}
	r.ctx.GetAgg2D().ClipBox(r.clipRect.Min.X, r.clipRect.Min.Y, r.clipRect.Max.X, r.clipRect.Max.Y)
}

// ClipPath sets a path-based clip region (not yet supported, no-op).
func (r *Renderer) ClipPath(_ geom.Path) {
	// AGG supports path clipping at the rasterizer level, but not through Agg2D's simple API.
	// For now this is a no-op, same as gobasic.
}

// Path draws a path with the given paint style.
func (r *Renderer) Path(p geom.Path, paint *render.Paint) {
	if !p.Validate() {
		return
	}

	agg := r.ctx.GetAgg2D()

	// Fill first if requested
	if paint.Fill.A > 0 {
		r.buildAGGPath(p)
		agg.FillColor(renderColorToAGG(paint.Fill))
		agg.NoLine()
		agg.DrawPath(agglib.FillOnly)
	}

	// Then stroke if requested
	if paint.Stroke.A > 0 && paint.LineWidth > 0 {
		r.buildAGGPath(p)
		agg.LineColor(renderColorToAGG(paint.Stroke))
		agg.NoFill()
		agg.LineWidth(paint.LineWidth)

		// Map line join
		switch paint.LineJoin {
		case render.JoinMiter:
			agg.LineJoin(agglib.JoinMiter)
		case render.JoinRound:
			agg.LineJoin(agglib.JoinRound)
		case render.JoinBevel:
			agg.LineJoin(agglib.JoinBevel)
		}

		// Map line cap
		switch paint.LineCap {
		case render.CapButt:
			agg.LineCap(agglib.CapButt)
		case render.CapRound:
			agg.LineCap(agglib.CapRound)
		case render.CapSquare:
			agg.LineCap(agglib.CapSquare)
		}

		// Set miter limit
		if paint.MiterLimit > 0 {
			agg.MiterLimit(paint.MiterLimit)
		}

		// Handle dashes
		agg.RemoveAllDashes()
		if len(paint.Dashes) >= 2 {
			for i := 0; i+1 < len(paint.Dashes); i += 2 {
				agg.AddDash(paint.Dashes[i], paint.Dashes[i+1])
			}
		}

		agg.DrawPath(agglib.StrokeOnly)

		// Clean up dashes
		if len(paint.Dashes) >= 2 {
			agg.RemoveAllDashes()
		}
	}
}

// buildAGGPath converts a geom.Path into AGG path commands on the current context.
// Coordinates are quantized to ensure deterministic rendering across platforms.
func (r *Renderer) buildAGGPath(p geom.Path) {
	agg := r.ctx.GetAgg2D()
	agg.ResetPath()

	vi := 0
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(p.V) {
				return
			}
			pt := quantizePt(p.V[vi])
			agg.MoveTo(pt.X, pt.Y)
			vi++
		case geom.LineTo:
			if vi >= len(p.V) {
				return
			}
			pt := quantizePt(p.V[vi])
			agg.LineTo(pt.X, pt.Y)
			vi++
		case geom.QuadTo:
			if vi+1 >= len(p.V) {
				return
			}
			ctrl := quantizePt(p.V[vi])
			to := quantizePt(p.V[vi+1])
			agg.QuadricCurveTo(ctrl.X, ctrl.Y, to.X, to.Y)
			vi += 2
		case geom.CubicTo:
			if vi+2 >= len(p.V) {
				return
			}
			c1 := quantizePt(p.V[vi])
			c2 := quantizePt(p.V[vi+1])
			to := quantizePt(p.V[vi+2])
			agg.CubicCurveTo(c1.X, c1.Y, c2.X, c2.Y, to.X, to.Y)
			vi += 3
		case geom.ClosePath:
			agg.ClosePolygon()
		}
	}
}

// Image draws an image within the destination rectangle.
func (r *Renderer) Image(_ render.Image, _ geom.Rect) {
	// Will be implemented in later phases when image support is needed.
}

// GlyphRun draws a run of glyphs.
func (r *Renderer) GlyphRun(_ render.GlyphRun, _ render.Color) {
	// GlyphRun requires glyph-ID-to-character mapping.
	// Text rendering is done through DrawText helper instead.
}

// setFont configures the font engine for the given size.
func (r *Renderer) setFont(size float64) {
	agg := r.ctx.GetAgg2D()
	if r.fontPath != "" {
		_ = agg.Font(r.fontPath, size, false, false, agglib.RasterFontCache, 0)
	} else {
		agg.FontGSV(size)
	}
}

// MeasureText measures text dimensions using the active font engine.
func (r *Renderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	if text == "" {
		return render.TextMetrics{}
	}

	r.setFont(size)
	agg := r.ctx.GetAgg2D()
	w := agg.TextWidth(text)
	h := agg.FontHeight()

	return render.TextMetrics{
		W:       w,
		H:       h,
		Ascent:  h * 0.8,
		Descent: h * 0.2,
	}
}

// DrawText renders text at the given position with the specified size and color.
// This is a helper method (not part of the Renderer interface).
func (r *Renderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "" {
		return
	}

	r.setFont(size)
	agg := r.ctx.GetAgg2D()
	agg.FillColor(renderColorToAGG(textColor))
	agg.LineColor(renderColorToAGG(textColor))
	agg.Text(origin.X, origin.Y, text, true, 0, 0)
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

// renderColorToAGG converts a render.Color (linear float64 0..1) to an agg.Color (uint8 0..255).
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
