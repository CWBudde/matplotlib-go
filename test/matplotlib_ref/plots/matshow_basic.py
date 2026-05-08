#!/usr/bin/env python3
"""Matplotlib reference for matshow parity."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def matshow_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.22, 0.12, 0.78, 0.9))
    ax.set_title("Matshow")
    ax.matshow(
        np.array([
            [0.10, 0.20, 0.35, 0.55],
            [0.18, 0.32, 0.48, 0.70],
            [0.28, 0.46, 0.66, 0.86],
            [0.40, 0.58, 0.78, 0.96],
        ]),
        cmap="cividis",
    )
    save(fig, out_dir, "matshow_basic")


PLOT = matshow_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
