# Python reference counterparts

Human-readable Matplotlib counterparts live next to the Go examples where a
matching reference plot exists. These scripts call the split reference modules
under `test/matplotlib_ref/plots/`, so the example comparison code and golden
reference generation stay in sync.

The canonical Matplotlib plot bodies are now split by plot name:

```text
test/matplotlib_ref/plots/<plot_name>.py
```

`test/matplotlib_ref/generate.py` is only the batch CLI and RNG-debug
dispatcher.

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
