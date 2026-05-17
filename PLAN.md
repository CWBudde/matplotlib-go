# Matplotlib-Go Development Plan

This plan prioritizes getting useful plotting functionality working quickly. The AGG backend (via `github.com/cwbudde/agg_go`) is now available and provides a high-quality AGG-backed raster renderer, while GoBasic remains the default pure-Go backend.

This roadmap is cross-checked against the local upstream Matplotlib snapshot in `third_party/matplotlib` so uncovered areas are tracked explicitly instead of being left in broad "future work" buckets.

---

# Plan Tracking

- `✅` = done and stable
- `🧪` = implemented but under hardening
- `⚪` = in progress
- `⚠️` = deferred/design decision required
- `[ ]` = not started

---

# ✅ Phase 0: Architecture Alignment (COMPLETED)

**Goal:** convert PoC decisions and the new upstream architecture comparison into concrete execution tracks before adding more surface area.

### 0.1 Architecture Baseline

- [x] Clone/record upstream Matplotlib into `third_party/matplotlib`.
- [x] Add upstream snapshot to `.gitignore`.
- [x] Author `docs/architecture-comparison-with-matplotlib.md` capturing the fundamental differences.

### 0.2 Sub-Phase A: Core Object Model Parity

- [x] Keep interface-based `Artist` model as the port design baseline.
- [x] Add explicit parity notes for state/callback/stale behaviors that differ from upstream (`Artist` callbacks, stale propagation, picker/query surface).
- [x] Document and codify lifecycle boundaries in `core` for parity-critical cases (animation, clipping, draw ordering, overlay).

### 0.3 Sub-Phase B: Transform & Coordinate System

- [x] Keep explicit transform combinators (`transform` package) as the operational model.
- [x] Build a minimal invalidation-capable `TransformNode`-style compatibility layer for non-affine/affine split and cache-friendly compositions where it matters for projection-heavy cases.
- [x] Expand composable coordinate helpers (`transData`, `transAxes`, `transFigure`, blended transforms) with explicit test coverage on nested projections.

### 0.4 Sub-Phase C: Renderer/Backend Runtime Parity

- [x] Keep compact `render.Renderer` interface for portability.
- [x] Add parity-facing façade for backend capability checks and optional methods in a single registry contract.
- [x] Add backend contract tests for raster/vector export behavior, text pipelines, clipping, and save/dispatch semantics.

### 0.5 Sub-Phase D: Canvas, Event, and State API

- [x] Keep current headless canvas and event dispatcher as the minimum baseline.
- [x] Add parity mapping for Matplotlib event categories (`mouse`, `key`, `resize`, `draw`, `close`) and cursor/pick interactions.
- [x] Introduce a stricter manager contract for interactive backends (tooling + host lifecycle hooks) without blocking current non-interactive flow.

### 0.6 Sub-Phase E: Style/RC and Pyplot/API Surface

- [x] Keep lightweight `style` RC defaults and stackable contexts.
- [x] Add a staged rc-system expansion plan keyed by `supportedMPLStyleKeys` and upstream validator parity.
- [x] Add `pyplot` parity buckets for wrappers that are high-value but currently absent or partial (e.g. figure/axes/window/axes property convenience, output dispatch, context helpers).

### 0.7 Sub-Phase F: Architecture-First Test Strategy

- [x] Keep golden/reference comparison loop alive for image behavior.
- [x] Add architecture tests that validate: backend capability behavior, transform semantics, event lifecycle, rc param/option precedence.
- [x] Add review points in plan milestones so feature work only starts when structural test debt is bounded.

---

### 0.8 Missing architectural parity not yet tracked

- [x] Add an explicit draw-state model closer to Matplotlib’s `GraphicsContext` split (`RendererBase` vs stateful graphics context), including centralized opacity, clip, transform, and path-state ownership.
- [x] Add first-class event object types and connection lifecycle matching Matplotlib event classes (`MouseEvent`, `KeyEvent`, `PickEvent`) instead of only generic canvas events.
- [x] Add event-loop and redraw queue semantics (`draw_idle`-style scheduling, timer callbacks) for interactive backends and widgets.
- [x] Add artist callback and dirty-state lifecycle (`add_callback`/`remove_callback`, stale propagation, and draw scheduling) in `core` to support interactive mutation.
- [x] Add backend format/router parity (`register_backend` / default format handler behavior) as a single dynamic registry instead of path-extension switch logic.

## Baseline Status (Stable to Keep Unless Broken)

The following phases have reached "foundational parity enough to continue feature expansion" and should be treated as stable:

- ✅ Foundation (PoC to working renderer + AGG integration)
- ✅ Phase 1: Core Plot Types
- ✅ Phase 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.8
- ✅ Phase 3 (except open edges noted)
- ✅ Phase 4 completed items shown as checked
- ✅ Phase 5.1, 5.2, 5.3, 5.4
- ✅ Phase 6 (including sections 6.1–6.6)
- ✅ Phase 7 core projection/toolkit scaffolding

## ⚪ Example, Fixture, and Web Demo Catalog

**Goal:** keep user examples, Go golden fixtures, Matplotlib references, and
browser demos aligned around one catalog instead of separate hand-maintained
lists.

