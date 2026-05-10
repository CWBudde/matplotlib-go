#!/usr/bin/env python3
from __future__ import annotations

import argparse

import matplotlib.pyplot as plt
import numpy as np


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib SVG export analogue for export/svg.go")
    parser.add_argument("--output", default="export_python.svg")
    parser.add_argument("--width", type=int, default=800)
    parser.add_argument("--height", type=int, default=500)
    args = parser.parse_args()

    fig = plt.figure(figsize=(args.width / 100, args.height / 100), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.18, 0.83, 0.70])

    x = np.linspace(0, 10, 80)
    y1 = 0.8 * (x - 5)
    y2 = 0.5 * (1 - x / 5)
    ax.plot(x, y1, label="line 1")
    ax.plot(x, y2, label="line 2")

    # Include direct text, a legend, and an annotation so the SVG keeps native text nodes.
    ax.plot([0, 2, 8, 10], [-0.8, 0.2, 0.2, -0.8], color=(0.5, 0.1, 0.1), linewidth=2, label="diag")
    ax.text(4.2, 0.4, "native SVG text")
    ax.annotate("marker", xy=(2.5, 0.3), ha="center")
    ax.set(title="SVG Export Demo", xlabel="x", ylabel="y", xlim=(0, 10), ylim=(-1, 1))
    ax.legend()
    save(fig, args.output)


if __name__ == "__main__":
    main()
