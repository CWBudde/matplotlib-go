#!/usr/bin/env python3
"""Matplotlib reference counterpart for examples/mesh/gouraud/main.go."""

from __future__ import annotations

from pathlib import Path
import argparse

import matplotlib

matplotlib.use("Agg")

import matplotlib.pyplot as plt
import numpy as np


def smooth_field(x, y):
    xx, yy = np.meshgrid(x, y)
    radius = np.hypot(xx * 0.82, yy * 1.10)
    return np.sin(1.8 * radius) * np.exp(-0.12 * radius * radius) + 0.18 * np.cos(
        1.6 * xx - 0.7 * yy
    )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=str(Path.cwd()))
    args = parser.parse_args()

    fig = plt.figure(figsize=(7.6, 4.6), dpi=100)
    ax = fig.add_axes([0.10, 0.14, 0.78, 0.76])
    ax.set_title("Gouraud PColorMesh")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    x = np.linspace(-3.0, 3.0, 9)
    y = np.linspace(-2.2, 2.2, 7)
    z = smooth_field(x, y)
    mesh = ax.pcolormesh(
        x,
        y,
        z,
        shading="gouraud",
        cmap="viridis",
        vmin=-0.85,
        vmax=0.85,
    )
    ax.set_xlim(x[0], x[-1])
    ax.set_ylim(y[0], y[-1])
    ax.grid(True)
    fig.colorbar(mesh, ax=ax, label="value")

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    fig.savefig(out_dir / "mesh_gouraud.png")


if __name__ == "__main__":
    main()
