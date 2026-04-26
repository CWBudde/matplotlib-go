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
