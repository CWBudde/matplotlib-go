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

def demo_histogram(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Latency Distribution")
    ax.set_xlabel("latency (ms)")
    ax.set_ylabel("density")
    ax.grid(True)
    ax.hist(deterministic_normal(400, 47.0, 8.5), bins=24, density=True, color=color(0.42, 0.23, 0.77, 0.7), edgecolor=color(0.17, 0.12, 0.33), linewidth=lw(0.8), label="requests")
    ax.margins(0.05)
    ax.legend()
    save(fig, out_dir, "histogram")

DEMO = demo_histogram


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
