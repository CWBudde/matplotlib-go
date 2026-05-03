package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	boxplotbasic "github.com/cwbudde/matplotlib-go/examples/boxplot/basic"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := boxplotbasic.Build()
	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      boxplotbasic.Width,
		Height:     boxplotbasic.Height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        boxplotbasic.DPI,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}

	if err := core.SavePNG(fig, r, "boxplot_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved boxplot_basic.png")
}
