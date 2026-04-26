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

from matplotlib.widgets import Button, CheckButtons, RadioButtons, Slider, TextBox


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for widgets/basic.go")
    parser.add_argument("--out", default="widgets_basic_python.png")
    args = parser.parse_args()
    fig = plt.figure(figsize=(10.8, 7.2), dpi=100, facecolor="white")
    plot = fig.add_axes([0.08, 0.44, 0.86, 0.48])
    x = np.linspace(0, 2 * math.pi, 240)
    plot.plot(x, np.sin(x), color=(0.13, 0.36, 0.72), linewidth=2.2, label="signal")
    plot.plot(x, 0.6 * np.cos(1.5 * x), color=(0.84, 0.34, 0.18), linewidth=1.8, label="modulation")
    plot.set(title="widgets family", xlabel="phase", ylabel="response", xlim=(0, 2 * math.pi), ylim=(-1.3, 1.3))
    plot.grid(axis="y"); plot.legend()
    plot.text(0.02, 0.98, "static widget showcase
Matplotlib-style control strip", transform=plot.transAxes, ha="left", va="top", bbox={"boxstyle": "round", "facecolor": "white"})
    Button(fig.add_axes([0.08, 0.28, 0.14, 0.10]), "Apply")
    Slider(fig.add_axes([0.26, 0.28, 0.36, 0.10]), "gain", 0, 1, valinit=0.68)
    CheckButtons(fig.add_axes([0.66, 0.18, 0.14, 0.20]), ["signal", "mod", "grid"], [True, True, False])
    RadioButtons(fig.add_axes([0.82, 0.18, 0.12, 0.20]), ["blue", "amber", "mono"], active=1)
    TextBox(fig.add_axes([0.08, 0.14, 0.54, 0.10]), "label", initial="phase scan")
    fig.text(0.98, 0.02, "widgets: Button, Slider, CheckButtons, RadioButtons, TextBox", ha="right", va="bottom", bbox={"boxstyle": "round", "facecolor": "white"})
    save(fig, args.out)


if __name__ == "__main__":
    main()
