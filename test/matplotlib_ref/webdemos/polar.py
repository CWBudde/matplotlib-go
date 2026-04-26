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

def demo_polar(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect(0.12, 0.10, 0.88, 0.88), projection="polar")
    ax.set_title("Polar Wave")
    ax.set_xlabel("theta")
    ax.set_ylabel("radius")
    theta = linspace(0, 2 * math.pi, 720)
    radius = 0.55 + 0.28 * np.cos(4 * theta) + 0.08 * np.sin(9 * theta)
    ax.set_ylim(0, 1.1)
    ax.grid(color=color(0.80, 0.82, 0.86), linewidth=lw(0.9))
    ax.fill(theta, radius, color=color(0.36, 0.56, 0.92), edgecolor=color(0.14, 0.25, 0.52), linewidth=lw(1.0), alpha=0.24, label="filled area")
    ax.plot(theta, radius, color=color(0.16, 0.33, 0.73), linewidth=lw(2.2), label="r(theta)")
    ax.legend()
    save(fig, out_dir, "polar")

DEMO = demo_polar


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
