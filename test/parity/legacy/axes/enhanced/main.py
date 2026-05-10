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


def enhanced(out):
    # Three functions share the same x samples so their shapes line up.
    fig = plt.figure(figsize=(10, 8), dpi=100, facecolor=(0.98, 0.98, 0.98))
    ax = fig.add_axes([0.12, 0.12, 0.83, 0.76])
    x = np.linspace(-5, 5, 200)
    ax.plot(x, 2 * np.sin(x), color=(0.8, 0.2, 0.2), linewidth=2.5)
    ax.plot(x, np.exp(-x * x / 10) * np.cos(3 * x), color=(0.2, 0.6, 0.2), linewidth=2, linestyle=(0, (8, 4)))
    ax.plot(x, 0.1 * x ** 3 - 0.5 * x + 1, color=(0.2, 0.2, 0.8), linewidth=2)
    cx = np.array([-3, -1, 0, 1, 3])
    ax.scatter(cx, 2 * np.sin(cx), s=[100, 144, 225, 144, 100], marker="D", c=[(1, .5, 0), (.5, 0, 1), (1, 0, .5), (0, 1, .5), (1, 1, 0)])
    ax.set_xlim(-5, 5); ax.set_ylim(-3, 4); ax.grid(color=(0.7, 0.7, 0.7), linewidth=0.5)
    ax.tick_params(length=6, width=1.5, colors="0.2")
    save(fig, out)


def logarithmic(out):
    # Same data as the Go example, shown on log/log axes.
    fig = plt.figure(figsize=(10, 8), dpi=100, facecolor="white")
    ax = fig.add_axes([0.15, 0.12, 0.80, 0.76])
    x = 0.1 * np.power(10000, np.linspace(0, 1, 50))
    ax.plot(x, 10 * np.power(x, 1.5), color=(0.8, 0.2, 0.2), linewidth=3)
    ax.plot(x, 5 * np.exp(0.01 * x), color=(0.2, 0.6, 0.8), linewidth=3, linestyle=(0, (10, 5)))
    ax.set_xscale("log"); ax.set_yscale("log")
    ax.set_xlim(0.1, 1000); ax.set_ylim(1, 10000); ax.grid(True, which="both")
    save(fig, out)


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/enhanced/main.go")
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    out = Path(args.output_dir); out.mkdir(parents=True, exist_ok=True)
    enhanced(out / "axes_enhanced_python.png")
    logarithmic(out / "axes_logarithmic_enhanced_python.png")


if __name__ == "__main__":
    main()
