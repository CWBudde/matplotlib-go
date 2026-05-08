#!/usr/bin/env python3
"""Matplotlib reference for marker-mode spy parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def sparse_fixture_data(rows, cols):
    data = np.zeros((rows, cols))
    for row in range(rows):
        for col in range(cols):
            if row == col or row + col == cols - 1 or (col + 2 * row) % 5 == 0 or (2 * col + row) % 9 == 0:
                data[row, col] = 1
    return data


def spy_marker(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.22, 0.14, 0.78, 0.9))
    ax.set_title("Spy Marker")
    ax.spy(sparse_fixture_data(14, 14), precision=0.1, marker="s", markersize=8, color=(0.16, 0.38, 0.72, 1))
    save(fig, out_dir, "spy_marker")


PLOT = spy_marker


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
