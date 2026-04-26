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

def demo_heatmap(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Heatmap Surface")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    data = heatmap_data()
    ax.imshow(data, origin="upper", cmap="inferno", extent=[0, data.shape[1], data.shape[0], 0], aspect="auto")
    ax.set_xlim(0, data.shape[1])
    ax.set_ylim(0, data.shape[0])
    save(fig, out_dir, "heatmap")

DEMO = demo_heatmap


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
