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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for agg-demo/main.go")
    parser.add_argument("--out", default="agg_demo_python.png")
    args = parser.parse_args()
    fig = plt.figure(figsize=(8, 5), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.18, 0.83, 0.70])
    x = np.linspace(0, 10, 200)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.plot(x, np.cos(x), label="cos(x)")
    ax.set(title="Sine and Cosine Waves", xlabel="x (radians)", ylabel="y", xlim=(0, 10), ylim=(-1.2, 1.2))
    ax.grid(True)
    save(fig, args.out)


if __name__ == "__main__":
    main()
