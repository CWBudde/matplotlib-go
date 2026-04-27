#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

from matplotlib.offsetbox import AnchoredText

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403

def axes_grid1_showcase(out_dir):
    fig = make_fig_px(1100, 720)

    grid = ImageGrid(
        fig,
        go_rect(0.06, 0.12, 0.60, 0.88),
        nrows_ncols=(2, 2),
        axes_pad=(0.18, 0.20),
        share_all=False,
    )

    for idx, ax in enumerate(grid):
        row = idx // 2
        col = idx % 2
        ax.set_title(f"Tile {row + 1},{col + 1}")
        rows, cols = 24, 24
        data = np.zeros((rows, cols))
        phase = float(row * 2 + col)
        for y in range(rows):
            yy = y / float(rows - 1)
            for x in range(cols):
                xx = x / float(cols - 1)
                data[y, x] = 0.5 + 0.25 * math.sin((xx + phase) * 2 * math.pi) + 0.25 * math.cos((yy + phase * 0.3) * 3 * math.pi)
        ax.imshow(data, origin="upper")
        ax.text(
            0.98,
            0.02,
            "image grid",
            transform=ax.transAxes,
            ha="right",
            va="bottom",
            fontsize=9,
            bbox=dict(boxstyle="round,pad=0.25", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
        )

    channel_cmaps = {
        "Red": mcolors.LinearSegmentedColormap.from_list("red channel", [(0.18, 0.02, 0.02), (1.00, 0.18, 0.12)]),
        "Green": mcolors.LinearSegmentedColormap.from_list("green channel", [(0.02, 0.14, 0.05), (0.20, 0.90, 0.28)]),
        "Blue": mcolors.LinearSegmentedColormap.from_list("blue channel", [(0.02, 0.05, 0.18), (0.18, 0.45, 1.00)]),
    }
    channels = [
        (fig.add_axes(go_rect(0.66, 0.34, 0.75, 0.56)), "Red", 0),
        (fig.add_axes(go_rect(0.775, 0.34, 0.865, 0.56)), "Green", 1),
        (fig.add_axes(go_rect(0.89, 0.34, 0.98, 0.56)), "Blue", 2),
    ]
    for ax, title, phase in channels:
        rows, cols = 28, 28
        data = np.zeros((rows, cols))
        for y in range(rows):
            yy = y / float(rows - 1)
            for x in range(cols):
                xx = x / float(cols - 1)
                if phase == 1:
                    data[y, x] = 0.5 + 0.32 * math.sin(yy * 4 * math.pi) + 0.18 * math.cos(xx * 2 * math.pi)
                elif phase == 2:
                    dx = xx - 0.5
                    dy = yy - 0.5
                    data[y, x] = 0.58 + 0.36 * math.sin((xx + yy) * 3 * math.pi) - 0.18 * math.hypot(dx, dy)
                else:
                    dx = xx - 0.35
                    dy = yy - 0.42
                    data[y, x] = 0.35 + 0.65 * math.exp(-7 * (dx * dx + dy * dy))
        ax.set_title(title)
        ax.set_xticks([0, 10, 20])
        ax.set_yticks([0, 10, 20])
        ax.imshow(data, origin="upper", cmap=channel_cmaps[title])

    note = AnchoredText(
        "axes_grid1-style layout\nImageGrid + RGB channel views",
        loc="upper right",
        prop={"size": 11},
        pad=0.35,
        borderpad=0.5,
        frameon=True,
        bbox_to_anchor=(0, 0, 1, 1),
        bbox_transform=fig.transFigure,
    )
    note.patch.set_boxstyle("round,pad=0.35")
    note.patch.set_facecolor("white")
    note.patch.set_edgecolor((0.75, 0.75, 0.75, 1.0))
    fig.add_artist(note)

    save(fig, out_dir, "axes_grid1_showcase")

PLOT = axes_grid1_showcase


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
