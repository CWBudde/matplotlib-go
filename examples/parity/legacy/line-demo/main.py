#!/usr/bin/env python3
from __future__ import annotations

import argparse

import matplotlib.pyplot as plt


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for line-demo/main.go")
    parser.add_argument("--out", default="output_python.png")
    args = parser.parse_args()

    fig = plt.figure(figsize=(6.4, 3.6), dpi=100, facecolor="white")
    ax = fig.add_axes([0.1, 0.15, 0.85, 0.75])
    # Same sample path as the Go example.
    ax.plot([0, 1, 3, 6, 10], [0, 0.2, 0.9, 0.4, 0.8], color="black", linewidth=2)
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 1)
    save(fig, args.out)


if __name__ == "__main__":
    main()
