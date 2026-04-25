# Matplotlib-Go Development Plan

This plan prioritizes getting useful plotting functionality working quickly. The AGG backend (via `github.com/cwbudde/agg_go`) is now available and provides a high-quality AGG-backed raster renderer, while GoBasic remains the default pure-Go backend.

This roadmap is also now cross-checked against the local upstream Matplotlib checkout in `/tmp/matplotlib` so uncovered areas are tracked explicitly instead of being left in broad "future work" buckets.

---

# ✅ Foundation (COMPLETED)

**What we have working:**

- ✅ Artist hierarchy (Figure→Axes→Artists) with proper traversal
- ✅ Transform system (Linear/Log scales, data→pixel transforms)
- ✅ GoBasic renderer using `golang.org/x/image/vector` (PoC)
- ✅ **AGG renderer** using `github.com/cwbudde/agg_go` — anti-aliased, sub-pixel accurate
- ✅ Line2D artist with stroke support (joins, caps, dashes)
- ✅ Golden image testing infrastructure
- ✅ Working example: `examples/lines/basic.go` produces clean line plots
- ✅ Backend registry system with GoBasic and AGG backends registered

**Current capabilities:**

```go
// AGG backend with anti-aliased rendering
r, err := agg.New(800, 500, render.Color{R: 1, G: 1, B: 1, A: 1})
if err != nil { log.Fatal(err) }
core.SavePNG(fig, r, "output.png")
```

---

# ✅ Phase 1: Core Plot Types (COMPLETED)

**Goal:** Get the most commonly used plot types working

### 1.1 Scatter Plots

- [x] `Scatter2D` artist with point/marker rendering
- [x] Basic marker shapes: circle, square, triangle, diamond, plus, cross
- [x] Variable marker sizes and colors per point
- [x] Edge colors and stroke support for marker outlines
- [x] Alpha transparency support
- [x] Proper bounds calculation
- [x] Comprehensive unit tests and golden tests
- [x] Example: `examples/scatter/basic.go`

### 1.2 Bar Charts

- [x] `Bar2D` artist using rectangle patches
- [x] Vertical and horizontal bars
- [x] Grouped bars (multiple series)
- [x] Comprehensive unit tests and golden tests
- [x] Edge colors and transparency support
- [x] Variable bar widths and colors per bar
- [x] Proper bounds calculation
- [x] Example: `examples/bar/basic.go`

### 1.3 Fill Operations

- [x] `Fill2D` artist for area plots and fill_between
- [x] Alpha transparency support
- [x] Edge colors and stroke support for fill outlines
- [x] Multiple fill regions on same axes
- [x] Proper bounds calculation
- [x] Comprehensive unit tests and golden tests
- [x] Performance optimization for large datasets
- [x] Example: `examples/fill/basic.go`

### 1.4 Multiple Series Support

- [x] Plot multiple lines/scatter/bars on same axes
- [x] Automatic color cycling for series
- [x] Series labels for legend preparation
- [x] Example: `examples/multi/basic.go`

---

# Phase 2: Axes, Scales, Ticks & Transforms

**Goal:** Move from the current rectilinear subset to a broader Matplotlib-compatible axis system while preserving the existing AGG-backed quality work

### 2.1 AGG Backend Integration ✅

- [x] Add `github.com/cwbudde/agg_go` v0.2.2 dependency
- [x] Implement `render.Renderer` interface using an AGG-backed raster backend
- [x] Path rendering (fill + stroke) with proper joins, caps, dashes
- [x] Move AGG backend drawing policy into backend-owned helpers instead of relying broadly on `Agg2D`
- [x] Add shared optional renderer extension interfaces in `render/` for text, transformed images, DPI, and PNG export
- [x] Register AGG backend in the backend registry
- [x] AGG backend unit tests
- [x] Example: `examples/agg-demo/main.go`

### 2.2 Update Golden Tests for AGG ✅

