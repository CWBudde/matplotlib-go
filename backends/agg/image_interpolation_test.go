package agg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// renderUpscaledImage builds a tiny 2x2 checkerboard, hands it to an AGG
// renderer with the given interpolation name, and returns the PNG bytes.
func renderUpscaledImage(t *testing.T, interp string, dstW, dstH int) []byte {
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
	r.Image(raster, geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: float64(dstW), Y: float64(dstH)}})
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
	pngNearest := renderUpscaledImage(t, "nearest", 64, 64)
	pngBilinear := renderUpscaledImage(t, "bilinear", 64, 64)
	if bytes.Equal(pngNearest, pngBilinear) {
		t.Fatal("expected different PNG bytes between nearest and bilinear; interpolation appears to be ignored")
	}
}

func TestAggImage_EmptyInterpolationMatchesNearest(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 64, 64)
	pngEmpty := renderUpscaledImage(t, "", 64, 64)

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

func TestAggImage_AutoInterpolationMatchesNearestForIntegerScale(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 4, 4)
	pngAuto := renderUpscaledImage(t, "auto", 4, 4)
	if !bytes.Equal(pngNearest, pngAuto) {
		t.Fatal("auto should resolve to nearest on integer-scale transforms")
	}
}

func TestAggImage_AutoInterpolationUsesHanningForNonIntegerScale(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 3, 3)
	pngAuto := renderUpscaledImage(t, "auto", 3, 3)
	if bytes.Equal(pngNearest, pngAuto) {
		t.Fatal("auto should prefer hanning for non-integer small upscales")
	}
}

func TestImageTransformDisplaySpan(t *testing.T) {
	raster := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 2, 3)))
	spanX, spanY := imageTransformDisplaySpan(raster, geom.Affine{
		A: 2,
		B: 1,
		C: 1,
		D: 2,
		E: 0,
		F: 0,
	})
	if spanX != 7 || spanY != 8 {
		t.Fatalf("imageTransformDisplaySpan = (%g, %g), want (7, 8)", spanX, spanY)
	}

	spanX, spanY = imageTransformDisplaySpan(render.NewImageData(image.NewRGBA(image.Rect(0, 0, 0, 0))), geom.Affine{A: 2})
	if spanX != 0 || spanY != 0 {
		t.Fatalf("empty image span should be zero, got (%g, %g)", spanX, spanY)
	}

	spanX, spanY = imageTransformDisplaySpan(raster, geom.Affine{
		A: 1,
		B: 0,
		C: -1,
		D: 1,
	})
	if spanX != 5 || spanY != 3 {
		t.Fatalf("imageTransformDisplaySpan with opposing sign axes should be (5,3), got (%g, %g)", spanX, spanY)
	}

	spanX, spanY = imageTransformDisplaySpan(raster, geom.Affine{
		A: 0,
		B: 1,
		C: -1,
		D: 0,
		E: 0,
		F: 0,
	})
	if spanX != 3 || spanY != 2 {
		t.Fatalf("imageTransformDisplaySpan with 90° rotation should be (3,2), got (%g, %g)", spanX, spanY)
	}
}

func TestAggImageRespectsImageAlphaState(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	data := render.NewImageData(src)
	data.SetAlpha(0.5)

	r, err := backends.Create(backends.AGG, backends.Config{Width: 10, Height: 10, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(data, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	aggR, ok := r.(*Renderer)
	if !ok {
		t.Fatalf("expected *agg.Renderer, got %T", r)
	}

	got := aggR.GetImage().RGBAAt(0, 0)
	if got.A != 255 {
		t.Fatalf("composited alpha = %d, want 255", got.A)
	}
	if got.R != 255 || math.Abs(float64(got.G)-128) > 2 || math.Abs(float64(got.B)-128) > 2 {
		t.Fatalf("expected red with 0.5 image alpha, got %+v", got)
	}
}
