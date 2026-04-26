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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/inset/main.go")
    parser.add_argument("--out", default="inset_python.png")
    args = parser.parse_args()
    fig = plt.figure(figsize=(7.2, 4.2), dpi=100, facecolor="white")
    ax = fig.add_axes([0.10, 0.14, 0.82, 0.72])
    x = np.linspace(0, 10, 400)
    y = np.sin(x) * (0.75 + 0.2 * np.cos(2 * x))
    ax.plot(x, y, color=(0.12, 0.36, 0.72), linewidth=2)
    ax.set(title="Inset Axes", xlabel="x", ylabel="sin(x)", xlim=(0, 10), ylim=(-1.2, 1.2))
    ax.grid(True)
    inset = ax.inset_axes([0.58, 0.55, 0.36, 0.38])
    inset.plot(x, y, color=(0.12, 0.36, 0.72), linewidth=2)
    inset.set(title="detail", xlim=(2, 4), ylim=(-0.2, 1.05))
    inset.grid(True)
    ax.indicate_inset_zoom(inset, edgecolor="0.4")
    save(fig, args.out)


if __name__ == "__main__":
    main()
