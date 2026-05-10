// CLI runner for the examples/annotation_composition showcase. Renders to annotation_composition.png using the
// backend selected by the BACKEND env var.
package main

import (
	"log"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	example "github.com/cwbudde/matplotlib-go/examples/annotation_composition"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := example.Plot()
	w := int(fig.SizePx.X)
	h := int(fig.SizePx.Y)
	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      w,
		Height:     h,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        fig.RC.DPI,
	}, backends.TextCapabilities)
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}
	if err := core.SavePNG(fig, r, "annotation_composition.png"); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Println("saved annotation_composition.png")
}
