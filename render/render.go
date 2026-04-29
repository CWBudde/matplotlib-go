package render

import (
	"errors"
	"image"

	"matplotlib-go/internal/geom"
)

// Paint configures drawing style for paths.
type Paint struct {
	LineWidth         float64
	LineJoin          LineJoin
	LineCap           LineCap
	MiterLimit        float64
	Stroke            Color
	Fill              Color
	Dashes            []float64 // on/off pairs, in user space units
	Antialias         AntialiasMode
	Snap              SnapMode // path snapping policy; zero preserves existing unsnapped behavior
	Simplify          bool     // simplify line-only paths before rasterization
	SimplifyThreshold float64  // simplification tolerance in display pixels
	MaxChunkVertices  int      // split very large stroke-only line paths; <=0 uses backend default
	Hatch             string
	HatchColor        Color
	HatchLineWidth    float64
	HatchSpacing      float64
	Sketch            SketchParams
	ForceAlpha        bool
	Alpha             float64
	ClipPathTransform geom.Affine
	HasClipPathTrans  bool
}

// AntialiasMode controls path antialiasing behavior.
type AntialiasMode uint8

const (
	AntialiasDefault AntialiasMode = iota
	AntialiasOn
	AntialiasOff
)

// SnapMode controls whether path vertices are aligned to device pixels.
type SnapMode uint8

const (
	// SnapOff disables path snapping.
	SnapOff SnapMode = iota
	// SnapAuto snaps simple horizontal/vertical paths, matching Matplotlib's
	// default path snap mode when callers opt into it.
	SnapAuto
	// SnapOn forces snapping for all path vertices.
	SnapOn
)

// SketchParams describes Matplotlib-style sketch/jitter rendering.
//
// Backends may ignore this until they implement a sketch pass.
type SketchParams struct {
	Scale      float64
	Length     float64
	Randomness float64
}

// LineJoin controls how path joins are rendered.
type LineJoin uint8

const (
	JoinMiter LineJoin = iota
	JoinRound
	JoinBevel
)

// LineCap controls how path endpoints are rendered.
type LineCap uint8

const (
	CapButt LineCap = iota
	CapRound
	CapSquare
)

// Color is a normalized sRGBA tuple in [0..1].
//
// In this codebase, callers generally provide plotting-style colors directly
// (for example Matplotlib face/edge tuples), so these channels are interpreted
// as display-encoded sRGB values with straight alpha unless a backend states
// otherwise.
type Color struct{ R, G, B, A float64 }

// Premultiply returns a color with RGB components premultiplied by alpha.
//
// The operation is applied to the stored channel values as-is. It does not
// perform any transfer-curve conversion.
func (c Color) Premultiply() Color {
	return Color{
		R: c.R * c.A,
		G: c.G * c.A,
		B: c.B * c.A,
		A: c.A,
	}
}

// ToPremultipliedRGBA converts normalized sRGBA values to 8-bit premultiplied
// RGBA suitable for raster backends in this repository.
func (c Color) ToPremultipliedRGBA() (uint8, uint8, uint8, uint8) {
	premul := c.Premultiply()
	return uint8(premul.R*255 + 0.5),
		uint8(premul.G*255 + 0.5),
		uint8(premul.B*255 + 0.5),
		uint8(premul.A*255 + 0.5)
}

// Glyph represents a single shaped glyph.
type Glyph struct {
	ID      uint32
	Advance float64
	Offset  geom.Pt
}

// GlyphRun represents a run of glyphs to render at a baseline origin.
type GlyphRun struct {
	Glyphs  []Glyph
	Origin  geom.Pt
	Size    float64
	FontKey string
}

// TextMetrics provides basic text measurements.
type TextMetrics struct{ W, H, Ascent, Descent float64 }

// FontHeightMetrics describes font-wide vertical line metrics for a text run.
// These are distinct from the actual ink extents of a specific string.
type FontHeightMetrics struct{ Ascent, Descent, LineGap float64 }

// TextBounds describes the rendered ink bounds of a string relative to the
// baseline origin used for DrawText.
type TextBounds struct{ X, Y, W, H float64 }

// Image is a minimal interface for raster images passed to renderers.
type Image interface {
	Size() (w, h int)
}

// RGBAImage is an optional renderer-facing image interface that exposes direct
// RGBA pixel access. Renderers may use this for efficient conversion and
// image-space transforms.
type RGBAImage interface {
	Image
	RGBA() *image.RGBA
}

// ImageData is a concrete raster image implementation that satisfies both Image
// and RGBAImage. It is the primary raster type used by heatmaps and image
// artists in this package.
type ImageData struct {
	rgba *image.RGBA
}

// NewImageData creates a new ImageData from an RGBA source image.
// A nil image results in an empty image.
func NewImageData(img *image.RGBA) *ImageData {
	if img == nil {
		return &ImageData{rgba: image.NewRGBA(image.Rect(0, 0, 0, 0))}
	}
	return &ImageData{rgba: img}
}

// Size returns the raster dimensions in pixels.
func (i *ImageData) Size() (w, h int) {
	if i == nil || i.rgba == nil {
		return 0, 0
	}
	return i.rgba.Bounds().Dx(), i.rgba.Bounds().Dy()
}

// RGBA returns the backing RGBA image.
func (i *ImageData) RGBA() *image.RGBA {
	if i == nil {
		return nil
	}
	return i.rgba
}

// Renderer defines the core drawing verbs.
type Renderer interface {
	Begin(viewport geom.Rect) error
	End() error

	// State stack
	Save()
	Restore()

	// Clipping
	ClipRect(r geom.Rect)
	ClipPath(p geom.Path)

	// Drawing
	Path(p geom.Path, paint *Paint)
	Image(img Image, dst geom.Rect)
	GlyphRun(run GlyphRun, color Color)
	MeasureText(text string, size float64, fontKey string) TextMetrics
}

// NullRenderer is a no-op renderer used for traversal/tests.
type NullRenderer struct {
	began  bool
	stack  int
	cstack int
}

var _ Renderer = (*NullRenderer)(nil)

// Begin starts a drawing session for the given viewport.
func (n *NullRenderer) Begin(_ geom.Rect) error {
	if n.began {
		return errors.New("Begin called twice")
	}
	n.began = true
	return nil
}

// End ends a drawing session.
func (n *NullRenderer) End() error {
	if !n.began {
		return errors.New("End called before Begin")
	}
	n.began = false
	n.stack = 0
	n.cstack = 0
	return nil
}

// Save pushes state.
func (n *NullRenderer) Save() { n.stack++ }

// Restore pops state; underflow is clamped to zero.
func (n *NullRenderer) Restore() {
	if n.stack > 0 {
		n.stack--
	}
}

// ClipRect pushes a rectangular clip.
func (n *NullRenderer) ClipRect(_ geom.Rect) { n.cstack++ }

// ClipPath pushes a path clip.
func (n *NullRenderer) ClipPath(_ geom.Path) { n.cstack++ }

// Path draws a path using the provided paint; no-op here.
func (n *NullRenderer) Path(_ geom.Path, _ *Paint) {}

// Image draws an image in the destination rectangle; no-op here.
func (n *NullRenderer) Image(_ Image, _ geom.Rect) {}

// GlyphRun draws a run of glyphs with the given color; no-op here.
func (n *NullRenderer) GlyphRun(_ GlyphRun, _ Color) {}

// MeasureText returns zero metrics in the null renderer.
func (n *NullRenderer) MeasureText(_ string, _ float64, _ string) TextMetrics { return TextMetrics{} }

// depth returns the current state stack depth (for tests).
func (n *NullRenderer) depth() int { return n.stack }
