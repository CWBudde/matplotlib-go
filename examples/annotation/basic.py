#!/usr/bin/env python3
"""Matplotlib reference counterpart for examples/annotation/basic.go.

The plot body lives in test/matplotlib_ref/plots/annotation_composition.py so reference
generation and the example counterpart use the same Python implementation.
"""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

ROOT = Path(__file__).resolve()
while ROOT.name != "matplotlib-go" and ROOT.parent != ROOT:
    ROOT = ROOT.parent
sys.path.insert(0, str(ROOT))

from test.matplotlib_ref.plots.annotation_composition import PLOT


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=str(Path.cwd()))
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
