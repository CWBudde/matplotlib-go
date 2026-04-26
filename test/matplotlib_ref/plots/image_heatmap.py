#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def image_heatmap(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.15, 0.95, 0.9))
    ax.set_title("Image Heatmap")
    ax.set_xlim(0, 3)
    ax.set_ylim(0, 3)
    data = np.array([
        [0, 1, 2],
        [3, 4, 5],
        [6, 7, 8],
    ], dtype=float)
    ax.imshow(
        data,
        cmap="viridis",
        interpolation="nearest",
        origin="lower",
        extent=(0, 3, 0, 3),
        aspect="auto",
        vmin=data.min(),
        vmax=data.max(),
    )
    save(fig, out_dir, "image_heatmap")

PLOT = image_heatmap


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
