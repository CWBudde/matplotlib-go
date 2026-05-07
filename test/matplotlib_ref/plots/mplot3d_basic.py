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


def mplot3d_basic(out_dir):
    fig = make_fig_px(760, 560)
    ax = fig.add_axes(go_rect(0.12, 0.14, 0.88, 0.88), projection="3d")
    ax.set(title="3D Toolkit Scaffold", xlabel="x", ylabel="y")
    ax.view_init(elev=30, azim=-60)

    x = np.array([0, 1])
    y = np.array([0, 1])
    xx, yy = np.meshgrid(x, y)
    zz = np.array([[0, 1], [1, 2]])
    ax.plot([0, 1], [0, 1], [0, 1])
    ax.scatter([0.5, 0.7], [0.2, 0.9], [0.1, 0.3])
    ax.plot_wireframe(xx, yy, zz, color="0.35")
    ax.plot_surface(xx, yy, zz, alpha=0.35, cmap="viridis")
    ax.contour(xx, yy, zz, zdir="z")
    ax.bar3d([0.2], [0.3], [0.4], [0.2], [0.2], [0.3], color="tab:orange", alpha=0.7)
    ax.text(0.2, 0.8, 0.6, "demo point")

    save(fig, out_dir, "mplot3d_basic")


PLOT = mplot3d_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
