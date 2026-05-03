package main

import (
	"flag"
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/arrays/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	out := flag.String("out", "arrays_basic.png", "output PNG path")
	flag.Parse()

	fig := showcase.ArraysShowcase()
	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      showcase.Width,
		Height:     showcase.Height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        showcase.DPI,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, *out); err != nil {
		fmt.Printf("error saving figure: %v\n", err)
		return
	}

	fmt.Printf("saved %s\n", *out)
}
