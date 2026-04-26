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
    parser = argparse.ArgumentParser(description="Matplotlib SVG export analogue for export/svg.go")
    parser.add_argument("--out", default="export_python.svg")
    args = parser.parse_args()
    fig, ax = plt.subplots(figsize=(6.4, 3.6), dpi=100)
    x = np.linspace(0, 10, 200)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.set(title="SVG export", xlabel="x", ylabel="sin(x)")
    ax.grid(True); ax.legend()
    save(fig, args.out)


if __name__ == "__main__":
    main()
