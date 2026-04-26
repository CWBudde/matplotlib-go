#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path
import argparse
import math

import matplotlib.pyplot as plt
import numpy as np


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/grid/main.go")
    parser.add_argument("--out", default="grid_python.png")
    args = parser.parse_args()
    # Major grid lines stay solid; minor grid lines are lighter and dashed.
    fig = plt.figure(figsize=(8, 5), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.15, 0.83, 0.73])
    x = np.linspace(0, 10, 200)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.plot(x, 0.7 * np.sin(2 * x + 0.5), label="0.7·sin(2x+0.5)")
    ax.set(title="Grid Lines: Major (solid) & Minor (dashed)", xlabel="x", ylabel="y", xlim=(0, 10), ylim=(-1.5, 1.5))
    ax.minorticks_on()
    ax.grid(which="major", color=(0.7, 0.7, 0.7, 0.8), linewidth=0.8)
    ax.grid(which="minor", color=(0.85, 0.85, 0.85, 0.6), linestyle=(0, (2, 3)))
    save(fig, args.out)


if __name__ == "__main__":
    main()
