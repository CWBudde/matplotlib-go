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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for axes/labels/main.go")
    parser.add_argument("--out", default="labels_python.png")
    args = parser.parse_args()
    # Single sine curve keeps attention on title, x-label, and rotated y-label.
    fig = plt.figure(figsize=(8, 5), dpi=100, facecolor="white")
    ax = fig.add_axes([0.10, 0.15, 0.85, 0.73])
    x = np.linspace(0, 2 * math.pi, 200)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.set(title="Text Labels: Title, X-Label, Rotated Y-Label", xlabel="Angle (radians)", ylabel="Amplitude", xlim=(0, 2 * math.pi), ylim=(-1.5, 1.5))
    ax.minorticks_on(); ax.grid(True)
    save(fig, args.out)


if __name__ == "__main__":
    main()
