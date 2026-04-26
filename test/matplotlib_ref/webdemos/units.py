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

def demo_units(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(1, 3, 0.06, 0.98, 0.17, 0.86, 0.10, 0.08)
    axs = [fig.add_axes(rects[0][col]) for col in range(3)]
    ax = axs[0]
    ax.set_title("Dates")
    ax.set_ylabel("requests")
    ax.grid(True, axis="y")
    dates = [dt.datetime(2024, 1, d) for d in [1, 3, 7, 10]]
    ax.plot(dates, [12, 18, 9, 21], color=color(0.12, 0.47, 0.71), linewidth=lw(2))
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%Y-%m-%d"))
    ax.tick_params(axis="x", rotation=30)
    ax = axs[1]
    ax.set_title("Categories")
    ax.set_ylabel("count")
    ax.grid(True, axis="y")
    ax.bar(["draft", "review", "ship", "watch"], [3, 8, 6, 4], color=color(1.0, 0.50, 0.05), edgecolor=color(0.60, 0.30, 0.03), linewidth=lw(1))
    ax = axs[2]
    ax.set_title("Categorical Y")
    ax.set_xlabel("hours")
    ax.grid(True, axis="x")
    ax.barh(["north", "south", "east"], [4, 7, 5], color=color(0.17, 0.63, 0.17), edgecolor=color(0.09, 0.36, 0.09), linewidth=lw(1))
    save(fig, out_dir, "units")

DEMO = demo_units


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
