#!/usr/bin/env python3
"""Split Matplotlib web demo module generated from test/matplotlib_ref/webdemo.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403

def demo_scatter(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Scatter Clusters")
    ax.set_xlabel("feature x")
    ax.set_ylabel("feature y")
    ax.grid(True)
    edge = color(1, 1, 1, 0.8)
    specs = [
        (scatter_cluster(1, 11, -1.2, 0.5, 64), "D", 10, color(0.13, 0.49, 0.92), "cluster a"),
        (scatter_cluster(2, 22, 1.0, 1.4, 64), "^", 12, color(0.93, 0.39, 0.26), "cluster b"),
        (scatter_cluster(3, 33, 2.4, -0.8, 64), "s", 11, color(0.24, 0.72, 0.42), "cluster c"),
    ]
    for (x, y), marker, size, c, label in specs:
        ax.scatter(x, y, marker=marker, s=ss(size), color=c, edgecolors=edge, linewidths=lw(1.25), alpha=0.8, label=label)
    ax.set_xlim(-3.2, 4.2)
    ax.set_ylim(-3.0, 3.4)
    ax.legend()
    save(fig, out_dir, "scatter")

DEMO = demo_scatter


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