- [x] Regenerate golden images using AGG backend
- [x] Verify anti-aliased output quality vs GoBasic
- [x] Update test infrastructure to use AGG backend (all 14 golden tests migrated)
- [x] Add incremental bar-baseline parity fixtures (`bar_basic_frame` → `bar_basic_ticks` → `bar_basic_tick_labels` → `bar_basic_title`)
- [x] Commit `testdata/matplotlib_ref` fixtures and document regeneration via `test/matplotlib_ref/generate.py`
- [x] Refresh golden/reference baselines after the AGG text-path refactor
- [x] Re-run Matplotlib reference comparisons and either accept the new baseline or tighten text/layout parity

### 2.3 Axis Rendering with AGG ✅

- [x] Draw actual axis lines (spines) using AGG's anti-aliased lines
- [x] Tick marks (major/minor) positioned correctly
- [x] Use existing LinearLocator/LogLocator for tick placement
- [x] MinorLinearLocator for subdividing between major ticks
- [x] Example: `examples/axes/spines/main.go`

### 2.4 Grid Lines ✅

- [x] Major and minor grid lines
- [x] Grid styling (color, alpha, line width, dash patterns)
- [x] Grid behind data (z-order -1000)
- [x] Custom locators per grid (major/minor independently)
- [x] Example: `examples/axes/grid/main.go`

### 2.5 Axis Limits and Scaling ✅

- [x] `SetXLim(min, max)` and `SetYLim(min, max)` methods
- [x] `SetXLimLog`/`SetYLimLog` with auto-configured locators/formatters
- [x] `AutoScale(margin)` — computes limits from artist bounds with configurable margin
- [x] `Line2D.Bounds()` now returns actual data extent (was stub)
- [x] Example: `examples/axes/limits/main.go`

### 2.6 Text Labels using AGG Text Engine ✅

- [x] Text rendering using AGG raster FreeType with GSV vector font fallback
- [x] Title, xlabel, ylabel placement with proper centering
- [x] Tick labels using existing formatters (ScalarFormatter, LogFormatter)
- [x] Vertical ylabel text via `DrawTextVertical` (character-by-character layout)
- [x] `MeasureText` for layout calculations
- [x] Example: `examples/axes/labels/main.go`

### 2.7 Axes Control Surface

- [x] Public API for top/right spines, ticks, and labels instead of only internal `AxisSide` support
- [x] Axis inversion helpers (`InvertX`, `InvertY`)
- [x] Visual regression fixture for the new top/right-axis + inversion slice (`axes_top_right_inverted`)
- [x] Broader explicit axis direction control beyond inversion
- [x] Aspect controls (`SetAspect`, `SetBoxAspect`, axis-equal helpers)
- [x] `TickParams`, `LocatorParams`, and minor tick toggles (`MinorticksOn` / `MinorticksOff`)
- [x] Twin/secondary axis support (`TwinX`, `TwinY`, `SecondaryXAxis`, `SecondaryYAxis`)
- [x] Visual regression fixture for the broader axes-control surface (`axes_control_surface`)

### 2.8 Scale System Parity ✅

- [x] Public `SetXScale` / `SetYScale` API instead of only linear/log limit helpers
- [x] Additional built-in scales: `symlog`, `asinh`, `logit`, and function-based scales
- [x] Scale-specific configuration (`base`, `subs`, `linthresh`, non-positive handling, etc.)
- [x] Scale registration hooks so projections/toolkits can contribute custom scales

### 2.9 Locator and Formatter Parity

- [x] Additional locators: `FixedLocator`, `NullLocator`, `MultipleLocator`, `MaxNLocator`, `AutoLocator`, `AutoMinorLocator`
- [x] Additional formatters: `FixedFormatter`, `NullFormatter`, `FuncFormatter`, `FormatStrFormatter`, `StrMethodFormatter`, `EngFormatter`, `PercentFormatter`
- [x] Axis-owned tick style/state instead of today's loose locator/formatter pairing only
- [x] Multi-level ticks, label rotation/alignment helpers, and top/right tick label placement

