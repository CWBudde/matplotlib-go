#!/usr/bin/env python3
"""Matplotlib reference for transformed imshow parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.transforms as mtransforms

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def wave_image_data(rows, cols):
    y = np.linspace(0, 1, rows)
    x = np.linspace(0, 1, cols)
    yy, xx = np.meshgrid(y, x, indexing="ij")
    return 0.5 + 0.25 * np.sin((xx * 2.4 + 0.15) * np.pi) + 0.22 * np.cos((yy * 2.7 - 0.2) * np.pi)


def imshow_transformed(out_dir):
    fig = make_fig_px(420, 420)
    ax = fig.add_axes(go_rect(0.16, 0.14, 0.9, 0.88))
    ax.set_title("Transformed Imshow")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(-1, 5)
    ax.set_ylim(-1, 5)
    ax.set_aspect("equal")
    trans = mtransforms.Affine2D().rotate_deg_around(2, 2, 28) + ax.transData
    ax.imshow(
        wave_image_data(6, 6),
        cmap="magma",
        vmin=0,
        vmax=1,
        interpolation="bilinear",
        origin="lower",
        extent=(0, 4, 0, 4),
        transform=trans,
    )
    save(fig, out_dir, "imshow_transformed")


PLOT = imshow_transformed


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
