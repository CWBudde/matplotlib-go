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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for mathtext/basic.go")
    parser.add_argument("--out", default="mathtext_basic_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(900 / 96, 540 / 96), dpi=96, facecolor="white")
    ax = fig.add_axes([0.10, 0.12, 0.82, 0.76])
    x = np.linspace(0, 4 * math.pi, 200)
    y = np.sin(x) * np.exp(-0.08 * x)
    ax.plot(x, y)
    ax.set_title(r"MathText $\alpha^2 + \beta_i$")
    ax.set_xlabel(r"phase $\theta$")
    ax.set_ylabel(r"amplitude $\frac{1}{\sqrt{2}}$")
    ax.text(0.98, 0.92, r"$x_{\mathrm{max}}$", transform=ax.transAxes, ha="right", va="top", fontsize=12)
    ax.annotate(r"$\Delta y \approx \frac{1}{2}$", xy=(3.2, 0.35), xytext=(34, -26), textcoords="offset points", fontsize=12, arrowprops={"arrowstyle": "->"})
    ax.text(0.02, 0.98, r"$\omega_n = 2\pi f_n$", transform=ax.transAxes, ha="left", va="top", fontsize=11, bbox={"boxstyle": "round,pad=0.3", "facecolor": "white", "edgecolor": "0.75"})
    save(fig, args.out)


if __name__ == "__main__":
    main()
