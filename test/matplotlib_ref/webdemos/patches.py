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

def demo_patches(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Patch Showcase")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 4)
    patches = [
        Rectangle((0.5, 0.6), 1.5, 1.0, facecolor=color(0.95, 0.70, 0.23, 0.86), edgecolor=color(0.48, 0.27, 0.08), linewidth=lw(1.1), label="rectangle"),
        Circle((2.8, 1.25), 0.56, facecolor=color(0.22, 0.57, 0.82, 0.82), edgecolor=color(0.11, 0.29, 0.44), linewidth=lw(1.0), label="circle"),
        Ellipse((4.9, 2.85), 1.4, 0.92, angle=24, facecolor=color(0.23, 0.72, 0.51, 0.80), edgecolor=color(0.10, 0.36, 0.24), linewidth=lw(1.0), label="ellipse"),
        Polygon([(1.6, 3.0), (2.2, 2.1), (0.9, 2.3)], facecolor=color(0.84, 0.34, 0.34, 0.82), edgecolor=color(0.48, 0.14, 0.14), linewidth=lw(1.0), label="polygon"),
    ]
    for patch in patches:
        ax.add_patch(patch)
    ax.add_patch(FancyArrow(3.4, 0.8, 1.4, 1.1, width=0.16, head_width=0.48, head_length=0.42, facecolor=color(0.91, 0.42, 0.22, 0.88), edgecolor=color(0.58, 0.22, 0.10), linewidth=lw(1.0), label="arrow"))
    ax.legend()
    save(fig, out_dir, "patches")

DEMO = demo_patches


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
