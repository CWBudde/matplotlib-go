# Python/Go Example Readability Gaps

This note tracks the architectural issues that keep Go examples from reading
like their Matplotlib Python counterparts. The immediate example is
`examples/arrays/basic.go` versus `examples/arrays/basic.py`, but the same
patterns show up across parity examples.

## Current friction points

### Renderer and output boilerplate

Python examples can end with `fig.savefig(out, dpi=DPI)`. Go examples currently
need explicit backend selection, capability checks, renderer construction, and
`core.SavePNG`. That is useful backend plumbing, but it distracts from the
plotting calls in human-facing examples.

Possible fix: add a small example helper or public convenience API that accepts
a figure, output path, dimensions, DPI, and background, then chooses the default
registered renderer.

### Pointer-backed options

Optional scalar settings such as colormap names, booleans, widths, colors, and
alpha values often require addressable locals in Go:

```go
cmap := "plasma"
mesh.PColorMesh(data, core.MeshOptions{Colormap: &cmap})
```

The arrays example now uses a local generic `ptr` helper to keep the call site
closer to Python keyword arguments, but this is still example-level glue.

Possible fix: introduce lightweight option constructors for common optional
values, or switch selected option fields to value-plus-zero-value semantics when
there is no ambiguity.

### Axes rectangle spelling

Matplotlib examples use compact list rectangles:

```python
fig.add_axes([0.05, 0.14, 0.26, 0.74])
```

Go examples use explicit `geom.Rect` and `geom.Pt` values. That is clearer for
core geometry, but verbose for examples that are trying to mirror Matplotlib.

Possible fix: add a public helper such as `core.AxesRect(x0, y0, x1, y1)` or an
example-local helper where exact Matplotlib reading parity matters.

### High-level helper mismatch

`core.AnnotatedHeatmap` and `core.Spy` intentionally package repeated
Matplotlib call sequences (`imshow` plus text labels, or `np.where` plus
`scatter`) into Go helpers. That keeps examples short, but the code no longer
looks like a line-by-line translation.

Possible fix: keep high-level helpers for normal Go usage, and prefer comments
or paired lower-level examples when the goal is didactic Matplotlib parity.

### Reference/example drift

Some Python reference implementations are shared with golden-image generation,
while others are standalone human examples. When the same plot is represented in
multiple files, titles, labels, and layout choices can drift.

Possible fix: keep each Python counterpart as a thin wrapper around the
canonical `test/matplotlib_ref/plots/<name>.py` module, or explicitly document
when a counterpart is optimized for reading rather than golden generation.

## Arrays example status

`examples/arrays/basic.go` now follows the same top-level shape as
`examples/arrays/basic.py`:

- constants for width, height, and DPI;
- compact layout helpers;
- shared data-generation helpers;
- `drawAnnotatedHeatmap`;
- `drawMeshAndContour`;
- `drawSpyMatrix`;
- an output flag/argument and one save step.

Remaining intentional differences are Go option pointers, renderer
construction, and the local glue needed to map figure/axes anchored text onto
each library's API.

## Rendering parity fixes from the arrays comparison

The side-by-side render exposed several differences that belonged in the port,
not in the example source:

- Matplotlib's default text sizes come from `font.size: 10`,
  `axes.titlesize: large`, and `axes.labelsize`/tick/legend `medium`
  (`third_party/matplotlib/lib/matplotlib/mpl-data/matplotlibrc`). The Go
  defaults now mirror those point sizes instead of deriving labels and legends
  from earlier local ratios.
- Matplotlib's `AutoLocator` is a `MaxNLocator` with automatic tick-space
  budgeting and `[1, 2, 2.5, 5, 10]` steps
  (`third_party/matplotlib/lib/matplotlib/ticker.py`). Go axes now default to
  `AutoLocator` and estimate tick capacity with the same x/y label-size
  heuristic used by Matplotlib's `XAxis.get_tick_space` and
  `YAxis.get_tick_space` (`third_party/matplotlib/lib/matplotlib/axis.py`).
- Matplotlib's title placement updates against the top x-axis tight bounding
  box, including top ticks and top xlabel
  (`third_party/matplotlib/lib/matplotlib/axes/_base.py`). Go title placement
  now includes the top xlabel extent, fixing the title/xlabel collision in
  matrix-style plots.
- Matplotlib scatter sizes are marker areas in points squared, and collections
  convert them with `sqrt(size) * dpi / 72`
  (`third_party/matplotlib/lib/matplotlib/axes/_axes.py` and
  `third_party/matplotlib/lib/matplotlib/collections.py`). Go `Scatter` now
  uses the same area semantics, which fixes oversized sparse-matrix markers.
- The AGG backend could fall back to local vector text whose size did not track
  renderer DPI. AGG text measurement and unrotated drawing now use the
  FreeType-backed Go raster path first, so 10pt/12pt text in examples scales
  like Matplotlib's 100 DPI Agg output.
- `AnchoredText` padding and border padding are font-relative in Matplotlib
  (`third_party/matplotlib/lib/matplotlib/offsetbox.py`). Go anchored text now
  merges user options with inherited defaults instead of replacing the whole
  box configuration, preserving default row gaps and frame/text defaults.
- The arrays matplotlib reference now uses the same lower-level source shape as
  the examples for the heatmap and sparse-matrix panels: image/scatter calls,
  explicit integer ticks, and an explicit matrix top axis. This avoids masking
  source drift with a Go-only post-processing workaround.
