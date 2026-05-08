package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func run(label string, img *render.ImageData) {
	r, err := agg.New(10, 10, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	_ = r
	r.Image(img, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	_ = r
	if err := r.End(); err != nil {
		panic(err)
	}
	fmt.Printf("%s center (5,5): %v\n", label, r.GetImage().RGBAAt(5, 5))
}

func main() {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	full := render.NewImageData(src)
	run("full", full)

	alpha := render.NewImageData(src)
	alpha.SetAlpha(0.5)
	run("half", alpha)

	srcHalf := image.NewRGBA(image.Rect(0, 0, 1, 1))
	srcHalf.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	halfPixel := render.NewImageData(srcHalf)
	run("halfPixel", halfPixel)
}
