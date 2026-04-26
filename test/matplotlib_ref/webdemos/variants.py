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

def demo_variants(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.11, 0.91, 0.10, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("Step + Stairs")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5.2)
    ax.grid(True, axis="y")
    ax.step([0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4], [1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6], where="post", color=color(0.15, 0.39, 0.78), linewidth=lw(2), label="step")
    ax.stairs([0.9, 1.7, 1.4, 2.6, 1.8, 2.2], [0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5], baseline=0.35, fill=True, color=color(0.91, 0.49, 0.20, 0.72), edgecolor=color(0.58, 0.26, 0.08), linewidth=lw(1.5), label="stairs")
    ax.legend()
    ax = axs[0][1]
    ax.set_title("FillBetweenX + Refs")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 6)
    ax.grid(True, axis="x")
    ax.fill_betweenx([0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2], [1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6], [3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3], color=color(0.24, 0.68, 0.54, 0.72), edgecolor=color(0.12, 0.38, 0.28), linewidth=lw(1.2))
    ax.axvspan(2.2, 3.1, color=color(0.92, 0.75, 0.18), alpha=0.20)
    ax.axhline(4.0, color=color(0.52, 0.18, 0.18), linewidth=lw(1.2), dashes=[4, 3])
    ax.axline((0.9, 0.3), (6.4, 5.6), color=color(0.22, 0.22, 0.22), linewidth=lw(1.1))
    ax = axs[1][0]
    ax.set_title("Broken BarH")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 4.4)
    ax.grid(True, axis="x")
    ax.broken_barh([(0.8, 1.6), (3.1, 2.2), (6.5, 1.3)], (0.7, 0.9), facecolors=color(0.21, 0.51, 0.76))
    ax.broken_barh([(1.6, 1.0), (4.0, 1.4), (7.1, 1.7)], (2.1, 0.9), facecolors=color(0.86, 0.38, 0.16))
    for x, y, txt in [(1.6, 1.15, "prep"), (4.2, 1.15, "run"), (7.15, 1.15, "cool"), (2.1, 2.55, "IO"), (4.7, 2.55, "fit"), (7.95, 2.55, "ship")]:
        ax.text(x, y, txt, ha="center", va="center", color="white", fontsize=10)
    ax = axs[1][1]
    ax.set_title("Stacked Bars")
    ax.set_xlim(0.4, 4.6)
    ax.set_ylim(0, 7.6)
    ax.grid(True, axis="y")
    x = np.array([1, 2, 3, 4])
    a = np.array([1.4, 2.2, 1.8, 2.5])
    b = np.array([2.1, 1.6, 2.4, 1.7])
    bottom = ax.bar(x, a, color=color(0.16, 0.59, 0.49))
    top = ax.bar(x, b, bottom=a, color=color(0.88, 0.47, 0.16))
    ax.bar_label(bottom, labels=["A1", "A2", "A3", "A4"], label_type="center", color="white", fontsize=10)
    ax.bar_label(top, fmt="%.1f", color=color(0.20, 0.20, 0.20))
    save(fig, out_dir, "variants")

DEMO = demo_variants


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
