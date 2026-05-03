#!/usr/bin/env python3
"""Matplotlib reference wrapper for examples/boxplot/basic/plot.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import importlib.util

ROOT = Path(__file__).resolve()
while ROOT.name != "matplotlib-go" and ROOT.parent != ROOT:
    ROOT = ROOT.parent

PLOT_PATH = ROOT / "examples" / "boxplot" / "basic" / "plot.py"
SPEC = importlib.util.spec_from_file_location("boxplot_basic_plot", PLOT_PATH)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError(f"could not load {PLOT_PATH}")
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)

PLOT = MODULE.PLOT


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
