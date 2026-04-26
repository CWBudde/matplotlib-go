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


def first_plot(out):
    fig = plt.figure(figsize=(8, 6), dpi=100, facecolor="white")
    ax = fig.add_axes([0.15, 0.15, 0.80, 0.70])
    x = np.linspace(0, 10, 50)
    y = np.sin(x) + 0.5 * np.cos(2 * x)
    ax.plot(x, y, color=(0.2, 0.4, 0.8), linewidth=2.5)
    sx = np.array([1, 3, 5, 7, 9])
    ax.scatter(sx, np.sin(sx) + 0.5 * np.cos(2 * sx), color=(0.8, 0.2, 0.2), s=64)
    ax.set_xlim(0, 10); ax.set_ylim(-2, 3)
    save(fig, out)


def second_plot(out):
    fig = plt.figure(figsize=(8, 6), dpi=100, facecolor="white")
    ax = fig.add_axes([0.15, 0.15, 0.80, 0.70])
    t = np.linspace(0, 1, 20)
    x = 0.1 + (100 - 0.1) * t
    y = 1 + np.exp(5 * t)
    ax.plot(x, y, color=(0.8, 0.5, 0.2), linewidth=3)
    ax.set_xlim(0.1, 100); ax.set_ylim(1, 1000)
    ax.set_xscale("log"); ax.set_yscale("log")
    ax.tick_params(length=8, width=1.5, colors="0.3")
    for spine in ax.spines.values():
        spine.set_color("0.3"); spine.set_linewidth(1.5)
    save(fig, out)


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/basic/main.go")
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    out = Path(args.output_dir)
    out.mkdir(parents=True, exist_ok=True)
    first_plot(out / "axes_basic_python.png")
    second_plot(out / "axes_logarithmic_python.png")


if __name__ == "__main__":
    main()
