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

def scatter_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Basic Scatter")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.scatter(
        [2, 4, 6, 8, 3, 7],
        [3, 6, 4, 7, 8, 2],
        s=ss(8),
        color=(0.8, 0.2, 0.2),
        linewidths=0,
    )
    save(fig, out_dir, "scatter_basic")

PLOT = scatter_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