### 2.10 Transform Graph and Coordinate Systems

- [x] Expose Matplotlib-like coordinate spaces: `transData`, `transAxes`, `transFigure`
- [x] Add blended transforms, offset transforms, and bbox-driven transforms
- [x] Refactor annotations/layout helpers to consume shared transform primitives instead of ad-hoc math
- [x] Make the transform stack projection-friendly so non-Cartesian axes do not require a redraw pipeline rewrite

### 2.11 Dates, Categories, and Units

- [x] Date locators/formatters and `time.Time`-friendly axis plumbing
- [x] Categorical axes instead of today's "categories as float positions" workaround
- [x] Units/converter support similar to Matplotlib's `units` machinery
- [x] Example coverage for dates, units, and category plots
- [x] Golden/parity coverage for dates, units, and category plots

**Exit Criteria:**

- All plots render with AGG anti-aliasing
- Plots have proper axis lines, ticks, and labels
- Grid lines work and look good (major + minor, dashed)
- Axis limits can be set manually or auto-computed

**Current follow-up after Phase 2:** package-level tests for `color`, `render`, `backends`, and `core` are green after the renderer-contract cleanup; the remaining short-term backend-quality work is in image golden/reference parity, while the largest functional gap versus Matplotlib is now the broader axes/scale/ticker surface above.

---

# Phase 3: Additional Plot Types (MEDIUM PRIORITY)

**Goal:** Expand the plotting vocabulary

### 3.1 Histograms ✅

- [x] `Hist2D` artist for histogram plots
- [x] Automatic binning strategies: Auto, Sturges, Scott, Freedman-Diaconis, Sqrt
- [x] Explicit bin edges supported via `BinEdges` field
- [x] Normalization modes: Count (default), Probability, Density
- [x] `Axes.Hist()` convenience method with color cycling
- [x] Comprehensive unit tests and golden tests (3 golden images)
- [x] Example: `examples/histogram/basic.go`

### 3.2 Box Plots

- [x] `BoxPlot` artist for statistical visualization
- [x] Quartiles, whiskers, outliers
- [x] Multiple box plots per axes
- [x] Create matplotlib reference in testdata/matplotlib_ref similar how the others got generated
- [x] Comprehensive unit tests and golden tests
- [x] Example: `examples/boxplot/basic.go`

### 3.3 Error Bars

- [x] `ErrorBar` artist for scientific plots
- [x] X and Y error bars with caps
- [x] Combine with scatter/line plots
- [x] Create matplotlib reference in testdata/matplotlib_ref similar how the others got generated
- [x] Comprehensive unit tests and golden tests
- [x] Example: `examples/errorbar/basic.go`

### 3.4 Images and Heatmaps

- [x] `Image2D` artist for imshow functionality
- [x] AGG image transformation for scaling/rotation
- [x] Colormaps (using AGG gradients)
- [x] Create matplotlib reference in testdata/matplotlib_ref similar how the others got generated
- [x] Comprehensive unit tests and golden tests
- [x] Example: `examples/image/basic.go`

### 3.5 Simple Plot Variants Built from Existing Primitives ✅

- [x] `Axes.Step()` and step draw styles
- [x] `Axes.Stairs()` for pre-binned histogram-style plots
- [x] `Axes.FillBetweenX()` / `fill_betweenx`
- [x] Infinite/reference line helpers: `axhline`, `axvline`, `axline`
- [x] Span helpers: `axhspan`, `axvspan`
- [x] `broken_barh` and stacked bar support
- [x] Bar labels and other simple bar-annotation helpers

### 3.6 Derived Statistical and Area Variants

- [x] `stackplot`
- [x] `ecdf`
- [x] Histogram presentation variants (cumulative, multi-hist, histtype variants)
- [x] Example and Matplotlib-reference coverage for each newly added convenience plot

