// Command example renders any cataloged showcase to a PNG file. It is the
// unified entry point that replaces the per-example runner files.
//
// Usage:
//
//	go run ./cmd/example -list
//	go run ./cmd/example -name basic_line -o /tmp/basic_line.png
//	BACKEND=agg go run ./cmd/example -name polar_axes -o polar.png
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	"github.com/cwbudde/matplotlib-go/render"

	annotation_composition "github.com/cwbudde/matplotlib-go/examples/annotation_composition"
	arrays_showcase "github.com/cwbudde/matplotlib-go/examples/arrays_showcase"
	axes_control_surface "github.com/cwbudde/matplotlib-go/examples/axes_control_surface"
	axes_grid1_showcase "github.com/cwbudde/matplotlib-go/examples/axes_grid1_showcase"
	axisartist_showcase "github.com/cwbudde/matplotlib-go/examples/axisartist_showcase"
	bar_basic "github.com/cwbudde/matplotlib-go/examples/bar_basic"
	basic_line "github.com/cwbudde/matplotlib-go/examples/basic_line"
	boxplot_basic "github.com/cwbudde/matplotlib-go/examples/boxplot_basic"
	colorbar_composition "github.com/cwbudde/matplotlib-go/examples/colorbar_composition"
	dashes "github.com/cwbudde/matplotlib-go/examples/dashes"
	errorbar_basic "github.com/cwbudde/matplotlib-go/examples/errorbar_basic"
	figure_labels_composition "github.com/cwbudde/matplotlib-go/examples/figure_labels_composition"
	fill_basic "github.com/cwbudde/matplotlib-go/examples/fill_basic"
	geo_aitoff_axes "github.com/cwbudde/matplotlib-go/examples/geo_aitoff_axes"
	geo_mollweide_axes "github.com/cwbudde/matplotlib-go/examples/geo_mollweide_axes"
	gridspec_composition "github.com/cwbudde/matplotlib-go/examples/gridspec_composition"
	hist_basic "github.com/cwbudde/matplotlib-go/examples/hist_basic"
	image_heatmap "github.com/cwbudde/matplotlib-go/examples/image_heatmap"
	mesh_contour_tri "github.com/cwbudde/matplotlib-go/examples/mesh_contour_tri"
	mplot3d_terrain "github.com/cwbudde/matplotlib-go/examples/mplot3d_terrain"
	multi_series_basic "github.com/cwbudde/matplotlib-go/examples/multi_series_basic"
	plot_variants "github.com/cwbudde/matplotlib-go/examples/plot_variants"
	polar_axes "github.com/cwbudde/matplotlib-go/examples/polar_axes"
	radar_basic "github.com/cwbudde/matplotlib-go/examples/radar_basic"
	scatter_basic "github.com/cwbudde/matplotlib-go/examples/scatter_basic"
	skewt_basic "github.com/cwbudde/matplotlib-go/examples/skewt_basic"
	specialty_artists "github.com/cwbudde/matplotlib-go/examples/specialty_artists"
	stat_variants "github.com/cwbudde/matplotlib-go/examples/stat_variants"
	units_overview "github.com/cwbudde/matplotlib-go/examples/units_overview"
	unstructured_showcase "github.com/cwbudde/matplotlib-go/examples/unstructured_showcase"
	vector_fields "github.com/cwbudde/matplotlib-go/examples/vector_fields"
)

// registry maps a catalog showcase ID to the Plot() function that builds the
// corresponding *core.Figure. Keep in sync with the Showcase: true rows in
// internal/examplecatalog/catalog.go.
var registry = map[string]func() *core.Figure{
	"annotation_composition":    annotation_composition.Plot,
	"arrays_showcase":           arrays_showcase.Plot,
	"axes_control_surface":      axes_control_surface.Plot,
	"axes_grid1_showcase":       axes_grid1_showcase.Plot,
	"axisartist_showcase":       axisartist_showcase.Plot,
	"bar_basic":                 bar_basic.Plot,
	"basic_line":                basic_line.Plot,
	"boxplot_basic":             boxplot_basic.Plot,
	"colorbar_composition":      colorbar_composition.Plot,
	"dashes":                    dashes.Plot,
	"errorbar_basic":            errorbar_basic.Plot,
	"figure_labels_composition": figure_labels_composition.Plot,
	"fill_basic":                fill_basic.Plot,
	"geo_aitoff_axes":           geo_aitoff_axes.Plot,
	"geo_mollweide_axes":        geo_mollweide_axes.Plot,
	"gridspec_composition":      gridspec_composition.Plot,
	"hist_basic":                hist_basic.Plot,
	"image_heatmap":             image_heatmap.Plot,
	"mesh_contour_tri":          mesh_contour_tri.Plot,
	"mplot3d_terrain":           mplot3d_terrain.Plot,
	"multi_series_basic":        multi_series_basic.Plot,
	"plot_variants":             plot_variants.Plot,
	"polar_axes":                polar_axes.Plot,
	"radar_basic":               radar_basic.Plot,
	"scatter_basic":             scatter_basic.Plot,
	"skewt_basic":               skewt_basic.Plot,
	"specialty_artists":         specialty_artists.Plot,
	"stat_variants":             stat_variants.Plot,
	"units_overview":            units_overview.Plot,
	"unstructured_showcase":     unstructured_showcase.Plot,
	"vector_fields":             vector_fields.Plot,
}

func main() {
	name := flag.String("name", "", "Catalog ID of the showcase to render")
	out := flag.String("o", "", "Output PNG path (default: <name>.png)")
	list := flag.Bool("list", false, "List all available showcase IDs and exit")
	flag.Parse()

	if *list {
		listShowcases()
		return
	}

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required (or use -list)")
		flag.Usage()
		os.Exit(2)
	}

	plot, ok := registry[*name]
	if !ok {
		log.Fatalf("unknown showcase %q (run with -list to see available IDs)", *name)
	}

	output := *out
	if output == "" {
		output = *name + ".png"
	}

	fig := plot()
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
	if err := core.SavePNG(fig, r, output); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Printf("saved %s", output)
}

func listShowcases() {
	cases := examplecatalog.Cases()
	type row struct {
		id, title string
	}
	var rows []row
	for i := range cases {
		c := cases[i]
		if !c.Showcase {
			continue
		}
		if _, ok := registry[c.ID]; !ok {
			continue
		}
		rows = append(rows, row{id: c.ID, title: c.Title})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })

	maxID := 0
	for _, r := range rows {
		if len(r.id) > maxID {
			maxID = len(r.id)
		}
	}
	for _, r := range rows {
		fmt.Printf("%-*s  %s\n", maxID, r.id, r.title)
	}
}
