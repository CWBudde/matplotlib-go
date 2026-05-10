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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/spines/main.go")
    parser.add_argument("--out", default="spines_python.png")
    args = parser.parse_args()
    # Sine and cosine share the same domain to make ticks/spines easy to compare.
    fig = plt.figure(figsize=(8, 5), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.15, 0.83, 0.73])
    x = np.linspace(0, 2 * math.pi, 200)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.plot(x, np.cos(x), label="cos(x)")
    ax.set(title="Axis Spines, Major & Minor Ticks", xlabel="x", ylabel="y", xlim=(0, 2 * math.pi), ylim=(-1.5, 1.5))
    ax.minorticks_on(); ax.grid(True)
    save(fig, args.out)


if __name__ == "__main__":
    main()