**Exit Criteria:**

- Common scientific plot types work
- Examples demonstrate real-world use cases

---

# Phase 4: Layout, Figure Composition & Annotation

**Goal:** Move beyond a fixed subplot grid into the figure/layout systems Matplotlib relies on for real multi-panel figures

### 4.1 Subplots ✅

- [x] `Subplot` functionality for multiple axes grids
- [x] Automatic spacing and layout
- [x] Shared axes between subplots
- [x] Example: `examples/subplots/basic.go`

### 4.2 Figure Composition and GridSpec

- [x] `add_subplot` / subplot-spec style axes creation
- [x] `GridSpec`, nested grids, width/height ratios, and `subplot2grid`
- [x] `subplot_mosaic`
- [x] `SubFigure` / subfigure composition
- [x] More granular share modes (`row`, `col`, selected peers) instead of all-or-nothing grid sharing

### 4.3 Layout Engines

- [x] `subplots_adjust`
- [x] `tight_layout`
- [x] `constrained_layout`
- [x] Automatic label/title alignment across axes
- [x] Margin computation driven by measured text extents instead of fixed subplot padding

### 4.4 Legends ✅

- [x] `Legend` artist with automatic entries
- [x] Legend placement and styling
- [x] Line/marker/patch legend entries
- [x] Example: `examples/legend/basic.go`

### 4.5 Text Annotations ✅

- [x] `Text` artist for arbitrary text placement
- [x] Arrow annotations pointing to data
- [x] Math symbols and Greek letters (basic)
- [x] Example: `examples/annotation/basic.go`

### 4.6 Figure-Level Labels and Annotation Helpers

- [x] `suptitle`, `supxlabel`, and `supylabel`
- [x] Figure-level legends
- [x] Annotation boxes / offset-box style anchored layout helpers
- [x] Better title/xlabel/ylabel alignment behavior across shared-axis figures
- [x] Example: `examples/figure_labels/basic.go`

### 4.7 Colorbars ✅

- [x] `Colorbar` artist for heatmaps
- [x] Automatic scaling and labels
- [x] Various colormap support
- [x] Example: `examples/colorbar/basic.go`

### 4.8 Phase 4 Visual Parity and Composition Hardening

- [x] Add committed golden and Matplotlib-reference fixtures for Phase 4 composition examples:
  - nested `GridSpec` / subfigure composition
  - figure-level labels plus figure legends and anchored text
  - heatmap colorbar composition
  - text/arrow annotation composition
- [x] Fix nested `GridSpec` / constrained-layout small-panel spacing so tick labels do not overlap in `examples/gridspec/main.go` (`Nested Top` currently compresses y tick labels).
- [x] Make figure-level labels participate in measured margins strongly enough to prevent clipping (`examples/figure_labels/basic.go` currently clips the left `supylabel`).
- [x] Include figure legends and figure-level anchored boxes in layout collision checks so they avoid the suptitle and plot area by default.
- [x] Tighten colorbar layout coverage so colorbar axes, ticks, and labels compose without hand-tuned parent axes padding.
- [x] Define Phase 4 visual acceptance checks: generated examples have no clipped labels, no overlapping tick labels, no legend/title collisions, and pass documented golden/reference tolerances.

**Exit Criteria:**

- [x] Multi-panel figures work well beyond simple uniform grids
- [x] Layout no longer depends on hand-tuned subplot padding for common cases
- [x] Figure-level labels, legends, and colorbars compose cleanly

---

# Phase 5: API Surface, Configuration & Output

**Goal:** Close the major non-artist API gaps between an object-only plotting core and a usable Matplotlib-like runtime

### 5.1 SVG Export

- [x] SVG backend using path recording
- [x] Vector output for publications
- [x] Text as actual text (not paths)
- [x] Example: `examples/export/svg.go`

### 5.2 Pyplot / Stateful API

