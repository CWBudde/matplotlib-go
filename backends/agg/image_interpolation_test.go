package agg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// renderUpscaledImage builds a tiny 2x2 checkerboard, hands it to an AGG
// renderer with the given interpolation name, and returns the PNG bytes.
func renderUpscaledImage(t *testing.T, interp string) []byte {
	t.Helper()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	src.Set(0, 0, black)
	src.Set(1, 1, black)
	// Other pixels stay zero (transparent).

	raster := render.NewImageData(src)
	raster.SetInterpolation(interp)

	r, err := backends.Create(backends.AGG, backends.Config{Width: 64, Height: 64, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 64, Y: 64}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(raster, geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 64, Y: 64}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	exporter, ok := r.(render.PNGExporter)
	if !ok {
		t.Fatal("AGG renderer should implement render.PNGExporter")
	}
	tmp := t.TempDir() + "/out.png"
	if err := exporter.SavePNG(tmp); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("PNG output is empty")
	}
	return data
}

func TestAggImage_NearestVsBilinear_DifferentBytes(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest")
	pngBilinear := renderUpscaledImage(t, "bilinear")
	if bytes.Equal(pngNearest, pngBilinear) {
		t.Fatal("expected different PNG bytes between nearest and bilinear; interpolation appears to be ignored")
	}
}

func TestAggImage_EmptyInterpolationMatchesNearest(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest")
	pngEmpty := renderUpscaledImage(t, "")

	decoded := func(data []byte) *image.RGBA {
		t.Helper()
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("png.Decode: %v", err)
		}
		if r, ok := img.(*image.RGBA); ok {
			return r
		}
		b := img.Bounds()
		rgba := image.NewRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				rgba.Set(x, y, img.At(x, y))
			}
		}
		return rgba
	}

	a := decoded(pngNearest)
	b := decoded(pngEmpty)
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("empty Interpolation should produce the same pixels as 'nearest'")
	}
}
