#!/usr/bin/env python3
"""Matplotlib reference wrapper for test/parity/boxplot_basic/plot.py."""

from __future__ import annotations

import argparse

from examples.parity.boxplot_basic.plot import PLOT


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
