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
- [ ] Refresh golden/reference baselines after the AGG text-path refactor
- [ ] Re-run Matplotlib reference comparisons and either accept the new baseline or tighten text/layout parity

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

- [ ] Date locators/formatters and `time.Time`-friendly axis plumbing
- [ ] Categorical axes instead of today's "categories as float positions" workaround
- [ ] Units/converter support similar to Matplotlib's `units` machinery
- [ ] Example and parity coverage for dates, units, and category plots

**Exit Criteria:**

- All plots render with AGG anti-aliasing
- Plots have proper axis lines, ticks, and labels
- Grid lines work and look good (major + minor, dashed)
- Axis limits can be set manually or auto-computed

**Current follow-up after Phase 2:** package-level tests for `color`, `render`, `backends`, and `core` are green after the renderer-contract cleanup; the remaining short-term backend-quality work is in image golden/reference parity, while the largest functional gap versus Matplotlib is now the broader axes/scale/ticker surface above.

---

# Phase X: Immediate AGG Text Parity Drilldown (DO THIS NOW)

**Goal:** Close the remaining text-rendering and text-positioning gap versus `/tmp/matplotlib` systematically, starting with correctness and then tightening visual parity.

**Why this is immediate:** tick labels, titles, and other text still differ from Matplotlib in small but visible ways. We now have enough source-level evidence to stop guessing and work through the real backend differences one by one.

### X.1 Lock Down The Matplotlib Reference Model

- [x] Capture and document the exact Matplotlib Agg text pipeline we are matching:
  - `RendererAgg.draw_text()`
  - `Text._get_layout()`
  - `FT2Font.set_size()`
  - `FT2Font.set_text()`
  - `FT2Font.draw_glyphs_to_bitmap()`
  - See [docs/text_parity_phase_x1.md](/mnt/projekte/Code/matplotlib-go/docs/text_parity_phase_x1.md)
- [x] Record the exact default knobs from `/tmp/matplotlib` that affect text output:
  - `text.hinting`
  - `text.hinting_factor`
  - tick alignments
  - tick pad and tick size
  - title pad and size
  - See [docs/text_parity_phase_x1.md](/mnt/projekte/Code/matplotlib-go/docs/text_parity_phase_x1.md)
- [x] Keep the dedicated strict fixtures (`title_strict`, `text_labels_strict`) and `bar_basic_tick_labels` as the primary parity probes for this phase.
  - Documented in [docs/text_parity_phase_x1.md](/mnt/projekte/Code/matplotlib-go/docs/text_parity_phase_x1.md)

### X.2 Correct The Remaining Backend Rendering Differences

- [x] Stop forcing FreeType autohinting in the AGG raster path when Matplotlib is using the default hinting mode.
- [x] Rework raster glyph compositing in `../agg_go` so normal text is blended glyph-by-glyph like Matplotlib Agg, instead of first OR-combining a full run mask.
- [x] Verify device-space origin snapping happens at the same stage as Matplotlib’s `round(0x40 * ...)` placement path.
- [x] Verify glyph placement uses the same coordinate convention as Matplotlib’s run bbox packing:
  - `bitmap.left`
  - `bitmap.top`
  - `bbox.xMin`
  - `bbox.yMax`
  - the `+1` bitmap row adjustment

### X.3 Match Matplotlib’s Text Layout Metrics Model

- [ ] Replace remaining simplified baseline math with a Matplotlib-style layout model that combines:
  - actual run width/height/descent
  - font-wide minimum ascent/descent
  - line gap
- [ ] Do not use string ink bounds as a proxy for vertical placement except where Matplotlib itself effectively does so.
- [ ] Keep text ink bounds for horizontal anchoring and sidebearing handling only.
- [ ] Confirm bottom x ticks, top x ticks, left y ticks, and right y ticks all use the same alignment semantics as Matplotlib:
  - `top`
  - `baseline`
  - `center_baseline`
  - `left` / `center` / `right`

### X.4 Reduce The Gap In `matplotlib-go` Only After Backend Corrections

- [ ] Only keep local `matplotlib-go` text positioning code that is clearly source-motivated by Matplotlib.
- [ ] Remove empirical nudges that are not traceable to upstream behavior once the backend path is corrected.
- [ ] Route tick labels, titles, axis labels, and strict text fixtures through one consistent text-measurement model.

### X.5 Validation And Release Discipline

- [ ] Add or keep focused regressions for:
  - duplicated-letter/space replay bugs
  - stable line metrics vs string ink bounds
  - tick-label anchor placement
  - title-only parity
