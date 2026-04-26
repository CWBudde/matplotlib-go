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


def draw(out, dark=False):
    if dark:
        plt.rcParams.update({"figure.facecolor": "#202733", "text.color": "#f7f3ea", "axes.facecolor": "#273142", "axes.edgecolor": "#d8c7a1", "grid.color": "#607087", "lines.color": "#ffb347"})
        title = "Temporary rc_context override"
    else:
        plt.rcParams.update({"figure.dpi": 144, "figure.facecolor": "#faf7f0", "axes.facecolor": "#fffdf8", "axes.edgecolor": "#3d342c", "axes.labelcolor": "#2c241d", "grid.color": "#d6cbbd", "grid.linewidth": 0.75})
        title = "Runtime rc defaults"
    fig, ax = plt.subplots(figsize=(6.4, 4.8), dpi=plt.rcParams["figure.dpi"])
    x = np.linspace(0, 2 * math.pi, 240)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.set(title=title, xlabel="x", ylabel="sin(x)")
    ax.grid(True); ax.legend()
    save(fig, out)


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for styling/rc/main.go")
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    out = Path(args.output_dir); out.mkdir(parents=True, exist_ok=True)
    draw(out / "rc_defaults_python.png", dark=False)
    draw(out / "rc_context_python.png", dark=True)


if __name__ == "__main__":
    main()
