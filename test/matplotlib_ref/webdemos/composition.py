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

def demo_composition(out_dir, width, height):
    fig = make_fig(width, height)
    fig.suptitle("GridSpec, Figure Labels, Legend, Colorbar")
    fig.supxlabel("shared figure x")
    fig.supylabel("shared figure y")
    rects = grid_rects(2, 3, 0.08, 0.92, 0.14, 0.86, 0.08, 0.12, width_ratios=[1.3, 1, 0.9])
    left = fig.add_axes(span_rect(rects, 0, 0, 2, 1))
    left.set_title("spanning axes")
    left.grid(True, axis="y")
    left.plot([0, 1, 2, 3, 4], [1.0, 1.6, 1.2, 2.2, 1.8], color=color(0.16, 0.42, 0.82), linewidth=lw(2), label="left")
    left.scatter([0, 1, 2, 3, 4], [1.0, 1.6, 1.2, 2.2, 1.8], color=color(0.91, 0.45, 0.16), s=ss(6), label="points")
    top = fig.add_axes(rects[0][1], sharex=left)
    top.set_title("shared x")
    top.plot([0, 1, 2, 3, 4], [2, 1, 2.4, 1.7, 2.8], color=color(0.23, 0.62, 0.34), linewidth=lw(1.8), label="top")
    bottom = fig.add_axes(rects[1][1])
    bottom.set_title("anchored")
    bottom.grid(True)
    bottom.plot([0, 1, 2, 3, 4], [3.0, 2.6, 2.9, 2.1, 1.9], color=color(0.69, 0.27, 0.67), linewidth=lw(1.8), label="bottom")
    bottom.text(0.02, 0.98, "axes note", transform=bottom.transAxes, ha="left", va="top", bbox={"facecolor": "white", "edgecolor": "0.7"})
    heat = fig.add_axes(span_rect(rects, 0, 2, 2, 1))
    heat.set_title("colorbar")
    im = heat.imshow([[0.2, 0.5, 0.7], [0.9, 0.4, 0.1], [0.3, 0.8, 0.6]], cmap="inferno", origin="upper")
    fig.colorbar(im, ax=heat, label="intensity")
    fig.legend()
    fig.text(0.98, 0.02, "figure note", ha="right", va="bottom", bbox={"facecolor": "white", "edgecolor": "0.7"})
    save(fig, out_dir, "composition")

DEMO = demo_composition


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
