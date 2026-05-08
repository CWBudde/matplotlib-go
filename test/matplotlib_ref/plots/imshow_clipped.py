#!/usr/bin/env python3
"""Matplotlib reference for clipped imshow parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

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


def imshow_clipped(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.92, 0.88))
    ax.set_title("Clipped Imshow")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.imshow(
        wave_image_data(8, 8),
        cmap="viridis",
        interpolation="nearest",
        origin="lower",
        extent=(0, 8, 0, 8),
        aspect="auto",
    )
    ax.set_xlim(2, 6)
    ax.set_ylim(1, 7)
    save(fig, out_dir, "imshow_clipped")


PLOT = imshow_clipped


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
