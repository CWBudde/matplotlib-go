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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for subplots/basic.go")
    parser.add_argument("--out", default="subplots_basic_python.png")
    args = parser.parse_args()
    # Four panels share limits; each varies frequency by column and damping by row.
    fig, axs = plt.subplots(2, 2, figsize=(12, 8), dpi=100, sharex=True, sharey=True)
    fig.subplots_adjust(left=0.08, right=0.96, bottom=0.08, top=0.92, wspace=0.08, hspace=0.08)
    x = np.linspace(0, 10, 128)
    for row in range(2):
        for col in range(2):
            ax = axs[row, col]
            v = np.linspace(0, 1, 128)
            y = np.sin((col + 1) * x) * np.exp(-0.05 * v * (row + 1))
            ax.plot(x, y)
            ax.set_title(f"Panel {row + 1}-{col + 1}")
            ax.set_xlabel("x"); ax.set_ylabel("y"); ax.grid(True)
    axs[0, 0].set_xlim(0, 10); axs[0, 0].set_ylim(-1.2, 1.2)
    save(fig, args.out)


if __name__ == "__main__":
    main()
