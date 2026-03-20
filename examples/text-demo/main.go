package main

import (
	"fmt"
	"log"

	"matplotlib-go/backends"
	_ "matplotlib-go/backends/all"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

type textRenderer interface {
	render.Renderer
	DrawText(text string, origin geom.Pt, size float64, textColor render.Color)
	SavePNG(path string) error
}

func main() {
	// Create a renderer from the registry with white background
	width, height := 400, 200
	bgColor := render.Color{R: 1, G: 1, B: 1, A: 1} // white
	renderer, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      width,
		Height:     height,
		Background: bgColor,
		DPI:        72.0,
	}, backends.TextCapabilities)
	if createErr != nil {
		log.Fatalf("Failed to create renderer: %v", createErr)
	}

	textRen, ok := renderer.(textRenderer)
	if !ok {
		log.Fatalf("Selected backend does not support direct text drawing")
	}

	// Begin rendering
	viewport := geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: float64(width), Y: float64(height)},
	}
	if err := renderer.Begin(viewport); err != nil {
		log.Fatalf("Failed to begin rendering: %v", err)
	}
	defer renderer.End()

	// Test text measurement
	text := "Hello, matplotlib-go!"
	fontSize := 13.0
	metrics := renderer.MeasureText(text, fontSize, "default")
	fmt.Printf("Text metrics for '%s':\n", text)
	fmt.Printf("  Width: %.2f pixels\n", metrics.W)
	fmt.Printf("  Height: %.2f pixels\n", metrics.H)
	fmt.Printf("  Ascent: %.2f pixels\n", metrics.Ascent)
	fmt.Printf("  Descent: %.2f pixels\n", metrics.Descent)

	// Draw some text
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1} // black

	// Draw text at different positions
	textRen.DrawText("matplotlib-go Text Rendering Demo", geom.Pt{X: 20, Y: 30}, 13, textColor)
	textRen.DrawText("Rendered with DejaVu Sans via AGG", geom.Pt{X: 20, Y: 60}, 13, textColor)
	textRen.DrawText("Supports basic text positioning", geom.Pt{X: 20, Y: 90}, 13, textColor)

	// Draw text with different "sizes" (scaling)
	textRen.DrawText("Small text (size 10)", geom.Pt{X: 20, Y: 120}, 10, textColor)
	textRen.DrawText("Large text (size 16)", geom.Pt{X: 20, Y: 150}, 16, textColor)

	// Draw colored text
	redColor := render.Color{R: 1, G: 0, B: 0, A: 1}
	blueColor := render.Color{R: 0, G: 0, B: 1, A: 1}
	textRen.DrawText("Red text", geom.Pt{X: 250, Y: 120}, 13, redColor)
	textRen.DrawText("Blue text", geom.Pt{X: 250, Y: 150}, 13, blueColor)

	// Save the result as PNG
	if err := textRen.SavePNG("text-demo.png"); err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	fmt.Println("Text rendering demo saved as 'text-demo.png'")
}
