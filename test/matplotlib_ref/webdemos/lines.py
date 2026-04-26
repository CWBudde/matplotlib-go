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

def demo_lines(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Signal Comparison")
    ax.set_xlabel("t")
    ax.set_ylabel("amplitude")
    ax.grid(True)
    x = linspace(0, 12, 160)
    ax.plot(x, np.sin(x), color=color(0.16, 0.42, 0.82), linewidth=lw(3.0), label="sin(t)")
    ax.plot(x, 0.7 * np.cos(0.7 * x + 0.3), color=color(0.91, 0.45, 0.16), linewidth=lw(2.2), dashes=[10, 6], label="0.7 cos(0.7t + 0.3)")
    ax.plot(x, np.sin(1.6 * x) * np.exp(-x / 11), color=color(0.13, 0.62, 0.38), linewidth=lw(2.2), label="damped")
    ax.set_xlim(0, 12)
    ax.set_ylim(-1.4, 1.4)
    ax.legend()
    save(fig, out_dir, "lines")

DEMO = demo_lines


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
