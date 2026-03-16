#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.7"]
# ///
"""Generate matplotlib reference images for matplotlib-go test comparisons.

Each function renders the same plot as the corresponding golden test in
test/golden_test.go, using real matplotlib as the reference renderer.

Usage:
    uv run generate.py --output-dir /path/to/output
    python3 generate.py --output-dir /path/to/output
"""

import argparse
import math
import os
import sys

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt

DPI = 100
W_PX, H_PX = 640, 360

# Tab10 palette — must match color/palette.go
TAB10 = [
    (0.12, 0.47, 0.71),  # blue
    (1.00, 0.50, 0.05),  # orange
    (0.17, 0.63, 0.17),  # green
    (0.84, 0.15, 0.16),  # red
    (0.58, 0.40, 0.74),  # purple
    (0.55, 0.34, 0.29),  # brown
    (0.89, 0.47, 0.76),  # pink
    (0.50, 0.50, 0.50),  # gray
    (0.74, 0.74, 0.13),  # olive
    (0.09, 0.75, 0.81),  # cyan
]


# ─── Helpers ─────────────────────────────────────────────────────────────────


def make_fig():
    """640×360 figure at 100 DPI with white background."""
    fig = plt.figure(figsize=(W_PX / DPI, H_PX / DPI), dpi=DPI)
    fig.patch.set_facecolor("white")
    return fig


def go_rect(min_x, min_y, max_x, max_y):
    """Convert Go RectFraction to matplotlib add_axes rect.

    Go uses top-left origin (min_y < max_y, y increases downward).
    Matplotlib add_axes uses [left, bottom, width, height] with bottom-left origin.
    """
    return [min_x, 1.0 - max_y, max_x - min_x, max_y - min_y]


def lw(go_width_px):
    """Go line width (pixels) → matplotlib linewidth (points)."""
    return go_width_px * 72.0 / DPI


def ss(go_radius_px):
    """Go scatter radius (pixels) → matplotlib scatter s (points²).

    Go Scatter2D.Size is radius in pixels.
    Matplotlib scatter s ≈ π·r² where r is the radius in points.
    """
    r_pt = go_radius_px * 72.0 / DPI
    return math.pi * r_pt * r_pt


def save(fig, out_dir, name):
    path = os.path.join(out_dir, f"{name}.png")
    fig.savefig(path, dpi=DPI, facecolor="white", bbox_inches=None)
    plt.close(fig)
    print(f"  wrote {path}")


# ─── Plot generators (one per golden test) ───────────────────────────────────


def basic_line(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.15, 0.95, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 1)
    ax.plot(
        [0, 1, 3, 6, 10],
        [0, 0.2, 0.9, 0.4, 0.8],
        color="black",
        linewidth=lw(2),
    )
    save(fig, out_dir, "basic_line")


def joins_caps(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)
    ax.plot([1, 3, 3, 5], [5, 5, 3, 3], color=(0.8, 0.2, 0.2), linewidth=lw(8))
    ax.plot([7, 9], [5, 5], color=(0.2, 0.2, 0.8), linewidth=lw(8))
    save(fig, out_dir, "joins_caps")


def dashes(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)

    specs = [
        (4, [],             (0,   0,   0)),
        (3, [5, 2],         (0.8, 0,   0)),
        (2, [3, 1, 1, 1],   (0,   0.6, 0)),
        (1, [1, 1],         (0,   0,   0.8)),
    ]
    for y_val, pattern, color in specs:
        (line,) = ax.plot([1, 9], [y_val, y_val], color=color, linewidth=lw(3))
        if pattern:
            line.set_dashes([p * 72.0 / DPI for p in pattern])

    save(fig, out_dir, "dashes")


def scatter_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.scatter(
        [2, 4, 6, 8, 3, 7],
        [3, 6, 4, 7, 8, 2],
        s=ss(8),
        color=(0.8, 0.2, 0.2),
        linewidths=0,
    )
    save(fig, out_dir, "scatter_basic")


def scatter_marker_types(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 8)
    markers = ["o", "s", "^", "D", "+", "x"]
    colors  = [(1,0,0), (0,1,0), (0,0,1), (1,1,0), (1,0,1), (0,1,1)]
    for i, (marker, color) in enumerate(zip(markers, colors)):
        ax.scatter([i + 1], [4], s=ss(12), c=[color], marker=marker, linewidths=0)
    save(fig, out_dir, "scatter_marker_types")


def scatter_advanced(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    xs = [2, 4, 6, 8, 2, 4, 6, 8]
    ys = [2, 4, 6, 8, 8, 6, 4, 2]
    radii = [6, 10, 14, 18, 8, 12, 16, 20]
    fills = [
        (1, 0.5, 0.5), (0.5, 1, 0.5), (0.5, 0.5, 1), (1, 1, 0.5),
        (1, 0.5, 1),   (0.5, 1, 1),   (0.8, 0.8, 0.8), (0.3, 0.3, 0.3),
    ]
    edges = [
        (0.5, 0, 0),   (0, 0.5, 0),   (0, 0, 0.5),   (0.5, 0.5, 0),
        (0.5, 0, 0.5), (0, 0.5, 0.5), (0.4, 0.4, 0.4), (0, 0, 0),
    ]
    ax.scatter(
        xs, ys,
        s=[ss(r) for r in radii],
        c=fills,
        edgecolors=edges,
        linewidths=lw(2),
        alpha=0.8,
    )
    save(fig, out_dir, "scatter_advanced")


def bar_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 10)
    ax.bar([1, 2, 3, 4, 5], [3, 7, 2, 8, 5], width=0.6, color=(0.2, 0.6, 0.8))
    save(fig, out_dir, "bar_basic")


