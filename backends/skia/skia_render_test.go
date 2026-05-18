//go:build skia

package skia

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSkiaTaggedRendererImplementsCPUBaseContract(t *testing.T) {
	r, err := New(backends.Config{
		Width:      32,
		Height:     24,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		Options:    backends.SkiaConfig{UseGPU: false, SampleCount: 1, ColorType: "RGBA8888"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if r.GPU() {
		t.Fatal("CPU config should not enable GPU mode")
	}
	if r.SampleCount() != 1 {
		t.Fatalf("SampleCount() = %d, want 1", r.SampleCount())
	}

	viewport := geom.Rect{Max: geom.Pt{X: 32, Y: 24}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	r.Save()
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 4, Y: 4}, Max: geom.Pt{X: 28, Y: 20}})
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 32, Y: 0}, {X: 32, Y: 24}, {X: 0, Y: 24}},
	}, &render.Paint{
		Fill: render.Color{R: 0.1, G: 0.2, B: 0.8, A: 1},
	})
	r.Restore()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	img.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	img.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})
	r.Image(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 12, Y: 8}, Max: geom.Pt{X: 20, Y: 16}})

	if err := r.End(); err != nil {
		t.Fatalf("End() error = %v", err)
	}

	out := r.GetImage()
	if out == nil {
		t.Fatal("GetImage() returned nil")
	}
	if got := out.RGBAAt(1, 1); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("pixel outside clip = %#v, want unchanged white", got)
	}
	if got := out.RGBAAt(6, 6); got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("pixel inside clip stayed background: %#v", got)
	}
	if got := out.RGBAAt(14, 10); got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("image draw pixel stayed background: %#v", got)
	}

	path := filepath.Join(t.TempDir(), "skia.png")
	if err := r.SavePNG(path); err != nil {
		t.Fatalf("SavePNG() error = %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open saved PNG: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode saved PNG: %v", err)
	}
	if decoded.Bounds().Dx() != 32 || decoded.Bounds().Dy() != 24 {
		t.Fatalf("saved PNG bounds = %v, want 32x24", decoded.Bounds())
	}
}

func TestSkiaTaggedRegistryAdvertisesImplementedCPUCapabilities(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.Skia)
	if !ok {
		t.Fatal("skia backend should be registered")
	}
	if !info.Available {
		t.Fatal("skia build-tag backend should be available")
	}
	if _, ok := info.SaveFormats[".png"]; !ok {
		t.Fatal("skia build-tag backend should advertise PNG save dispatch")
	}

	renderer, err := backends.Create(backends.Skia, backends.TestDefaultConfig(32, 24))
	if err != nil {
		t.Fatalf("Create(skia) error = %v", err)
	}
	if err := backends.VerifyRendererCapabilities(backends.Skia, renderer); err != nil {
		t.Fatalf("VerifyRendererCapabilities(skia) error = %v", err)
	}

	for _, cap := range []backends.Capability{
		backends.AntiAliasing,
		backends.PathClip,
		backends.DPIAware,
		backends.TextShaping,
		backends.TextPathing,
		backends.RotatedText,
		backends.VerticalText,
		backends.RGBABuffer,
		backends.PNGExport,
	} {
		if !backends.HasCapability(backends.Skia, cap) {
			t.Fatalf("skia backend should advertise %s", cap)
		}
	}
	if backends.HasCapability(backends.Skia, backends.GPUAccel) {
		t.Fatal("CPU skia backend should not advertise GPU acceleration")
	}
}