- [x] Optional `pyplot`-style figure registry (`Figure`, `GCF`, `GCA`, `Subplot`, `Subplots`)
- [x] Stateful helpers for `title`, `xlabel`, `ylabel`, `legend`, `colorbar`, `savefig`, `show`, and `pause`
- [x] Keep the explicit object API first-class while offering a migration-friendly convenience layer

### 5.3 Styling and Configuration Parity

- [x] Style sheets and themes
- [x] Color palettes and defaults
- [x] Publication-ready themes
- [x] Example: `examples/styling/themes.go`
- [x] Runtime rc defaults wired through new figures and the `pyplot` stateful API
- [x] `rcParams`, `rc`, `rc_context`, `rcdefaults`, and rc-file loading semantics
- [x] Example: `examples/styling/rc/main.go`
- [x] Much broader `.mplstyle` coverage than the current supported subset
- [x] Broader `.mplstyle` coverage for typography, tick styling, grid defaults/styles, legend frame controls, and `figure.figsize`
- [x] Style-library discovery beyond the built-in named theme registry

### 5.4 Backend Runtime, Canvas, and Tooling

- [x] `FigureCanvas` / `FigureManager` abstraction instead of only renderer factories
- [x] Event model shared across backends (mouse, keyboard, resize, draw, close)
- [x] Tool manager / toolbar abstractions
- [x] Embedding/runtime hosts for desktop or web backends
- [x] Minimal browser runtime host for Go `js/wasm` demos using the GoBasic backend and an HTML canvas bridge
- [x] Web demo stabilization for browser-hosted WASM callbacks: preserve the runtime on canvas focus/mouse input and fail clearly on stale generated web assets
- [x] Broaden the WASM web demo gallery with showcase examples for fills, error bars, patches, and polar axes

### 5.5 Additional Export and Embedding Targets

- [ ] PDF/PS/PGF export for publication workflows
- [ ] Filetype dispatch through backend/canvas capabilities instead of hard-coded `SavePNG` / `SaveSVG`
- [ ] GUI and web embedding examples (SDL/GTK/Qt/Tk/WebAgg-style)
- [x] GitHub Pages deployment workflow for the WASM-backed web demo

**Exit Criteria:**

- [ ] Both object-oriented and optional stateful APIs are usable
- [x] Configuration is no longer limited to hard-coded theme structs plus a tiny `.mplstyle` subset
- [x] Output/runtime capabilities are organized around backend canvases rather than ad-hoc helpers

---

# Phase 6: Advanced Artists, Collections, Meshes & Specialty Plots

**Goal:** Add the missing artist families that Matplotlib builds many higher-level plot types on top of

### 6.1 Patch Hierarchy

- [x] Introduce a `Patch` base artist instead of baking patch logic into one-off plot types
- [x] `Rectangle`, `Circle`, `Ellipse`, `Polygon`, and `PathPatch`
- [x] Arrow/fancy-box support (`arrow`, `FancyArrow`, box styles)
- [x] Hatch fill support

### 6.2 Collections and Result Containers

- [x] `Collection`, `PathCollection`, `LineCollection`, `PatchCollection`, `PolyCollection`
- [x] `QuadMesh` and `FillBetweenPolyCollection`-style primitives
- [x] Generalize scatter onto collection semantics for arbitrary marker paths and better batching
- [x] Matplotlib-style result containers (`BarContainer`, `ErrorbarContainer`, `StemContainer`)

### 6.3 Mesh, Contour, and Triangulation Plots

- [x] `pcolor` / `pcolormesh`
- [x] `contour` / `contourf` and contour labels
- [x] `hist2d`
- [x] Unstructured triangulation family: `triplot`, `tricontour`, `tricontourf`, `tripcolor`

### 6.4 Vector Fields and Field Visualization

- [x] `quiver`
- [x] `quiverkey`
- [x] `barbs`
- [x] `streamplot`

### 6.5 Statistical and Specialty Artists

