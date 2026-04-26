#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def scatter_marker_types(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Scatter Marker Types")
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 8)
    markers = ["o", "s", "^", "D", "+", "x"]
    colors  = [(1,0,0), (0,1,0), (0,0,1), (1,1,0), (1,0,1), (0,1,1)]
    for i, (marker, color) in enumerate(zip(markers, colors)):
        # Plus (+) and cross (x) are line-only markers; use linewidths to make them visible
        line_w = lw(2) if marker in ("+", "x") else 0
        ax.scatter([i + 1], [4], s=ss(12), c=[color], marker=marker, linewidths=line_w)
    save(fig, out_dir, "scatter_marker_types")

PLOT = scatter_marker_types


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
