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

def title_strict(out_dir):
    """Minimal title-only fixture for strict title font regression testing."""
    fig = make_fig_px(320, 280)
    titles = [
        "Histogram Strategies",
        "Fill to Baseline",
        "Dash Patterns",
        "Box Plots",
        "Text Labels",
    ]
    rows = [
        (0.05, 0.20, 0.95, 0.28),
        (0.05, 0.36, 0.95, 0.44),
        (0.05, 0.52, 0.95, 0.60),
        (0.05, 0.68, 0.95, 0.76),
        (0.05, 0.84, 0.95, 0.92),
    ]
    for title, rect in zip(titles, rows):
        ax = fig.add_axes(go_rect(*rect))
        ax.set_xlim(0, 1)
        ax.set_ylim(0, 1)
        ax.set_title(title)
        ax.set_xticks([])
        ax.set_yticks([])
        ax.patch.set_visible(False)
        for spine in ax.spines.values():
            spine.set_visible(False)
    save(fig, out_dir, "title_strict")

PLOT = title_strict


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default=".")
    args = parser.parse_args()
    PLOT(args.output_dir)


if __name__ == "__main__":
    main()
