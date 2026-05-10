#!/usr/bin/env python3
from __future__ import annotations

import argparse

import matplotlib.pyplot as plt


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for lines/basic-backend/main.go")
    parser.add_argument("--output", default="out_python.png")
    parser.add_argument("--width", type=int, default=640)
    parser.add_argument("--height", type=int, default=360)
    args = parser.parse_args()
    fig = plt.figure(figsize=(args.width / 100, args.height / 100), dpi=100, facecolor="white")
    ax = fig.add_axes([0.1, 0.15, 0.85, 0.75])
    # Backend selection is Go-specific; the plot body mirrors the Go example.
    ax.set_title("Basic Line")
    ax.plot([0, 1, 3, 6, 10], [0, 0.2, 0.9, 0.4, 0.8], color="black", linewidth=2)
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 1)
    save(fig, args.output)


if __name__ == "__main__":
    main()
