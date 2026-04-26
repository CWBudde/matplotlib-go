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

def demo_projections(out_dir, width, height):
    fig = make_fig(width, height)
    geo = fig.add_axes(rect(0.06, 0.16, 0.48, 0.84), projection="mollweide")
    geo.set_title("Mollweide Projection")
    geo.set_xlabel("longitude")
    geo.set_ylabel("latitude")
    geo.grid(color=color(0.78, 0.80, 0.84), linewidth=lw(0.8))
    lon = linspace(-math.pi, math.pi, 241)
    geo.plot(lon, 0.35 * np.sin(3 * lon), color=color(0.14, 0.34, 0.70), linewidth=lw(2.0))
    ax = fig.add_axes(rect(0.57, 0.16, 0.96, 0.84))
    ax.set_title("Zoomed Inset")
    ax.set_xlabel("x")
    ax.set_ylabel("response")
    ax.set_xlim(0, 10)
    ax.set_ylim(-1.2, 1.2)
    ax.grid(True)
    x = linspace(0, 10, 320)
    y = np.sin(x) * (0.75 + 0.20 * np.cos(2 * x))
    ax.plot(x, y, color=color(0.12, 0.36, 0.72), linewidth=lw(2.0))
    ins = inset_axes(ax, width="43%", height="40%", loc="upper right", borderpad=0.5)
    ins.set_title("detail")
    ins.plot(x, y, color=color(0.12, 0.36, 0.72), linewidth=lw(1.6))
    ins.set_xlim(2, 4)
    ins.set_ylim(-0.2, 1.05)
    ins.grid(True)
    mark_inset(ax, ins, loc1=2, loc2=4, fc="none", ec="0.5")
    save(fig, out_dir, "projections")

DEMO = demo_projections


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
