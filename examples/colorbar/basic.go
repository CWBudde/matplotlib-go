package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	// Python reference:
	//     fig, ax = plt.subplots(figsize=(1000 / DPI, 700 / DPI), dpi=DPI,
	//                             constrained_layout=True)
	fig := core.NewFigure(1000, 700)
	fig.ConstrainedLayout()
	ax := fig.AddSubplot(1, 1, 1)

	// Python reference:
	//     rows, cols = 80, 120
	//     data = np.zeros((rows, cols))
	//     for row in range(rows):
	//         for col in range(cols):
	//             x = (col / (cols - 1)) * 4 - 2
	//             y = (row / (rows - 1)) * 4 - 2
	//             r = math.hypot(x, y)
	//             data[row, col] = math.sin(3 * r) * math.exp(-0.6 * r)
	const (
		rows = 80
		cols = 120
	)
	data := make([][]float64, rows)
	for row := 0; row < rows; row++ {
		data[row] = make([]float64, cols)
		for col := 0; col < cols; col++ {
			x := (float64(col)/float64(cols-1))*4 - 2
			y := (float64(row)/float64(rows-1))*4 - 2
			r := math.Hypot(x, y)
			data[row][col] = math.Sin(3*r) * math.Exp(-0.6*r)
		}
	}

	// Python reference:
	//     im = ax.imshow(data, cmap="inferno", origin="lower",
	//                    extent=[0, cols, 0, rows], aspect="auto")
	//
	// Image() is the lower-level Go equivalent of imshow here. The explicit
	// extent and lower origin mirror the Python call; aspect="auto" is the
	// default axes aspect, so no extra SetAspect call is needed.
	cmap := "inferno"
	img := ax.Image(data, core.ImageOptions{
		Colormap: &cmap,
		XMin:     ptr(0.0),
		XMax:     ptr(float64(cols)),
		YMin:     ptr(0.0),
		YMax:     ptr(float64(rows)),
		Origin:   core.ImageOriginLower,
	})

	// Python reference:
	//     ax.set_title("Heatmap with Colorbar")
	//     ax.set_xlabel("x")
	//     ax.set_ylabel("y")
	//     ax.set_xlim(0, cols)
	//     ax.set_ylim(0, rows)
	//     ax.set_yticks(np.arange(0, rows + 1, 20))
	ax.SetTitle("Heatmap with Colorbar")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(0, cols)
	ax.SetYLim(0, rows)
	ax.YAxis.Locator = core.FixedLocator{TicksList: []float64{0, 20, 40, 60, 80}}

	// Python reference:
	//     ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
		grid.LineWidth = 0.5
	}

	// Python reference:
	//     cbar = fig.colorbar(im, ax=ax)
	//     cbar.set_label("Intensity")
	fig.AddColorbar(ax, img, core.ColorbarOptions{Label: "Intensity"})

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      1000,
		Height:     700,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
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

func ptr[T any](v T) *T {
	return &v
}
