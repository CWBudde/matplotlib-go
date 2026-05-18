//go:build skia

package skia

import (
	"errors"
	"image"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/render"
)

// Renderer implements the Skia backend's CPU raster contract.
//
// The current skia-tagged implementation uses the shared Go raster surface as a
// compatibility layer while the external Skia C ABI is still being wired. This
// keeps backend registration, save dispatch, and renderer contract tests
// functional without claiming GPU acceleration or Skia-native optional paths.
type Renderer struct {
	*gobasic.Renderer
	width       int
	height      int
	background  render.Color
	useGPU      bool
	sampleCount int
	colorType   string
}

var (
	_ render.Renderer           = (*Renderer)(nil)
	_ render.DPIAware           = (*Renderer)(nil)
	_ render.TextDrawer         = (*Renderer)(nil)
	_ render.TextPather         = (*Renderer)(nil)
	_ render.RotatedTextDrawer  = (*Renderer)(nil)
	_ render.VerticalTextDrawer = (*Renderer)(nil)
	_ render.RGBAExporter       = (*Renderer)(nil)
	_ render.PNGExporter        = (*Renderer)(nil)
)

// New creates a new Skia renderer with the given configuration.
func New(config backends.Config) (*Renderer, error) {
	skiaConfig, ok := config.Options.(backends.SkiaConfig)
	if !ok {
		// Use defaults if no Skia-specific config provided
		skiaConfig = backends.SkiaConfig{
			UseGPU:      false,
			SampleCount: 1,
			ColorType:   "RGBA8888",
		}
	}
	if skiaConfig.UseGPU {
		return nil, errors.New("skia backend GPU mode is not implemented")
	}
	if config.Width <= 0 || config.Height <= 0 {
		return nil, errors.New("skia backend requires positive width and height")
	}
	if skiaConfig.SampleCount <= 0 {
		skiaConfig.SampleCount = 1
	}
	if skiaConfig.ColorType == "" {
		skiaConfig.ColorType = "RGBA8888"
	}

	cpu := gobasic.New(config.Width, config.Height, config.Background)
	if config.DPI > 0 {
		cpu.SetResolution(uint(config.DPI))
	}
	return &Renderer{
		Renderer:    cpu,
		width:       config.Width,
		height:      config.Height,
		background:  config.Background,
		useGPU:      false,
		sampleCount: skiaConfig.SampleCount,
		colorType:   skiaConfig.ColorType,
	}, nil
}

// GetSurface returns the underlying Skia surface for advanced operations.
func (r *Renderer) GetSurface() interface{} {
	// No Skia-native surface exists until the external C ABI lands.
	return nil
}

// GetImage returns the CPU raster output buffer.
func (r *Renderer) GetImage() *image.RGBA {
	if r == nil || r.Renderer == nil {
		return nil
	}
	return r.Renderer.GetImage()
}

// FlushGPU flushes pending GPU operations (if using GPU backend).
func (r *Renderer) FlushGPU() {
	if !r.useGPU {
		return
	}
	// TODO: Call GrDirectContext::flushAndSubmit()
}

// GPU returns true if this renderer is using GPU acceleration.
func (r *Renderer) GPU() bool {
	return r.useGPU
}

// SampleCount returns the MSAA sample count.
func (r *Renderer) SampleCount() int {
	return r.sampleCount
}

// ColorType returns the configured Skia color type name.
func (r *Renderer) ColorType() string {
	return r.colorType
}

func buildTagAvailable() bool {
	return true
}