- [x] Add `internal/examplecatalog` as the shared metadata source for parity cases and the curated web demo subset.
- [x] Reduce the browser demo catalog to high-signal showcase/composition cases selected from parity fixtures.
- [x] Add catalog invariants for committed Go goldens, Matplotlib references, web-demo references, and reference-compare registration.
- [x] Mark renderer/backend stress cases as fixture-only so they are excluded from human-example migration.
- [x] Relocate the parity Go tree from `examples/parity/` to `test/parity/` to physically separate user-facing showcase code from test fixtures.
- [x] Move shared parity helpers from `test/parity/internal/common` to `internal/parityutil` so showcase examples can reuse them.
- [x] Migrate plot bodies into importable showcase packages at `examples/{id}/example.go`; parity wrappers at `test/parity/{id}/plot.go` import the showcase `Render()` to avoid duplication.
- [x] Inline the legacy delegating wrappers (`dashes`, `arrays_showcase`, `axes_grid1_showcase`, `axisartist_showcase`, `boxplot_basic`, `colorbar_composition`, `unstructured_showcase`) so each `examples/{id}/example.go` is self-contained.
- [x] Add a `Plot() *core.Figure` accessor to every non-fixture showcase package so the figure body is backend-agnostic and the AGG `Render()` is a thin wrapper.
- [x] Split shared `internal/parityutil.RenderBarBasicScaffold` / `RenderGeoProjectionAxes` into `Plot…` + `Render…` pairs so the bar-progression and geo-projection cases also expose a backend-agnostic `Plot()`.
- [x] Thin out duplicate top-level CLI runners (`examples/scatter/basic.go`, `examples/lines/basic.go`, etc.) so each one calls `examples/{id}.Plot()` instead of carrying its own copy of the figure body.
- [x] Curate the showcase set to 31 polished examples tagged `Showcase: true` in the catalog. Bodies of dropped non-fixture cases (joins*caps, scatter_marker_types, scatter_advanced, the bar_basic*\* progression, fill_between, fill_stacked, hist_density/strategies, units_dates/categories/custom_converter, geo_hammer/lambert, etc.) moved back into `test/parity/{id}/plot.go` only, with `examples/{id}/` deleted.
- [x] Retire the top-level `package main` wrapper directories (`examples/annotation/`, `examples/scatter/`, `examples/lines/basic.go`, etc.) — duplicated thin runners deleted; nested unique educational demos (`examples/axes/limits/`, `examples/lines/styles/`, `examples/lines/dash/`, `examples/geo/aitoff/`, `examples/mplot3d/terrain/`, `examples/backends/`, etc.) kept.
- [x] Replace the 22 hand-maintained `buildXxxDemo()` builders in `internal/webdemo/demo.go` with calls to `examples/{id}.Plot()` for the 11 cataloged web demos. `demo.go` is now a ~210-line dispatcher importing the 11 showcase packages; `demo_test.go` was simplified to smoke-test that every cataloged web demo produces a non-blank figure and renders without error.
- [x] Add a unified `cmd/example -name <id> -o out.png` runner with a `-list` mode that enumerates every `Showcase: true` catalog row. Uses `MATPLOTLIB_BACKEND` env selection. The remaining nested topic runners (`examples/lines/styles`, `examples/mplot3d/terrain`, etc.) are kept as topical educational demos but are no longer the only way to render a showcase.
- [x] De-duplicate the parity Python sources: 89 `test/parity/{id}/plot.py` files that were byte-identical (or behaviourally identical) to `test/matplotlib_ref/plots/{id}.py` are now relative symlinks pointing at the canonical reference. Two parity-only cases (`imshow_bilinear`, `imshow_bicubic`) keep their own standalone `plot.py` because no canonical counterpart exists. Normalised `boxplot_basic` so the implementation lives in `test/matplotlib_ref/plots/boxplot_basic.py` like every other case. Updated `test/matplotlib_ref/generate.py` to import the case registry from `test.parity` instead of the retired `examples.parity`.
- [x] Restructure `test/` from 19 to 8 files (~3.4k LOC removed via deduplication and subtest loops). Added `MinPSNR`/`MaxMeanAbs`/`MaxRMSE` fields to `examplecatalog.Case` so reference-compare tolerances are now stored on the catalog row; the duplicated `referenceCompareCases` slice and its sync-check (`example_catalog_test.go`) are gone. The 92 hand-written `TestX_Golden` one-liners collapsed into a single `TestGolden` that subtests over the catalog and skips cases without a committed PNG; the 77 `TestMpl_X` one-liners similarly collapsed into `TestMatplotlibRef`. The seven small fixture/showcase files (`image_fixtures_test.go`, `mesh_fixtures_test.go`, `agg_batch_fixtures_test.go`, `color_norm_fixtures_test.go`, `imshow_interpolation_test.go`, `showcase_parity_test.go`, `text_strict_test.go`) and the `golden_flavor_test.go` / `optional_visual_test.go` helpers were folded into `helpers_test.go` and `golden_test.go`. Three of the four diagnostic files (`alpha_residual_diagnostics`, `histogram_height_profile`, `rng_parity`) merged into `diagnostics_test.go`; `bar_text_diagnostic_test.go` stays separate because of its `cgo && !purego` build tag. Per-case invocation works via `go test -run TestGolden/basic_line` (and regex equivalents).

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

### 2.7 Axes Control Surface ✅

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

### 2.9 Locator and Formatter Parity ✅

- [x] Additional locators: `FixedLocator`, `NullLocator`, `MultipleLocator`, `MaxNLocator`, `AutoLocator`, `AutoMinorLocator`
- [x] Additional formatters: `FixedFormatter`, `NullFormatter`, `FuncFormatter`, `FormatStrFormatter`, `StrMethodFormatter`, `EngFormatter`, `PercentFormatter`
- [x] Axis-owned tick style/state instead of today's loose locator/formatter pairing only
- [x] Multi-level ticks, label rotation/alignment helpers, and top/right tick label placement

### 2.10 ⚪ Transform Graph and Coordinate Systems

- [x] Expose Matplotlib-like coordinate spaces: `transData`, `transAxes`, `transFigure`
- [x] Add blended transforms, offset transforms, and bbox-driven transforms
- [x] Refactor annotations/layout helpers to consume shared transform primitives instead of ad-hoc math
- [x] Make the transform stack projection-friendly so non-Cartesian axes do not require a redraw pipeline rewrite

### 2.11 Dates, Categories, and Units ✅

- [x] Date locators/formatters and `time.Time`-friendly axis plumbing
- [x] Categorical axes instead of today's "categories as float positions" workaround
- [x] Units/converter support similar to Matplotlib's `units` machinery
- [x] Example coverage for dates, units, and category plots
- [x] Golden/parity coverage for dates, units, and category plots
- [x] Tighten web-demo units parity for Matplotlib-style bar sticky baselines, default bar margins, daily date ticks, and rotated tick anchoring

### 2.12 ✅ Architecture Gates for Axes/Transforms

- [x] Add automated assertions for axis state transitions in non-affine/projection-heavy paths.
- [x] Add focused tests for transform-space APIs (`CoordData`, `CoordAxes`, `CoordFigure`) before adding new transform-specific plot APIs.
- [x] Add parity acceptance checks for coordinate-space helpers used by annotations and inset-like features.

**Exit Criteria:**

- [x] All plots render with AGG anti-aliasing
- [x] Plots have proper axis lines, ticks, and labels
- [x] Grid lines work and look good (major + minor, dashed)
- [x] Axis limits can be set manually or auto-computed
- [x] Architectural gates for transforms/coordinate systems are validated

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
- [x] Add a persisted light/dark/auto theme switch to the WASM web demo host

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
- [x] Phase 6.5 parity hardening for the specialty artist fixture: Matplotlib-style tick density, table alignment/row labels, pie label distance, and reference-matched demo construction

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
- [x] Curvelinear grids and grid helpers
- [x] Axis line styles and tick-direction control beyond standard Cartesian spines

Current slice landed:

- `AxisArtist` and `Axes.ExtraAxes` provide host-linked auxiliary axes that render through the normal figure draw path.
- `Axes.FloatingXAxis(...)` and `Axes.FloatingYAxis(...)` create data-positioned floating axes on rectilinear plots.
- Axis spines now support data-position overrides, and axes support explicit tick direction (`out`, `in`, `inout`) across standard and auxiliary axes.
- Projection grids now sample constant-coordinate paths for non-separable projection transforms, and curvilinear grids inherit axis locators by default so `skewx`/future projection helpers get the right grid geometry without bespoke grid artists.
- Axis stroke styling now exposes cap/join/dash control through `Axis.SetLineStyle(...)` and `Axes.SetAxisLineStyle(...)`, and `TickParams` now carries tick-direction control through the normal axes API.

### 7.5 Gallery Parity and Showcase Coverage

- [x] Add example coverage for the remaining upstream gallery families still missing here (`widgets`) and deepen the newer showcase families where coverage is still thin (`axes_grid1`, `axisartist`, `unstructured`, `arrays`)
- [x] Add parity fixtures for each newly ported toolkit/projection family

Current slice landed:

- Added showcase examples for `axisartist` and `axes_grid1` using the new floating-axis, parasite-axis, anchored-box, `ImageGrid`, and `RGBAxes` helpers.
- Added showcase examples for `unstructured` and `arrays` using `TriPlot`, `TriColor`, `TriContour`, `TriContourf`, `AnnotatedHeatmap`, `PColorMesh`, `Contour`, and `Spy`.
- Added golden and Matplotlib parity fixtures for the new `unstructured` and `arrays` showcase cases, including cross-reference thresholds and reference-image generator entries.
- Added golden and Matplotlib parity fixtures for the `axisartist` and `axes_grid1` showcase cases so the new toolkit-style examples are covered by the same visual-regression pipeline.
- Added a first static widget artist surface in `core` (`Button`, `Slider`, `CheckButtons`, `RadioButtons`, `TextBox`) plus a `widgets` showcase example to close the remaining gallery-family example gap in Phase 7.
- Documented the Python/Go example readability gaps in `docs/example-python-go-readability-gaps.md`, starting from the `examples/arrays` counterpart pair.
- Aligned the Go example bodies with their Python counterparts across the gallery, keeping language-specific setup idiomatic while matching data, call order, constants, layout, and explanatory comments where possible.

