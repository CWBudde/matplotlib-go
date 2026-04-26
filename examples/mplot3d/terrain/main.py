#!/usr/bin/env python3
from __future__ import annotations

import argparse
import math

import matplotlib.pyplot as plt
import numpy as np


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")

from mpl_toolkits.mplot3d import Axes3D  # noqa: F401


def sinusoidal_terrain(x_count, y_count):
    # Use the same deterministic terrain formula as the Go counterpart so
    # surface, contour, and contourf behavior can be compared directly.
    x = np.linspace(-math.pi, math.pi, x_count)
    y = np.linspace(-math.pi, math.pi, y_count)
    xx, yy = np.meshgrid(x, y)
    z = 0.5 * np.sin(xx) * np.cos(yy) + 0.35 * np.sin(2 * xx + 0.6) * np.cos(yy / 2) + 0.15 * np.cos(3 * yy - xx)
    return x, y, xx, yy, z


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for mplot3d/terrain/main.go")
    parser.add_argument("--out", default="mplot3d_terrain_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(9, 6.4), dpi=100, facecolor="white")
    ax = fig.add_axes([0.08, 0.08, 0.84, 0.80], projection="3d")
    ax.set(title="3D Surface + Filled Contours", xlabel="x", ylabel="y")
    ax.view_init(elev=35, azim=-60)
    x, y, xx, yy, z = sinusoidal_terrain(90, 70)
    ax.plot_surface(xx, yy, z, cmap="viridis", linewidth=0, alpha=0.85)

    # Additional primitives exercise mixed 3D artist ordering on top of the
    # surface: a floor outline, sample points, a triangular patch, and text.
    ax.plot([0, 0.9, 0.9, 0, 0], [0, 0, 0.9, 0.9, 0], [-0.2] * 5, color="black")
    ax.scatter([0.2, 0.5, 0.8], [0.2, 0.5, 0.8], [0.3, 0.35, 0.2])
    ax.contour(xx, yy, z, levels=8, colors="black", linewidths=0.6)
    ax.contourf(xx, yy, z, levels=8, zdir="z", offset=z.min() - 0.2, cmap="viridis", alpha=0.45)
    ax.plot_trisurf([0, 0.5, 1], [0, 0, 0.4], [0.1, 0.4, 0.9], triangles=[[0, 1, 2]], color="tab:orange", alpha=0.7)
    ax.text(0.9, 0.1, 0.65, "3D demo")
    save(fig, args.out)


if __name__ == "__main__":
    main()
