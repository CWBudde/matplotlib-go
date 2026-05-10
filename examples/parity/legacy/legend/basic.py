#!/usr/bin/env python3
from __future__ import annotations

import argparse
import math

import matplotlib.pyplot as plt
import numpy as np


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for legend/basic.go")
    parser.add_argument("--out", default="legend_basic_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(1000 / 96, 700 / 96), dpi=96, facecolor="white")
    ax = fig.add_axes([0.11, 0.12, 0.83, 0.76])
    x = np.linspace(0, 2 * math.pi, 120)
    # Legend entries come from labels on each artist.
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.plot(x, 0.7 * np.cos(2 * x), linestyle=(0, (8, 5)), label="0.7 cos(2x)")
    ax.scatter([0.6, 1.9, 3.4, 5.1], [0.56, 0.95, -0.26, -0.93], label="samples")
    ax.set(title="Legend Entries", xlabel="x", ylabel="y", xlim=(0, 2 * math.pi), ylim=(-1.3, 1.3))
    ax.grid(True)
    # Match the Go example's LegendUpperLeft setting.
    ax.legend(loc="upper left")
    save(fig, args.out)


if __name__ == "__main__":
    main()
