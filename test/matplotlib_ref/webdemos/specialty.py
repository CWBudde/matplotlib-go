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

def demo_specialty(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 3, 0.07, 0.98, 0.09, 0.91, 0.10, 0.14)
    axs = [[fig.add_axes(rects[row][col]) for col in range(3)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("Eventplot")
    ax.eventplot([[0.8, 1.4, 3.1, 4.6, 7.3], [1.2, 2.9, 4.0, 6.4, 8.6], [0.5, 2.2, 5.4, 6.8, 9.1]], lineoffsets=[1, 2, 3], linelengths=[0.6, 0.7, 0.8], colors=[color(0.18, 0.44, 0.74), color(0.84, 0.38, 0.16), color(0.20, 0.63, 0.42)])
    ax.set_xlim(0, 10)
    ax.set_ylim(0.4, 3.6)
    ax = axs[0][1]
    ax.set_title("Hexbin")
    ax.hexbin([0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88], [0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58], C=[1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1], gridsize=7, reduce_C_function=np.mean)
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax = axs[0][2]
    ax.set_title("Pie")
    ax.pie([28, 22, 18, 32], labels=["Core", "I/O", "Render", "Docs"], autopct="%.0f%%", startangle=90, labeldistance=1.08, explode=[0, 0.04, 0, 0.02])
    ax = axs[1][0]
    ax.set_title("Stem")
    markerline, stemlines, baseline = ax.stem([1, 2, 3, 4, 5, 6, 7], [0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1], basefmt=" ")
    plt.setp(stemlines, color=color(0.15, 0.42, 0.73))
    plt.setp(markerline, color=color(0.15, 0.42, 0.73), markersize=7)
    ax.axhline(0.3, color="0.4")
    ax.set_xlim(0.5, 7.5)
    ax.set_ylim(-0.2, 4.2)
    ax.grid(True, axis="y")
    ax = axs[1][1]
    ax.set_title("Table")
    ax.axis("off")
    ax.table(cellText=[["Latency", "18ms", "14ms"], ["Throughput", "220/s", "265/s"]], rowLabels=["A", "B"], colLabels=["Metric", "Q1", "Q2"], bbox=[0.04, 0.18, 0.92, 0.64])
    ax = axs[1][2]
    ax.set_title("Sankey")
    ax.axis("off")
    Sankey(ax=ax, scale=0.16, offset=0.2).add(flows=[-2, 3, 1.5], labels=["Waste", "CPU", "Cache"], orientations=[-1, 1, 1]).finish()
    save(fig, out_dir, "specialty")

DEMO = demo_specialty


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
