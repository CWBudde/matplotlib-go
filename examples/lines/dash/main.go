// Package demonstrates dash patterns with matplotlib-go.
package main

import (
	"log"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/examples/lines/dashes"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := dashes.Dashes()
	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      dashes.Width,
		Height:     dashes.Height,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        dashes.DPI,
	}, backends.TextCapabilities)
	if createErr != nil {
		log.Fatalf("Failed to create renderer: %v", createErr)
	}

	err := core.SavePNG(fig, r, "dashes.png")
	if err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	log.Println("saved dashes.png")
}