**Exit Criteria:**

- [ ] The port is no longer limited to rectangular 2D axes
- [ ] Toolkit-driven layouts and projections have first-class support instead of bespoke examples

---

# Phase 8: Advanced Rendering, Text & Backend Depth

**Goal:** finish the renderer-depth cleanup that still belongs in Phase 8. Backend-specific parity programs now live in Phase 14, plot-category raster parity lives in Phase 10, and interaction/runtime behavior lives in Phase 9.

Phase 8 has been condensed to remaining work only. Completed AGG, text, TeX, shaping, and performance history has been collapsed out of this section; use git history and the dedicated later phases for detailed implementation provenance.

### 8.1 Renderer Contract and Effects Cleanup

- [ ] Expose renderer-neutral pattern fill, gradient fill, and richer alpha-compositing contracts beyond AGG-local hatch support.
- [ ] Surface path effects and post-stroke/post-fill rendering passes to artists and backends, using AGG offscreen filter support where native.
- [ ] Add mixed raster/vector rendering controls at the artist/save-dispatch level, with backend capability checks instead of backend-name conditionals.

### 8.2 AGG Text and Font Pipeline Paydown

- [x] Replace the temporary DejaVu font-file bootstrap with an explicit font-resource strategy that can use embedded, shipped, and system fonts without leaking backend policy into draw calls.
- [x] Remove the GSV text fallback as a normal parity path; keep it only as an explicit emergency fallback with tests and diagnostics.
- [x] Match Matplotlib glyph bitmap compositing for antialiased and mono glyphs, including color-alpha application and clipping semantics.
- [x] Isolate text, path, and stroke state in the AGG surface so hidden draw-state resets and legacy font state no longer affect correctness.

### 8.3 Hatch, Pattern, and Fallback Cleanup

- [x] Move residual hatch clipping/fallback logic out of `core` once backend-native hatches and renderer-neutral hatch fallbacks are fully covered.
- [x] Split renderer-neutral fallback tests from AGG-native parity tests so missing native backend behavior cannot be hidden by fallback drawing.
- [x] Add focused diagnostics for remaining non-text AGG residuals that are not already owned by Phase 10 or Phase 14.

### 8.4 MathText Packaging

- [ ] Promote `internal/mathtext` to a focused standalone module/repo once the grammar and cache APIs stabilize.

### 8.5 Work Moved to Dedicated Phases

- [x] Backend parity, Skia, SVG, GoBasic, and cross-backend capability/save dispatch now live in Phase 14.
- [x] Plot-category AGG raster parity now lives in Phase 10.
- [x] Interactive redraw, widgets, animation, and runtime host behavior now live in Phase 9.

**Exit Criteria:**

- [ ] Remaining Phase 8 renderer cleanup is either complete or explicitly moved to a dedicated later phase.
- [ ] Renderer APIs expose effects, pattern, and mixed-output controls without backend-name conditionals.
- [ ] AGG text/font cleanup no longer depends on temporary font-file bootstrap behavior or normal GSV fallback rendering.

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
- [x] Web/server embedding examples
- [x] Direct web-demo PNG export and Matplotlib parity-viewer workflow
- [ ] Interactive example gallery parity for events, widgets, and animation

**Exit Criteria:**

- [ ] Interactive backends are usable for exploration instead of export-only rendering
- [ ] Widgets and animation work on top of the shared backend/event runtime instead of custom demos

---

# Phase 10: AGG-First Raster Plot Parity

**Goal:** make the existing raster-heavy 2D plot categories behave like Matplotlib on the AGG backend before expanding the public plot vocabulary further.

This phase tracks the plot categories that already exist in the Go port but are not yet fully supported at Matplotlib/RendererAgg fidelity. The reference source for these behaviors is `third_party/matplotlib/lib/matplotlib/backends/backend_agg.py`, `third_party/matplotlib/src/_backend_agg.*`, and the plot-type examples under `third_party/matplotlib/galleries/plot_types`.

### 10.1 QuadMesh, PColor, PColorMesh, and Hist2D

- [x] Match Matplotlib `pcolor` / `pcolormesh` input-shape validation for the supported 1-D rectilinear scalar-data API, including explicit edge coordinates, center coordinates, and mismatched dimensions.
- [x] Add shading policy parity for `flat`, `nearest`, `auto`, and rectilinear `gouraud`, including how each mode derives cell geometry from input coordinates.
- [x] Preserve NaN/Inf bad-cell transparency through `QuadMesh` color mapping and AGG draw batches; add weighted/density `hist2d` bin semantics.
- [x] Extend masked, bad, under, and over cell semantics beyond the current linear colormap model through `QuadMesh`, `Colorbar`, and AGG draw batches.
- [x] Match AGG edge rendering for antialiasing, linewidth, snap, join, cap, and alpha accumulation on dense quad meshes.
- [x] Add focused Matplotlib-reference fixtures for `pcolor`, `pcolormesh(shading="nearest")`, `pcolormesh(shading="gouraud")`, masked quad meshes, and `hist2d` with weights/density.
- [x] Tighten `quad_mesh` thresholds once the native batch path matches upstream cell placement and blending.

Completed notes:

- Added `MeshShading` support for rectilinear `flat`, `nearest`, `auto`, and `gouraud` mesh construction.
- `nearest` now interprets supplied coordinates as centers and expands them to cell edges using Matplotlib's midpoint policy; `flat` rejects center-shaped coordinate inputs.
- Rectilinear `gouraud` meshes route through native `render.GouraudTriangleDrawer` batches when available and fall back to averaged flat cells otherwise.
- Non-finite mesh scalar values now remain transparent while finite values drive scalar range resolution.
- `Hist2D` now supports per-sample weights plus probability/density normalization over 2D bin area.
- Colormaps now preserve explicit bad, under, and over colors for mesh scalar mapping; `MeshOptions.Mask` treats masked cells as bad values and excludes them from scalar range calculation.
- Added committed Go goldens and Matplotlib references for `pcolor_flat`, `pcolormesh_nearest`, `pcolormesh_gouraud`, `pcolormesh_masked`, and `hist2d_weighted_density`.
- Matplotlib-style colorbar slot placement now preserves manual-axes default fraction slots while constrained managed axes use the resolved aspect-width slot; `hist2d_weighted_density` and `pcolormesh_gouraud` are RMSE-gated below 30 against Matplotlib references.

### 10.2 PathCollection and Large Scatter

- [x] Audit `PathCollection` batch routing against upstream `RendererAgg.draw_markers` and `draw_path_collection` fast-path selection for homogeneous and heterogeneous markers.
- [x] Match per-point facecolor, edgecolor, linewidth, alpha, snap, hatch, antialiasing, and transform handling for large scatter and mixed collections.
- [x] Add coverage for empty collections, all-masked collections, unfilled markers, `edgecolors="face"`-style behavior, and collection-level alpha combined with per-item alpha.
- [x] Tighten `large_scatter` and `mixed_collection` thresholds after collection placement and alpha accumulation match upstream.
- [x] Keep fallback-renderer behavior tested separately so GoBasic/SVG fallback paths do not hide missing AGG-native behavior.

