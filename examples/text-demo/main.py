#!/usr/bin/env python3
from __future__ import annotations

import argparse
from pathlib import Path

import matplotlib.pyplot as plt


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for text-demo/main.go")
    parser.add_argument("--out", default="text-demo-python.png")
    args = parser.parse_args()
    fig = plt.figure(figsize=(400 / 72, 200 / 72), dpi=72, facecolor="white")
    ax = fig.add_axes([0, 0, 1, 1])
    ax.set_axis_off()
    ax.text(20 / 400, 1 - 30 / 200, "matplotlib-go Text Rendering Demo", transform=ax.transAxes, fontsize=13, color="black")
    ax.text(20 / 400, 1 - 60 / 200, "Rendered with DejaVu Sans via Matplotlib", transform=ax.transAxes, fontsize=13, color="black")
    ax.text(20 / 400, 1 - 90 / 200, "Supports basic text positioning", transform=ax.transAxes, fontsize=13, color="black")
    ax.text(20 / 400, 1 - 120 / 200, "Small text (size 10)", transform=ax.transAxes, fontsize=10, color="black")
    ax.text(20 / 400, 1 - 150 / 200, "Large text (size 16)", transform=ax.transAxes, fontsize=16, color="black")
    ax.text(250 / 400, 1 - 120 / 200, "Red text", transform=ax.transAxes, fontsize=13, color="red")
    ax.text(250 / 400, 1 - 150 / 200, "Blue text", transform=ax.transAxes, fontsize=13, color="blue")
    fig.savefig(args.out, dpi=72, facecolor="white")
    plt.close(fig)
    print(f"saved {args.out}")


if __name__ == "__main__":
    main()
