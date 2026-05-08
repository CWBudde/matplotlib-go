package agg

import (
	"image"
	"image/color"
	"testing"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestDebugImageBlendColorBehavior(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	r := mustNew(t, 10, 10)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	defer func() {
		if err := r.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
	}()

	img := render.NewImageData(src)
	r.Image(img, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	full := r.GetImage().RGBAAt(5, 5)
	t.Logf("full=%v", full)

	r2 := mustNew(t, 10, 10)
	_ = r2.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	defer func() {
		if err := r2.End(); err != nil {
			t.Fatalf("End failed: %v", err)
		}
	}()

	img2 := render.NewImageData(src)
	img2.SetAlpha(0.5)
	r2.ctx.SetImageBlendColor(agglib.NewColor(255, 255, 255, 128))
	r2.ctx.SetImageBlendMode(agglib.BlendSrcOver)
	r2.Image(img2, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	half := r2.GetImage().RGBAAt(5, 5)

	t.Logf("full=%v half=%v", full, half)
}