Completed notes:

- `PathCollection` routing is now explicitly covered against the upstream selection policy: homogeneous single-marker state can use `draw_markers`, while varying transforms, colors, linewidths, or paths route through `draw_path_collection`.
- Added focused collection tests for empty and all-invisible collections, line-only/unfilled marker stroke fallback, face-colored edges, and collection alpha combined with per-item face/edge alpha.
- Collection batches now propagate `SnapAuto`, collection-level antialias-off, native hatch metadata, hatch alpha, path transforms, and per-item paint state through marker/path collection batches and fallback paths.
- Added separate fallback coverage for heterogeneous path collections when native batch drawing declines.
- Tightened `large_scatter` to PSNR >= 55, MeanAbs <= 0.5, RMSE <= 4 and `mixed_collection` to PSNR >= 60, MeanAbs <= 0.5, RMSE <= 2 against Matplotlib references.

### 10.3 Images, Imshow, Matshow, and Spy

- [x] Complete AGG image clipping for clip boxes and non-rectangular clip paths, including projected axes frames.
- [x] Align image interpolation and resampling controls with Matplotlib names and fallback behavior (`nearest`, `none`, `bilinear`, `bicubic`, and rc/default policy); add scale-aware `auto`/`antialiased` resolution for direct and transformed draws.
- [x] Preserve image alpha, GC alpha, premultiplication, and background compositing semantics in the backend instead of pre-flattening in callers.
- [x] Match transformed image orientation, affine sampling, clipping, and interpolation against upstream `RendererAgg.draw_image`.
- [x] Add reference fixtures for clipped `imshow`, transformed `imshow`, alpha images, `matshow`, marker-mode `spy`, and image-mode `spy`.
- [x] Add direct buffer tests for RGBA/ARGB ordering, transparent backgrounds, and raw image export behavior.

Completed notes:

- Ported upstream `RendererAgg::draw_image` alpha behavior from `third_party/matplotlib/src/_backend_agg.h`: image-level alpha is applied to source alpha only before source-over compositing; RGB channels are not recolored by image alpha.
- AGG image draws now use source-over image compositing for direct and transformed image paths, preserving transparent-background behavior and matching Matplotlib's straight-RGBA buffer semantics.
- Fixed transformed-image parallelogram orientation so affine image sampling uses top-left, top-right, bottom-right corner order as AGG expects.
- Added committed Go goldens and Matplotlib references for `imshow_clipped`, `imshow_transformed`, `image_alpha`, `matshow_basic`, `spy_marker`, and `spy_image`.
- Added backend buffer/export tests covering RGBA byte order, transparent backgrounds, PNG round-tripping from the renderer buffer, source-alpha scaling, and transformed image orientation.
- `spy` image mode now uses Matplotlib's binary white/black image defaults with nearest interpolation, and transformed image rotation now follows Matplotlib's data-coordinate positive-angle convention; both `spy_image` and `imshow_transformed` are RMSE-gated below 30.

### 10.4 Filled Areas, Contours, and Alpha Accumulation

- [x] Track down remaining fill/hist alpha residuals called out in the current AGG notes and add fixture-specific diagnostics for repeated translucent overlaps.
- [x] Match filled polygon winding, self-intersection, clipping, and alpha accumulation for `fill_between`, `stackplot`, histogram fill variants, and filled contours.
- [x] Add contour topology cases for saddle points, masked triangles, holes, and contour bands that touch plot boundaries.
- [x] Make inline contour label erasure and rotated label placement match upstream display-space behavior across dense and sparse contours.

Completed notes:

- Added `TestAlphaResidualDiagnostics` for repeated translucent overlap cases, currently reporting axes-local residuals for `fill_stacked` and `hist_strategies` against Matplotlib references.
- Fixed explicit alpha handling for filled areas and histograms so `Alpha` overrides embedded RGBA alpha consistently; omitted histogram alpha now preserves color alpha through the `Axes.Hist` wrapper.
- Added structured contour topology coverage for filled saddle quads, interior holes, and boundary-touching contour bands, plus masked-triangle `tricontourf` coverage.
- Split ambiguous filled saddle bands into separate Matplotlib-style polygons instead of emitting a self-intersecting hourglass path, and close boundary-touching band paths where Matplotlib does.
- Inline contour label erasure now handles closed contours by cutting the wrapped segment across the path boundary, matching Matplotlib's display-space split behavior for labels near a closed-path seam.
- Added focused inline contour label coverage for dense, sparse, and too-short contours; label erasure, style preservation, and rotated placement are verified in display-space units.
- Verified filled-area, histogram, stackplot/stat, and filled-contour reference paths against Matplotlib with documented residuals: `fill_between` PSNR 53.13 / MeanAbs 0.21, `fill_stacked` PSNR 50.08 / MeanAbs 0.58, `hist_strategies` PSNR 50.98 / MeanAbs 0.24, `stat_variants` PSNR 53.34 / MeanAbs 0.23, and `mesh_contour_tri` PSNR 47.63 / MeanAbs 0.92.

**Exit Criteria:**

- [x] Raster-heavy 2D fixtures compare against Matplotlib references with the same strict thresholds used by line/bar/scatter basics, or have documented fixture-specific tolerances with root causes.
- [x] Existing plot categories no longer depend on backend fallbacks to appear complete when the AGG-native path is missing behavior.
- [x] `RendererAgg` batch fixtures cover marker, path collection, quad mesh, Gouraud, image, and clip-path paths with committed Go goldens and Matplotlib references.

Phase 10 closure notes:

- Added `clip_path_batch` as the missing non-rectangular clip-path visual fixture. It applies a data-space path clip to a native AGG quad-mesh batch and has committed Go golden and Matplotlib reference images.
- Phase 10 reference fixtures now cover pcolor/pcolormesh/Hist2D, large scatter/path collections, native quad mesh, Gouraud triangles, imshow/matshow/spy/image alpha, fill/hist/stack/contour, and clip-path batch behavior.
- Fresh reference comparison for `clip_path_batch`: PSNR 50.40 / MeanAbs 0.34 / RMSE 3.93, thresholded at PSNR >= 45, MeanAbs <= 1, RMSE <= 6.

---

# Phase 11: Color Mapping, Normalization & ScalarMappable Parity

**Goal:** make color-mapped plot categories share a Matplotlib-like scalar mapping model instead of each artist carrying a partial `vmin`/`vmax` implementation.

### 11.1 Normalization Model

- [x] Add a backend-independent `Normalize` model covering linear `Normalize`, `NoNorm`, `LogNorm`, `SymLogNorm`, `PowerNorm`, `TwoSlopeNorm`, `CenteredNorm`, and `BoundaryNorm`.
- [x] Preserve Matplotlib validation behavior for conflicting `norm`, `vmin`, and `vmax` inputs.
- [x] Add masked, bad, under, and over color handling in scalar mapping; colorbar extension rendering remains tracked in 11.2.
- [x] Make `Image2D`, `QuadMesh`, `PolyCollection`, `ContourSet`, `HexbinCollection`, `TriColor`, and `StreamplotSet` expose consistent scalar-mappable state.

Completed notes:

