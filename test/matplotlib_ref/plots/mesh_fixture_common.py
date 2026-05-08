from __future__ import annotations

import math

import numpy as np

try:
    from test.matplotlib_ref.common import make_fig_px, go_rect
except ModuleNotFoundError:
    from common import make_fig_px, go_rect


def mesh_fixture_axes(title):
    fig = make_fig_px(640, 360)
    ax = fig.add_axes(go_rect(0.12, 0.16, 0.90, 0.88))
    ax.set_title(title)
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    return fig, ax


def mesh_fixture_data(rows, cols):
    data = np.empty((rows, cols), dtype=float)
    for yi in range(rows):
        y = yi / max(1, rows - 1)
        for xi in range(cols):
            x = xi / max(1, cols - 1)
            data[yi, xi] = 0.65 * math.sin((x * 2.3 + 0.15) * math.pi) + 0.28 * math.cos(
                (y * 2.1 - 0.25) * math.pi
            )
    return data


def hist2d_weighted_data():
    x = [-1.8, -1.4, -0.8, -0.3, 0.2, 0.7, 1.1, 1.6, 2.1, 2.5, -0.6, 0.4, 1.3, 2.7]
    y = [-1.1, -0.4, -0.8, 0.1, 0.5, 0.9, 1.2, 1.7, 2.0, 2.2, 0.7, -0.2, 0.4, 1.3]
    weights = [0.8, 1.3, 0.7, 1.1, 1.6, 0.9, 1.4, 1.2, 1.8, 0.6, 1.5, 0.9, 1.1, 1.7]
    return x, y, weights
