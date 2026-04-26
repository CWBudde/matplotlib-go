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


def render_theme(style_name, title, out):
    if style_name != "default":
        plt.style.use("ggplot" if style_name == "ggplot" else "default")

    # Keep the data and axes rectangle matched to themes.go; the style is the variable.
    fig = plt.figure(figsize=(9, 5.2), dpi=100, facecolor="white")
    ax = fig.add_axes([0.1, 0.14, 0.84, 0.74])
    x = np.linspace(0, 10, 220)
    sin_wave = np.sin(x)
    cos_wave = 0.7 * np.cos(0.7 * x)
    envelope = 0.18 + 0.12 * np.sin(0.4 * x)
    ax.fill_between(x, sin_wave + envelope, sin_wave - envelope, alpha=0.25, label="Band")
    ax.plot(x, sin_wave, label="sin(x)")
    ax.plot(x, cos_wave, label="0.7 cos(0.7x)")
    ax.scatter([1.5, 4.8, 8.2], [0.99, -0.73, 0.94], label="Samples")
    ax.annotate("peak", xy=(1.5, 0.99), xytext=(18, 22), textcoords="offset points", arrowprops={"arrowstyle": "->"})
    ax.set(title=title, xlabel="Time", ylabel="Signal", xlim=(0, 10), ylim=(-1.4, 1.4))
    ax.grid(True, which="both"); ax.legend(loc="upper left")
    save(fig, out)


def main():
    parser = argparse.ArgumentParser(description="Matplotlib counterpart for styling/themes.go")
    parser.add_argument("--output-dir", default="examples/styling")
    args = parser.parse_args()
    out = Path(args.output_dir); out.mkdir(parents=True, exist_ok=True)
    for name, title in [("default", "Default Theme"), ("ggplot", "GGPlot Theme"), ("publication", "Publication Theme")]:
        render_theme(name, title, out / f"{name}_theme_python.png")


if __name__ == "__main__":
    main()
