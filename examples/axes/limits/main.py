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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/limits/main.go")
    parser.add_argument("--out", default="limits_python.png")
    args = parser.parse_args()
    fig = plt.figure(figsize=(8, 5), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.15, 0.83, 0.73])
    x = np.linspace(3, 7, 100)
    ax.plot(x, np.sin(x) * 2.5, label="2.5·sin(x)")
    ax.plot(x, np.cos(x) * 1.5 + 0.5, label="1.5·cos(x)+0.5")
    ax.margins(0.05)
    ax.set(title="Auto-Scaled Axes with 5% Margin", xlabel="x", ylabel="y")
    ax.grid(True)
    save(fig, args.out)


if __name__ == "__main__":
    main()
