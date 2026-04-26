#!/usr/bin/env python3
from __future__ import annotations

import argparse

import matplotlib.pyplot as plt


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for lines/styles/main.go")
    parser.add_argument("--out", default="line_styles_python.png")
    args = parser.parse_args()
    fig = plt.figure(figsize=(8, 6), dpi=100, facecolor="white")
    ax = fig.add_axes([0.1, 0.1, 0.8, 0.8])
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 8)
    # Top row compares join styles on the same L-shaped path.
    for offset, color, join in [(0, (0.8, 0.2, 0.2), "miter"), (3, (0.2, 0.8, 0.2), "round"), (6, (0.2, 0.2, 0.8), "bevel")]:
        ax.plot([1 + offset, 3 + offset, 3 + offset], [6, 6, 4], color=color, linewidth=12, solid_joinstyle=join)
    # Bottom row compares cap styles on short horizontal strokes.
    for x0, x1, color, cap in [(1, 3, (0.8, 0.2, 0.2), "butt"), (4, 6, (0.2, 0.8, 0.2), "round"), (7, 9, (0.2, 0.2, 0.8), "projecting")]:
        ax.plot([x0, x1], [2, 2], color=color, linewidth=12, solid_capstyle=cap)
    save(fig, args.out)


if __name__ == "__main__":
    main()
