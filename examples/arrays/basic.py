#!/usr/bin/env python3

import argparse
import math

import matplotlib.pyplot as plt
import numpy as np
from matplotlib.offsetbox import AnchoredText


WIDTH_PX = 1240
HEIGHT_PX = 620
DPI = 100


def axes_rect(x0, y0, x1, y1):
    # Keep axes placement readable as lower-left and upper-right figure
    # fractions, matching the Go helper.
    return [x0, y0, x1 - x0, y1 - y0]


def arange(n):
    # Mirror the Go arange helper for integer tick and edge coordinates.
    return np.arange(n)


def annotated_data():
    # Scalar matrix shown in the first panel.
    return np.array(
        [
            [0.12, 0.28, 0.46, 0.64, 0.82],
            [0.18, 0.34, 0.58, 0.74, 0.88],
            [0.24, 0.42, 0.63, 0.79, 0.91],
            [0.16, 0.38, 0.61, 0.83, 0.97],
        ]
    )


def wave_grid(rows, cols, phase):
    # Shared scalar field for pcolormesh and contour.
    values = np.zeros((rows, cols))
    for y in range(rows):
        yy = y / (rows - 1)
        for x in range(cols):
            xx = x / (cols - 1)
            values[y, x] = (
                0.55
                + 0.25 * math.sin((xx * 2.3 + phase) * math.pi)
                + 0.20 * math.cos((yy * 2.8 - phase * 0.4) * math.pi)
            )
    return values


def sparse_pattern(rows, cols):
    # Deterministic non-zero pattern for the spy panel.
    values = np.zeros((rows, cols))
    for y in range(rows):
        for x in range(cols):
            if (
                x == y
                or x + y == cols - 1
                or (x + 2 * y) % 7 == 0
                or (2 * x + y) % 11 == 0
            ):
                values[y, x] = 1
    return values


def use_matrix_top_axis(ax):
    # Matrix examples read column labels more naturally above the image.
    ax.xaxis.tick_top()
    ax.xaxis.set_label_position("top")


def set_matrix_ticks(ax, rows, cols):
    # Pin matrix axes to integer row/column indices.
    ax.set_xticks(arange(cols))
    ax.set_yticks(arange(rows))


def add_anchored_text(target, text, location):
    # Centralize the boxed-note style so figure-level and axes-level notes match.
    kwargs = {}
    if hasattr(target, "transFigure"):
        kwargs = {
            "bbox_to_anchor": (0, 0, 1, 1),
            "bbox_transform": target.transFigure,
        }

    anchored = AnchoredText(
        text,
        loc=location,
        prop={"size": 10},
        pad=0.3,
        borderpad=0.5,
        frameon=True,
        **kwargs,
    )
    anchored.patch.set_boxstyle("round,pad=0.3")
    anchored.patch.set_facecolor("white")
    anchored.patch.set_edgecolor((0.75, 0.75, 0.75, 1.0))
    target.add_artist(anchored)


def draw_annotated_heatmap(fig):
    data = annotated_data()
    rows, cols = data.shape

    # Add axes explicitly so the three examples share the same layout in Go and
    # Python.
    ax = fig.add_axes(axes_rect(0.05, 0.14, 0.31, 0.88))
    ax.set_title("MatShow Annotated Heatmap")
    ax.set_xlabel("column index")
    ax.set_ylabel("row")

    # imshow places matrix cell centers on integer row/column coordinates.
    ax.imshow(data, cmap="viridis", aspect="equal", origin="upper")
    ax.set_xlim(-0.5, cols - 0.5)
    ax.set_ylim(rows - 0.5, -0.5)
    ax.set_aspect("equal")
    set_matrix_ticks(ax, rows, cols)
    use_matrix_top_axis(ax)

    threshold = 0.5 * (0.12 + 0.97)
    # Cell labels use a contrasting color for readability on the colormap.
    for row in range(rows):
        for col in range(cols):
            text_color = (
                "white"
                if data[row, col] >= threshold
                else (0.12, 0.12, 0.14, 1.0)
            )
            ax.text(
                col,
                row,
                f"{data[row, col]:.2f}",
                ha="center",
                va="center",
                fontsize=10,
                color=text_color,
            )


def draw_mesh_and_contour(fig):
    rows, cols = 8, 10
    data = wave_grid(rows, cols, 0.35)

    ax = fig.add_axes(axes_rect(0.37, 0.14, 0.63, 0.88))
    ax.set_title("PColorMesh + Contour")
    ax.set_xlabel("x bin")
    ax.set_ylabel("y bin")

    # PColorMesh consumes edge coordinates, so both axes have one more value
    # than the data has cells.
    ax.pcolormesh(
        arange(cols + 1),
        arange(rows + 1),
        data,
        cmap="plasma",
        edgecolors=(1.0, 1.0, 1.0, 0.48),
        linewidth=0.65,
        shading="flat",
    )

    # Contour overlays use cell-center coordinates for the same data.
    contour = ax.contour(
        arange(cols),
        arange(rows),
        data,
        levels=6,
        colors=[(0.14, 0.10, 0.16, 0.95)],
        linewidths=1.1,
    )
    ax.clabel(
        contour,
        inline=True,
        fmt="%.3g",
        fontsize=10,
        colors=[(0.14, 0.10, 0.16, 0.95)],
    )
    # Match the visible mesh extent exactly, including the outer cell edges.
    ax.set_xlim(0, cols)
    ax.set_ylim(0, rows)


def draw_spy_matrix(fig):
    data = sparse_pattern(18, 18)

    # The spy panel is another matrix view, so it reuses the same top-axis
    # helpers as the heatmap.
    ax = fig.add_axes(axes_rect(0.69, 0.14, 0.95, 0.88))
    ax.set_title("Spy Matrix")
    ax.set_xlabel("column index")
    ax.set_ylabel("row")

    # spy converts non-zero matrix entries into square markers and applies the
    # same matrix-style limits, aspect, y inversion, and tick locator as Go.
    ax.spy(
        data,
        precision=0.1,
        marker="s",
        markersize=10,
        color=(0.16, 0.38, 0.72, 1.0),
    )

    add_anchored_text(ax, "sparse structure view", "lower right")


def save_figure(fig, out):
    # Python hides backend construction inside savefig; Go does this explicitly.
    fig.savefig(out, dpi=DPI)
    plt.close(fig)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", default="arrays_basic_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(WIDTH_PX / DPI, HEIGHT_PX / DPI), dpi=DPI)
    fig.patch.set_facecolor("white")
    draw_annotated_heatmap(fig)
    draw_mesh_and_contour(fig)
    draw_spy_matrix(fig)
    # Figure-level anchored text mirrors the small gallery label in Go.
    add_anchored_text(
        fig,
        "arrays gallery family\nheatmap, quad mesh, sparse matrix",
        "upper right",
    )
    save_figure(fig, args.out)


if __name__ == "__main__":
    main()
