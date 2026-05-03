# Python reference counterparts

Human-readable Matplotlib counterparts live next to the Go examples where a
matching reference plot exists. The shared catalog in
`internal/examplecatalog` records the relationship between example sources,
Go golden images, Matplotlib reference images, and the curated web demo subset.

Many older counterparts still call the split reference modules while the
examples are being migrated into importable side-by-side Go/Python case
directories:

```text
test/matplotlib_ref/plots/<plot_name>.py
```

For new or reshaped plotting examples, put the Go plot body and Python plot
body beside each other and add the case to `internal/examplecatalog`. The Go
runner should stay thin; parity fixes belong in core/rendering code, not in
example-only workarounds.

Run an individual counterpart, for example:

```bash
python3 examples/scatter/basic.py --output-dir /tmp/mpl-go-reference
```

Regenerate all reference images with:

```bash
python3 test/matplotlib_ref/generate.py --output-dir testdata/matplotlib_ref
```

Some Go examples are backend, export, styling, or interactive utility demos
rather than golden-reference plots. Those do not have a Matplotlib reference
counterpart unless a comparable plot exists under `test/matplotlib_ref/plots/`.
