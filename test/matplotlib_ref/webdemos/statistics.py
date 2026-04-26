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

def demo_statistics(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.91, 0.10, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    data = [[1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2], [1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0], [2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6]]
    ax = axs[0][0]
    ax.set_title("Box + Violin")
    ax.set_xlim(0.4, 3.6)
    ax.set_ylim(0.6, 5.4)
    ax.grid(True, axis="y")
    ax.boxplot(data, positions=[1, 2, 3], widths=0.42, patch_artist=True, boxprops={"facecolor": color(0.39, 0.62, 0.84, 0.38)})
    ax.violinplot(data, showmeans=True, showmedians=True)
    ax = axs[0][1]
    ax.set_title("ECDF")
    vals = np.array([1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0])
    ax.step(np.sort(vals), np.arange(1, len(vals) + 1) / len(vals), where="post", color=color(0.18, 0.36, 0.75), linewidth=lw(2))
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 1.05)
    ax.grid(True, axis="y")
    ax = axs[1][0]
    ax.set_title("StackPlot")
    ax.stackplot([0, 1, 2, 3, 4, 5], [[1.0, 1.4, 1.3, 1.8, 1.6, 2.0], [0.8, 1.1, 1.4, 1.2, 1.6, 1.8], [0.5, 0.8, 1.0, 1.4, 1.1, 1.5]], colors=[color(0.20, 0.55, 0.75, 0.76), color(0.90, 0.48, 0.18, 0.76), color(0.35, 0.66, 0.42, 0.76)])
    ax.set_xlim(0, 5)
    ax.set_ylim(0, 7)
    ax.grid(True, axis="y")
    ax = axs[1][1]
    ax.set_title("Cumulative Multi-Hist")
    ax.hist([[0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2], [0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0], [1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4]], bins=[0, 1, 2, 3, 4, 5, 6], stacked=True, color=[color(0.22, 0.55, 0.70, 0.8), color(0.86, 0.42, 0.19, 0.8), color(0.36, 0.62, 0.36, 0.8)])
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 6)
    ax.grid(True, axis="y")
    save(fig, out_dir, "statistics")

DEMO = demo_statistics


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
