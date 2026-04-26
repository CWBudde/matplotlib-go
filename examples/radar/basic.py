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
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for radar/basic.go")
    parser.add_argument("--out", default="radar_basic_python.png")
    args = parser.parse_args()

    labels = ["Speed", "Power", "Range", "Handling", "Comfort"]
    values = np.array([0.72, 0.88, 0.64, 0.79, 0.58])
    angles = np.linspace(0, 2 * math.pi, len(labels), endpoint=False)
    closed_angles = np.r_[angles, angles[0]]
    closed_values = np.r_[values, values[0]]

    fig = plt.figure(figsize=(7.2, 7.2), dpi=100, facecolor="white")
    ax = fig.add_axes([0.12, 0.10, 0.76, 0.78], projection="polar")
    ax.set_title("Radar Projection")
    ax.set_thetagrids(np.degrees(angles), labels)
    ax.set_ylim(0, 1)
    ax.set_yticks([0.25, 0.5, 0.75, 1.0])
    ax.set_yticklabels(["25%", "50%", "75%", "100%"])

    # Split theta and radius grids so their styling matches the Go projection
    # calls, while the radial ticks use percent labels.
    ax.xaxis.grid(True, color=(0.78, 0.80, 0.84), linewidth=0.8)
    ax.yaxis.grid(True, color=(0.80, 0.83, 0.88), linewidth=0.8)

    # Radar data is closed explicitly by repeating the first angle/value pair.
    ax.fill(closed_angles, closed_values, color=(0.18, 0.50, 0.82, 0.22))
    ax.plot(closed_angles, closed_values, color=(0.15, 0.35, 0.70), linewidth=2.2, label="model A")
    save(fig, args.out)


if __name__ == "__main__":
    main()
