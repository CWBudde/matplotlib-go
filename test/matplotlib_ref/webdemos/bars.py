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

def demo_bars(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Quarterly Revenue")
    ax.set_xlabel("quarter")
    ax.set_ylabel("EUR million")
    ax.grid(True, axis="y")
    x_a = np.array([-0.18, 0.82, 1.82, 2.82])
    x_b = np.array([0.18, 1.18, 2.18, 3.18])
    a = np.array([18, 24, 29, 34])
    b = np.array([14, 20, 27, 31])
    edge = color(0.18, 0.18, 0.22, 0.7)
    bars_a = ax.bar(x_a, a, width=0.34, color=color(0.16, 0.42, 0.82, 0.9), edgecolor=edge, linewidth=lw(1), label="Product A")
    bars_b = ax.bar(x_b, b, width=0.34, color=color(0.91, 0.45, 0.16, 0.9), edgecolor=edge, linewidth=lw(1), label="Product B")
    ax.bar_label(bars_a, labels=["18", "24", "29", "34"])
    ax.bar_label(bars_b, labels=["14", "20", "27", "31"])
    for i, label in enumerate(["Q1", "Q2", "Q3", "Q4"]):
        ax.text(i, -2.5, label, ha="center")
    ax.set_xlim(-0.75, 3.75)
    ax.set_ylim(-4, 38)
    ax.legend()
    save(fig, out_dir, "bars")

DEMO = demo_bars


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
