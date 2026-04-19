package main

import (
	"fmt"
	"image/png"
	"os"

	agglib "github.com/cwbudde/agg_go"
)

func main() {
	// Exact replication of dashes golden test at 640x360 with small dash values
	const (
		width  = 640
		height = 360
	)

	img := agglib.NewImage(make([]uint8, width*height*4), width, height, width*4)
	agg := agglib.NewAgg2D()
	agg.AttachImage(img)
	agg.ClearAll(agglib.NewColor(255, 255, 255, 255))
	agg.ClipBox(0, 0, width, height)

	// Solid black (y=4, from pixel ~114 to ~525 based on axes 10-90% of 640, ylim 0-5)
	agg.LineColor(agglib.Color{R: 0, G: 0, B: 0, A: 255})
	agg.LineWidth(3.0)
	agg.RemoveAllDashes()
	agg.ResetPath()
	agg.MoveTo(114, 80)
	agg.LineTo(525, 80)
	agg.DrawPath(agglib.StrokeOnly)

	// Red {5, 2} dashes (y=3)
	agg.LineColor(agglib.Color{R: 204, G: 0, B: 0, A: 255})
	agg.LineWidth(3.0)
	agg.RemoveAllDashes()
	agg.AddDash(5, 2)
	agg.ResetPath()
	agg.MoveTo(114, 151)
	agg.LineTo(525, 151)
	agg.DrawPath(agglib.StrokeOnly)
	agg.RemoveAllDashes()

	f, _ := os.Create("/tmp/test_small_dashes.png")
	png.Encode(f, img.ToGoImage())
	f.Close()
	fmt.Println("Small dashes (5,2) - 640x360")
}
