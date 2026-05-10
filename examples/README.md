Examples live here.

- `examples/subplots/basic.go`: uniform subplot grids with shared axes.
- `examples/gridspec/main.go`: `GridSpec`, nested subplot specs, and subfigure composition.
- `examples/mesh/gouraud/main.go`: AGG-backed Gouraud-shaded `PColorMesh`.
- `examples/mathtext/basic.go`: inline MathText fallback rendering across titles, labels, text, and annotations.
- `core.Figure.TightLayout` / `core.Figure.ConstrainedLayout`: measured layout engines for subplot margins and spacing.

Parity examples live in `examples/parity/<case-id>/` with matching Go and
Python rendering source files. Use `go run ./examples/parity/cmd --list` or
`python3 examples/parity/run.py --list` to inspect the canonical case IDs.

All Python example/reference source under `examples/` is kept below
`examples/parity/`. Historical Python scripts that are not cataloged parity
cases live in `examples/parity/legacy/` with their original relative paths.
