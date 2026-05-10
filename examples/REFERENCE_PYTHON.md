# Python reference counterparts

Canonical parity examples live under:

```text
examples/parity/<case-id>/plot.go
examples/parity/<case-id>/plot.py
```

Python source under `examples/` is centralized here. Files that are not part of
the canonical parity catalog are preserved under `examples/parity/legacy/`
using their previous relative paths.

The shared catalog in `internal/examplecatalog` records the relationship
between case IDs, Go source, Python source, committed Go goldens, Matplotlib
reference images, and the curated web demo subset.

Render one Python counterpart:

```bash
python3 examples/parity/run.py --id lognorm_imshow --output-dir /tmp/mpl-go-reference
```

Render one Go counterpart:

```bash
go run ./examples/parity/cmd --id lognorm_imshow --output-dir /tmp/mpl-go-reference
```

Regenerate all reference images with:

```bash
python3 test/matplotlib_ref/generate.py --output-dir testdata/matplotlib_ref
```

The legacy `test/matplotlib_ref/plots/<case-id>.py` modules remain as
compatibility/reference sources while the user-facing example counterparts are
normalized around `examples/parity`.
