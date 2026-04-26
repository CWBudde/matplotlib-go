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

def demo_annotations(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Coordinate Text + Arrow Annotation")
    ax.set_xlabel("x")
    ax.set_ylabel("response")
    ax.grid(True)
    x = linspace(0, 8, 120)
    y = np.sin(x) * np.exp(-x / 8)
    peak = int(np.argmax(y))
    ax.plot(x, y, color=color(0.13, 0.43, 0.72), linewidth=lw(2.2), label="signal")
    ax.annotate("peak", xy=(x[peak], y[peak]), xytext=(44, -34), textcoords="offset points", arrowprops={"arrowstyle": "->"})
    ax.text(0.03, 0.94, "axes coords", transform=ax.transAxes, ha="left", va="top", fontsize=11)
    fig.text(0.50, 0.10, "figure coords", ha="center", va="bottom", fontsize=11)
    ax.text(6.0, -0.55, "data coords", fontsize=11, color=color(0.56, 0.22, 0.18))
    ax.text(0.98, 0.02, "anchored\ntext box", transform=ax.transAxes, ha="right", va="bottom", bbox={"facecolor": "white", "edgecolor": "0.7"})
    ax.set_xlim(0, 8)
    ax.set_ylim(-0.8, 1.1)
    ax.legend()
    save(fig, out_dir, "annotations")

DEMO = demo_annotations


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
