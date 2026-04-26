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

def demo_subplots(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.92, 0.08, 0.12)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    palette = [color(0.16, 0.42, 0.82), color(0.91, 0.45, 0.16), color(0.24, 0.72, 0.42), color(0.80, 0.20, 0.42)]
    x = linspace(0, 10, 120)
    for idx, ax in enumerate([axs[0][0], axs[0][1], axs[1][0], axs[1][1]]):
        row, col = divmod(idx, 2)
        ax.set_title(f"Panel {idx + 1}")
        ax.set_xlabel("x")
        ax.set_ylabel("y")
        ax.grid(True)
        y = np.sin((col + 1) * x + idx * 0.4) * np.exp(-0.12 * (row + 1) * x / 10)
        ax.plot(x, y, color=palette[idx], linewidth=lw(2.4))
    axs[0][0].set_xlim(0, 10)
    axs[0][0].set_ylim(-1.25, 1.25)
    save(fig, out_dir, "subplots")

DEMO = demo_subplots


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
