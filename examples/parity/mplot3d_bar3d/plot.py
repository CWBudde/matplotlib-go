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


def mplot3d_bar3d(out_dir):
    fig = make_fig_px(720, 560)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.88, 0.88), projection="3d")

    x = [1, 1, 2, 2]
    y = [1, 2, 1, 2]
    z = [0, 0, 0, 0]
    dx = np.ones_like(x) * 0.5
    dy = np.ones_like(x) * 0.5
    dz = [2, 3, 1, 4]
    ax.bar3d(x, y, z, dx, dy, dz)
    ax.set(xticklabels=[], yticklabels=[], zticklabels=[])

    save(fig, out_dir, "mplot3d_bar3d")


PLOT = mplot3d_bar3d


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
