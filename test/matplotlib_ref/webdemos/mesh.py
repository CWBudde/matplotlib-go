#!/usr/bin/env python3
"""Split Matplotlib web demo module generated from test/matplotlib_ref/webdemo.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403

def demo_mesh(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.91, 0.12, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("PColorMesh")
    ax.pcolormesh([0, 1, 2, 3, 4], [0, 1, 2, 3], [[0.2, 0.6, 0.3, 0.9], [0.4, 0.8, 0.5, 0.7], [0.1, 0.3, 0.9, 0.6]], edgecolors=color(0.95, 0.95, 0.95), linewidth=lw(0.8))
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    ax = axs[0][1]
    ax.set_title("Contour + Contourf")
    z = np.array([[0, 0.4, 0.8, 0.4, 0], [0.2, 0.8, 1.3, 0.8, 0.2], [0.3, 1.0, 1.7, 1.0, 0.3], [0.2, 0.8, 1.3, 0.8, 0.2], [0, 0.4, 0.8, 0.4, 0]])
    ax.contourf(z, levels=[0.2, 0.6, 1.0, 1.4, 1.8])
    ax.contour(z, levels=[0.4, 0.8, 1.2, 1.6], colors=[color(0.18, 0.18, 0.18)])
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 4)
    ax = axs[1][0]
    ax.set_title("Hist2D")
    ax.hist2d([0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6], [0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4], bins=[[0, 1, 2, 3, 4], [0, 1, 2, 3, 4]])
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 4)
    ax = axs[1][1]
    ax.set_title("Triangulation")
    tx = [0.4, 1.6, 3.0, 0.8, 2.1, 3.5]
    ty = [0.5, 0.4, 0.7, 2.2, 2.8, 2.1]
    triangles = [[0, 1, 3], [1, 4, 3], [1, 2, 4], [2, 5, 4]]
    vals = [0.2, 0.8, 1.0, 1.5, 1.1, 0.6]
    ax.tripcolor(tx, ty, triangles, vals)
    ax.triplot(tx, ty, triangles, color=color(0.15, 0.15, 0.15), linewidth=lw(1))
    ax.tricontour(tx, ty, triangles, vals, levels=[0.7, 1.1], colors=[color(0.98, 0.98, 0.98)])
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 4)
    save(fig, out_dir, "mesh")

DEMO = demo_mesh


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
