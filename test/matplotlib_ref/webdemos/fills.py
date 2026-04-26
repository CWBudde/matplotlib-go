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

def demo_fills(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Filled Signals")
    ax.set_xlabel("t")
    ax.set_ylabel("value")
    ax.grid(True)
    x = linspace(0, 2 * math.pi, 180)
    upper = 0.85 * np.sin(x) + 0.22 * np.cos(2 * x - 0.4)
    lower = -0.45 * np.cos(x - 0.2) - 0.18 * np.sin(2.4 * x)
    ax.fill_between(x, upper, lower, color=color(0.22, 0.60, 0.88), edgecolor=color(0.09, 0.30, 0.48), linewidth=lw(1.1), alpha=0.30, label="band")
    ax.plot(x, upper, color=color(0.10, 0.24, 0.62), linewidth=lw(2.2), label="upper")
    ax.plot(x, lower, color=color(0.86, 0.34, 0.18), linewidth=lw(2.2), dashes=[9, 5], label="lower")
    ax.set_xlim(0, 2 * math.pi)
    ax.set_ylim(-1.25, 1.25)
    ax.legend()
    save(fig, out_dir, "fills")

DEMO = demo_fills


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
