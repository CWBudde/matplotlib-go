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

def demo_axes(out_dir, width, height):
    fig = make_fig(width, height)
    left = fig.add_axes(rect(0.08, 0.14, 0.42, 0.86))
    left.set_title("Top/Right + Equal Aspect")
    left.set_xlabel("top x")
    left.set_ylabel("right y")
    left.set_xlim(-1, 5)
    left.set_ylim(-1, 5)
    left.grid(True)
    left.xaxis.set_label_position("top")
    left.xaxis.tick_top()
    left.yaxis.set_label_position("right")
    left.yaxis.tick_right()
    left.set_aspect("equal", adjustable="box")
    left.minorticks_on()
    left.plot([-0.5, 0.8, 2.2, 4.2], [-0.2, 1.0, 2.1, 4.4], color=color(0.10, 0.32, 0.76), linewidth=lw(2))
    left.scatter([0, 1.5, 3.5, 4.5], [0, 1.8, 3.2, 4.6], color=color(0.92, 0.48, 0.20, 0.92), edgecolor=color(0.52, 0.22, 0.08), linewidth=lw(1), s=ss(8))
    right = fig.add_axes(rect(0.56, 0.14, 0.94, 0.86))
    right.set_title("Log, Twin, Secondary")
    right.set_xlabel("seconds")
    right.set_ylabel("count")
    right.set_xlim(0, 10)
    right.set_ylim(1, 100)
    right.set_yscale("log")
    right.grid(True)
    right.plot([0, 2, 4, 6, 8, 10], [2, 6, 9, 18, 40, 82], color=color(0.12, 0.45, 0.72), linewidth=lw(2), label="log series")
    twin = right.twinx()
    twin.set_ylim(0, 100)
    twin.plot([0, 2, 4, 6, 8, 10], [10, 22, 38, 58, 81, 96], color=color(0.80, 0.22, 0.22), linewidth=lw(1.8), label="twin")
    right.secondary_xaxis("top", functions=(lambda x: x * 10, lambda x: x / 10))
    right.legend()
    save(fig, out_dir, "axes")

DEMO = demo_axes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
