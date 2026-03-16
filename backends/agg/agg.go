package agg

import (
	"errors"
	"image"
	"image/png"
	"math"
	"os"

	"agg_go"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// quantizationEpsilon ensures deterministic rendering across platforms.
// All coordinates are quantized to this precision to avoid floating-point drift.
const quantizationEpsilon = 1e-6

// AggConfig holds AGG-specific configuration options.
type AggConfig struct {
	AntiAliasing    bool    // Enable high-quality anti-aliasing
	SubPixelPrec    bool    // Enable sub-pixel precision
	GammaCorrection float64 // Gamma correction for anti-aliasing (typically 2.2)
	// TODO: Add more configuration options as AGG features are implemented
	// FilterType      string  // Image scaling filter type
	// ThreadCount     int     // Number of rendering threads
}

// AggRenderer implements the render.Renderer interface using Anti-Grain Geometry
// for high-quality anti-aliased rendering with sub-pixel precision.
type AggRenderer struct {
	// AGG context and rendering state
	ctx    *agg.Agg2D
	buffer []uint8  // RGBA pixel buffer
	width  int
	height int
	config AggConfig
	
	// State management
	began   bool
	stateStack []renderState
	
	// Current rendering state
	transform geom.Affine
	clipStack []clipRegion
}

// renderState captures the complete rendering state for Save/Restore operations.
type renderState struct {
	transform geom.Affine
	clipStack []clipRegion
}

// clipRegion represents a clipping area (rectangular or path-based).
type clipRegion struct {
	rect *geom.Rect
	path *geom.Path
}

// Ensure AggRenderer implements render.Renderer interface at compile time.
var _ render.Renderer = (*AggRenderer)(nil)

// New creates a new AGG renderer with the specified dimensions.
func New(width, height int) (*AggRenderer, error) {
	if width <= 0 || height <= 0 {
		return nil, errors.New("invalid dimensions: width and height must be positive")
	}
	
	// Create AGG2D context
	ctx := agg.NewAgg2D()
	if ctx == nil {
		return nil, errors.New("failed to create AGG context")
	}
	
	// Create RGBA buffer (4 bytes per pixel)
	bufferSize := width * height * 4
	buffer := make([]uint8, bufferSize)
	
	// Attach buffer to AGG context (stride = width * 4 for RGBA)
	ctx.Attach(buffer, width, height, width*4)
	
	return &AggRenderer{
		ctx:       ctx,
		buffer:    buffer,
		width:     width,
		height:    height,
		transform: geom.Identity(),
	}, nil
}

// Begin initializes the renderer for a drawing session with the given viewport.
func (a *AggRenderer) Begin(viewport geom.Rect) error {
	if a.began {
		return errors.New("Begin called twice without End")
	}
	
	// Clear the canvas with white (matching matplotlib convention)
	// AGG Color takes RGBA values as uint8 [0..255]
	clearColor := agg.Color{R: 255, G: 255, B: 255, A: 255}
	a.ctx.ClearAll(clearColor)
	
	// Set up coordinate system transformation
	// Convert from matplotlib-go coordinates (top-left origin, Y down)
	// to AGG coordinates (bottom-left origin, Y up)
	a.transform = geom.Affine{
		A: 1, B: 0, C: 0, D: -1,
		E: 0, F: float64(a.height),
	}
	
	a.began = true
	return nil
}

// End finalizes the drawing session and returns the rendered image.
func (a *AggRenderer) End() error {
	if !a.began {
		return errors.New("End called before Begin")
	}
	
	// Reset state
	a.began = false
	a.stateStack = a.stateStack[:0]
	a.clipStack = a.clipStack[:0]
	a.transform = geom.Identity()
	
	return nil
}

// Save pushes the current rendering state onto the stack.
func (a *AggRenderer) Save() {
	state := renderState{
		transform: a.transform,
		clipStack: make([]clipRegion, len(a.clipStack)),
	}
	copy(state.clipStack, a.clipStack)
	a.stateStack = append(a.stateStack, state)
}

// Restore pops the rendering state from the stack.
func (a *AggRenderer) Restore() {
	if len(a.stateStack) == 0 {
		return // Gracefully handle underflow like GoBasic
	}
	
	// Restore state
	state := a.stateStack[len(a.stateStack)-1]
	a.stateStack = a.stateStack[:len(a.stateStack)-1]
	
	a.transform = state.transform
	a.clipStack = make([]clipRegion, len(state.clipStack))
	copy(a.clipStack, state.clipStack)
}

// ClipRect adds a rectangular clipping region.
func (a *AggRenderer) ClipRect(r geom.Rect) {
	// Quantize coordinates for determinism
	r.Min = quantizePt(r.Min)
	r.Max = quantizePt(r.Max)
	
	a.clipStack = append(a.clipStack, clipRegion{rect: &r})
}

// ClipPath adds a path-based clipping region.
func (a *AggRenderer) ClipPath(p geom.Path) {
	// Quantize path for determinism
	quantizedPath := quantizePath(p)
	a.clipStack = append(a.clipStack, clipRegion{path: &quantizedPath})
}

// Path renders a path with the specified paint style.
func (a *AggRenderer) Path(p geom.Path, paint *render.Paint) {
	if paint == nil {
		return
	}
	
	// Convert path to AGG format and apply quantization
	quantizedPath := quantizePath(p)
	
	// Reset AGG path and convert from geom.Path to AGG path
	a.ctx.ResetPath()
	a.convertPathToAGG(quantizedPath)
	
	// Set up paint properties and render
	if paint.Stroke.A > 0 { // Has stroke
		a.setupStrokePaint(paint)
		a.ctx.DrawPath(agg.StrokeOnly)
	}
	
	if paint.Fill.A > 0 { // Has fill
		a.setupFillPaint(paint)
		a.ctx.DrawPath(agg.FillOnly)
	}
}

// Image renders a raster image at the specified destination rectangle.
func (a *AggRenderer) Image(img render.Image, dst geom.Rect) {
	// Quantize destination rectangle
	dst.Min = quantizePt(dst.Min)
	dst.Max = quantizePt(dst.Max)
	
	// TODO: Implement image rendering with AGG
	// This will include:
	// 1. Convert render.Image to AGG-compatible format
	// 2. Apply transformation and scaling
	// 3. Use AGG's image filtering for quality scaling
}

// GlyphRun renders a run of text glyphs with the specified color.
func (a *AggRenderer) GlyphRun(run render.GlyphRun, color render.Color) {
	// Quantize glyph positions
	run.Origin = quantizePt(run.Origin)
	
	// TODO: Implement text rendering with AGG
	// This will include:
	// 1. Font loading and management
	// 2. Glyph rasterization with anti-aliasing
	// 3. Proper baseline and advance handling
}

// MeasureText calculates text metrics for layout purposes.
func (a *AggRenderer) MeasureText(text string, size float64, fontKey string) render.TextMetrics {
	// TODO: Implement text measurement
	// For now, return zero metrics
	return render.TextMetrics{}
}

// GetImage returns the current rendered image as an RGBA image.
// This method is AGG-specific and not part of the render.Renderer interface.
func (a *AggRenderer) GetImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, a.width, a.height))
	
	// Copy pixel data from AGG buffer to Go image
	// AGG buffer is in RGBA format, same as Go's image.RGBA
	copy(img.Pix, a.buffer)
	
	return img
}