- Added `core.ScalarNormalizer` and backend-independent norm implementations for linear, no-op/index, log, symlog, power, diverging/two-slope, centered, and boundary normalization.
- Added `ScalarMapConfig`/`ResolveScalarMapValues`/`ResolveScalarMapGrid` so explicit norms autoscale through a shared path and `norm` with `vmin`/`vmax` is rejected consistently.
- Routed `Image2D`, `PColorMesh`/`QuadMesh`, `Hist2D`, `TriColor`, `ContourSet` fills, `HexbinCollection`, collection scalar metadata, and `StreamplotSet` scalar metadata through shared scalar-map state.
- Reused colormap bad/under/over support from scalar mapping so masked and non-finite mesh/image values no longer require plot-specific fallbacks; colorbar under/over extension geometry and norm-specific ticks remain in 11.2.

### 11.2 Colormap and Colorbar Depth

- [x] Add colormap registration, reversal, copying, and bad/under/over color mutation APIs close enough for common Matplotlib migration paths.
- [x] Match colorbar tick locator/formatter behavior for linear, log, boundary, and categorical-like norms.
- [x] Support colorbar extension triangles/rectangles for under/over ranges.
- [x] Add reference fixtures for `BoundaryNorm` pcolormesh, `LogNorm` imshow, diverging `TwoSlopeNorm`, and colorbar extension behavior.

Completed notes:

- Added `Colormap.Copy`, `Colormap.Reversed`, and in-place `SetBad`/`SetUnder`/`SetOver` methods while preserving the existing immutable `WithBad`/`WithUnder`/`WithOver` helpers and runtime registration path.
- Colorbars now retain scalar-mappable `ScalarMapInfo`, configure right-side log ticks/formatters for `LogNorm`, and use boundary values as fixed colorbar ticks for `BoundaryNorm`.
- Added colorbar extension drawing for `Extend: "min"`, `"max"`, and `"both"` with under/over colormap colors.
- Added focused unit coverage for colormap copy/reversal/mutation APIs, log and boundary colorbar tick configuration, and colorbar extension geometry.
- Added committed Go golden and Matplotlib reference fixtures for `boundarynorm_pcolormesh`, `lognorm_imshow`, `twoslope_norm_image`, and `colorbar_extensions`. Fresh reference comparisons: `boundarynorm_pcolormesh` PSNR 42.08 / MeanAbs 4.49, `lognorm_imshow` PSNR 41.95 / MeanAbs 5.03, `twoslope_norm_image` PSNR 41.25 / MeanAbs 3.81, `colorbar_extensions` PSNR 42.79 / MeanAbs 3.14.

### 11.3 Plot Category Integration

- [x] Route `imshow`, `matshow`, `pcolor`, `pcolormesh`, `contourf`, `tripcolor`, `hexbin`, `hist2d`, `quiver`, `barbs`, and `streamplot` through the shared normalizer.
- [x] Ensure legends and colorbars can infer scalar-mappable metadata consistently from all color-mapped artists.
- [x] Document unsupported normalization modes explicitly until they are implemented.

Completed notes:

- Added `Norm` forwarding to `MatShowOptions`/`ImShowOptions` and routed both helpers through the shared `Image2D` scalar-map path.
- Routed scalar-colored `Quiver`, `QuiverGrid`, `Barbs`, and `BarbsGrid` through `ResolveScalarMapValues`, preserving norm metadata for legends and colorbars.
- Added streamplot scalar coloring via `CGrid`, with scalar interpolation along streamline segments and arrows, and a line-collection scalar-map fallback when arrows are disabled.
- The Phase 11 built-in normalizer set is implemented: linear `Normalize`, `NoNorm`, `LogNorm`, `SymLogNorm`, `PowerNorm`, `TwoSlopeNorm`, `CenteredNorm`, and `BoundaryNorm`. Custom Matplotlib norm subclasses do not have a direct porting layer; Go callers can implement `ScalarNormalizer` explicitly.

**Exit Criteria:**

- [x] Color-mapped plot categories use the same normalization and colormap semantics.
- [x] Colorbar behavior is driven by scalar-mappable state rather than plot-specific shortcuts.
- [x] Matplotlib-reference fixtures cover linear, log, boundary, diverging, masked, bad, under, and over color behavior.

---

# Phase 12: Remaining 2D Plot API Surface

**Goal:** close the remaining high-value 2D API gaps where Matplotlib has plot-category entry points but the Go port only has lower-level building blocks or no direct equivalent.

### 12.1 Convenience Plot Entry Points

- [x] Add `SemilogX`, `SemilogY`, and `LogLog` helpers that mirror Matplotlib's scale-setting side effects while reusing `Axes.Plot`.
- [x] Add `PlotDate` or an explicit date-plot helper that preserves existing units/date converter behavior while matching Matplotlib's common migration path.
- [x] Add `Fill` for arbitrary closed polygon fills, distinct from `FillBetween` and patch helpers.
- [x] Add direct `BarH` convenience API if horizontal bars remain option-only in `Axes.Bar`.
- [x] Add pyplot wrappers for any object-oriented helpers already present but missing from `pyplot`.

### 12.2 Signal and Spectrum Variants

- [x] Add `MagnitudeSpectrum`, `AngleSpectrum`, and `PhaseSpectrum` equivalents alongside existing `Specgram`, `PSD`, `CSD`, `Cohere`, `XCorr`, and `ACorr`.
- [x] Align FFT windowing, detrending, scaling, sides, and dB behavior with upstream `matplotlib.mlab`/`Axes` helper behavior where practical.
- [x] Add Matplotlib-reference fixtures for spectrum variants and representative parameter combinations.

Current slice:

- Added object-oriented and pyplot convenience wrappers for `semilogx`, `semilogy`, `loglog`, `plot_date`, `fill`, and `barh`, backed by existing plot, units, polygon collection, and bar implementations.
- Added one-sided `magnitude_spectrum`, `angle_spectrum`, and unwrapped `phase_spectrum` helpers over the existing FFT utility path, with focused unit coverage for frequency bins and phase unwrapping.
- Aligned spectrum variants with Matplotlib's single-segment helper path: full-input FFT by default, `Fs`/`Fc`, Hanning or named windowing, mean/linear detrending, one-sided/two-sided frequency selection, and linear/dB magnitude scaling. FFT execution is now backed by the local `../algo-fft` module, and the repository Go floor is raised to 1.25 accordingly.
- Added `spectrum_variants` Go golden and Matplotlib-reference fixture covering magnitude dB scaling, two-sided angle spectra with `Fc`, and one-sided unwrapped phase spectra.

### 12.3 Statistical and Specialty Depth

- [x] Expand `ErrorBar` to support Matplotlib limit indicators (`uplims`, `lolims`, `xuplims`, `xlolims`) and asymmetric error shape validation.
- [x] Add deeper `BoxPlot` options: notch behavior, bootstrap/confidence intervals, custom medians/confidence intervals, whisker percentiles, and flier customization.
- [x] Expand `Violinplot` options for side selection, quantiles, custom bandwidth methods, and orientation aliases.
- [x] Expand `Pie` with label-rotation, normalization controls, shadow dictionaries, hatch support, and `pie_label`-style post-labeling.
- [x] Expand `Hexbin` with log bins, `xscale`/`yscale` behavior, marginal histograms, and reducer behavior beyond the current common reducers.

Current slice:

- Added asymmetric `ErrorBar` range fields, per-point limit flags, and validation for negative or shape-mismatched error arrays.
- Added `BoxPlot` notch/stat override fields, custom confidence intervals and medians, percentile whiskers, and flier marker/edge styling.
- Added `Violinplot` horizontal orientation aliases, low/high/both side selection, quantile line collections, and Scott/Silverman/numeric bandwidth method selection.
- Added `Pie` normalization control, wedge hatching, simple shadow wedges, stored label rotation angles, and `PieLabel`/`pyplot.PieLabel` post-labeling.
- Added `Hexbin` log-bin normalization, log-scale axis side effects, min/max reducers, bin discretization, and optional marginal bar collections.

