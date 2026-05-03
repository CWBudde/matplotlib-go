package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/unstructured/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := showcase.UnstructuredShowcase()
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

	if err := core.SavePNG(fig, r, "unstructured_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved unstructured_basic.png")
}
