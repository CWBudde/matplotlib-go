#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import math

import matplotlib.pyplot as plt
import numpy as np


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")

from mpl_toolkits.mplot3d import Axes3D  # noqa: F401


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for mplot3d/basic.go")
    parser.add_argument("--out", default="mplot3d_basic_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(7.6, 5.6), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.14, 0.76, 0.74], projection="3d")
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
    ax.bar3d([0.2], [0.3], [0.4], [0.2], [0.2], [0.3], alpha=0.7)
    ax.text(0.2, 0.8, 0.6, "demo point")
    save(fig, args.out)


if __name__ == "__main__":
    main()