- [x] `stem`
- [x] `eventplot`
- [x] `hexbin`
- [x] `pie`
- [x] `violinplot`
- [x] `table`
- [x] `sankey`
- [x] Stateful `pyplot` wrappers for the Phase 6.5 artists/builders (`Stem`, `Eventplot`, `Hexbin`, `Pie`, `Violinplot`, `Table`, `Sankey`)
- [x] Focused unit coverage for the new Phase 6.5 artist families and the Sankey builder
- [x] Combined specialty example: `examples/specialty/main.go`
- [x] Golden and Matplotlib-reference parity fixtures for the new specialty artists

### 6.6 Derived Image and Signal Helpers

- [x] `matshow`
- [x] `spy`
- [x] `specgram`
- [x] Signal-analysis helpers such as `psd`, `csd`, `cohere`, `xcorr`, and `acorr`
- [x] Annotated-heatmap / matrix-display helpers built on top of image + text primitives

**Exit Criteria:**

- [x] The port covers more than the "basic chart" subset and includes the artist families Matplotlib uses for mesh, contour, collection, and specialty plots
- [x] New plot families are backed by reusable artist infrastructure instead of one-off helpers

---

# Phase 7: Advanced Axes, Projections & Toolkits

**Goal:** Cover the non-Cartesian and toolkit-driven areas that make upstream Matplotlib much broader than a simple 2D plotting core

### 7.1 Non-Cartesian Projections

- [x] Polar axes
- [x] Geographic / geo projections (initial built-in `mollweide` projection)
- [x] Projection registry and `projection=`-style axes creation
- [x] Projection-specific ticks, labels, and transforms
- [x] Specialty projections built on top of the registry (`radar`, `skew-T`, other curvilinear examples)

Current slice landed:

- Built-in `polar` projection with circular spines, angular/radial grids, and polar tick labels
- Built-in `radar` projection on the projection registry with polygon frames, polygon radial grids, configurable spoke labels, and `Figure.AddRadarAxes`
- Built-in `skewx` projection for Skew-T style meteorological plots with pressure-axis defaults, top x-axis support, configurable skew angle, and `Figure.AddSkewXAxes`
- Projection registry plus `Figure.AddAxesProjection`, `Figure.AddPolarAxes`, and subplot `WithProjection(...)`
- Polar-specific theta zero-location, theta-direction, radial-label-angle controls, and projection-stage transform access via `DrawContext.TransProjection()`
- Built-in `mollweide` projection with oval frame/clipping, longitude/latitude degree ticks, sampled graticule lines, inverse pixel-to-data support, and Matplotlib reference coverage

### 7.2 3D Toolkit

- [x] `Axes3D` and projection math
- [x] 3D line, scatter, bar, surface, wireframe, contour, trisurf, voxel, and text artists (initial scaffold)
- [x] 3D examples corresponding to the upstream `mplot3d` and `plot_types/3D` gallery families

Current slice landed:

- `examples/mplot3d/basic.go` for base 3D artist coverage
- `examples/mplot3d/terrain/main.go` for surface/filled-contour/trisurf workflow

### 7.3 Axes Composition Helpers

- [x] Inset axes and zoomed-inset helpers
- [x] `AxesDivider`, `ImageGrid`, and RGB axes composition
- [x] Parasite axes / multi-view axes composition helpers
- [x] Anchored artists and locator-driven placement helpers

Current slice landed:

- Draw-time `AxesLocator` support with parent-relative `Axes.InsetAxes(...)` and `Axes.ZoomedInset(...)`
- Inset style/projection/share options and example coverage in `examples/axes/inset`
- `AxesDivider`, `ImageGrid`, and `RGBAxes` helpers now compose structured grids and channel-axis layout.
- `ParasiteAxes` adds overlay host-linked axes with optional shared x/y scale-root wiring for multi-view workflows in `core` and pyplot-facing helpers.
- Shared `AnchoredBoxLocator` helpers now drive placement for `AnchoredTextBox` and `Legend`, including normalized relative placement and corner-offset locator helpers.

