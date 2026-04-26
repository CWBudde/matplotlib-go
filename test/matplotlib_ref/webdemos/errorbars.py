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

def demo_errorbars(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Measured Trend With Error Bars")
    ax.set_xlabel("sample")
    ax.set_ylabel("response")
    ax.grid(True)
    x = np.array([1, 2, 3, 4, 5, 6])
    y = np.array([1.8, 2.5, 2.2, 3.1, 2.8, 3.7])
    ax.plot(x, y, color=color(0.12, 0.47, 0.71), linewidth=lw(2.1), label="trend")
    ax.scatter(x, y, color=color(0.17, 0.63, 0.17), s=ss(5), label="samples")
    ax.errorbar(x, y, xerr=[0.20, 0.25, 0.15, 0.22, 0.30, 0.18], yerr=[0.28, 0.20, 0.35, 0.24, 0.30, 0.22], fmt="none", ecolor=color(0.10, 0.12, 0.16), linewidth=lw(1.2), capsize=7, label="1sigma")
    ax.set_xlim(0.4, 6.6)
    ax.set_ylim(1.0, 4.3)
    ax.legend()
    save(fig, out_dir, "errorbars")

DEMO = demo_errorbars


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
