#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

from mpl_toolkits.mplot3d import Axes3D  # noqa: F401

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def mplot3d_fill_between3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    n = 50
    theta = np.linspace(0, 2 * np.pi, n)
    x1 = np.cos(theta)
    y1 = np.sin(theta)
    z1 = np.linspace(0, 1, n)
    x2 = np.cos(theta + np.pi)
    y2 = np.sin(theta + np.pi)
    z2 = z1

    ax.plot(x1, y1, z1, linewidth=2, color="C0")
    ax.plot(x2, y2, z2, linewidth=2, color="C0")
    ax.fill_between(x1, y1, z1, x2, y2, z2, alpha=0.5)

    save(fig, out_dir, "mplot3d_fill_between3d")


PLOT = mplot3d_fill_between3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
