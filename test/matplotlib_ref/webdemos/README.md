# Web demo Matplotlib references

`test/matplotlib_ref/webdemo.py` is the batch CLI only.

The actual browser-demo reference plots live here, one module per demo ID:

```text
test/matplotlib_ref/webdemos/<demo_id>.py
```

Shared helpers are in:

```text
test/matplotlib_ref/webdemo_common.py
```

Run all web demo references:

```bash
python3 test/matplotlib_ref/webdemo.py --output-dir /tmp/mpl-webdemo
```

Run one demo:

```bash
python3 test/matplotlib_ref/webdemo.py --output-dir /tmp/mpl-webdemo --plots matrix
```
