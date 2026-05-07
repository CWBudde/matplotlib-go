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


def sinusoidal_terrain(x_count, y_count):
    x = np.linspace(-math.pi, math.pi, x_count)
    y = np.linspace(-math.pi, math.pi, y_count)
    xx, yy = np.meshgrid(x, y)
    z = (
        0.5 * np.sin(xx) * np.cos(yy)
        + 0.35 * np.sin(2 * xx + 0.6) * np.cos(yy / 2)
        + 0.15 * np.cos(3 * yy - xx)
    )
    return x, y, xx, yy, z


def mplot3d_terrain(out_dir):
    fig = make_fig_px(900, 640)
    ax = fig.add_axes(go_rect(0.08, 0.08, 0.92, 0.88), projection="3d")
    ax.set(title="3D Surface + Filled Contours", xlabel="x", ylabel="y")
    ax.view_init(elev=35, azim=-60)

    x, y, xx, yy, z = sinusoidal_terrain(90, 70)
    ax.plot_surface(xx, yy, z, cmap="viridis", linewidth=0, alpha=0.85)
    ax.plot([0, 0.9, 0.9, 0, 0], [0, 0, 0.9, 0.9, 0], [-0.2] * 5, color="black")
    ax.scatter([0.2, 0.5, 0.8], [0.2, 0.5, 0.8], [0.3, 0.35, 0.2])
    ax.contour(xx, yy, z, levels=8, colors="black", linewidths=0.6)
    ax.contourf(xx, yy, z, levels=8, zdir="z", offset=z.min() - 0.2, cmap="viridis", alpha=0.45)
    ax.plot_trisurf([0, 0.5, 1], [0, 0, 0.4], [0.1, 0.4, 0.9], triangles=[[0, 1, 2]], color="tab:orange", alpha=0.7)
    ax.text(0.9, 0.1, 0.65, "3D demo")

    save(fig, out_dir, "mplot3d_terrain")


PLOT = mplot3d_terrain


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