### 7.4 AxisArtist and Floating-Axis Systems

- [x] Alternate axisartist-style axis subsystem
- [x] Floating axes
- [ ] Curvelinear grids and grid helpers
- [ ] Axis line styles and tick-direction control beyond standard Cartesian spines

Current slice landed:

- `AxisArtist` and `Axes.ExtraAxes` provide host-linked auxiliary axes that render through the normal figure draw path.
- `Axes.FloatingXAxis(...)` and `Axes.FloatingYAxis(...)` create data-positioned floating axes on rectilinear plots.
- Axis spines now support data-position overrides, and axes support explicit tick direction (`out`, `in`, `inout`) even though the broader line-style surface is still incomplete.

### 7.5 Gallery Parity and Showcase Coverage

- [ ] Add example coverage for major upstream gallery families still missing here (`axes_grid1`, `axisartist`, `mplot3d`, `pie_and_polar_charts`, `units`, `widgets`, `unstructured`, `arrays`)
- [ ] Add parity fixtures for each newly ported toolkit/projection family

**Exit Criteria:**

- [ ] The port is no longer limited to rectangular 2D axes
- [ ] Toolkit-driven layouts and projections have first-class support instead of bespoke examples

---

# Phase 8: Advanced Rendering, Text & Backend Depth

**Goal:** Deepen the renderer/backends so the broader artist surface above can be implemented with fidelity and performance

### 8.1 Renderer Contract Parity

- [ ] Clip paths implemented consistently across raster and vector backends
- [ ] Batch drawing primitives for collections/path collections
- [ ] Gouraud-shaded triangles and mixed raster/vector rendering controls
- [ ] Pattern fills, gradients, and richer alpha compositing
- [ ] Path effects and other post-stroke/post-fill rendering passes

### 8.2 Font Manager and Text Layout

- [x] Real font discovery/cache and `FontProperties`-style selection instead of a fixed fallback story
- [x] Shared single-line text layout metrics helper used by core layout
- [x] TTF/OTF loading across raster text backends
- [x] SVG font embedding and broader backend font-file parity
- [x] Better fallback/substitution and text-path support

### 8.3 MathText and TeX

#### 8.3.1 MathText parser/layout and renderer integration

- [x] Inline MathText fallback normalization for `$...$` segments and common commands
- [x] Lightweight MathText parser/layout model for scripts, fractions, roots, operator names, and simple accents
- [x] Full-expression horizontal MathText rendering in text artists, annotations, legends, axis labels, and figure labels
- [x] Full-expression rotated MathText rendering through text paths when backend text-path support is available
- [x] True mixed inline layout for strings that combine plain text and MathText in one line
- [x] Vertical full-expression MathText rendering through structured layout/path output
- [x] Renderer-neutral internal MathText engine boundary extracted to `internal/mathtext`; `core` now only adapts it to renderer/font APIs
- [ ] Broader MathText grammar
  - [x] Limits on large operators such as `\sum`, `\prod`, and `\lim`
  - [x] Basic spacing commands such as `\,`, `\:`, `\;`, `\quad`, and `\qquad`
  - [x] Font/style switches with layout consequences for `\mathrm`, `\mathsf`, `\mathtt`, `\mathit`, and `\mathbf`
  - [x] Basic fenced delimiters via `\left...\right` with size-aware delimiter rendering
  - [x] Matrices/arrays
  - [ ] More complete stretchy delimiter behavior beyond the current basic `\left...\right` handling
    - [x] `\middle` and omitted `.` delimiters within `\left...\right`
    - [x] Extensible rule-based rendering for vertical bar and bracket-style delimiters
  - [ ] Richer TeX spacing/control semantics beyond the current small command subset
    - [x] Named spacing commands and explicit `\hspace{...}` / `\kern{...}` dimensions
