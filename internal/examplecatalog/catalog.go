package examplecatalog

// Case describes a plotting example that participates in parity testing.
//
// The catalog is the shared source of truth for the relationship between
// examples, committed Go goldens, Matplotlib references, and the curated web
// demo subset. GoPath and PythonPath are normalized to the canonical
// test/parity/<case-id>/ source files.
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
	{ID: "basic_line", Topic: "lines", Title: "Basic Line"},
	{ID: "joins_caps", Topic: "lines", Title: "Line Joins and Caps"},
	{ID: "dashes", Topic: "lines", Title: "Dash Patterns"},
	{ID: "scatter_basic", Topic: "scatter", Title: "Basic Scatter"},
	{ID: "scatter_marker_types", Topic: "scatter", Title: "Scatter Marker Types"},
	{ID: "scatter_advanced", Topic: "scatter", Title: "Advanced Scatter"},
	{ID: "bar_basic_frame", Topic: "bar", Title: "Bar Frame"},
	{ID: "bar_basic_ticks", Topic: "bar", Title: "Bar Ticks"},
	{ID: "bar_basic_tick_labels", Topic: "bar", Title: "Bar Tick Labels"},
	{ID: "bar_basic_title", Topic: "bar", Title: "Bar Title"},
	{ID: "bar_basic", Topic: "bar", Title: "Basic Bars"},
	{ID: "bar_horizontal", Topic: "bar", Title: "Horizontal Bars"},
	{ID: "bar_grouped", Topic: "bar", Title: "Grouped Bars"},
	{ID: "fill_basic", Topic: "fill", Title: "Fill to Baseline"},
	{ID: "fill_between", Topic: "fill", Title: "Fill Between Curves"},
	{ID: "fill_stacked", Topic: "fill", Title: "Stacked Fill"},
	{ID: "errorbar_basic", Topic: "errorbar", Title: "Error Bars"},
	{ID: "multi_series_basic", Topic: "multi", Title: "Multiple Series"},
	{ID: "multi_series_color_cycle", Topic: "multi", Title: "Color Cycle"},
	{ID: "hist_basic", Topic: "histogram", Title: "Histogram Counts"},
	{ID: "hist_density", Topic: "histogram", Title: "Histogram Density"},
	{ID: "hist_strategies", Topic: "histogram", Title: "Histogram Strategies"},
	{ID: "boxplot_basic", Topic: "boxplot", Title: "Box Plot", Optional: true},
	{ID: "text_labels_strict", Topic: "text", Title: "Strict Text Labels", Optional: true},
	{ID: "title_strict", Topic: "text", Title: "Strict Title"},
	{ID: "image_heatmap", Topic: "image", Title: "Heatmap Image"},
	{ID: "imshow_clipped", Topic: "image", Title: "Clipped Imshow", FixtureOnly: true},
	{ID: "imshow_transformed", Topic: "image", Title: "Transformed Imshow", FixtureOnly: true, Width: 420, Height: 420},
	{ID: "imshow_bilinear", Topic: "image", Title: "Bilinear Imshow", FixtureOnly: true, Width: 256, Height: 256},
	{ID: "imshow_bicubic", Topic: "image", Title: "Bicubic Imshow", FixtureOnly: true, Width: 256, Height: 256},
	{ID: "image_alpha", Topic: "image", Title: "Image Alpha", FixtureOnly: true},
	{ID: "matshow_basic", Topic: "image", Title: "Matshow", FixtureOnly: true},
	{ID: "spy_marker", Topic: "image", Title: "Spy Marker Mode", FixtureOnly: true},
	{ID: "spy_image", Topic: "image", Title: "Spy Image Mode", FixtureOnly: true},
	{ID: "axes_top_right_inverted", Topic: "axes", Title: "Top/Right Inverted Axes", Optional: true},
	{ID: "axes_control_surface", Topic: "axes", Title: "Axes, Scales, and Twins", Optional: true, WebDemoID: "axes", Description: "Minor ticks, top/right axes, aspect controls, log scale, twin axes, and secondary axes."},
	{ID: "transform_coordinates", Topic: "axes", Title: "Transform Coordinates", Optional: true},
	{ID: "gridspec_composition", Topic: "composition", Title: "Figure Composition", WebDemoID: "composition", Description: "GridSpec spans, figure-level labels, figure legends, anchored text, and colorbars."},
	{ID: "figure_labels_composition", Topic: "composition", Title: "Figure Labels"},
	{ID: "colorbar_composition", Topic: "colorbar", Title: "Colorbar Composition"},
	{ID: "annotation_composition", Topic: "annotation", Title: "Annotations"},
	{ID: "patch_showcase", Topic: "patches", Title: "Patch Showcase", Optional: true},
	{ID: "mesh_contour_tri", Topic: "mesh", Title: "Meshes and Contours", Optional: true, WebDemoID: "mesh", Description: "PColorMesh, contour/contourf, Hist2D, triplot, tripcolor, and tricontour."},
	{ID: "plot_variants", Topic: "variants", Title: "Plot Variants", Optional: true, WebDemoID: "variants", Description: "Step, stairs, reference lines, spans, broken bars, and stacked bars."},
	{ID: "spectrum_variants", Topic: "signal", Title: "Spectrum Variants", FixtureOnly: true},
	{ID: "stat_variants", Topic: "statistics", Title: "Statistical Views", Optional: true, WebDemoID: "statistics", Description: "Box plots, violin plots, empirical CDFs, and stack plots."},
	{ID: "phase12_specialty_depth", Topic: "statistics", Title: "Phase 12 Specialty Depth", FixtureOnly: true},
	{ID: "stem_plot", Topic: "specialty", Title: "Stem Plot", Optional: true},
	{ID: "specialty_artists", Topic: "specialty", Title: "Specialty Artists", Optional: true, WebDemoID: "specialty", Description: "Event plots, hexbin, pie charts, stem plots, tables, and Sankey-style flows."},
	{ID: "units_overview", Topic: "units", Title: "Dates and Categories", Optional: true, WebDemoID: "units", Description: "Time-aware axes, categorical bars, and horizontal categorical bars."},
	{ID: "units_dates", Topic: "units", Title: "Date Units", Optional: true},
	{ID: "units_categories", Topic: "units", Title: "Category Units", Optional: true},
	{ID: "units_custom_converter", Topic: "units", Title: "Custom Unit Converter", Optional: true},
	{ID: "vector_fields", Topic: "vectors", Title: "Vector Fields", Optional: true, WebDemoID: "vectors", Description: "Quiver, quiver keys, barbs, streamplots, and grid-based vector input."},
	{ID: "polar_axes", Topic: "polar", Title: "Polar Wave", WebDemoID: "polar", Description: "A filled polar curve with custom radial and angular grid styling."},
	{ID: "geo_mollweide_axes", Topic: "geo", Title: "Projections and Insets", WebDemoID: "projections", Description: "Mollweide geo projection plus a zoomed inset axes."},
	{ID: "geo_aitoff_axes", Topic: "geo", Title: "Aitoff Projection", Optional: true},
	{ID: "geo_hammer_axes", Topic: "geo", Title: "Hammer Projection", Optional: true},
	{ID: "geo_lambert_axes", Topic: "geo", Title: "Lambert Projection", Optional: true},
	{ID: "radar_basic", Topic: "radar", Title: "Radar Projection", Optional: true},
	{ID: "skewt_basic", Topic: "skewt", Title: "Skew-T Projection", Optional: true},
	{ID: "mplot3d_basic", Topic: "mplot3d", Title: "3D Toolkit Scaffold", Optional: true},
	{ID: "mplot3d_terrain", Topic: "mplot3d", Title: "3D Terrain", Optional: true, Width: 900, Height: 640},
	{ID: "mplot3d_plot3d", Topic: "mplot3d", Title: "3D Plot", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_scatter3d", Topic: "mplot3d", Title: "3D Scatter", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_surface3d", Topic: "mplot3d", Title: "3D Surface", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_wire3d", Topic: "mplot3d", Title: "3D Wireframe", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_trisurf3d", Topic: "mplot3d", Title: "3D Triangulated Surface", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_bar3d", Topic: "mplot3d", Title: "3D Bars", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_voxels", Topic: "mplot3d", Title: "3D Voxels", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_quiver3d", Topic: "mplot3d", Title: "3D Quiver", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_stem3d", Topic: "mplot3d", Title: "3D Stem", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "mplot3d_fill_between3d", Topic: "mplot3d", Title: "3D Fill Between", FixtureOnly: true, Width: 720, Height: 560},
	{ID: "unstructured_showcase", Topic: "unstructured", Title: "Unstructured Showcase", Optional: true},
	{ID: "arrays_showcase", Topic: "arrays", Title: "Matrix Helpers", Optional: true, WebDemoID: "matrix", Description: "MatShow, sparsity spy plots, annotated heatmaps, and colorbars.", Width: 1240, Height: 620},
	{ID: "axisartist_showcase", Topic: "axisartist", Title: "AxisArtist Showcase", Optional: true},
	{ID: "axes_grid1_showcase", Topic: "axes_grid1", Title: "Axes Grid1 Showcase", Optional: true},
	{ID: "pcolor_flat", Topic: "mesh", Title: "PColor Flat", FixtureOnly: true},
	{ID: "pcolormesh_nearest", Topic: "mesh", Title: "PColorMesh Nearest", FixtureOnly: true},
	{ID: "pcolormesh_gouraud", Topic: "mesh", Title: "PColorMesh Gouraud", FixtureOnly: true},
	{ID: "pcolormesh_masked", Topic: "mesh", Title: "PColorMesh Masked", FixtureOnly: true},
	{ID: "hist2d_weighted_density", Topic: "mesh", Title: "Hist2D Weighted Density", FixtureOnly: true},
	{ID: "boundarynorm_pcolormesh", Topic: "colorbar", Title: "BoundaryNorm PColorMesh", FixtureOnly: true},
	{ID: "lognorm_imshow", Topic: "colorbar", Title: "LogNorm Imshow", FixtureOnly: true},
	{ID: "twoslope_norm_image", Topic: "colorbar", Title: "TwoSlopeNorm Image", FixtureOnly: true},
	{ID: "colorbar_extensions", Topic: "colorbar", Title: "Colorbar Extensions", FixtureOnly: true},
	{ID: "large_scatter", Topic: "raster", Title: "Large Scatter Batch", FixtureOnly: true},
	{ID: "mixed_collection", Topic: "raster", Title: "Mixed Path Collection", FixtureOnly: true},
	{ID: "quad_mesh", Topic: "raster", Title: "Quad Mesh Batch", FixtureOnly: true},
	{ID: "gouraud_triangles", Topic: "raster", Title: "Gouraud Triangles", FixtureOnly: true},
	{ID: "clip_path_batch", Topic: "raster", Title: "Clip Path Batch", FixtureOnly: true},
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
	if c.ID != "" {
		if c.GoPath == "" {
			if c.FixtureOnly {
				// Fixture-only cases live exclusively under test/parity/ and are
				// not surfaced as user-facing showcase examples.
				c.GoPath = "test/parity/" + c.ID + "/plot.go"
			} else {
				c.GoPath = "examples/" + c.ID + "/example.go"
			}
		}
		c.PythonPath = "test/parity/" + c.ID + "/plot.py"
	}
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