def bar_horizontal(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)
    ax.barh([1, 2, 3, 4, 5], [3, 7, 2, 8, 5], height=0.6, color=(0.8, 0.4, 0.2))
    save(fig, out_dir, "bar_horizontal")


def bar_grouped(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 10)
    ax.bar(
        [1.2, 2.2, 3.2, 4.2, 5.2], [3, 7, 2, 8, 5],
        width=0.35, color=(0.8, 0.2, 0.2), edgecolor=(0.5, 0, 0),
        linewidth=lw(1),
    )
    ax.bar(
        [1.8, 2.8, 3.8, 4.8, 5.8], [5, 4, 6, 3, 7],
        width=0.35, color=(0.2, 0.8, 0.2), edgecolor=(0, 0.5, 0),
        linewidth=lw(1),
    )
    save(fig, out_dir, "bar_grouped")


def fill_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 10)
    ax.set_ylim(-1, 3)
    x  = [1, 2, 3, 4, 5, 6, 7, 8, 9]
    y  = [0.5, 1.8, 2.3, 1.2, 2.8, 1.9, 2.1, 1.5, 0.8]
    ax.fill_between(
        x, 0, y,
        facecolor=(0.3, 0.7, 0.9, 0.7),
        edgecolor=(0.1, 0.3, 0.5, 1.0),
        linewidth=lw(2),
    )
    save(fig, out_dir, "fill_basic")


def fill_between(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 6.28)
    ax.set_ylim(-1.5, 1.5)
    n  = 50
    x  = [6.28 * i / (n - 1) for i in range(n)]
    y1 = [math.sin(v) for v in x]
    y2 = [0.8 * math.cos(v) for v in x]
    ax.fill_between(
        x, y1, y2,
        facecolor=(0.8, 0.3, 0.3, 0.6),
        edgecolor=(0.5, 0.1, 0.1, 1.0),
        linewidth=lw(1.5),
    )
    ax.plot(x, y1, color=(1, 0, 0), linewidth=lw(2))
    ax.plot(x, y2, color=(0, 0, 1), linewidth=lw(2))
    save(fig, out_dir, "fill_between")


def fill_stacked(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 8)
    x      = [1, 2, 3, 4, 5, 6, 7]
    layer1 = [1, 1.5, 2, 1.8, 2.2, 1.9, 1.6]
    layer2 = [layer1[i] + 1.5 + 0.3 * math.sin(i) for i in range(len(layer1))]
    layer3 = [layer2[i] + 1.2 + 0.4 * math.cos(i) for i in range(len(layer2))]
    ax.fill_between(x, 0,      layer1, facecolor=(0.8, 0.2, 0.2, 0.8), edgecolor=(0.5, 0,   0,   1), linewidth=lw(1))
    ax.fill_between(x, layer1, layer2, facecolor=(0.2, 0.8, 0.2, 0.8), edgecolor=(0,   0.5, 0,   1), linewidth=lw(1))
    ax.fill_between(x, layer2, layer3, facecolor=(0.2, 0.2, 0.8, 0.8), edgecolor=(0,   0,   0.5, 1), linewidth=lw(1))
    save(fig, out_dir, "fill_stacked")


def multi_series_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 6)
    ax.plot([1, 2, 3, 4, 5, 6], [1.5, 2.8, 2.2, 3.5, 3.8, 4.2], color=TAB10[0], linewidth=lw(2))
    ax.scatter([1.5, 2.5, 3.5, 4.5, 5.5], [2.2, 3.1, 2.9, 4.1, 4.5],
               s=ss(8), c=[TAB10[1]], linewidths=0)
    ax.bar([2, 3, 4, 5], [3.8, 2.5, 4.8, 3.2], width=0.4, color=TAB10[2])
    save(fig, out_dir, "multi_series_basic")


def multi_series_color_cycle(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 2 * math.pi)
    ax.set_ylim(-1.2, 1.2)
    n = 50
    x = [2 * math.pi * i / (n - 1) for i in range(n)]
    for i, freq in enumerate([1, 2, 3, 4]):
        y = [math.sin(freq * v) for v in x]
        ax.plot(x, y, color=TAB10[i], linewidth=lw(2))
    save(fig, out_dir, "multi_series_color_cycle")


# ─── Entry point ─────────────────────────────────────────────────────────────

ALL_PLOTS = [
    basic_line, joins_caps, dashes,
    scatter_basic, scatter_marker_types, scatter_advanced,
    bar_basic, bar_horizontal, bar_grouped,
    fill_basic, fill_between, fill_stacked,
    multi_series_basic, multi_series_color_cycle,
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True, help="Directory to write PNG files")
    parser.add_argument("--plots", nargs="*", help="Subset of plot names to generate (default: all)")
    args = parser.parse_args()

    os.makedirs(args.output_dir, exist_ok=True)

    by_name = {f.__name__: f for f in ALL_PLOTS}
    to_run = ALL_PLOTS
    if args.plots:
        unknown = set(args.plots) - by_name.keys()
        if unknown:
            print(f"Unknown plots: {', '.join(sorted(unknown))}", file=sys.stderr)
            print(f"Available: {', '.join(sorted(by_name))}", file=sys.stderr)
            sys.exit(1)
        to_run = [by_name[n] for n in args.plots]

    print(f"Generating {len(to_run)} matplotlib reference image(s) → {args.output_dir}/")
    for fn in to_run:
        fn(args.output_dir)
    print("Done.")


if __name__ == "__main__":
    main()
