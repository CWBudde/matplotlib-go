#!/usr/bin/env python3
"""Matplotlib reference plot module generated from test/matplotlib_ref/generate.py."""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def radar_basic(out_dir):
    labels = ["Speed", "Power", "Range", "Handling", "Comfort"]
    values = np.array([0.72, 0.88, 0.64, 0.79, 0.58])
    angles = np.linspace(0, 2 * math.pi, len(labels), endpoint=False)
    closed_angles = np.r_[angles, angles[0]]
    closed_values = np.r_[values, values[0]]

    fig = make_fig_px(720, 720)
    ax = fig.add_axes(go_rect(0.12, 0.10, 0.88, 0.88), projection="polar")
    ax.set_title("Radar Projection")
    ax.set_thetagrids(np.degrees(angles), labels)
    ax.set_ylim(0, 1)
    ax.set_yticks([0.25, 0.5, 0.75, 1.0])
    ax.set_yticklabels(["25%", "50%", "75%", "100%"])
    ax.xaxis.grid(True, color=(0.78, 0.80, 0.84, 1.0), linewidth=lw(0.8))
    ax.yaxis.grid(True, color=(0.80, 0.83, 0.88, 1.0), linewidth=lw(0.8))

    ax.fill(closed_angles, closed_values, color=(0.18, 0.50, 0.82, 0.22))
    ax.plot(closed_angles, closed_values, color=(0.15, 0.35, 0.70, 1.0), linewidth=lw(2.2), label="model A")

    save(fig, out_dir, "radar_basic")


PLOT = radar_basic


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