- [ ] Re-run and inspect visually after every meaningful backend text change:
  - `bar_basic_tick_labels`
  - `bar_basic_title`
  - `hist_strategies`
  - `text_labels_strict`
  - `title_strict`
- [ ] When `../agg_go` changes are required:
  - commit them cleanly
  - tag a new release
  - update `matplotlib-go` to that tag
  - regenerate only the affected goldens

**Exit Criteria:**

- No duplicated or replayed glyph artifacts.
- Tick labels, titles, and labels no longer look systematically too high or too low compared with Matplotlib.
- `bar_basic_tick_labels`, `text_labels_strict`, and the broad reference comparison pass comfortably.
- `title_strict` is improved materially enough that the remaining gap is explainable by known backend limitations, not unknown positioning errors.

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

### 3.5 Simple Plot Variants Built from Existing Primitives

- [ ] `Axes.Step()` and step draw styles
- [ ] `Axes.Stairs()` for pre-binned histogram-style plots
- [ ] `Axes.FillBetweenX()` / `fill_betweenx`
- [ ] Infinite/reference line helpers: `axhline`, `axvline`, `axline`
- [ ] Span helpers: `axhspan`, `axvspan`
- [ ] `broken_barh` and stacked bar support
- [ ] Bar labels and other simple bar-annotation helpers

### 3.6 Derived Statistical and Area Variants

- [ ] `stackplot`
- [ ] `ecdf`
- [ ] Histogram presentation variants (cumulative, multi-hist, histtype variants)
- [ ] Example and Matplotlib-reference coverage for each newly added convenience plot

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

- [ ] `add_subplot` / subplot-spec style axes creation
- [ ] `GridSpec`, nested grids, width/height ratios, and `subplot2grid`
- [ ] `subplot_mosaic`
- [ ] `SubFigure` / subfigure composition
- [ ] More granular share modes (`row`, `col`, selected peers) instead of all-or-nothing grid sharing

### 4.3 Layout Engines

- [ ] `subplots_adjust`
- [ ] `tight_layout`
- [ ] `constrained_layout`
- [ ] Automatic label/title alignment across axes
- [ ] Margin computation driven by measured text extents instead of fixed subplot padding

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

- [ ] `suptitle`, `supxlabel`, and `supylabel`
- [ ] Figure-level legends
- [ ] Annotation boxes / offset-box style anchored layout helpers
- [ ] Better title/xlabel/ylabel alignment behavior across shared-axis figures

### 4.7 Colorbars ✅

- [x] `Colorbar` artist for heatmaps
- [x] Automatic scaling and labels
- [x] Various colormap support
- [x] Example: `examples/colorbar/basic.go`

**Exit Criteria:**

- [ ] Multi-panel figures work well beyond simple uniform grids
- [ ] Layout no longer depends on hand-tuned subplot padding for common cases
- [ ] Figure-level labels, legends, and colorbars compose cleanly

---

# Phase 5: API Surface, Configuration & Output

**Goal:** Close the major non-artist API gaps between an object-only plotting core and a usable Matplotlib-like runtime

### 5.1 SVG Export

- [x] SVG backend using path recording
- [x] Vector output for publications
- [x] Text as actual text (not paths)
- [x] Example: `examples/export/svg.go`

### 5.2 Pyplot / Stateful API

- [ ] Optional `pyplot`-style figure registry (`Figure`, `GCF`, `GCA`, `Subplot`, `Subplots`)
- [ ] Stateful helpers for `title`, `xlabel`, `ylabel`, `legend`, `colorbar`, `savefig`, `show`, and `pause`
- [ ] Keep the explicit object API first-class while offering a migration-friendly convenience layer

### 5.3 Styling and Configuration Parity

- [x] Style sheets and themes
- [x] Color palettes and defaults
- [x] Publication-ready themes
- [x] Example: `examples/styling/themes.go`
- [ ] `rcParams`, `rc`, `rc_context`, `rcdefaults`, and rc-file loading semantics
- [ ] Much broader `.mplstyle` coverage than the current supported subset
- [ ] Style-library discovery and context-scoped temporary overrides

### 5.4 Backend Runtime, Canvas, and Tooling

- [ ] `FigureCanvas` / `FigureManager` abstraction instead of only renderer factories
- [ ] Event model shared across backends (mouse, keyboard, resize, draw, close)
- [ ] Tool manager / toolbar abstractions
- [ ] Embedding/runtime hosts for desktop or web backends

### 5.5 Additional Export and Embedding Targets