// SavePNG implements the PNGExporter interface for core.SavePNG compatibility.
func (a *AggRenderer) SavePNG(path string) error {
	// Get the rendered image
	img := a.GetImage()
	
	// Create the output file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Encode as PNG
	return png.Encode(file, img)
}

// Utility functions for coordinate quantization (matching GoBasic approach)

// quantize snaps a float64 value to quantizationEpsilon precision.
func quantize(v float64) float64 {
	return math.Round(v/quantizationEpsilon) * quantizationEpsilon
}

// quantizePt quantizes both X and Y coordinates of a point.
func quantizePt(p geom.Pt) geom.Pt {
	return geom.Pt{
		X: quantize(p.X),
		Y: quantize(p.Y),
	}
}

// quantizePath quantizes all vertices in a path for deterministic rendering.
func quantizePath(p geom.Path) geom.Path {
	result := geom.Path{
		C: make([]geom.Cmd, len(p.C)),
		V: make([]geom.Pt, len(p.V)),
	}
	
	copy(result.C, p.C)
	for i, v := range p.V {
		result.V[i] = quantizePt(v)
	}
	
	return result
}

// convertPathToAGG converts a geom.Path to AGG path commands.
func (a *AggRenderer) convertPathToAGG(p geom.Path) {
	vIndex := 0 // Index into vertices array
	
	for _, cmd := range p.C {
		switch cmd {
		case geom.MoveTo:
			if vIndex < len(p.V) {
				pt := a.transformPoint(p.V[vIndex])
				a.ctx.MoveTo(pt.X, pt.Y)
				vIndex++
			}
			
		case geom.LineTo:
			if vIndex < len(p.V) {
				pt := a.transformPoint(p.V[vIndex])
				a.ctx.LineTo(pt.X, pt.Y)
				vIndex++
			}
			
		case geom.QuadTo:
			// AGG doesn't have native quadratic curves, convert to cubic
			if vIndex+1 < len(p.V) {
				// Get current position from last point (approximation)
				// For proper implementation, we'd need to track current position
				ctrl := a.transformPoint(p.V[vIndex])
				end := a.transformPoint(p.V[vIndex+1])
				
				// Convert quadratic to cubic Bézier (approximate)
				// This is a simplified conversion - for exact conversion we'd need current position
				a.ctx.CubicCurveTo(ctrl.X, ctrl.Y, ctrl.X, ctrl.Y, end.X, end.Y)
				vIndex += 2
			}
			
		case geom.CubicTo:
			if vIndex+2 < len(p.V) {
				ctrl1 := a.transformPoint(p.V[vIndex])
				ctrl2 := a.transformPoint(p.V[vIndex+1])
				end := a.transformPoint(p.V[vIndex+2])
				a.ctx.CubicCurveTo(ctrl1.X, ctrl1.Y, ctrl2.X, ctrl2.Y, end.X, end.Y)
				vIndex += 3
			}
			
		case geom.ClosePath:
			a.ctx.ClosePolygon()
		}
	}
}

