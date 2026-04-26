#!/usr/bin/env python3
from __future__ import annotations

import argparse

import matplotlib.pyplot as plt
import numpy as np


def save(fig, path):
    fig.savefig(path, dpi=fig.dpi, facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"saved {path}")


def main():
    parser = argparse.ArgumentParser(description="Matplotlib backend demo analogue for backends/demo/main.go")
    parser.add_argument("--backend", default="Agg")
    parser.add_argument("--output", default="backend_demo_python.png")
    parser.add_argument("--width", type=int, default=800)
    parser.add_argument("--height", type=int, default=600)
    parser.add_argument("--usecase", default="basic", help="accepted for parity with the Go example")
    parser.add_argument("--list", action="store_true", help="show the active Matplotlib backend and exit")
    parser.add_argument("--capabilities", action="store_true")
    args = parser.parse_args()

    if args.list or args.capabilities:
        # Matplotlib chooses a single active backend at runtime instead of the Go registry model.
        print(f"Matplotlib backend: {plt.get_backend()}")
        return

    fig = plt.figure(figsize=(args.width / 100, args.height / 100), dpi=100, facecolor="white")
    ax = fig.add_subplot(111)
    x = np.linspace(0, 10, 200)
    ax.plot(x, np.sin(x), label="sin(x)")
    ax.plot(x, np.cos(x), label="cos(x)")
    ax.set_title(f"Matplotlib backend: {plt.get_backend()}")
    ax.grid(True); ax.legend()
    save(fig, args.output)


if __name__ == "__main__":
    main()
