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

def demo_vectors(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.91, 0.10, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("Quiver + Key")
    qx, qy, qu, qv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.8 + col
            y = 0.8 + row * 0.95
            qx.append(x)
            qy.append(y)
            qu.append(0.55 + 0.08 * math.sin(y * 0.9))
            qv.append(0.22 * math.cos(x * 0.8))
    q = ax.quiver(qx, qy, qu, qv, color=color(0.14, 0.42, 0.73), scale=10, units="dots", width=2.2)
    ax.quiverkey(q, 0.78, 0.12, 0.5, "0.5", coordinates="axes", labelpos="E")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax.grid(True)
    ax = axs[0][1]
    ax.set_title("Barbs")
    bx, by, bu, bv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.9 + col * 0.95
            y = 0.8 + row * 0.95
            bx.append(x)
            by.append(y)
            bu.append(14 + 5 * math.sin(y * 0.8))
            bv.append(8 * math.cos(x * 0.7))
    ax.barbs(bx, by, bu, bv, barbcolor=color(0.47, 0.23, 0.12), flagcolor=color(0.86, 0.52, 0.24), length=6)
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax.grid(True)
    ax = axs[1][0]
    ax.set_title("Streamplot")
    sx = np.array([0, 1, 2, 3, 4, 5, 6])
    sy = np.array([0, 1, 2, 3, 4, 5])
    su, sv = vector_grid(sx, sy)
    ax.streamplot(sx, sy, su, sv, start_points=np.array([[0.4, 0.8], [0.4, 2.2], [0.4, 3.6]]), color=color(0.13, 0.53, 0.39), integration_direction="forward", broken_streamlines=False)
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax = axs[1][1]
    ax.set_title("Quiver Grid XY")
    xg = np.array([0.8, 1.8, 2.8, 3.8, 4.8])
    yg = np.array([0.8, 1.8, 2.8, 3.8])
    xx, yy = np.meshgrid(xg, yg)
    ax.quiver(xx, yy, -(yy - 2.4) * 0.35, (xx - 2.8) * 0.35, color=color(0.74, 0.23, 0.27), pivot="middle", angles="xy", scale=9, units="dots", width=1.9)
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    save(fig, out_dir, "vectors")

DEMO = demo_vectors


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