// transformPoint applies the current transformation matrix to a point.
func (a *AggRenderer) transformPoint(pt geom.Pt) geom.Pt {
	return a.transform.Apply(pt)
}

// setupStrokePaint configures AGG for stroke rendering.
func (a *AggRenderer) setupStrokePaint(paint *render.Paint) {
	// Set stroke color (convert from float64 [0..1] to uint8 [0..255])
	color := agg.Color{
		R: uint8(paint.Stroke.R*255 + 0.5),
		G: uint8(paint.Stroke.G*255 + 0.5), 
		B: uint8(paint.Stroke.B*255 + 0.5),
		A: uint8(paint.Stroke.A*255 + 0.5),
	}
	a.ctx.LineColor(color)
	
	// Set stroke width
	a.ctx.LineWidth(paint.LineWidth)
	
	// Set line cap style
	switch paint.LineCap {
	case render.CapButt:
		a.ctx.LineCap(agg.CapButt)
	case render.CapRound:
		a.ctx.LineCap(agg.CapRound)
	case render.CapSquare:
		a.ctx.LineCap(agg.CapSquare)
	default:
		a.ctx.LineCap(agg.CapButt)
	}
	
	// Set line join style  
	switch paint.LineJoin {
	case render.JoinMiter:
		a.ctx.LineJoin(agg.JoinMiter)
	case render.JoinRound:
		a.ctx.LineJoin(agg.JoinRound)
	case render.JoinBevel:
		a.ctx.LineJoin(agg.JoinBevel)
	default:
		a.ctx.LineJoin(agg.JoinMiter)
	}
}

// setupFillPaint configures AGG for fill rendering.
func (a *AggRenderer) setupFillPaint(paint *render.Paint) {
	// Set fill color (convert from float64 [0..1] to uint8 [0..255])
	color := agg.Color{
		R: uint8(paint.Fill.R*255 + 0.5),
		G: uint8(paint.Fill.G*255 + 0.5),
		B: uint8(paint.Fill.B*255 + 0.5), 
		A: uint8(paint.Fill.A*255 + 0.5),
	}
	a.ctx.FillColor(color)
}