- [x] Caching/performance pass for parsed and laid-out MathText expressions
  - [x] Shared parse cache plus opt-in layout cache keyed by renderer measurement context
- [ ] Promote `internal/mathtext` to a focused standalone module/repo once the grammar/cache API stabilizes

#### 8.3.2 LaTeX / `usetex` integration

- [x] `text.usetex` rcParam support in style defaults, mplstyle parsing, and runtime context handling
- [ ] External TeX command pipeline
- [ ] Artifact caching and reproducible invalidation for TeX output
- [ ] Error reporting and fallback behavior when TeX toolchain execution fails
- [ ] Font/package coordination between TeX output and renderer expectations
- [ ] Raster/vector import of TeX output back into the renderer pipeline
- [ ] Backend-specific integration behavior and tests across raster and vector backends

#### 8.3.3 Complex text shaping beyond the current basic glyph flow

- [ ] OpenType shaping for scripts that need glyph substitution and positioning
- [ ] Ligatures, bidi handling, combining-mark placement, and script-aware glyph selection
- [ ] Backend-independent shaping layer or shaping-library integration
- [ ] Measurement and text-path parity for shaped output

### 8.4 Performance Optimization

- [ ] Path simplification and culling
- [ ] Large dataset handling (100k+ points)
- [ ] Parallel rendering

### 8.5 Additional Backends

- [ ] Finish the Skia backend beyond the current scaffold
- [ ] GPU-accelerated or browser-native backends if they prove necessary
- [ ] Backend capability matrix kept aligned with actual implementation depth

**Exit Criteria:**

- [ ] Renderer/backends are no longer the limiting factor for advanced artists and projections
- [ ] Text fidelity and performance are good enough to support a much larger Matplotlib surface

---

# Phase 9: Interactivity, Widgets & Animation

**Goal:** Add the runtime behaviors that depend on the backend/event infrastructure introduced earlier

### 9.1 Interactive Navigation and Events

- [ ] Basic pan/zoom using mouse
- [ ] Picking / hit testing
- [ ] Coordinate inspection, cursors, and callback registration
- [ ] Real-time plot updates / redraw scheduling

### 9.2 Widgets

- [ ] Buttons, check buttons, radio buttons
- [ ] Sliders and range sliders
- [ ] Text boxes and selection widgets
- [ ] Span/rectangle/polygon selectors

### 9.3 Animation

- [ ] `FuncAnimation`-style API
- [ ] Frame capture / writer backends (GIF, video, HTML)
- [ ] Efficient redraw/blitting paths for animated artists

### 9.4 Embedding Examples

- [ ] Desktop embedding examples
- [ ] Web/server embedding examples
- [ ] Interactive example gallery parity for events, widgets, and animation

**Exit Criteria:**

- [ ] Interactive backends are usable for exploration instead of export-only rendering
- [ ] Widgets and animation work on top of the shared backend/event runtime instead of custom demos

---

# Development Guidelines

## Backend Strategy

- **Primary backend:** AGG (`backends/agg/`) — anti-aliased, sub-pixel accurate
- **PoC backend:** GoBasic (`backends/gobasic/`) — retained for reference and simple use cases
- **Future:** Skia (GPU), SVG (vector), PDF (print)

## Testing Strategy

- Golden image tests for all plot types
- Property-based tests for data ranges
- Visual regression testing
- `go test ./...` runs all tests

## API Design Principles

- Follow matplotlib conventions where sensible
- Use functional options for configuration
- Keep simple cases simple
- Provide escape hatches for complex cases

## Performance Goals

- Handle datasets up to 100k points smoothly
- Sub-second rendering for typical plots
- Memory efficient for long-running applications

## Examples-Driven Development

- Every feature gets a working example
- Examples serve as integration tests
- README showcases example gallery
- Examples demonstrate real-world usage

---

This plan gets you to a fully functional plotting library quickly while keeping the foundation solid for future enhancements.
