package examplecatalog

// Case describes a plotting example that participates in parity testing.
//
// The catalog is the shared source of truth for the relationship between
// examples, committed Go goldens, Matplotlib references, and the curated web
// demo subset. Rendering code still lives in the existing example/fixture
// packages while the examples are migrated case by case into importable
// side-by-side Go/Python directories.
type Case struct {
	ID          string
	Topic       string
	Title       string
	Description string
	Width       int
	Height      int
	DPI         int
	GoPath      string
	PythonPath  string
	Optional    bool
	FixtureOnly bool
	WebDemoID   string
}

const (
	DefaultWidth  = 640
	DefaultHeight = 360
	DefaultDPI    = 100
)

var cases = []Case{
	{ID: "basic_line", Topic: "lines", Title: "Basic Line", GoPath: "examples/lines/basic.go", PythonPath: "examples/lines/basic.py"},
	{ID: "joins_caps", Topic: "lines", Title: "Line Joins and Caps", PythonPath: "examples/lines/joins_caps.py"},
	{ID: "dashes", Topic: "lines", Title: "Dash Patterns", GoPath: "examples/lines/dashes/dashes.go", PythonPath: "examples/lines/dash/main.py"},
	{ID: "scatter_basic", Topic: "scatter", Title: "Basic Scatter", GoPath: "examples/scatter/basic.go", PythonPath: "examples/scatter/basic.py"},
	{ID: "scatter_marker_types", Topic: "scatter", Title: "Scatter Marker Types", PythonPath: "examples/scatter/marker_types.py"},
	{ID: "scatter_advanced", Topic: "scatter", Title: "Advanced Scatter", PythonPath: "examples/scatter/advanced.py"},
	{ID: "bar_basic_frame", Topic: "bar", Title: "Bar Frame", PythonPath: "examples/bar/basic_frame.py"},
	{ID: "bar_basic_ticks", Topic: "bar", Title: "Bar Ticks", PythonPath: "examples/bar/basic_ticks.py"},
	{ID: "bar_basic_tick_labels", Topic: "bar", Title: "Bar Tick Labels", PythonPath: "examples/bar/basic_tick_labels.py"},
	{ID: "bar_basic_title", Topic: "bar", Title: "Bar Title", PythonPath: "examples/bar/basic_title.py"},
	{ID: "bar_basic", Topic: "bar", Title: "Basic Bars", GoPath: "examples/bar/basic.go", PythonPath: "examples/bar/basic.py"},
	{ID: "bar_horizontal", Topic: "bar", Title: "Horizontal Bars", GoPath: "examples/bar/horizontal/main.go", PythonPath: "examples/bar/horizontal.py"},
	{ID: "bar_grouped", Topic: "bar", Title: "Grouped Bars", PythonPath: "examples/bar/grouped.py"},
	{ID: "fill_basic", Topic: "fill", Title: "Fill to Baseline", GoPath: "examples/fill/basic.go", PythonPath: "examples/fill/basic.py"},
	{ID: "fill_between", Topic: "fill", Title: "Fill Between Curves", PythonPath: "examples/fill/between.py"},
	{ID: "fill_stacked", Topic: "fill", Title: "Stacked Fill", PythonPath: "examples/fill/stacked.py"},
	{ID: "errorbar_basic", Topic: "errorbar", Title: "Error Bars", GoPath: "examples/errorbar/basic.go", PythonPath: "examples/errorbar/basic.py"},
	{ID: "multi_series_basic", Topic: "multi", Title: "Multiple Series", GoPath: "examples/multi/basic.go", PythonPath: "examples/multi/basic.py"},
	{ID: "multi_series_color_cycle", Topic: "multi", Title: "Color Cycle", PythonPath: "examples/multi/color_cycle.py"},
	{ID: "hist_basic", Topic: "histogram", Title: "Histogram Counts", GoPath: "examples/histogram/basic.go", PythonPath: "examples/histogram/basic.py"},
	{ID: "hist_density", Topic: "histogram", Title: "Histogram Density", PythonPath: "examples/histogram/density.py"},
	{ID: "hist_strategies", Topic: "histogram", Title: "Histogram Strategies", PythonPath: "examples/histogram/strategies.py"},
	{ID: "boxplot_basic", Topic: "boxplot", Title: "Box Plot", GoPath: "examples/boxplot/basic/plot.go", PythonPath: "examples/boxplot/basic/plot.py", Optional: true},
	{ID: "text_labels_strict", Topic: "text", Title: "Strict Text Labels", PythonPath: "examples/text-demo/labels_strict.py", Optional: true},
	{ID: "title_strict", Topic: "text", Title: "Strict Title", PythonPath: "examples/text-demo/title_strict.py"},
	{ID: "image_heatmap", Topic: "image", Title: "Heatmap Image", GoPath: "examples/image/basic.go", PythonPath: "examples/image/basic.py"},
	{ID: "axes_top_right_inverted", Topic: "axes", Title: "Top/Right Inverted Axes", PythonPath: "examples/axes/top_right_inverted.py", Optional: true},
	{ID: "axes_control_surface", Topic: "axes", Title: "Axes, Scales, and Twins", PythonPath: "examples/axes/control_surface.py", Optional: true, WebDemoID: "axes", Description: "Minor ticks, top/right axes, aspect controls, log scale, twin axes, and secondary axes."},
	{ID: "transform_coordinates", Topic: "axes", Title: "Transform Coordinates", PythonPath: "examples/axes/transform_coordinates.py", Optional: true},
	{ID: "gridspec_composition", Topic: "composition", Title: "Figure Composition", GoPath: "examples/gridspec/main.go", PythonPath: "examples/gridspec/main.py", WebDemoID: "composition", Description: "GridSpec spans, figure-level labels, figure legends, anchored text, and colorbars."},
	{ID: "figure_labels_composition", Topic: "composition", Title: "Figure Labels", GoPath: "examples/figure_labels/basic.go", PythonPath: "examples/figure_labels/basic.py"},
	{ID: "colorbar_composition", Topic: "colorbar", Title: "Colorbar Composition", GoPath: "examples/colorbar/composition/composition.go", PythonPath: "examples/colorbar/basic.py"},
	{ID: "annotation_composition", Topic: "annotation", Title: "Annotations", GoPath: "examples/annotation/basic.go", PythonPath: "examples/annotation/basic.py"},
	{ID: "patch_showcase", Topic: "patches", Title: "Patch Showcase", PythonPath: "examples/patch_showcase.py", Optional: true},
	{ID: "mesh_contour_tri", Topic: "mesh", Title: "Meshes and Contours", PythonPath: "examples/mesh_contour_tri.py", Optional: true, WebDemoID: "mesh", Description: "PColorMesh, contour/contourf, Hist2D, triplot, tripcolor, and tricontour."},
	{ID: "plot_variants", Topic: "variants", Title: "Plot Variants", GoPath: "examples/plot_variants/basic.go", PythonPath: "examples/plot_variants/basic.py", Optional: true, WebDemoID: "variants", Description: "Step, stairs, reference lines, spans, broken bars, and stacked bars."},
	{ID: "stat_variants", Topic: "statistics", Title: "Statistical Views", GoPath: "examples/stat_variants/basic.go", PythonPath: "examples/stat_variants/basic.py", Optional: true, WebDemoID: "statistics", Description: "Box plots, violin plots, empirical CDFs, and stack plots."},
	{ID: "stem_plot", Topic: "specialty", Title: "Stem Plot", PythonPath: "examples/stem_plot.py", Optional: true},
	{ID: "specialty_artists", Topic: "specialty", Title: "Specialty Artists", GoPath: "examples/specialty/main.go", PythonPath: "examples/specialty/main.py", Optional: true, WebDemoID: "specialty", Description: "Event plots, hexbin, pie charts, stem plots, tables, and Sankey-style flows."},
	{ID: "units_overview", Topic: "units", Title: "Dates and Categories", GoPath: "examples/units/basic.go", PythonPath: "examples/units/basic.py", Optional: true, WebDemoID: "units", Description: "Time-aware axes, categorical bars, and horizontal categorical bars."},
	{ID: "units_dates", Topic: "units", Title: "Date Units", PythonPath: "examples/units/dates.py", Optional: true},
	{ID: "units_categories", Topic: "units", Title: "Category Units", PythonPath: "examples/units/categories.py", Optional: true},
	{ID: "units_custom_converter", Topic: "units", Title: "Custom Unit Converter", PythonPath: "examples/units/custom_converter.py", Optional: true},
	{ID: "vector_fields", Topic: "vectors", Title: "Vector Fields", PythonPath: "examples/vector_fields.py", Optional: true, WebDemoID: "vectors", Description: "Quiver, quiver keys, barbs, streamplots, and grid-based vector input."},
	{ID: "polar_axes", Topic: "polar", Title: "Polar Wave", GoPath: "examples/polar/basic.go", PythonPath: "examples/polar/basic.py", WebDemoID: "polar", Description: "A filled polar curve with custom radial and angular grid styling."},
	{ID: "geo_mollweide_axes", Topic: "geo", Title: "Projections and Insets", GoPath: "examples/geo/mollweide.go", PythonPath: "examples/geo/mollweide.py", WebDemoID: "projections", Description: "Mollweide geo projection plus a zoomed inset axes."},
	{ID: "unstructured_showcase", Topic: "unstructured", Title: "Unstructured Showcase", GoPath: "examples/unstructured/showcase/showcase.go", PythonPath: "examples/unstructured/basic.py", Optional: true},
	{ID: "arrays_showcase", Topic: "arrays", Title: "Matrix Helpers", GoPath: "examples/arrays/showcase/showcase.go", PythonPath: "examples/arrays/basic.py", Optional: true, WebDemoID: "matrix", Description: "MatShow, sparsity spy plots, annotated heatmaps, and colorbars.", Width: 1240, Height: 620},
	{ID: "axisartist_showcase", Topic: "axisartist", Title: "AxisArtist Showcase", GoPath: "examples/axisartist/showcase/showcase.go", PythonPath: "examples/axisartist/basic.py", Optional: true},
	{ID: "axes_grid1_showcase", Topic: "axes_grid1", Title: "Axes Grid1 Showcase", GoPath: "examples/axes_grid1/showcase/showcase.go", PythonPath: "examples/axes_grid1/basic.py", Optional: true},
	{ID: "rendereragg_large_scatter", Topic: "rendereragg", Title: "RendererAgg Marker Batch", FixtureOnly: true},
	{ID: "rendereragg_mixed_collection", Topic: "rendereragg", Title: "RendererAgg Mixed Path Collection", FixtureOnly: true},
	{ID: "rendereragg_quad_mesh", Topic: "rendereragg", Title: "RendererAgg Quad Mesh", FixtureOnly: true},
	{ID: "rendereragg_gouraud_triangles", Topic: "rendereragg", Title: "RendererAgg Gouraud Triangles", FixtureOnly: true},
}

// Cases returns every cataloged parity example/fixture.
func Cases() []Case {
	out := make([]Case, len(cases))
	copy(out, cases)
	for i := range out {
		applyDefaults(&out[i])
	}
	return out
}

// Lookup finds a parity case by canonical case ID.
func Lookup(id string) (Case, bool) {
	for _, c := range Cases() {
		if c.ID == id {
			return c, true
		}
	}
	return Case{}, false
}

// WebDemos returns the curated web-demo subset in display order.
func WebDemos() []Case {
	var out []Case
	for _, c := range Cases() {
		if c.WebDemoID != "" {
			out = append(out, c)
		}
	}
	return out
}

// LookupWebDemo finds a catalog case by browser demo ID.
func LookupWebDemo(id string) (Case, bool) {
	for _, c := range WebDemos() {
		if c.WebDemoID == id {
			return c, true
		}
	}
	return Case{}, false
}

func applyDefaults(c *Case) {
	if c.Width == 0 {
		c.Width = DefaultWidth
	}
	if c.Height == 0 {
		c.Height = DefaultHeight
	}
	if c.DPI == 0 {
		c.DPI = DefaultDPI
	}
	if c.Description == "" {
		c.Description = c.Title
	}
}
