#!/usr/bin/env python3
"""Matplotlib reference for the radialforce parity fixture.

Mirrors internal/webdemo/demo.go::buildRadialforceDemo. The data resembles a
typical radial-force test rig output (close + open branches of a hysteresis
loop) so the parity viewer exercises scatter, dash-dot grid lines, integer x
ticks, and a basic legend.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

import numpy as np

try:
    from test.matplotlib_ref.webdemo_common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from webdemo_common import *  # noqa: F401,F403


D_MIN = 8.0
D_MAX = 16.0
N = 81


def _force(diameter, scale, offset):
    delta = np.clip(16.0 - diameter, 0.0, None)
    return scale * np.power(delta, 1.4) + offset


def demo_radialforce(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Radial Force Hysteresis")
    ax.set_xlabel("Diameter [mm]")
    ax.set_ylabel("Radial Force [N]")

    close_d = linspace(D_MAX, D_MIN, N)
    close_f = _force(close_d, 1.20, 0.30)
    open_d = linspace(D_MIN, D_MAX, N)
    open_f = _force(open_d, 0.80, 0.10)

    ax.scatter(
        close_d, close_f,
        s=ss(2.5), color=color(0.16, 0.42, 0.82), alpha=0.85, label="close",
    )
    ax.scatter(
        open_d, open_f,
        s=ss(2.5), color=color(0.93, 0.39, 0.26), alpha=0.85, label="open",
    )

    ax.set_xlim(D_MIN - 1, D_MAX)
    ax.set_ylim(0, 18)
    ax.set_xticks(np.arange(D_MIN - 1, D_MAX + 1, 1))
    ax.grid(True, linestyle="-.")
    ax.legend(loc="upper right")
    save(fig, out_dir, "radialforce")


DEMO = demo_radialforce


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True)
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH)
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT)
    args = parser.parse_args()
    DEMO(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
