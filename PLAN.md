# Matplotlib-Go Development Plan

This plan prioritizes getting useful plotting functionality working quickly. The AGG backend (via `github.com/cwbudde/agg_go`) is now available and provides a high-quality AGG-backed raster renderer, while GoBasic remains the default pure-Go backend.

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

# Phase 2: Axes Features — AGG Migration (CURRENT PRIORITY)

**Goal:** Improve AGG-backed rendering quality and make plots look professional

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

**Exit Criteria:** ✅

- All plots render with AGG anti-aliasing
- Plots have proper axis lines, ticks, and labels
- Grid lines work and look good (major + minor, dashed)
- Axis limits can be set manually or auto-computed

**Current follow-up after Phase 2:** package-level tests for `color`, `render`, `backends`, and `core` are green after the renderer-contract cleanup; the remaining backend-quality work is in image golden/reference parity, not API correctness.

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

**Exit Criteria:**

- Common scientific plot types work
- Examples demonstrate real-world use cases

---

# ✅ Phase 4: Layout & Annotation (COMPLETED)

**Goal:** Polish and professional presentation

### 4.1 Subplots

- [x] `Subplot` functionality for multiple axes grids
- [x] Automatic spacing and layout
- [x] Shared axes between subplots
- [x] Example: `examples/subplots/basic.go`

### 4.2 Legends

- [x] `Legend` artist with automatic entries
- [x] Legend placement and styling
- [x] Line/marker/patch legend entries
- [x] Example: `examples/legend/basic.go`

### 4.3 Text Annotations

- [x] `Text` artist for arbitrary text placement
- [x] Arrow annotations pointing to data
- [x] Math symbols and Greek letters (basic)
- [x] Example: `examples/annotation/basic.go`

### 4.4 Colorbars

- [x] `Colorbar` artist for heatmaps
- [x] Automatic scaling and labels
- [x] Various colormap support
- [x] Example: `examples/colorbar/basic.go`

**Exit Criteria:**

- [x] Multi-panel figures work well
- [x] Plots are publication-ready with legends, annotations, and colorbars

---

# Phase 5: Export & Polish (LOW PRIORITY)

**Goal:** Multiple output formats and refinements

### 5.1 SVG Export

- [x] SVG backend using path recording
- [x] Vector output for publications
- [x] Text as actual text (not paths)
- [x] Example: `examples/export/svg.go`

### 5.2 Interactive Features

- [ ] Basic pan/zoom using mouse (leveraging AGG's SDL2 support)
- [ ] Simple event handling
- [ ] Real-time plot updates
- [ ] Example: `examples/interactive/basic.go`

### 5.3 Styling and Themes

- [x] Style sheets and themes
- [x] Color palettes and defaults
- [x] Publication-ready themes
- [x] Example: `examples/styling/themes.go`

**Exit Criteria:**

- Multiple export formats work
- Library feels polished and complete

---

# Phase 6: Advanced Rendering (FUTURE)

**Goal:** High-performance and specialized rendering

### 6.1 AGG Advanced Features

- [ ] Gradient fills using AGG's linear/radial gradients
- [ ] Image transformations and pattern fills
- [ ] Gouraud-shaded triangles for smooth colormaps
- [ ] Custom blend modes and alpha compositing

### 6.2 TrueType Font Support

- [ ] Load TTF/OTF fonts via AGG's FreeType integration
- [ ] Complex text shaping
- [ ] LaTeX-style math rendering

### 6.3 Performance Optimization

- [ ] Path simplification and culling
- [ ] Large dataset handling (100k+ points)
- [ ] Parallel rendering

### 6.4 Additional Backends

- [ ] Skia backend for GPU acceleration (if needed)
- [ ] PDF export for publications (if needed)

**Exit Criteria:**

- Only implement if performance or quality demands it
- AGG should handle most use cases well

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
