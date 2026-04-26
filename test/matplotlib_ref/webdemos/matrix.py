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

def demo_matrix(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(1, 3, 0.07, 0.92, 0.14, 0.86, 0.10, 0.06)
    axs = [fig.add_axes(rects[0][col]) for col in range(3)]
    axs[0].set_title("MatShow")
    axs[0].imshow([[0.1, 0.5, 0.9], [0.7, 0.3, 0.2], [0.4, 0.8, 0.6]], cmap="viridis", origin="upper")
    axs[1].set_title("Spy")
    axs[1].spy([[1, 0, 0, 2, 0], [0, 0, 3, 0, 0], [4, 0, 0, 0, 5], [0, 6, 0, 0, 0]], color=color(0.13, 0.43, 0.72), markersize=8)
    axs[2].set_title("Annotated Heatmap")
    data = np.array([[0.1, 0.7, 0.4], [0.9, 0.2, 0.5], [0.3, 0.8, 0.6]])
    im = axs[2].imshow(data, cmap="magma", origin="upper")
    for y in range(3):
        for x in range(3):
            axs[2].text(x, y, f"{data[y, x]:.1f}", ha="center", va="center", color="white" if data[y, x] > 0.5 else color(0.05, 0.05, 0.05), fontsize=9)
    fig.colorbar(im, ax=axs[2], label="value")
    save(fig, out_dir, "matrix")

DEMO = demo_matrix


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
