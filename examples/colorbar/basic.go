package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/colorbar/composition"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := composition.ColorbarComposition()
	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      composition.Width,
		Height:     composition.Height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        composition.DPI,
	}, backends.TextCapabilities)
	if createErr != nil {
		fmt.Printf("error creating renderer: %v\n", createErr)
		return
	}

	if err := core.SavePNG(fig, r, "colorbar_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved colorbar_basic.png")
}
