#!/usr/bin/env python3
"""Matplotlib reference counterpart for examples/boxplot/basic.go."""

from __future__ import annotations

from pathlib import Path
import importlib.util

PLOT_PATH = Path(__file__).with_name("basic") / "plot.py"
SPEC = importlib.util.spec_from_file_location("boxplot_basic_plot", PLOT_PATH)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError(f"could not load {PLOT_PATH}")
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)

PLOT = MODULE.PLOT
main = MODULE.main


if __name__ == "__main__":
    main()