**Exit Criteria:**

- [x] Common Matplotlib 2D plot-type entry points have either a direct Go API or an explicitly documented lower-level migration path.
- [x] New helpers are covered by unit tests and at least one Matplotlib-reference fixture per plot family.
- [x] Existing lower-level implementations are not duplicated by convenience wrappers.

Exit notes:

- Phase 12 plot families now have direct object-oriented APIs and pyplot wrappers where stateful coverage is expected: semilog/loglog/date/fill/barh, spectrum variants, errorbar depth, boxplot depth, violin depth, pie label helpers, and hexbin depth.
- Unit coverage spans the new direct helpers and option paths, including asymmetric errorbar limits, boxplot statistical overrides, violin side/orientation/quantiles, pie post-labeling, and hexbin log/reducer/marginal handling.
- Matplotlib-reference coverage includes existing basic family fixtures plus `spectrum_variants` and `specialty_depth`, the latter exercising the Phase 12.3 errorbar, boxplot, violin, pie, and hexbin depth paths in one reference image.
- Convenience wrappers continue to delegate to the lower-level artists/builders rather than duplicating rendering implementations.

---

# Phase 13: mplot3d Plot Category Completion

**Goal:** move 3D support from a projection scaffold to first-class coverage of Matplotlib's mplot3d plot categories.

The current implementation projects 3D data into 2D artists, which is useful for static AGG output but still falls short of the upstream `mpl_toolkits.mplot3d.axes3d.Axes3D` plot surface.

### 13.1 Missing 3D Plot Families

- [x] Add `Axes3D.Quiver` for 3D vector fields with length normalization, arrow ratios, pivot behavior, and `axlim_clip`.
- [x] Add `Axes3D.ErrorBar` for x/y/z error ranges, caps, limit markers, and depth-aware drawing order.
- [x] Add `Axes3D.Stem` with line, marker, and baseline styling.
- [x] Add `Axes3D.FillBetween` for polygon bands between two 3D curves.
- [x] Add `Axes3D.Bar` compatibility for 2D bars projected into selected z directions, separate from full cuboid `Bar3D`.

### 13.2 Existing 3D Plot Depth

- [x] Replace placeholder/simplified contour and contourf projection with Matplotlib-like level selection, `zdir`, `offset`, filled bands, and clipping behavior.
- [x] Expand `Surface` / `PlotSurfaceGrid` with stride/count sampling, facecolors, lighting, shade, antialiasing, `vmin`/`vmax`/`norm`, and colorbar-compatible scalar state.
- [x] Expand `Wireframe` with row/column stride and count behavior.
- [x] Expand `Trisurf` with colormap/norm support, triangle masking, edge/face styling, and depth sorting compatible with upstream examples.
- [x] Replace `Voxel` as bar-like prisms with Matplotlib-style boolean grid voxel input, facecolors, edgecolors, shade, and internal-face culling.

### 13.3 3D Rendering and Axes Behavior

- [x] Add depth sorting and z-order rules for mixed 3D collections that match upstream static AGG output.
- [x] Add 3D axis limit, aspect, pane, grid, tick locator, label placement, and view-init parity for common gallery examples.
- [x] Add Matplotlib-reference fixtures for every upstream `galleries/plot_types/3D` family: plot3d, scatter3d, surface3d, wire3d, trisurf3d, bar3d, voxels, quiver3d, stem3d, and fill_between3d.
- [x] Keep 3D tests focused on static AGG output; interactive rotation belongs to Phase 9 unless a backend-specific viewer requires it.

Current parity-hardening slice:

- Matplotlib-style 3D scatter marker depth sorting and depth-shaded alpha for the existing `mplot3d_basic` fixture.
- Filled-contour autoscaling from filled level midpoints, plus same-band contour path boundary merging for the existing `mplot3d_terrain` fixture.
- Unicode-minus scalar tick labels for default 3D z-axis formatter parity in negative tick ranges.

**Exit Criteria:**

- [x] Every Matplotlib plot-type gallery 3D category has a Go example, a golden image, and a Matplotlib reference image.
- [x] Existing 3D helpers expose scalar-mappable state where Matplotlib would support colorbars.
- [x] Mixed 3D scenes have deterministic depth ordering suitable for AGG golden tests.

---

# Phase 14: Backend Parity Program

**Goal:** make backend behavior explicit, testable, and Matplotlib-compatible across AGG, GoBasic, SVG, and Skia, in that order.

This phase consolidates the remaining backend work that was previously spread across Phase 8 renderer-depth notes and backend strategy notes. The ordering is intentional: AGG remains the reference raster backend, GoBasic is the pure-Go fallback, SVG is the first-class vector backend, and Skia follows after the shared contracts are stable enough to avoid duplicating backend-specific work.