- [ ] PDF/PS/PGF export for publication workflows
- [ ] Filetype dispatch through backend/canvas capabilities instead of hard-coded `SavePNG` / `SaveSVG`
- [ ] GUI and web embedding examples (SDL/GTK/Qt/Tk/WebAgg-style)

**Exit Criteria:**

- [ ] Both object-oriented and optional stateful APIs are usable
- [ ] Configuration is no longer limited to hard-coded theme structs plus a tiny `.mplstyle` subset
- [ ] Output/runtime capabilities are organized around backend canvases rather than ad-hoc helpers

---

# Phase 6: Advanced Artists, Collections, Meshes & Specialty Plots

**Goal:** Add the missing artist families that Matplotlib builds many higher-level plot types on top of

### 6.1 Patch Hierarchy

- [ ] Introduce a `Patch` base artist instead of baking patch logic into one-off plot types
- [ ] `Rectangle`, `Circle`, `Ellipse`, `Polygon`, and `PathPatch`
- [ ] Arrow/fancy-box support (`arrow`, `FancyArrow`, box styles)
- [ ] Hatch fill support

### 6.2 Collections and Result Containers

- [ ] `Collection`, `PathCollection`, `LineCollection`, `PatchCollection`, `PolyCollection`
- [ ] `QuadMesh` and `FillBetweenPolyCollection`-style primitives
- [ ] Generalize scatter onto collection semantics for arbitrary marker paths and better batching
- [ ] Matplotlib-style result containers (`BarContainer`, `ErrorbarContainer`, `StemContainer`)

### 6.3 Mesh, Contour, and Triangulation Plots

- [ ] `pcolor` / `pcolormesh`
- [ ] `contour` / `contourf` and contour labels
- [ ] `hist2d`
- [ ] Unstructured triangulation family: `triplot`, `tricontour`, `tricontourf`, `tripcolor`

### 6.4 Vector Fields and Field Visualization

- [ ] `quiver`
- [ ] `quiverkey`
- [ ] `barbs`
- [ ] `streamplot`

### 6.5 Statistical and Specialty Artists

- [ ] `hexbin`
- [ ] `pie`
- [ ] `violinplot`
- [ ] `eventplot`
- [ ] `stem`
- [ ] `table`
- [ ] `sankey`

### 6.6 Derived Image and Signal Helpers

- [ ] `matshow`
- [ ] `spy`
- [ ] `specgram`
- [ ] Signal-analysis helpers such as `psd`, `csd`, `cohere`, `xcorr`, and `acorr`
- [ ] Annotated-heatmap / matrix-display helpers built on top of image + text primitives

**Exit Criteria:**

- [ ] The port covers more than the "basic chart" subset and includes the artist families Matplotlib uses for mesh, contour, collection, and specialty plots
- [ ] New plot families are backed by reusable artist infrastructure instead of one-off helpers

---

# Phase 7: Advanced Axes, Projections & Toolkits

**Goal:** Cover the non-Cartesian and toolkit-driven areas that make upstream Matplotlib much broader than a simple 2D plotting core

### 7.1 Non-Cartesian Projections

- [ ] Polar axes
- [ ] Geographic / geo projections
- [ ] Projection registry and `projection=`-style axes creation
- [ ] Projection-specific ticks, labels, and transforms
- [ ] Specialty projections built on top of the registry (`radar`, `skew-T`, other curvilinear examples)

### 7.2 3D Toolkit

- [ ] `Axes3D` and projection math
- [ ] 3D line, scatter, bar, surface, wireframe, contour, trisurf, voxel, and text artists
- [ ] 3D examples corresponding to the upstream `mplot3d` and `plot_types/3D` gallery families

### 7.3 Axes Composition Helpers

- [ ] Inset axes and zoomed-inset helpers
- [ ] `AxesDivider`, `ImageGrid`, and RGB axes composition
- [ ] Parasite axes / multi-view axes composition helpers
- [ ] Anchored artists and locator-driven placement helpers

### 7.4 AxisArtist and Floating-Axis Systems

- [ ] Alternate axisartist-style axis subsystem
- [ ] Floating axes
- [ ] Curvelinear grids and grid helpers
- [ ] Axis line styles and tick-direction control beyond standard Cartesian spines

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

- [ ] Real font discovery/cache and `FontProperties`-style selection instead of a fixed fallback story
- [ ] TTF/OTF loading across backends
- [ ] Better fallback/substitution and text-path support

### 8.3 MathText and TeX

- [ ] MathText parser/layout
- [ ] LaTeX / `usetex` integration
- [ ] Complex text shaping beyond the current basic glyph flow

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
