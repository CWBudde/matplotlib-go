#!/usr/bin/env python3
"""Matplotlib reference plot for RendererAgg Gouraud triangle coverage."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

import matplotlib.ticker as mticker

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def rendereragg_gouraud_triangles(out_dir):
    fig = make_fig_px(980, 620)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.94, 0.88))
    ax.set_title("RendererAgg Gouraud triangles")
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3.2)
    ax.xaxis.set_major_locator(mticker.MultipleLocator(0.5))
    ax.yaxis.set_major_locator(mticker.MultipleLocator(0.5))

    x = np.array([0.35, 1.80, 3.40, 0.80, 2.20, 3.55], dtype=float)
    y = np.array([0.35, 0.30, 0.55, 1.70, 2.70, 1.75], dtype=float)
    triangles = np.array([[0, 1, 3], [1, 4, 3], [1, 2, 4], [2, 5, 4]], dtype=int)
    values = np.array([0.05, 0.38, 0.82, 0.62, 1.00, 0.28], dtype=float)
    tri = mtri.Triangulation(x, y, triangles)
    ax.tripcolor(tri, values, shading="gouraud", cmap="viridis", vmin=0, vmax=1)
    save(fig, out_dir, "rendereragg_gouraud_triangles")


PLOT = rendereragg_gouraud_triangles


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
