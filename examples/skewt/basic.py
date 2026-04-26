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

from matplotlib.transforms import Affine2D


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for skewt/basic.go")
    parser.add_argument("--out", default="skewt_basic_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(7.2, 6.4), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.14, 0.74, 0.74])
    skew = Affine2D().skew_deg(28, 0)
    ax.set(title="Skew-T Style Projection", xlabel="Temperature (deg C)", ylabel="Pressure (hPa)")
    ax.set_xlim(-70, 35)
    ax.set_ylim(1050, 180)
    ax.grid(color=(0.82, 0.84, 0.88), linewidth=0.8)
    pressure = np.array([1000, 925, 850, 700, 600, 500, 400, 300, 250, 200])
    temperature = np.array([24, 20, 15, 5, -4, -14, -28, -43, -51, -58])
    dewpoint = np.array([18, 14, 8, -4, -14, -25, -38, -50, -57, -64])
    transform = skew + ax.transData
    ax.plot(temperature, pressure, color=(0.78, 0.13, 0.16), linewidth=2.4, label="temperature", transform=transform)
    ax.plot(dewpoint, pressure, color=(0.05, 0.48, 0.28), linewidth=2.4, label="dewpoint", transform=transform)
    ax.legend()
    save(fig, args.out)


if __name__ == "__main__":
    main()