### 14.1 AGG Reference Backend Parity

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backends/backend_agg.py`, `third_party/matplotlib/src/_backend_agg.*`, `backends/agg/`, `render/`, `test/`.

- [x] Audit `backends/agg` against upstream `RendererAgg` method coverage and record any intentionally unsupported methods in backend docs.
- [ ] Finish the shared shaping layer tracked in 8.1C so AGG text draw, measurement, bounds, and text-path output all consume the same shaped glyph runs.
- [ ] Complete AGG MathText and `usetex` import paths so raster text, path text, MathText, and TeX output share the same clipping, alpha, and DPI semantics.
- [x] Add buffer-region APIs equivalent to `copy_from_bbox` / `restore_region` for animation, blitting, and interactive redraw (`backends/agg.CopyFromBBox` / `backends/agg.RestoreRegion`).
- [x] Add `start_filter` / `stop_filter`-style offscreen rendering support for path effects and filtered artist output (`backends/agg.StartFilter` / `backends/agg.StopFilter`).
- [ ] Expand AGG parity diagnostics for remaining non-text residuals: dense path collections, repeated translucent overlaps, image interpolation modes, hatch clipping, and mixed raster/vector fallbacks.
- [ ] Split AGG-native parity fixtures from renderer-neutral fallback fixtures so missing native AGG behavior cannot be hidden by fallback drawing.

Exit criteria:

- [x] AGG is the canonical raster reference backend for parity fixtures and passes the strictest committed golden/reference thresholds.
- [x] AGG exposes native or explicitly unsupported status for every optional renderer capability in `render/extensions.go`.
- [x] AGG text, image, path, collection, hatching, clipping, and buffer behavior have targeted unit coverage plus representative visual fixtures.

Verification reference:

- [x] Added AGG optional-surface capability registration and status wiring in `backends/registry.go` and `backends/capabilities.go`.
- [x] Marked AGG implementation surface (`backends/agg/init.go`) for DPI-aware, text-bounds/path/rotated/vertical, image-transform, native-hatch, and PNG export capabilities.
- [x] Added AGG capability status assertions in `backends/agg/registry_test.go` for all declared native capabilities and explicitly unsupported SVG export.

### 14.2 GoBasic Backend Parity

**Reference sources:** `backends/gobasic/`, `backends/test_suite.go`, `backends/contract_test.go`, `render/`.

- [x] Define GoBasic's supported scope as a pure-Go correctness fallback rather than a pixel-identical Matplotlib renderer.
- [x] Bring GoBasic capability reporting into exact agreement with runtime interfaces: text, clipping, image transforms, batch fallbacks, hatching, export formats, and DPI behavior.
- [ ] Make GoBasic implement all renderer-neutral fallback paths required by `core` without silently dropping paint state such as alpha, line joins/caps, dashes, clipping, hatches, and antialiasing flags.
- [ ] Add GoBasic contract tests for path state save/restore, clip stack behavior, image drawing, transformed image fallback, text metrics, and collection fallback routing.
- [ ] Add a small GoBasic visual smoke fixture set that checks semantic output stability without using AGG-level pixel thresholds.
- [ ] Document every known GoBasic fidelity limitation in `backends/gobasic/doc.go` and surface those limitations through the capability matrix.

Current contract checklist:

- [x] Capability/status reporting: native DPI/text/text-path/rotated/vertical/PNG/path-clip support, fallback marker/path-collection/quad-mesh/hatch support, and unsupported image-transform/font-bound/vector/Gouraud/SVG support are covered.
- [x] Shared backend lifecycle/path contract suite runs against GoBasic.
- [x] `Path` tolerates nil paint without panicking.
- [x] Image drawing preserves source pixel sampling and applies `render.ImageAlpha` before bitmap blending.
- [ ] Path paint-state contract audit: alpha, line joins/caps, dashes, clip rect/path stacking, hatch fallback, and antialias flags all need explicit GoBasic tests that inspect rendered semantics rather than only no-panic behavior.
- [ ] Transformed image fallback contract: rotated/transformed images should have an explicit GoBasic test documenting that `core` falls back to axis-aligned `Image` when `render.ImageTransformer` is absent.
- [ ] Text metrics contract: basic `MeasureText`, DPI scaling, missing-font fallback, rotated text, vertical text, and text-path behavior need grouped contract coverage; unsupported font-wide metrics and ink bounds should stay explicitly unsupported.
- [ ] Collection fallback routing contract: marker, path collection, patch collection, quad mesh, hatches, and Gouraud fallback/unsupported behavior need tests proving GoBasic receives renderer-neutral `Path`/`Image` calls instead of silently dropping output.
- [ ] Example smoke contract: render every committed non-interactive showcase/parity case with GoBasic and assert no panic plus non-blank output under semantic, non-AGG thresholds.

Exit criteria:

- [ ] GoBasic can render every committed non-interactive example without panics or missing mandatory artist output.
- [ ] GoBasic capability reports match actual behavior and fail tests when a claimed capability is absent.
- [ ] GoBasic remains dependency-light and pure Go while sharing as much renderer-neutral logic as possible with AGG/SVG/Skia.

Verification reference:

- [x] Added GoBasic capability-status coverage for native DPI/text/text-path/rotated/vertical/PNG/path-clip support, renderer-neutral marker/path-collection/quad-mesh/hatch fallback status, and unsupported image-transform/font-bound/vector/Gouraud/SVG capabilities.
- [x] Wired GoBasic capability registration to the runtime interfaces it actually implements and added compile-time interface assertions.
- [x] Ran the shared backend contract suite against GoBasic; fixed `Path` to tolerate nil paint without panicking.
- [x] Added a GoBasic image drawing contract for `render.ImageAlpha` and applied image-level alpha before bitmap blending.

### 14.3 SVG Vector Backend Parity

**Reference sources:** `third_party/matplotlib/lib/matplotlib/backends/backend_svg.py`, `backends/svg/`, `render/`, `test/`.

Bring the SVG backend to functional parity with upstream Matplotlib's `RendererSVG`: deterministic serialization, real vector clip paths, native marker/hatch defs, an explicit text-as-text vs text-as-path policy, and a structural test surface that does not depend on rasterized screenshots. Work is split into six dependency-ordered subtasks; each can ship independently and any subtask audits the matching upstream behavior as part of its own work item.

**14.3.1 Deterministic serialization foundation (landed):**

- [x] Compact float formatter mirroring matplotlib's `_short_float_fmt` (6 decimals max, trailing zeros stripped, `-0` normalized, NaN/Inf clamped to `0`).
- [x] Centralized rotation formatter (`rotateTransform`) so the eventual matrix-transform switch only touches one helper.
- [x] Deterministic ID assignment across rect clips, path clips, and marker defs via a single shared counter and registration-ordered slices (no map iteration in the writer).
- [x] Byte-for-byte reproducibility test (`TestSerializationDeterministic`) and table-driven `TestShortFloat` / `TestRotateTransformUsesShortFloat` pinning formatter semantics.

**14.3.2 Clip paths, nested clip groups, transformed clips (mostly landed):**

- [x] `ClipPath(geom.Path)` registers a deduped `<clipPath><path d="…"/></clipPath>` def in `<defs>` keyed by path data.
- [x] Nested clip stacks: `Save`/`Restore` snapshot/truncate the path-clip stack; each rendered node captures the full active chain (rect outer, path inner).
- [x] Serializer wraps content in nested `<g clip-path="url(#…)">` groups in outer-to-inner order.
- [x] Tests for nested rect+path composition, dedup across nodes, restore-pop unwind, empty-path rejection.
- [ ] Switch per-element rotation strings to canonical `matrix(a b c d e f)` transforms everywhere a transform is applied (text, image, marker `<use>`). Single helper (`matrixTransform`) already exists from 14.3.3 and is reused.
- [ ] Transformed clip paths: apply an affine to a stored clip-path def via `<clipPath><path … transform="…"/></clipPath>` once matrix transforms are universal.

**14.3.3 Native batches, image transforms, hatches, alpha (in progress):**

- [x] `ImageTransformed` emits `transform="matrix(a b c d e f)"` on `<image>`; advertised native `ImageTransform`.
- [x] Native `DrawMarkers`: one `<defs><path id="markerN" d="…"/>` per unique marker geometry; short `<use href="#markerN" transform=… fill=… stroke=…/>` per item, all wrapped in a single clip-group node; advertised native `MarkerBatch`.
- [ ] Native `DrawPathCollection` via `<defs>` + `<use>` per unique path geometry (mirrors the marker treatment); advertise native `PathCollectionBatch`.
- [ ] Native hatches via `<pattern>` defs (72×72 user-space, keyed by `(hatch, faceColor, edgeColor, lineWidth)`); advertise `NativeHatch` once `<pattern>` reuse is structural and the existing core-side rasterizer is no longer invoked for SVG.
- [ ] Forced-alpha group wrapping: when `paint.ForceAlpha && paint.Alpha < 1`, emit `opacity="alpha"` on the element (or a `<g opacity>` wrapper for batches) and skip per-color `*-opacity`. Requires coordinated changes to `render/graphics_context.go` so colors are not double-multiplied.
- [ ] Gouraud triangle fallback: continue rasterizing via core but document the limitation explicitly; native `<linearGradient>` emission deferred until use cases warrant it.

**14.3.4 Font policy, metadata, and SVG-specific save options (not started):**

- [ ] `FontPolicy` option on `svg.Renderer` (`"none"` text-as-text default, `"path"` text-as-path via `render.TextPather`); analogue of matplotlib's `svg.fonttype` rcParam.
- [ ] Plumb `FontPolicy`, `Metadata`, and `HashSalt` options through `core.SaveSVG`/canvas save dispatch via a typed options interface.
- [ ] Default empty `<metadata>` block; opt-in via explicit `Metadata` option or `SOURCE_DATE_EPOCH` env var for date stability.
- [ ] Opt-in `HashSalt`: when set, switch IDs to `SHA256(salt+content)[:10]` (matches matplotlib's `svg.hashsalt`); default keeps the existing sequential `clipN`/`markerN` IDs.
- [ ] Document the option set in `backends/svg/doc.go`.

**14.3.5 Reference fixtures and structural diff helper (not started):**

- [ ] `testdata/golden/svg/` populated with structural goldens for canonical plot families: line, scatter, bar, errorbar, hist, collection, image, clipped polar, hatch_bars, text_layout, mathtext.
- [ ] New `internal/svgcompare/` helper: parse, normalize attribute ordering, structural diff. Unit-test the normalizer before any golden lands.
- [ ] `-update` flag on the golden test to re-bake fixtures when intentional output changes ship.

**14.3.6 Capability matrix contract and audit coverage (in progress):**

- [x] `TestSVGBackend_AdvertisedCapabilitiesAreImplemented` invokes `VerifyRendererCapabilities`; over-advertised `FontHinting` removed (its runtime check expected `render.TextFontMetricer` which SVG never implemented).
- [ ] Cross-backend semantic tests: render the same figure through AGG and SVG; compare clip/marker/text-glyph counts rather than pixel/byte parity.
- [ ] `backends/svg/audit_test.go` exercising every `render.Renderer` surface method so silent-drop regressions surface as test failures rather than missing artist output.

**Save dispatch (already routed via registry, completed before 14.3 began):**

- [x] `BackendInfo.SaveFormats[".svg"] = backends.SaveSVG` is wired in `backends/svg/init.go` and exercised end-to-end by `backends/registry_save_test.go` and `core/savesvg_test.go`. The original "route save dispatch through backend capabilities" bullet shipped before this task started; promoted to exit criterion.

Exit criteria:

- [ ] SVG output is deterministic, standards-compliant enough for browser viewing, and covered by structural tests instead of only rasterized screenshots.
- [ ] SVG supports the same high-level artist surface as AGG, with documented vector-specific fallbacks for unsupported raster-only effects.
- [x] `.svg` save dispatch works through the backend registry/canvas path.

Progress notes:

- The original "Audit SVG output against upstream `RendererSVG`" bullet has been folded into the per-subtask design rather than tracked separately — each subtask cites the upstream behavior it is matching, so the audit ships incrementally with the implementation.
- 14.3.2 and 14.3.3 share a transitive dependency: transformed clip paths (14.3.2) and `<pattern>`-based hatch defs (14.3.3) both benefit from the matrix-transform helper introduced for `ImageTransformed`/`DrawMarkers`. The remaining matrix-transform refactor in 14.3.2 unblocks both.
- Moving `MarkerBatch` from fallback to native (14.3.3) exposed that `PathCollectionBatch` and `QuadMeshBatch` remain core-side fallbacks; native treatment of path collections is the next natural step.
- Default-no-metadata in 14.3.4 diverges intentionally from matplotlib (which always emits a generated date) to maximize byte-for-byte determinism out of the box; users opt in via `SOURCE_DATE_EPOCH` or explicit options.

Verification reference:

- [x] Formatter pinning: `TestShortFloat`, `TestRotateTransformUsesShortFloat`, `TestSerializationDeterministic`.
- [x] Clip paths: `TestClipPathNestsInsideActiveRectClip`, `TestClipPathDedupesIdenticalPathsAcrossNodes`, `TestClipPathStackUnwindsOnRestore`, `TestClipPathRejectsEmptyPath`, plus the existing `TestRenderSVGPreservesClipStackAcrossSaveRestore` for rect-only intersection semantics.
- [x] Image transforms: `TestImageTransformedEmitsMatrixAttribute`, `TestImageTransformedHonorsClip`, `TestImageTransformedSkipsUnsupportedImage`, `TestMatrixTransformFormat`.
- [x] Marker batches: `TestDrawMarkersEmitsDefAndUseElements`, `TestDrawMarkersDedupesIdenticalMarkerGeometry`, `TestDrawMarkersHonorsActiveClip`, `TestDrawMarkersRejectsEmptyBatchOrInvalidMarker`.
- [x] Capability advertising: `TestSVGBackend_AdvertisedCapabilitiesAreImplemented` (registry-level contract).

### 14.4 Skia Backend Parity

**Reference sources:** `backends/skia/`, `backends/registry.go`, `render/`, `backends/test_suite.go`.

- [x] Decide and document the Skia binding strategy, build tags, dependency expectations, CPU/GPU mode split, and CI support model.
- [ ] Replace the current scaffold with a functional Skia renderer that implements the base `render.Renderer` contract: paths, images, save/restore, clip rect/path, and PNG export.
- [ ] Add Skia text support through the shared font/shaping pipeline instead of inventing backend-local text metrics.
- [ ] Add native Skia support for optional capabilities where practical: marker batches, path collections, quad meshes, Gouraud triangles, transformed images, hatching or pattern fills, and GPU acceleration.
- [ ] Implement Skia capability reporting separately for CPU and GPU modes so tests can distinguish native, fallback, and unavailable paths.
- [ ] Add Skia backend contract tests and a small visual fixture set gated by build tags or environment checks when Skia dependencies are unavailable.
- [ ] Compare Skia output against AGG semantic fixtures and only use Matplotlib pixel thresholds where Skia is expected to match the raster reference closely.

Exit criteria:

- [ ] Skia is usable as an opt-in backend for static raster output.
- [ ] Skia's capability matrix is truthful for both CPU and GPU configurations.
- [ ] Skia test coverage can run deterministically in CI or skip with explicit dependency diagnostics.

Progress notes:

- Added `backends/skia.BackendStrategy()` as the source of truth for the Skia integration policy: `skia` build tag, controlled external C ABI wrapper around Skia, CPU raster first, GPU deferred, and default CI staying on the unavailable stub path until dependencies are explicitly enabled. `backends/skia/doc.go` now mirrors that policy and no longer claims current GPU/text/output capabilities before the renderer implements them.
- Skia registry metadata now stays conservative while the backend is unavailable: no native capabilities and no save formats are advertised until `isAvailable()` can be backed by a real renderer implementation. The unavailable scaffold remains registered so backend-selection diagnostics can still name it.

### 14.5 Cross-Backend Capability and Save Dispatch

- [ ] Replace remaining hard-coded `SavePNG` / `SaveSVG` paths with registry/canvas save dispatch keyed by backend capabilities and file extension.
- [ ] Expand `backends.CapabilityMatrix()` to include all optional renderer capabilities from `render/extensions.go`, not just the original coarse capability set.
- [ ] Add tests that instantiate each registered backend and verify advertised native capabilities against concrete runtime interfaces.
- [ ] Add a backend comparison report command or test helper that lists unsupported/fallback/native status for AGG, GoBasic, SVG, and Skia.
- [ ] Keep backend docs aligned with actual capabilities whenever a renderer interface is added or removed.

Exit criteria:

- [ ] Backend selection, save dispatch, and capability reporting are the single source of truth for AGG, GoBasic, SVG, and Skia.
- [ ] New artist work can rely on capability checks instead of backend-name conditionals.
- [ ] The backend parity matrix is reviewed before marking future rendering phases complete.

---

# Development Guidelines

## Backend Strategy

- **Primary backend:** AGG (`backends/agg/`) — anti-aliased, sub-pixel accurate
- **Pure-Go fallback:** GoBasic (`backends/gobasic/`) — retained for dependency-light semantic coverage
- **Vector backend:** SVG (`backends/svg/`) — deterministic vector export and browser-readable output
- **Future accelerated backend:** Skia (`backends/skia/`) — CPU/GPU raster path after shared backend contracts are stable
- **Future print/export backends:** PDF and other formats once SVG/vector contracts are mature

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
