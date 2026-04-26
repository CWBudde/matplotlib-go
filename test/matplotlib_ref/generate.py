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
import datetime as dt
import json
import math
import os
import sys

import numpy as np
import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import matplotlib.colors as mcolors
import matplotlib.patches as mpatches
import matplotlib.path as mpath
import matplotlib.tri as mtri
from mpl_toolkits.axes_grid1 import ImageGrid, make_axes_locatable

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

MASK64 = (1 << 64) - 1


def _pcg_step(hi: int, lo: int) -> tuple[int, int]:
    """Advance Go's math/rand/v2.PCG state and return the next tuple.

    Mirrors the exact state transition from /usr/local/go1.25.0/src/math/rand/v2/pcg.go.
    """
    mul_hi = 2549297995355413924
    mul_lo = 4865540595714422341
    inc_hi = 6364136223846793005
    inc_lo = 1442695040888963407

    # bits.Mul64(p.lo, mulLo)
    mul_lo_lo = (lo * mul_lo) & MASK64
    mul_lo_hi = (lo * mul_lo) >> 64

    hi_new = (mul_lo_hi + (hi * mul_lo) + (lo * mul_hi)) & MASK64

    lo_new = mul_lo_lo + inc_lo
    carry = 1 if lo_new > MASK64 else 0
    lo_new &= MASK64
    hi_new = (hi_new + inc_hi + carry) & MASK64
    return hi_new, lo_new


def _pcg_uint64(hi: int, lo: int) -> tuple[int, int, int]:
    """Generate one PCG output value and return (value, next_hi, next_lo)."""
    hi, lo = _pcg_step(hi, lo)
    cheap_mul = 0xDA942042E4DD58B5
    val = hi ^ (hi >> 32)
    val = (val * cheap_mul) & MASK64
    val ^= val >> 48
    val = (val * (lo | 1)) & MASK64
    return val, hi, lo


def _pcg_float64(hi: int, lo: int) -> tuple[float, int, int]:
    """Return a Go-compatible rand.Float64 and updated PCG state.

    Go's Float64 uses:
    float64(r.Uint64()<<11>>11) / (1 << 53)
    """
    value, hi, lo = _pcg_uint64(hi, lo)
    return float((value & ((1 << 53) - 1)) / float(1 << 53)), hi, lo


def normal_data(seed1: int, seed2: int, n: int, mean: float, std: float) -> np.ndarray:
    """Generate normally distributed samples using the Go Box-Muller path."""
    hi, lo = seed1 & MASK64, seed2 & MASK64
    out = np.empty(n, dtype=np.float64)
    for _ in range(n):
        u1, hi, lo = _pcg_float64(hi, lo)
        u2, hi, lo = _pcg_float64(hi, lo)
        out[_] = math.sqrt(-2 * math.log(u1)) * math.cos(2 * math.pi * u2) * std + mean
    return out


def pcg_float64_values(seed1: int, seed2: int, n: int) -> list[float]:
    """Generate n Go-compatible Float64 values from the PCG stream."""
    hi, lo = seed1 & MASK64, seed2 & MASK64
    out = []
    for _ in range(n):
        value, hi, lo = _pcg_float64(hi, lo)
        out.append(value)
    return out


def histogram_payload(data: list[float], bins, density: bool = False, weights: np.ndarray | None = None) -> dict[str, list[float]]:
    """Return histogram bin edges and bin heights for parity comparisons."""
    counts, edges = np.histogram(data, bins=bins, density=density, weights=weights)
    return {
        "edges": edges.tolist(),
        "heights": counts.tolist(),
    }


def _to_list(values: np.ndarray | list[float]) -> list[float]:
    if isinstance(values, np.ndarray):
        return values.tolist()
    return values


# ─── Helpers ─────────────────────────────────────────────────────────────────


def make_fig_px(width_px, height_px):
    """Figure with explicit pixel dimensions at the shared DPI."""
    fig = plt.figure(figsize=(width_px / DPI, height_px / DPI), dpi=DPI)
    fig.patch.set_facecolor("white")
    return fig


def make_fig():
    """640×360 figure at 100 DPI with white background."""
    return make_fig_px(W_PX, H_PX)


def go_rect(min_x, min_y, max_x, max_y):
    """Convert Go RectFraction to matplotlib add_axes rect.

    Go RectFraction and matplotlib add_axes both use bottom-left figure
    fractions. Matplotlib wants [left, bottom, width, height].
    """
    return [min_x, min_y, max_x - min_x, max_y - min_y]


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
    ax.set_title("Basic Line")
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
    ax.set_title("Line Joins and Caps")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)
    ax.plot([1, 3, 3, 5], [5, 5, 3, 3], color=(0.8, 0.2, 0.2), linewidth=lw(8))
    ax.plot([7, 9], [5, 5], color=(0.2, 0.2, 0.8), linewidth=lw(8))
    save(fig, out_dir, "joins_caps")


def dashes(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Dash Patterns")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)

    # Dash values match golden_test.go (pixels in Go renderer).
    # Empirically: set_dashes units map to ~2.8 pixels at DPI=100,
    # so the conversion is p * 36 / DPI (= p / 2.78 at DPI=100).
    specs = [
        (4, [],               (0,   0,   0)),
        (3, [10, 4],          (0.8, 0,   0)),
        (2, [6, 2, 2, 2],     (0,   0.6, 0)),
        (1, [2, 2],           (0,   0,   0.8)),
    ]
    for y_val, pattern, color in specs:
        (line,) = ax.plot([1, 9], [y_val, y_val], color=color, linewidth=lw(3))
        if pattern:
            line.set_dashes([p * 36.0 / DPI for p in pattern])

    save(fig, out_dir, "dashes")


def scatter_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Basic Scatter")
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
    ax.set_title("Scatter Marker Types")
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 8)
    markers = ["o", "s", "^", "D", "+", "x"]
    colors  = [(1,0,0), (0,1,0), (0,0,1), (1,1,0), (1,0,1), (0,1,1)]
    for i, (marker, color) in enumerate(zip(markers, colors)):
        # Plus (+) and cross (x) are line-only markers; use linewidths to make them visible
        line_w = lw(2) if marker in ("+", "x") else 0
        ax.scatter([i + 1], [4], s=ss(12), c=[color], marker=marker, linewidths=line_w)
    save(fig, out_dir, "scatter_marker_types")


def scatter_advanced(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Advanced Scatter")
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


def _bar_basic_scaffold(show_ticks: bool, show_tick_labels: bool, show_title: bool):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 10)
    ax.tick_params(
        axis="both",
        which="both",
        bottom=show_ticks,
        left=show_ticks,
        labelbottom=show_tick_labels,
        labelleft=show_tick_labels,
    )
    if show_title:
        ax.set_title("Basic Bars")
    return fig, ax


def bar_basic_frame(out_dir):
    fig, ax = _bar_basic_scaffold(show_ticks=False, show_tick_labels=False, show_title=False)
    save(fig, out_dir, "bar_basic_frame")


def bar_basic_ticks(out_dir):
    fig, ax = _bar_basic_scaffold(show_ticks=True, show_tick_labels=False, show_title=False)
    save(fig, out_dir, "bar_basic_ticks")


def bar_basic_tick_labels(out_dir):
    fig, ax = _bar_basic_scaffold(show_ticks=True, show_tick_labels=True, show_title=False)
    save(fig, out_dir, "bar_basic_tick_labels")


def bar_basic_title(out_dir):
    fig, ax = _bar_basic_scaffold(show_ticks=True, show_tick_labels=True, show_title=True)
    save(fig, out_dir, "bar_basic_title")


def bar_basic(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Basic Bars")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 10)
    ax.bar([1, 2, 3, 4, 5], [3, 7, 2, 8, 5], width=0.6, color=(0.2, 0.6, 0.8))
    save(fig, out_dir, "bar_basic")


def bar_horizontal(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Horizontal Bars")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 6)
    ax.barh([1, 2, 3, 4, 5], [3, 7, 2, 8, 5], height=0.6, color=(0.8, 0.4, 0.2))
    save(fig, out_dir, "bar_horizontal")


def bar_grouped(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Grouped Bars")
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
    ax.set_title("Fill to Baseline")
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
    ax.set_title("Fill Between Curves")
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
    ax.set_title("Stacked Fills")
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
    ax.set_title("Multi-Series Plot")
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 6)
    ax.plot([1, 2, 3, 4, 5, 6], [1.5, 2.8, 2.2, 3.5, 3.8, 4.2], color=TAB10[0], linewidth=lw(2))
    ax.scatter([1.5, 2.5, 3.5, 4.5, 5.5], [2.2, 3.1, 2.9, 4.1, 4.5],
               s=ss(8), c=[TAB10[1]], linewidths=0)
    ax.bar([2, 3, 4, 5], [3.8, 2.5, 4.8, 3.2], width=0.4, color=TAB10[2])
    save(fig, out_dir, "multi_series_basic")


def hist_basic(out_dir):
    """Count histogram matching renderHistBasic in golden_test.go."""
    data = normal_data(42, 0, 500, 5.0, 1.5)
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.95, 0.90))
    ax.set_title("Basic Histogram")
    n, bins, patches = ax.hist(
        data,
        bins="sturges",
        color=(0.26, 0.53, 0.80, 0.8),
        edgecolor=(0, 0, 0, 1),
        linewidth=lw(0.8),
        rwidth=1.0,
    )
    # match AutoScale(0.05) margin
    margin = 0.05 * (data.max() - data.min())
    ax.set_xlim(data.min() - margin, data.max() + margin)
    count_max = n.max()
    ax.set_ylim(0, count_max * 1.05)
    save(fig, out_dir, "hist_basic")


def hist_density(out_dir):
    """Density histogram matching renderHistDensity in golden_test.go."""
    data = normal_data(42, 0, 500, 5.0, 1.5)
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.95, 0.90))
    ax.set_title("Density Histogram")
    n, bins, patches = ax.hist(
        data,
        bins=20,
        density=True,
        color=(0.20, 0.65, 0.30, 0.8),
        edgecolor=(0, 0, 0, 1),
        linewidth=lw(0.8),
        rwidth=1.0,
    )
    margin = 0.05 * (data.max() - data.min())
    ax.set_xlim(data.min() - margin, data.max() + margin)
    density_max = n.max()
    ax.set_ylim(0, density_max * 1.05)
    save(fig, out_dir, "hist_density")


def hist_strategies(out_dir):
    """Two overlapping probability histograms matching renderHistStrategies."""
    data1 = normal_data(42, 0, 300, 4.0, 1.0)
    data2 = normal_data(7, 0, 300, 7.0, 1.2)
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.12, 0.12, 0.95, 0.90))
    ax.set_title("Histogram Strategies")
    n1, _, _ = ax.hist(
        data1, bins=15, density=False, weights=np.ones(len(data1)) / len(data1),
        color=(0.26, 0.53, 0.80, 0.6),
        edgecolor=(0, 0, 0, 1),
        linewidth=lw(0.5),
        rwidth=1.0,
    )
    n2, _, _ = ax.hist(
        data2, bins=15, density=False, weights=np.ones(len(data2)) / len(data2),
        color=(0.90, 0.50, 0.10, 0.6),
        edgecolor=(0, 0, 0, 1),
        linewidth=lw(0.5),
        rwidth=1.0,
    )
    all_data = np.concatenate([data1, data2])
    margin = 0.05 * (all_data.max() - all_data.min())
    ax.set_xlim(all_data.min() - margin, all_data.max() + margin)
    ax.set_ylim(0, max(n1.max(), n2.max()) * 1.05)
    save(fig, out_dir, "hist_strategies")


def errorbar_basic(out_dir):
    """Combined line+scatter with symmetric error bars."""
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Error Bars")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 6)

    x = [1, 2, 3, 4, 5, 6]
    y = [1.8, 2.5, 2.2, 3.1, 2.8, 3.7]
    xerr = [0.20, 0.25, 0.15, 0.22, 0.30, 0.18]
    yerr = [0.28, 0.20, 0.35, 0.24, 0.30, 0.22]

    ax.plot(x, y, color=TAB10[0], linewidth=lw(2))
    ax.scatter(
        x,
        y,
        s=ss(4.5),
        c=[TAB10[2]],
        linewidths=0,
    )
    cap = (6 * 72.0 / DPI) / 2
    ax.errorbar(
        x,
        y,
        xerr=xerr,
        yerr=yerr,
        fmt="none",
        ecolor=(0, 0, 0, 1),
        elinewidth=lw(1.2),
        capsize=cap,
    )
    save(fig, out_dir, "errorbar_basic")


def boxplot_basic(out_dir):
    """Multi-series box plot matching renderBoxPlotBasic."""
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 10)
    ax.set_title("Box Plots")
    ax.set_xlabel("Group")
    ax.set_ylabel("Value")

    datasets = [
        [0.9, 1.0, 1.1, 1.2, 1.3, 1.45, 1.5, 1.7, 1.8],
        [4.0, 4.2, 4.3, 4.5, 4.8, 5.0, 5.4, 5.8, 9.4],
        [2.0, 2.1, 2.1, 2.2, 2.3, 2.4, 2.4, 2.6, 3.8],
    ]
    positions = [1.0, 2.0, 3.0]
    colors = [
        (0.25, 0.55, 0.82, 0.75),
        (0.80, 0.45, 0.20, 0.75),
        (0.35, 0.70, 0.35, 0.75),
    ]

    bp = ax.boxplot(
        datasets,
        positions=positions,
        widths=0.55,
        patch_artist=True,
        showfliers=True,
        boxprops=dict(linewidth=lw(1.2), color=(0, 0, 0, 1)),
        whiskerprops=dict(linewidth=lw(1.2), color=(0, 0, 0, 1)),
        capprops=dict(linewidth=lw(1.2), color=(0, 0, 0, 1)),
        medianprops=dict(linewidth=lw(1.8), color=(0, 0, 0, 1)),
        flierprops=dict(
            marker="o",
            markerfacecolor=(0, 0, 0, 1),
            markeredgecolor=(0, 0, 0, 1),
            markersize=4,
        ),
        manage_ticks=False,
    )
    ax.set_axisbelow(True)
    ax.yaxis.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    for patch, color in zip(bp["boxes"], colors):
        patch.set_facecolor(color)
        patch.set_alpha(color[3])

    save(fig, out_dir, "boxplot_basic")


def text_labels_strict(out_dir):
    """Text-only fixture for strict font/baseline regression testing."""
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.set_title("Text Labels")
    ax.set_xlabel("Group")
    ax.set_ylabel("Value")
    save(fig, out_dir, "text_labels_strict")


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


def multi_series_color_cycle(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.1, 0.9, 0.9))
    ax.set_title("Color Cycle")
    ax.set_xlim(0, 2 * math.pi)
    ax.set_ylim(-1.2, 1.2)
    n = 50
    x = [2 * math.pi * i / (n - 1) for i in range(n)]
    for i, freq in enumerate([1, 2, 3, 4]):
        y = [math.sin(freq * v) for v in x]
        ax.plot(x, y, color=TAB10[i], linewidth=lw(2))
    save(fig, out_dir, "multi_series_color_cycle")


def image_heatmap(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.15, 0.95, 0.9))
    ax.set_title("Image Heatmap")
    ax.set_xlim(0, 3)
    ax.set_ylim(0, 3)
    data = np.array([
        [0, 1, 2],
        [3, 4, 5],
        [6, 7, 8],
    ], dtype=float)
    ax.imshow(
        data,
        cmap="viridis",
        interpolation="nearest",
        origin="lower",
        extent=(0, 3, 0, 3),
        aspect="auto",
        vmin=data.min(),
        vmax=data.max(),
    )
    save(fig, out_dir, "image_heatmap")


def axes_top_right_inverted(out_dir):
    fig = make_fig()
    ax = fig.add_axes(go_rect(0.1, 0.12, 0.9, 0.9))
    ax.set_title("Top/Right Axes + Inversion")
    ax.set_xlabel("Bottom X")
    ax.set_ylabel("Left Y")
    ax.set_xlim(10, 0)
    ax.set_ylim(10, 0)
    ax.tick_params(top=True, labeltop=True, right=True, labelright=True)
    ax.minorticks_off()

    ax.plot(
        [1, 3, 6, 8.5],
        [2, 4, 6.5, 8],
        color=(0.15, 0.35, 0.75),
        linewidth=lw(2.2),
    )
    ax.scatter(
        [2, 5, 8],
        [8, 5, 2],
        s=ss(9),
        marker="D",
        c=[(0.85, 0.35, 0.20, 0.9)],
        edgecolors=[(0.45, 0.15, 0.05, 1.0)],
        linewidths=lw(1.0),
    )
    save(fig, out_dir, "axes_top_right_inverted")


def axes_control_surface(out_dir):
    fig = make_fig_px(760, 360)

    left = fig.add_axes(go_rect(0.07, 0.14, 0.47, 0.90))
    left.set_title("Moved Axes + Aspect")
    left.set_xlabel("Top X")
    left.set_ylabel("Right Y")
    left.set_xlim(-1, 5)
    left.set_ylim(-1, 5)
    left.xaxis.set_label_position("top")
    left.xaxis.tick_top()
    left.yaxis.set_label_position("right")
    left.yaxis.tick_right()
    left.set_aspect("equal", adjustable="box")
    left.set_box_aspect(1)
    left.minorticks_on()
    left.locator_params(axis="both", nbins=6)
    tick_color = (0.18, 0.42, 0.55, 1.0)
    left.tick_params(
        axis="both",
        which="major",
        colors=tick_color,
        length=lw(7),
        width=lw(1.2),
    )
    left.tick_params(
        axis="both",
        which="minor",
        colors=tick_color,
        length=lw(4),
        width=lw(0.9),
    )
    left.plot(
        [-0.5, 0.8, 2.2, 4.2],
        [-0.2, 1.0, 2.1, 4.4],
        color=(0.10, 0.32, 0.76),
        linewidth=lw(2.0),
    )
    left.scatter(
        [0.0, 1.5, 3.5, 4.5],
        [0.0, 1.8, 3.2, 4.6],
        s=ss(8),
        c=[(0.92, 0.48, 0.20, 0.92)],
        edgecolors=[(0.52, 0.22, 0.08, 1.0)],
        linewidths=lw(1.0),
    )

    right = fig.add_axes(go_rect(0.58, 0.14, 0.95, 0.90))
    right.set_title("Twin + Secondary")
    right.set_xlim(0, 10)
    right.set_ylim(0, 20)
    right.plot(
        [0, 2, 4, 6, 8, 10],
        [2, 6, 9, 13, 16, 19],
        color=(0.12, 0.45, 0.72),
        linewidth=lw(2.0),
    )

    twin = right.twinx()
    twin.set_ylim(0, 100)
    twin.tick_params(axis="y", colors=(0.80, 0.22, 0.22, 1.0))
    twin.spines["right"].set_color((0.80, 0.22, 0.22, 1.0))
    twin.plot(
        [0, 2, 4, 6, 8, 10],
        [10, 22, 38, 58, 81, 96],
        color=(0.80, 0.22, 0.22),
        linewidth=lw(1.8),
    )

    sec = right.secondary_xaxis("top", functions=(lambda x: x * 10, lambda x: x / 10))
    sec.tick_params(axis="x", colors=(0.16, 0.42, 0.30, 1.0))
    sec.spines["top"].set_color((0.16, 0.42, 0.30, 1.0))

    save(fig, out_dir, "axes_control_surface")


def transform_coordinates(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.13, 0.16, 0.90, 0.84))
    ax.set_title("Transform Coordinates")
    ax.set_xlabel("X")
    ax.set_ylabel("Y")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 10)
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    ax.plot(
        [1.0, 2.5, 4.5, 7.0, 8.8],
        [1.5, 3.2, 5.6, 6.4, 8.2],
        color=(0.14, 0.37, 0.74),
        linewidth=lw(2.2),
    )
    ax.scatter(
        [2.5, 7.0, 8.8],
        [3.2, 6.4, 8.2],
        s=ss(8),
        marker="D",
        c=[(0.88, 0.42, 0.16, 0.92)],
        edgecolors=[(0.45, 0.18, 0.05, 1.0)],
        linewidths=lw(1.0),
    )

    text_kwargs = dict(fontsize=11, color=(0.10, 0.10, 0.10))
    ax.text(1.3, 1.1, "data", transform=ax.transData, ha="left", va="baseline", **text_kwargs)
    ax.text(0.03, 0.97, "axes", transform=ax.transAxes, ha="left", va="top", **text_kwargs)
    fig.text(0.07, 0.08, "figure", ha="left", va="bottom", **text_kwargs)
    ax.text(
        0.50,
        0.22,
        "blend",
        transform=matplotlib.transforms.blended_transform_factory(fig.transFigure, ax.transAxes),
        ha="center",
        va="bottom",
        **text_kwargs,
    )
    ax.annotate(
        "axes note",
        xy=(0.82, 0.78),
        xycoords="axes fraction",
        xytext=(-48, -26),
        textcoords="offset pixels",
        fontsize=10,
        color=(0.10, 0.10, 0.10),
        ha="right",
        va="top",
        arrowprops=dict(arrowstyle="->", color=(0.10, 0.10, 0.10), lw=lw(1.25)),
    )

    save(fig, out_dir, "transform_coordinates")


def _composition_configure_axes(ax, title, x, y, color):
    ax.set_title(title, fontsize=12)
    ax.set_xlabel("x", fontsize=10)
    ax.set_ylabel("y", fontsize=10)
    ax.tick_params(labelsize=10)
    ax.plot(x, y, color=color, linewidth=lw(2.0), label=title)
    ax.margins(0.10)


def gridspec_composition(out_dir):
    fig = make_fig_px(960, 640)
    outer = fig.add_gridspec(
        2,
        2,
        left=0.08,
        right=0.96,
        bottom=0.10,
        top=0.92,
        wspace=0.06,
        hspace=0.28,
        width_ratios=[2, 1],
    )

    main_ax = fig.add_subplot(outer[:, 0])
    _composition_configure_axes(main_ax, "Main Span", [0, 1, 2, 3, 4], [1.2, 2.8, 2.1, 3.6, 3.1], (0.15, 0.35, 0.72))
    main_ax.set_xticks([0, 1, 2, 3, 4])
    main_ax.set_yticks([1.0, 1.5, 2.0, 2.5, 3.0, 3.5])

    nested = outer[0, 1].subgridspec(2, 1, hspace=0.75)
    top_right = fig.add_subplot(nested[0, 0])
    _composition_configure_axes(top_right, "Nested Top", [0, 1, 2, 3], [3.4, 2.6, 2.9, 1.8], (0.72, 0.32, 0.18))
    top_right.set_xticks([0, 1, 2, 3])
    top_right.set_yticks([2, 3])

    bottom_right = fig.add_subplot(nested[1, 0], sharex=top_right)
    _composition_configure_axes(bottom_right, "Nested Bottom", [0, 1, 2, 3], [1.0, 1.6, 1.3, 2.2], (0.18, 0.55, 0.34))
    bottom_right.set_xticks([0, 1, 2, 3])
    bottom_right.set_yticks([1, 2])

    # Pin the subfigure inset to the same GridSpec-derived rectangle as the Go
    # port. Matplotlib's add_subfigure(outer[1, 1]) uses the raw lower-right
    # quadrant here rather than the explicitly padded GridSpec cell.
    inset = fig.add_axes(go_rect(0.703689, 0.135965, 0.945761, 0.423684))
    _composition_configure_axes(inset, "SubFigure", [0, 1, 2, 3], [2.0, 2.4, 1.9, 2.7], (0.55, 0.22, 0.50))
    inset.set_xticks([0, 1, 2, 3])
    inset.set_yticks([2.0, 2.2, 2.4, 2.6])

    save(fig, out_dir, "gridspec_composition")


def figure_labels_composition(out_dir):
    fig, axs = plt.subplots(2, 2, figsize=(1100 / DPI, 720 / DPI), dpi=DPI, constrained_layout=True)
    fig.suptitle("Shared-Axis Figure Labels")
    fig.supxlabel("time [s]")
    fig.supylabel("amplitude")

    handles = []
    labels = []
    for row in range(2):
        for col in range(2):
            ax = axs[row, col]
            x = np.linspace(0, 2 * math.pi, 180)
            y = np.sin(x + row * 0.5) * (1 + col * 0.2)
            (line,) = ax.plot(x, y, label=f"series {row * 2 + col + 1}")
            handles.append(line)
            labels.append(line.get_label())
            ax.set_title(f"Panel {row * 2 + col + 1}")
            ax.set_xlabel("local x")
            ax.set_ylabel("local y")
            ax.set_xlim(0, 2 * math.pi)
            ax.set_ylim(-1.6, 1.6)
            ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
            ax.set_axisbelow(True)

    axs[0, 0].text(
        0.02,
        0.92,
        "upper-left\nnote",
        transform=axs[0, 0].transAxes,
        va="top",
        bbox=dict(facecolor="white", edgecolor=(0.5, 0.5, 0.5)),
    )
    axs[1, 1].text(
        0.98,
        0.08,
        "lower-right",
        transform=axs[1, 1].transAxes,
        ha="right",
        va="bottom",
        bbox=dict(facecolor="white", edgecolor=(0.5, 0.5, 0.5)),
    )
    fig.text(
        0.985,
        0.94,
        "Figure note",
        ha="right",
        va="top",
        bbox=dict(facecolor="white", edgecolor=(0.5, 0.5, 0.5)),
    )
    fig.legend(handles, labels, loc="upper right", bbox_to_anchor=(0.99, 0.90))

    save(fig, out_dir, "figure_labels_composition")


def colorbar_composition(out_dir):
    fig, ax = plt.subplots(figsize=(1000 / DPI, 700 / DPI), dpi=DPI, constrained_layout=True)
    rows, cols = 80, 120
    data = np.zeros((rows, cols))
    for row in range(rows):
        for col in range(cols):
            x = (col / (cols - 1)) * 4 - 2
            y = (row / (rows - 1)) * 4 - 2
            r = math.hypot(x, y)
            data[row, col] = math.sin(3 * r) * math.exp(-0.6 * r)

    im = ax.imshow(data, cmap="inferno", origin="lower", extent=[0, cols, 0, rows], aspect="auto")
    ax.set_title("Heatmap with Colorbar")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(0, cols)
    ax.set_ylim(0, rows)
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    cbar = fig.colorbar(im, ax=ax)
    cbar.set_label("Intensity")

    save(fig, out_dir, "colorbar_composition")


def annotation_composition(out_dir):
    fig = make_fig_px(1040, 720)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.90, 0.88))
    ax.set_title("Text and Arrow Annotations")
    ax.set_xlabel("phase")
    ax.set_ylabel("response")
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    x = np.linspace(0, 6 * math.pi, 240)
    y = np.sin(x) * np.exp(-0.015 * x) + 0.2 * np.cos(0.5 * x)
    ax.plot(x, y, label="signal")
    ax.set_xlim(0, 6 * math.pi)
    ax.set_ylim(-1.2, 1.2)
    ax.legend(loc="upper right")

    peak_x = math.pi / 2
    peak_y = math.sin(peak_x) * math.exp(-0.015 * peak_x) + 0.2 * math.cos(0.5 * peak_x)
    ax.annotate(
        "Peak\n= 0.42",
        xy=(peak_x, peak_y),
        xytext=(48, -42),
        textcoords="offset pixels",
        fontsize=12,
        arrowprops=dict(arrowstyle="->", color="black", lw=lw(1.0)),
    )
    ax.text(0.20, 0.90, "m∫T  φ x =  λ/4", transform=ax.transAxes, fontsize=12)

    save(fig, out_dir, "annotation_composition")


def plot_variants(out_dir):
    fig = make_fig_px(840, 620)

    axes = {
        "step": fig.add_axes(go_rect(0.08, 0.585, 0.475, 0.93)),
        "fill": fig.add_axes(go_rect(0.575, 0.585, 0.97, 0.93)),
        "broken": fig.add_axes(go_rect(0.08, 0.10, 0.475, 0.445)),
        "stack": fig.add_axes(go_rect(0.575, 0.10, 0.97, 0.445)),
    }

    step_ax = axes["step"]
    step_ax.set_title("Step + Stairs")
    step_ax.set_xlim(0, 6)
    step_ax.set_ylim(0, 5.2)
    step_ax.grid(axis="y")
    step_ax.set_axisbelow(True)
    step_ax.step(
        [0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4],
        [1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6],
        where="post",
        color=(0.15, 0.39, 0.78),
        linewidth=lw(2.0),
    )
    step_ax.stairs(
        [0.9, 1.7, 1.4, 2.6, 1.8, 2.2],
        [0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5],
        baseline=0.35,
        fill=True,
        facecolor=(0.91, 0.49, 0.20, 0.72),
        edgecolor=(0.58, 0.26, 0.08, 1.0),
        linewidth=lw(1.5),
    )

    fill_ax = axes["fill"]
    fill_ax.set_title("FillBetweenX + Refs")
    fill_ax.set_xlim(0, 7)
    fill_ax.set_ylim(0, 6)
    fill_ax.grid(axis="x")
    fill_ax.set_axisbelow(True)
    fill_ax.fill_betweenx(
        [0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2],
        [1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6],
        [3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3],
        facecolor=(0.24, 0.68, 0.54, 0.72),
        edgecolor=(0.12, 0.38, 0.28, 1.0),
        linewidth=lw(1.2),
    )
    fill_ax.axvspan(2.2, 3.1, color=(0.92, 0.75, 0.18), alpha=0.20)
    fill_ax.axhline(4.0, color=(0.52, 0.18, 0.18), linewidth=lw(1.2), dashes=[4 * 36.0 / DPI, 3 * 36.0 / DPI])
    fill_ax.axvline(5.3, color=(0.18, 0.22, 0.55), linewidth=lw(1.2), dashes=[2 * 36.0 / DPI, 2 * 36.0 / DPI])
    fill_ax.axline((0.9, 0.3), (6.4, 5.6), color=(0.22, 0.22, 0.22), linewidth=lw(1.1))

    broken_ax = axes["broken"]
    broken_ax.set_title("broken_barh")
    broken_ax.set_xlim(0, 10)
    broken_ax.set_ylim(0, 4.4)
    broken_ax.grid(axis="x")
    broken_ax.set_axisbelow(True)
    broken_ax.broken_barh([(0.8, 1.6), (3.1, 2.2), (6.5, 1.3)], (0.7, 0.9), facecolors=(0.21, 0.51, 0.76))
    broken_ax.broken_barh([(1.6, 1.0), (4.0, 1.4), (7.1, 1.7)], (2.1, 0.9), facecolors=(0.86, 0.38, 0.16))
    for x, y, label in [(1.6, 1.15, "prep"), (4.2, 1.15, "run"), (7.15, 1.15, "cool"), (2.1, 2.55, "IO"), (4.7, 2.55, "fit"), (7.95, 2.55, "ship")]:
        broken_ax.text(x, y, label, ha="center", va="center", fontsize=10, color="white")

    stack_ax = axes["stack"]
    stack_ax.set_title("Stacked Bars + Labels")
    stack_ax.set_xlim(0.4, 4.6)
    stack_ax.set_ylim(0, 7.6)
    stack_ax.grid(axis="y")
    stack_ax.set_axisbelow(True)
    xs = [1, 2, 3, 4]
    series_a = [1.4, 2.2, 1.8, 2.5]
    series_b = [2.1, 1.6, 2.4, 1.7]
    bottom = stack_ax.bar(xs, series_a, color=(0.16, 0.59, 0.49), width=0.8)
    top = stack_ax.bar(xs, series_b, bottom=series_a, color=(0.88, 0.47, 0.16), width=0.8)
    stack_ax.bar_label(bottom, labels=["A1", "A2", "A3", "A4"], label_type="center", color="white", fontsize=10)
    stack_ax.bar_label(top, fmt="%.1f", color=(0.20, 0.20, 0.20), fontsize=10, padding=4)

    save(fig, out_dir, "plot_variants")


def stat_variants(out_dir):
    fig = make_fig_px(840, 620)

    axes = {
        "stackplot": fig.add_axes(go_rect(0.08, 0.585, 0.475, 0.93)),
        "ecdf": fig.add_axes(go_rect(0.575, 0.585, 0.97, 0.93)),
        "cumulative": fig.add_axes(go_rect(0.08, 0.10, 0.475, 0.445)),
        "multi": fig.add_axes(go_rect(0.575, 0.10, 0.97, 0.445)),
    }

    stack_ax = axes["stackplot"]
    stack_ax.set_title("StackPlot")
    stack_ax.set_xlim(0, 5)
    stack_ax.set_ylim(0, 7)
    stack_ax.grid(axis="y")
    stack_ax.set_axisbelow(True)
    stack_ax.stackplot(
        [0, 1, 2, 3, 4, 5],
        [1.0, 1.4, 1.3, 1.8, 1.6, 2.0],
        [0.8, 1.1, 1.4, 1.2, 1.6, 1.8],
        [0.5, 0.8, 1.0, 1.4, 1.1, 1.5],
        colors=[
            (0.20, 0.55, 0.75, 0.76),
            (0.90, 0.48, 0.18, 0.76),
            (0.35, 0.66, 0.42, 0.76),
        ],
    )

    ecdf_ax = axes["ecdf"]
    ecdf_ax.set_title("ECDF")
    ecdf_ax.set_xlim(0, 8)
    ecdf_ax.set_ylim(0, 1.05)
    ecdf_ax.grid(axis="y")
    ecdf_ax.set_axisbelow(True)
    ecdf_samples = np.array([1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0])
    ecdf_values = np.sort(ecdf_samples)
    unique_values = []
    probabilities = []
    for idx, value in enumerate(ecdf_values):
        if idx + 1 < len(ecdf_values) and ecdf_values[idx + 1] == value:
            continue
        unique_values.append(value)
        probabilities.append((idx + 1) / len(ecdf_values))
    ecdf_ax.step(
        [unique_values[0]] + unique_values,
        [0] + probabilities,
        where="post",
        color=(0.18, 0.36, 0.75, 1),
        linewidth=lw(2),
    )

    cumulative_ax = axes["cumulative"]
    cumulative_ax.set_title("Cumulative Step Hist")
    cumulative_ax.set_xlim(0, 6)
    cumulative_ax.set_ylim(0, 1.05)
    cumulative_ax.grid(axis="y")
    cumulative_ax.set_axisbelow(True)
    cumulative_data = [0.4, 0.7, 1.2, 1.4, 2.1, 2.6, 3.1, 3.2, 4.0, 4.8, 5.2]
    cumulative_ax.hist(
        cumulative_data,
        bins=[0, 1, 2, 3, 4, 5, 6],
        weights=np.ones(len(cumulative_data)) / len(cumulative_data),
        cumulative=True,
        histtype="stepfilled",
        facecolor=(0.42, 0.62, 0.90, 0.55),
        edgecolor=(0.12, 0.25, 0.55, 1),
        linewidth=lw(1.4),
    )

    multi_ax = axes["multi"]
    multi_ax.set_title("Stacked Multi-Hist")
    multi_ax.set_xlim(0, 6)
    multi_ax.set_ylim(0, 6)
    multi_ax.grid(axis="y")
    multi_ax.set_axisbelow(True)
    multi_ax.hist(
        [
            [0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2],
            [0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0],
            [1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4],
        ],
        bins=[0, 1, 2, 3, 4, 5, 6],
        stacked=True,
        color=[
            (0.22, 0.55, 0.70, 0.8),
            (0.86, 0.42, 0.19, 0.8),
            (0.36, 0.62, 0.36, 0.8),
        ],
        edgecolor=(0.10, 0.10, 0.10, 1),
        linewidth=lw(0.7),
    )

    save(fig, out_dir, "stat_variants")


def units_overview(out_dir):
    fig = make_fig_px(1200, 420)

    date_ax = fig.add_axes(go_rect(0.06, 0.18, 0.32, 0.86))
    date_ax.set_title("Dates")
    date_ax.set_ylabel("Requests")
    date_ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    date_ax.set_axisbelow(True)
    date_ax.plot(
        [
            dt.datetime(2024, 1, 1),
            dt.datetime(2024, 1, 3),
            dt.datetime(2024, 1, 7),
            dt.datetime(2024, 1, 10),
        ],
        [12, 18, 9, 21],
        color=(0.12, 0.47, 0.71),
        linewidth=lw(2.0),
    )
    date_ax.margins(x=0.05, y=0.05)

    category_ax = fig.add_axes(go_rect(0.38, 0.18, 0.64, 0.86))
    category_ax.set_title("Categories")
    category_ax.set_ylabel("Count")
    category_ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    category_ax.set_axisbelow(True)
    category_ax.bar(
        ["alpha", "beta", "gamma", "delta"],
        [4, 9, 6, 7],
        color=(1.0, 0.50, 0.05),
        edgecolor=(0.60, 0.30, 0.03),
        linewidth=lw(1.0),
        width=0.8,
    )
    category_ax.margins(x=0.10, y=0.10)

    unit_ax = fig.add_axes(go_rect(0.70, 0.18, 0.96, 0.86))
    unit_ax.set_title("Custom Units")
    unit_ax.set_xlabel("Distance")
    unit_ax.set_ylabel("Pace")
    unit_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    unit_ax.set_axisbelow(True)
    unit_ax.scatter(
        [5, 10, 21.1, 42.2],
        [6.4, 5.8, 5.2, 5.5],
        s=ss(8),
        c=[(0.17, 0.63, 0.17, 0.92)],
        edgecolors=[(0.09, 0.36, 0.09, 1.0)],
        linewidths=lw(1.0),
    )
    unit_ax.xaxis.set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x, _: f"{x:.0f} km"))
    unit_ax.margins(x=0.08, y=0.08)

    save(fig, out_dir, "units_overview")


def units_dates(out_dir):
    fig = make_fig_px(720, 380)
    ax = fig.add_axes(go_rect(0.10, 0.18, 0.94, 0.88))
    ax.set_title("Date Units")
    ax.set_xlabel("Date")
    ax.set_ylabel("Requests")
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    dates = [
        dt.datetime(2024, 2, 1),
        dt.datetime(2024, 2, 5),
        dt.datetime(2024, 2, 9),
        dt.datetime(2024, 2, 14),
        dt.datetime(2024, 2, 20),
    ]
    lower = [6, 7, 5, 8, 7]
    upper = [10, 15, 13, 18, 16]
    ax.fill_between(dates, lower, upper, color=(0.85, 0.91, 0.96), linewidth=0)
    ax.plot(
        dates,
        [8, 12, 9, 15, 13],
        color=(0.12, 0.47, 0.71),
        linewidth=lw(2.0),
    )
    ax.xaxis.set_major_locator(mdates.DayLocator(bymonthday=[5, 12, 19]))
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%d %b"))
    ax.margins(x=0.06, y=0.06)

    save(fig, out_dir, "units_dates")


def units_categories(out_dir):
    fig = make_fig_px(760, 360)
    left = fig.add_axes(go_rect(0.08, 0.20, 0.47, 0.86))
    left.set_title("Categorical X")
    left.set_ylabel("Count")
    left.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    left.set_axisbelow(True)
    left.bar(
        ["draft", "review", "ship", "watch"],
        [3, 8, 6, 4],
        color=(1.0, 0.50, 0.05),
        edgecolor=(0.60, 0.30, 0.03),
        linewidth=lw(1.0),
        width=0.8,
    )
    left.margins(x=0.10, y=0.10)

    right = fig.add_axes(go_rect(0.58, 0.20, 0.94, 0.86))
    right.set_title("Categorical Y")
    right.set_xlabel("Hours")
    right.grid(True, axis="x", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    right.set_axisbelow(True)
    right.barh(
        ["north", "south", "east"],
        [4, 7, 5],
        color=(0.17, 0.63, 0.17),
        edgecolor=(0.09, 0.36, 0.09),
        linewidth=lw(1.0),
        height=0.8,
    )
    right.margins(x=0.10, y=0.10)

    save(fig, out_dir, "units_categories")


def units_custom_converter(out_dir):
    fig = make_fig_px(680, 380)
    ax = fig.add_axes(go_rect(0.10, 0.18, 0.94, 0.88))
    ax.set_title("Custom Distance Units")
    ax.set_xlabel("Distance")
    ax.set_ylabel("Pace")
    ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)

    distances = [5, 10, 21.1, 30, 42.2]
    pace = [6.4, 5.9, 5.3, 5.1, 5.4]
    ax.plot(distances, pace, color=(0.55, 0.34, 0.29), linewidth=lw(1.4))
    ax.scatter(
        distances,
        pace,
        s=ss(8),
        c=[(0.17, 0.63, 0.17, 0.92)],
        edgecolors=[(0.09, 0.36, 0.09, 1.0)],
        linewidths=lw(1.0),
    )
    ax.xaxis.set_major_formatter(matplotlib.ticker.FuncFormatter(lambda x, _: f"{x:.0f} km"))
    ax.margins(x=0.08, y=0.08)

    save(fig, out_dir, "units_custom_converter")


def patch_showcase(out_dir):
    fig = make_fig_px(930, 340)

    left = fig.add_axes(go_rect(0.05, 0.16, 0.31, 0.88))
    left.set_title("Patch Primitives")
    left.set_xlim(0, 6)
    left.set_ylim(0, 4)
    left.add_patch(mpatches.Rectangle(
        (0.6, 0.7), 1.5, 1.0,
        facecolor=(0.95, 0.70, 0.23, 0.86),
        edgecolor=(0.48, 0.27, 0.08, 1.0),
        linewidth=lw(1.1),
        hatch="/",
    ))
    left.add_patch(mpatches.Circle(
        (3.0, 1.25), 0.56,
        facecolor=(0.22, 0.57, 0.82, 0.82),
        edgecolor=(0.11, 0.29, 0.44, 1.0),
        linewidth=lw(1.0),
    ))
    left.add_patch(mpatches.Ellipse(
        (4.8, 2.75), 1.55, 0.95, angle=28,
        facecolor=(0.23, 0.72, 0.51, 0.80),
        edgecolor=(0.10, 0.36, 0.24, 1.0),
        linewidth=lw(1.0),
    ))
    left.add_patch(mpatches.Polygon(
        [[2.15, 3.2], [2.85, 2.25], [1.35, 2.45]],
        closed=True,
        facecolor=(0.84, 0.34, 0.34, 0.82),
        edgecolor=(0.48, 0.14, 0.14, 1.0),
        linewidth=lw(1.0),
    ))

    middle = fig.add_axes(go_rect(0.37, 0.16, 0.63, 0.88))
    middle.set_title("Fancy Arrow + Path")
    middle.set_xlim(0, 6)
    middle.set_ylim(0, 4)
    middle.add_patch(mpatches.FancyArrow(
        0.9, 3.2, 2.2, -1.0,
        width=0.18,
        head_width=0.62,
        head_length=0.62,
        facecolor=(0.91, 0.42, 0.22, 0.88),
        edgecolor=(0.58, 0.22, 0.10, 1.0),
        linewidth=lw(1.0),
        length_includes_head=True,
    ))
    star_vertices = [
        (4.15, 0.95), (4.45, 1.70), (5.22, 1.75), (4.63, 2.22), (4.84, 2.96),
        (4.15, 2.54), (3.46, 2.96), (3.67, 2.22), (3.08, 1.75), (3.85, 1.70), (4.15, 0.95),
    ]
    star_codes = [mpath.Path.MOVETO] + [mpath.Path.LINETO] * 9 + [mpath.Path.CLOSEPOLY]
    middle.add_patch(mpatches.PathPatch(
        mpath.Path(star_vertices, star_codes),
        facecolor=(0.76, 0.76, 0.86, 0.72),
        edgecolor=(0.29, 0.29, 0.45, 1.0),
        linewidth=lw(1.0),
        hatch="x",
    ))

    right = fig.add_axes(go_rect(0.69, 0.16, 0.95, 0.88))
    right.set_title("Fancy Boxes")
    right.set_xlim(0, 6)
    right.set_ylim(0, 4)
    right.add_patch(mpatches.FancyBboxPatch(
        (0.9, 0.8), 2.1, 1.25,
        boxstyle="round,pad=0.14,rounding_size=0.24",
        facecolor=(0.29, 0.67, 0.78, 0.28),
        edgecolor=(0.10, 0.37, 0.45, 1.0),
        linewidth=lw(1.0),
        hatch="/",
    ))
    right.add_patch(mpatches.FancyBboxPatch(
        (3.35, 1.55), 1.65, 1.05,
        boxstyle="square,pad=0.10",
        facecolor=(0.96, 0.87, 0.60, 0.82),
        edgecolor=(0.50, 0.39, 0.12, 1.0),
        linewidth=lw(1.0),
    ))

    save(fig, out_dir, "patch_showcase")


def mesh_contour_tri(out_dir):
    fig = make_fig_px(980, 620)
    axes = {
        "mesh": fig.add_axes(go_rect(0.07, 0.57, 0.46, 0.93)),
        "contour": fig.add_axes(go_rect(0.57, 0.57, 0.96, 0.93)),
        "hist2d": fig.add_axes(go_rect(0.07, 0.10, 0.46, 0.46)),
        "tri": fig.add_axes(go_rect(0.57, 0.10, 0.96, 0.46)),
    }

    mesh_ax = axes["mesh"]
    mesh_ax.set_title("PColorMesh")
    mesh_ax.set_xlim(0, 4)
    mesh_ax.set_ylim(0, 3)
    data = np.array([
        [0.2, 0.6, 0.3, 0.9],
        [0.4, 0.8, 0.5, 0.7],
        [0.1, 0.3, 0.9, 0.6],
    ])
    xedges = np.array([0, 1, 2, 3, 4], dtype=float)
    yedges = np.array([0, 1, 2, 3], dtype=float)
    mesh_ax.pcolormesh(
        xedges, yedges, data,
        shading="flat",
        edgecolors=[(0.95, 0.95, 0.95, 1.0)],
        linewidth=lw(0.8),
    )

    contour_ax = axes["contour"]
    contour_ax.set_title("Contour + Contourf")
    contour_ax.set_xlim(0, 4)
    contour_ax.set_ylim(0, 4)
    contour_data = np.array([
        [0.0, 0.4, 0.8, 0.4, 0.0],
        [0.2, 0.8, 1.3, 0.8, 0.2],
        [0.3, 1.0, 1.7, 1.0, 0.3],
        [0.2, 0.8, 1.3, 0.8, 0.2],
        [0.0, 0.4, 0.8, 0.4, 0.0],
    ])
    xx = np.arange(5, dtype=float)
    yy = np.arange(5, dtype=float)
    contour_levels = [0.2, 0.6, 1.0, 1.4, 1.8]
    contour_ax.contourf(xx, yy, contour_data, levels=contour_levels)
    lines = contour_ax.contour(
        xx, yy, contour_data,
        levels=[0.4, 0.8, 1.2, 1.6],
        colors=[(0.18, 0.18, 0.18, 1.0)],
        linewidths=lw(1.0),
    )
    contour_ax.clabel(lines, fmt="%g", fontsize=10)

    hist_ax = axes["hist2d"]
    hist_ax.set_title("Hist2D")
    hist_ax.set_xlim(0, 4)
    hist_ax.set_ylim(0, 4)
    hx = [0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6]
    hy = [0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4]
    hist_ax.hist2d(
        hx, hy,
        bins=[np.array([0, 1, 2, 3, 4], dtype=float), np.array([0, 1, 2, 3, 4], dtype=float)],
        edgecolor=(0.95, 0.95, 0.95, 1.0),
        linewidth=lw(0.8),
    )

    tri_ax = axes["tri"]
    tri_ax.set_title("Triangulation")
    tri_ax.set_xlim(0, 4)
    tri_ax.set_ylim(0, 4)
    tx = np.array([0.4, 1.6, 3.0, 0.8, 2.1, 3.5], dtype=float)
    ty = np.array([0.5, 0.4, 0.7, 2.2, 2.8, 2.1], dtype=float)
    tri = mtri.Triangulation(tx, ty, triangles=[[0, 1, 3], [1, 4, 3], [1, 2, 4], [2, 5, 4]])
    values = np.array([0.2, 0.8, 1.0, 1.5, 1.1, 0.6], dtype=float)
    tri_ax.tripcolor(tri, values, shading="flat")
    tri_ax.triplot(tri, color=(0.15, 0.15, 0.15), linewidth=lw(1.0))
    tri_ax.tricontour(tri, values, levels=[0.7, 1.1], colors=[(0.98, 0.98, 0.98, 1.0)], linewidths=lw(1.0))

    save(fig, out_dir, "mesh_contour_tri")


def stem_plot(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.10, 0.16, 0.94, 0.86))
    ax.set_title("Stem")
    ax.set_xlabel("Sample")
    ax.set_ylabel("Amplitude")
    ax.set_xlim(0.5, 7.5)
    ax.set_ylim(-0.2, 4.2)
    ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    ax.set_axisbelow(True)
    markerline, stemlines, baseline = ax.stem(
        [1, 2, 3, 4, 5, 6, 7],
        [0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1],
        basefmt="-",
        bottom=0.3,
    )
    stem_color = (0.15, 0.42, 0.73)
    plt.setp(stemlines, color=stem_color, linewidth=lw(1.5))
    plt.setp(markerline, color=stem_color, markerfacecolor=stem_color, markeredgecolor=stem_color, markersize=7)
    plt.setp(baseline, color=(0.32, 0.32, 0.32), linewidth=lw(1.5))

    save(fig, out_dir, "stem_plot")


def specialty_artists(out_dir):
    fig = make_fig_px(980, 720)
    axes = {
        "event": fig.add_axes(go_rect(0.07, 0.57, 0.34, 0.94)),
        "hexbin": fig.add_axes(go_rect(0.39, 0.57, 0.66, 0.94)),
        "pie": fig.add_axes(go_rect(0.73, 0.57, 0.98, 0.94)),
        "violin": fig.add_axes(go_rect(0.07, 0.08, 0.34, 0.45)),
        "table": fig.add_axes(go_rect(0.39, 0.08, 0.66, 0.45)),
        "sankey": fig.add_axes(go_rect(0.73, 0.08, 0.98, 0.45)),
    }

    event_ax = axes["event"]
    event_ax.set_title("Eventplot")
    event_ax.set_xlim(0, 10)
    event_ax.set_ylim(0.4, 3.6)
    event_ax.grid(True, axis="x", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    event_ax.set_axisbelow(True)
    event_ax.eventplot(
        [
            [0.8, 1.4, 3.1, 4.6, 7.3],
            [1.2, 2.9, 4.0, 6.4, 8.6],
            [0.5, 2.2, 5.4, 6.8, 9.1],
        ],
        lineoffsets=[1, 2, 3],
        linelengths=[0.6, 0.7, 0.8],
        colors=[
            (0.18, 0.44, 0.74, 1.0),
            (0.84, 0.38, 0.16, 1.0),
            (0.20, 0.63, 0.42, 1.0),
        ],
        linewidths=lw(1.5),
    )

    hex_ax = axes["hexbin"]
    hex_ax.set_title("Hexbin")
    hex_ax.set_xlim(0, 1)
    hex_ax.set_ylim(0, 1)
    hx = [0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88]
    hy = [0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58]
    hc = [1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1]
    hex_ax.hexbin(hx, hy, C=hc, gridsize=7, reduce_C_function=np.mean, mincnt=1, extent=(0, 1, 0, 1))

    pie_ax = axes["pie"]
    pie_ax.set_title("Pie")
    pie_ax.pie(
        [28, 22, 18, 32],
        labels=["Core", "I/O", "Render", "Docs"],
        autopct="%.0f%%",
        startangle=90,
        labeldistance=1.08,
        explode=[0, 0.04, 0, 0.02],
        colors=TAB10[:4],
        wedgeprops={"linewidth": lw(1.0), "edgecolor": "white"},
    )

    violin_ax = axes["violin"]
    violin_ax.set_title("Violin")
    violin_ax.set_xlim(0.4, 3.6)
    violin_ax.set_ylim(0.8, 5.2)
    violin_ax.grid(True, axis="y", color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    violin_ax.set_axisbelow(True)
    parts = violin_ax.violinplot(
        [
            [1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2],
            [1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0],
            [2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6],
        ],
        showmeans=True,
        showmedians=False,
        showextrema=True,
    )
    for body in parts["bodies"]:
        body.set_facecolor(TAB10[0])
        body.set_edgecolor((0.20, 0.20, 0.20))
        body.set_alpha(0.45)

    table_ax = axes["table"]
    table_ax.set_title("Table")
    table_ax.axis("off")
    table = table_ax.table(
        cellText=[["Latency", "18ms", "14ms"], ["Throughput", "220/s", "265/s"]],
        rowLabels=["A", "B"],
        colLabels=["Metric", "Q1", "Q2"],
        bbox=[0.04, 0.18, 0.92, 0.64],
        cellLoc="center",
    )
    table.auto_set_font_size(False)
    table.set_fontsize(10)

    sankey_ax = axes["sankey"]
    sankey_ax.set_title("Sankey")
    sankey_ax.axis("off")
    sankey_ax.set_xlim(0, 1)
    sankey_ax.set_ylim(0, 1)
    trunk = mpatches.Rectangle(
        (0.18, 0.47),
        0.18,
        0.06,
        transform=sankey_ax.transAxes,
        facecolor=(0.12, 0.47, 0.71, 0.75),
        edgecolor=(0.10, 0.10, 0.10, 1.0),
        linewidth=lw(1.0),
    )
    sankey_ax.add_patch(trunk)
    flows = [
        ("Waste", -2, -1, (0.84, 0.15, 0.16, 0.75)),
        ("CPU", 3, 1, (0.17, 0.63, 0.17, 0.75)),
        ("Cache", 1.5, 1, (1.00, 0.50, 0.05, 0.75)),
    ]
    for idx, (label, flow, orient, color) in enumerate(flows):
        width = abs(flow) * 0.018
        y = 0.40 + idx * 0.095
        x0 = 0.36
        x1 = 0.66
        verts = [
            (x0, 0.50 - width / 2),
            (x0 + 0.10, y - width / 2),
            (x1, y - width / 2),
            (x1, y + width / 2),
            (x0 + 0.10, y + width / 2),
            (x0, 0.50 + width / 2),
            (x0, 0.50 - width / 2),
        ]
        path = mpath.Path(verts)
        sankey_ax.add_patch(mpatches.PathPatch(path, transform=sankey_ax.transAxes, facecolor=color, edgecolor=(0.10, 0.10, 0.10), linewidth=lw(1.0)))
        sankey_ax.text(0.70, y, label, transform=sankey_ax.transAxes, va="center", fontsize=10)

    save(fig, out_dir, "specialty_artists")


def vector_fields(out_dir):
    fig = make_fig_px(919, 620)
    axes = {
        "quiver": fig.add_axes(go_rect(0.07, 0.58, 0.47, 0.92)),
        "barbs": fig.add_axes(go_rect(0.57, 0.58, 0.97, 0.92)),
        "stream": fig.add_axes(go_rect(0.07, 0.10, 0.47, 0.44)),
        "xy": fig.add_axes(go_rect(0.57, 0.10, 0.97, 0.44)),
    }

    quiver_ax = axes["quiver"]
    quiver_ax.set_title("Quiver + Key")
    quiver_ax.set_xlim(0, 6)
    quiver_ax.set_ylim(0, 5)
    quiver_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    quiver_ax.set_axisbelow(True)
    qx, qy, qu, qv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.8 + col * 1.0
            y = 0.8 + row * 0.95
            qx.append(x)
            qy.append(y)
            qu.append(0.55 + 0.08 * math.sin(y * 0.9))
            qv.append(0.22 * math.cos(x * 0.8))
    q = quiver_ax.quiver(
        qx, qy, qu, qv,
        color=(0.14, 0.42, 0.73),
        scale=10.0,
        scale_units="width",
        units="dots",
        width=2.2,
    )
    quiver_ax.quiverkey(q, 0.78, 0.12, 0.5, "0.5", coordinates="axes", labelpos="E")

    barb_ax = axes["barbs"]
    barb_ax.set_title("Barbs")
    barb_ax.set_xlim(0, 6)
    barb_ax.set_ylim(0, 5)
    barb_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    barb_ax.set_axisbelow(True)
    bx, by, bu, bv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.9 + col * 0.95
            y = 0.8 + row * 0.95
            bx.append(x)
            by.append(y)
            bu.append(14 + 5 * math.sin(y * 0.8))
            bv.append(8 * math.cos(x * 0.7))
    barb_ax.barbs(
        bx, by, bu, bv,
        barbcolor=(0.47, 0.23, 0.12),
        flagcolor=(0.86, 0.52, 0.24),
        length=6.0,
        linewidth=lw(1.0),
    )

    stream_ax = axes["stream"]
    stream_ax.set_title("Streamplot")
    stream_ax.set_xlim(0, 6)
    stream_ax.set_ylim(0, 5)
    stream_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    stream_ax.set_axisbelow(True)
    sx = np.array([0, 1, 2, 3, 4, 5, 6], dtype=float)
    sy = np.array([0, 1, 2, 3, 4, 5], dtype=float)
    su = np.zeros((len(sy), len(sx)))
    sv = np.zeros((len(sy), len(sx)))
    for yi, y in enumerate(sy):
        for xi, x in enumerate(sx):
            su[yi, xi] = 1.0 + 0.12 * math.cos(y * 0.7)
            sv[yi, xi] = 0.35 * math.sin((x - 3) * 0.8) - 0.10 * (y - 2.5)
    stream_ax.streamplot(
        sx, sy, su, sv,
        start_points=np.array([[0.4, 0.8], [0.4, 2.2], [0.4, 3.6]], dtype=float),
        broken_streamlines=False,
        integration_direction="forward",
        color=(0.13, 0.53, 0.39),
        linewidth=lw(1.5),
        arrowsize=1.0,
    )

    xy_ax = axes["xy"]
    xy_ax.set_title("Quiver XY")
    xy_ax.set_xlim(0, 6)
    xy_ax.set_ylim(0, 5)
    xy_ax.grid(True, color=(0.8, 0.8, 0.8), linewidth=lw(0.5))
    xy_ax.set_axisbelow(True)
    xg = np.array([0.8, 1.8, 2.8, 3.8, 4.8], dtype=float)
    yg = np.array([0.8, 1.8, 2.8, 3.8], dtype=float)
    ugu = np.zeros((len(yg), len(xg)))
    ugv = np.zeros((len(yg), len(xg)))
    for yi, y in enumerate(yg):
        for xi, x in enumerate(xg):
            ugu[yi, xi] = -(y - 2.4) * 0.35
            ugv[yi, xi] = (x - 2.8) * 0.35
    xy_ax.quiver(
        xg, yg, ugu, ugv,
        color=(0.74, 0.23, 0.27),
        pivot="middle",
        angles="xy",
        scale=9.0,
        scale_units="width",
        units="dots",
        width=1.9,
    )

    save(fig, out_dir, "vector_fields")


def polar_axes(out_dir):
    fig = make_fig_px(720, 720)
    ax = fig.add_axes(go_rect(0.12, 0.10, 0.88, 0.88), projection="polar")
    ax.set_title("Polar Axes")
    ax.set_xlabel("theta")
    ax.set_ylabel("radius")
    ax.set_ylim(0, 1.1)

    thetas = np.linspace(0.0, 2.0 * math.pi, 720)
    radii = 0.55 + 0.35 * np.cos(5.0 * thetas)

    ax.xaxis.grid(True, color=(0.8, 0.82, 0.86, 1.0), linewidth=lw(0.9))
    ax.yaxis.grid(True, color=(0.82, 0.84, 0.88, 0.9), linewidth=lw(0.8))
    ax.plot(
        thetas,
        radii,
        color=(0.16, 0.33, 0.73, 1.0),
        linewidth=lw(2.2),
        label="r = 0.55 + 0.35 cos(5theta)",
    )
    ax.fill_between(
        thetas,
        radii,
        0,
        color=(0.36, 0.56, 0.92, 0.2),
    )
    save(fig, out_dir, "polar_axes")


def geo_mollweide_axes(out_dir):
    fig = make_fig_px(720, 420)
    ax = fig.add_axes(go_rect(0.10, 0.14, 0.92, 0.86), projection="mollweide")
    ax.set_title("Mollweide Projection")
    ax.set_xlabel("longitude")
    ax.set_ylabel("latitude")
    ax.set_xticks(np.arange(-150, 151, 30) * math.pi / 180.0)
    ax.set_yticks(np.arange(-75, 76, 15) * math.pi / 180.0)
    degree_formatter = matplotlib.ticker.FuncFormatter(lambda x, _: f"{round(x * 180.0 / math.pi):.0f}")
    ax.xaxis.set_major_formatter(degree_formatter)
    ax.yaxis.set_major_formatter(degree_formatter)

    ax.xaxis.grid(True, color=(0.78, 0.80, 0.84, 1.0), linewidth=lw(0.8))
    ax.yaxis.grid(True, color=(0.78, 0.80, 0.84, 1.0), linewidth=lw(0.8))

    lon = np.linspace(-math.pi, math.pi, 361)
    lat = 0.35 * np.sin(3.0 * lon)
    ax.plot(lon, lat, color=(0.14, 0.34, 0.70, 1.0), linewidth=lw(2.0))

    save(fig, out_dir, "geo_mollweide_axes")


def unstructured_showcase(out_dir):
    fig = make_fig_px(1320, 520)

    tri_x = np.array([0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55])
    tri_y = np.array([0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05])
    triangles = np.array([
        [0, 1, 4],
        [1, 5, 4],
        [1, 2, 5],
        [2, 6, 5],
        [2, 3, 6],
        [4, 5, 7],
        [5, 8, 7],
        [5, 6, 8],
        [6, 9, 8],
    ])
    tri = mtri.Triangulation(tri_x, tri_y, triangles)
    values = np.sin(tri_x * 1.4) + 0.7 * np.cos((tri_y + 0.15) * 2.1)

    ax_mesh = fig.add_axes(go_rect(0.05, 0.16, 0.31, 0.88))
    ax_mesh.set_title("Triangulation")
    ax_mesh.set_xlabel("x")
    ax_mesh.set_ylabel("y")
    ax_mesh.set_xlim(-0.1, 3.1)
    ax_mesh.set_ylim(-0.15, 2.65)
    ax_mesh.set_aspect("equal")
    ax_mesh.triplot(tri, color=(0.18, 0.24, 0.34, 1.0), linewidth=lw(1.35))
    ax_mesh.text(
        0.98,
        0.02,
        "explicit triangular mesh",
        transform=ax_mesh.transAxes,
        ha="right",
        va="bottom",
        fontsize=10,
        bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    ax_color = fig.add_axes(go_rect(0.37, 0.16, 0.63, 0.88))
    ax_color.set_title("Tripcolor + Tricontour")
    ax_color.set_xlabel("x")
    ax_color.set_ylabel("y")
    ax_color.set_xlim(-0.1, 3.1)
    ax_color.set_ylim(-0.15, 2.65)
    ax_color.set_aspect("equal")
    ax_color.tripcolor(
        tri,
        values,
        cmap="viridis",
        edgecolors="white",
        linewidth=lw(0.6),
        shading="flat",
    )
    contour = ax_color.tricontour(
        tri,
        values,
        levels=6,
        colors=[(0.08, 0.12, 0.18, 0.95)],
        linewidths=lw(1.15),
    )
    ax_color.clabel(contour, inline=True, fmt="%.3g", fontsize=10, colors=[(0.08, 0.12, 0.18, 0.95)])

    ax_fill = fig.add_axes(go_rect(0.69, 0.16, 0.95, 0.88))
    ax_fill.set_title("Filled Tricontour")
    ax_fill.set_xlabel("x")
    ax_fill.set_ylabel("y")
    ax_fill.set_xlim(-0.1, 3.1)
    ax_fill.set_ylim(-0.15, 2.65)
    ax_fill.set_aspect("equal")
    ax_fill.tricontourf(tri, values, levels=7, cmap="plasma")
    ax_fill.tricontour(
        tri,
        values,
        levels=7,
        colors=[(1.0, 1.0, 1.0, 0.88)],
        linewidths=lw(0.95),
    )

    fig.text(
        0.98,
        0.98,
        "unstructured gallery family\ntriangulation, tripcolor, tricontour",
        ha="right",
        va="top",
        fontsize=11,
        bbox=dict(boxstyle="round,pad=0.35", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    save(fig, out_dir, "unstructured_showcase")


def arrays_showcase(out_dir):
    fig = make_fig_px(1240, 620)

    annotated = np.array([
        [0.12, 0.28, 0.46, 0.64, 0.82],
        [0.18, 0.34, 0.58, 0.74, 0.88],
        [0.24, 0.42, 0.63, 0.79, 0.91],
        [0.16, 0.38, 0.61, 0.83, 0.97],
    ])

    ax_heat = fig.add_axes(go_rect(0.05, 0.14, 0.31, 0.88))
    ax_heat.set_title("Annotated Heatmap")
    ax_heat.set_xlabel("column")
    ax_heat.set_ylabel("row")
    img = ax_heat.imshow(annotated, cmap="viridis", aspect="equal", origin="upper")
    ax_heat.set_xticks(np.arange(annotated.shape[1]))
    ax_heat.set_yticks(np.arange(annotated.shape[0]))
    threshold = (annotated.min() + annotated.max()) / 2.0
    for row in range(annotated.shape[0]):
        for col in range(annotated.shape[1]):
            color = (1.0, 1.0, 1.0, 1.0) if annotated[row, col] >= threshold else (0.12, 0.12, 0.14, 1.0)
            ax_heat.text(col, row, f"{annotated[row, col]:.2f}", ha="center", va="center", fontsize=10, color=color)

    rows, cols = 8, 10
    xx = np.linspace(0.0, 1.0, cols)
    yy = np.linspace(0.0, 1.0, rows)
    mesh_data = np.zeros((rows, cols))
    for y in range(rows):
        for x in range(cols):
            mesh_data[y, x] = 0.55 + 0.25 * math.sin((xx[x] * 2.3 + 0.35) * math.pi) + 0.20 * math.cos((yy[y] * 2.8 - 0.35 * 0.4) * math.pi)

    ax_mesh = fig.add_axes(go_rect(0.37, 0.14, 0.63, 0.88))
    ax_mesh.set_title("PColorMesh + Contour")
    ax_mesh.set_xlabel("x bin")
    ax_mesh.set_ylabel("y bin")
    x_edges = np.arange(cols + 1)
    y_edges = np.arange(rows + 1)
    quad = ax_mesh.pcolormesh(
        x_edges,
        y_edges,
        mesh_data,
        cmap="plasma",
        edgecolors="white",
        linewidth=lw(0.65),
        shading="flat",
    )
    contour = ax_mesh.contour(
        np.arange(cols),
        np.arange(rows),
        mesh_data,
        levels=6,
        colors=[(0.14, 0.10, 0.16, 0.95)],
        linewidths=lw(1.1),
    )
    ax_mesh.clabel(contour, inline=True, fmt="%.3g", fontsize=10, colors=[(0.14, 0.10, 0.16, 0.95)])

    spy = np.zeros((18, 18))
    for y in range(18):
        for x in range(18):
            if x == y or x + y == 17 or (x + 2 * y) % 7 == 0 or (2 * x + y) % 11 == 0:
                spy[y, x] = 1

    ax_spy = fig.add_axes(go_rect(0.69, 0.14, 0.95, 0.88))
    ax_spy.set_title("Spy")
    ax_spy.set_xlabel("column")
    ax_spy.set_ylabel("row")
    yy_spy, xx_spy = np.where(spy > 0.1)
    ax_spy.scatter(xx_spy, yy_spy, s=ss(10), color=(0.16, 0.38, 0.72, 1.0), marker="s", linewidths=0)
    ax_spy.set_xlim(-0.5, 17.5)
    ax_spy.set_ylim(17.5, -0.5)
    ax_spy.set_aspect("equal")
    ax_spy.set_xticks(np.arange(18))
    ax_spy.set_yticks(np.arange(18))
    ax_spy.text(
        0.98,
        0.02,
        "sparse structure view",
        transform=ax_spy.transAxes,
        ha="right",
        va="bottom",
        fontsize=10,
        bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    fig.text(
        0.98,
        0.98,
        "arrays gallery family\nheatmap, quad mesh, sparse matrix",
        ha="right",
        va="top",
        fontsize=11,
        bbox=dict(boxstyle="round,pad=0.35", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    save(fig, out_dir, "arrays_showcase")


def axisartist_showcase(out_dir):
    fig = make_fig_px(980, 640)

    host = fig.add_axes(go_rect(0.08, 0.14, 0.56, 0.88))
    host.set_title("AxisArtist / Parasite")
    host.set_xlabel("phase")
    host.set_ylabel("signal")
    host.set_xlim(-3.5, 3.5)
    host.set_ylim(-1.3, 1.3)
    host.grid(axis="y", color=(0.78, 0.80, 0.84, 1.0), linewidth=lw(0.8))

    x = np.linspace(-3.5, 3.5, 240)
    sine = np.sin(x)
    cos_scaled = 55 + 35 * np.cos(x * 0.8)

    host.plot(x, sine, color=(0.14, 0.34, 0.72, 1.0), linewidth=lw(2.2), label="sin(x)")
    host.axhline(0.0, color=(0.26, 0.26, 0.30, 1.0), linewidth=lw(1.4), dashes=[5 * 36.0 / DPI, 3 * 36.0 / DPI])
    host.axvline(0.0, color=(0.26, 0.26, 0.30, 1.0), linewidth=lw(1.4), dashes=[5 * 36.0 / DPI, 3 * 36.0 / DPI])
    host.tick_params(direction="inout")

    right = host.twinx()
    right.set_ylim(0, 100)
    right.plot(x, cos_scaled, color=(0.74, 0.28, 0.18, 1.0), linewidth=lw(1.8), label="55 + 35 cos(0.8x)")
    right.spines["right"].set_color((0.74, 0.28, 0.18, 1.0))
    right.tick_params(axis="y", colors=(0.74, 0.28, 0.18, 1.0))
    right.spines["top"].set_visible(False)
    right.spines["left"].set_visible(False)
    right.spines["bottom"].set_visible(False)

    host.text(
        0.02,
        0.98,
        "floating axes at x=0 / y=0\nparasite right scale",
        transform=host.transAxes,
        ha="left",
        va="top",
        fontsize=10,
        bbox=dict(boxstyle="round,pad=0.3", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )
    host.legend(loc="upper center")

    save(fig, out_dir, "axisartist_showcase")


def axes_grid1_showcase(out_dir):
    fig = make_fig_px(1100, 720)

    grid = ImageGrid(
        fig,
        go_rect(0.06, 0.12, 0.60, 0.88),
        nrows_ncols=(2, 2),
        axes_pad=(0.18, 0.20),
        share_all=False,
    )

    for idx, ax in enumerate(grid):
        row = idx // 2
        col = idx % 2
        ax.set_title(f"Tile {row + 1},{col + 1}")
        rows, cols = 24, 24
        data = np.zeros((rows, cols))
        phase = float(row * 2 + col)
        for y in range(rows):
            yy = y / float(rows - 1)
            for x in range(cols):
                xx = x / float(cols - 1)
                data[y, x] = 0.5 + 0.25 * math.sin((xx + phase) * 2 * math.pi) + 0.25 * math.cos((yy + phase * 0.3) * 3 * math.pi)
        ax.imshow(data, origin="upper")
        ax.text(
            0.98,
            0.02,
            "image grid",
            transform=ax.transAxes,
            ha="right",
            va="bottom",
            fontsize=9,
            bbox=dict(boxstyle="round,pad=0.25", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
        )

    channel_cmaps = {
        "Red": mcolors.LinearSegmentedColormap.from_list("red channel", [(0.18, 0.02, 0.02), (1.00, 0.18, 0.12)]),
        "Green": mcolors.LinearSegmentedColormap.from_list("green channel", [(0.02, 0.14, 0.05), (0.20, 0.90, 0.28)]),
        "Blue": mcolors.LinearSegmentedColormap.from_list("blue channel", [(0.02, 0.05, 0.18), (0.18, 0.45, 1.00)]),
    }
    channels = [
        (fig.add_axes(go_rect(0.66, 0.34, 0.75, 0.56)), "Red", 0),
        (fig.add_axes(go_rect(0.775, 0.34, 0.865, 0.56)), "Green", 1),
        (fig.add_axes(go_rect(0.89, 0.34, 0.98, 0.56)), "Blue", 2),
    ]
    for ax, title, phase in channels:
        rows, cols = 28, 28
        data = np.zeros((rows, cols))
        for y in range(rows):
            yy = y / float(rows - 1)
            for x in range(cols):
                xx = x / float(cols - 1)
                if phase == 1:
                    data[y, x] = 0.5 + 0.32 * math.sin(yy * 4 * math.pi) + 0.18 * math.cos(xx * 2 * math.pi)
                elif phase == 2:
                    dx = xx - 0.5
                    dy = yy - 0.5
                    data[y, x] = 0.58 + 0.36 * math.sin((xx + yy) * 3 * math.pi) - 0.18 * math.hypot(dx, dy)
                else:
                    dx = xx - 0.35
                    dy = yy - 0.42
                    data[y, x] = 0.35 + 0.65 * math.exp(-7 * (dx * dx + dy * dy))
        ax.set_title(title)
        ax.set_xticks([0, 10, 20])
        ax.set_yticks([0, 10, 20])
        ax.imshow(data, origin="upper", cmap=channel_cmaps[title])

    fig.text(
        0.98,
        0.98,
        "axes_grid1-style layout\nImageGrid + RGB channel views",
        ha="right",
        va="top",
        fontsize=11,
        bbox=dict(boxstyle="round,pad=0.35", facecolor="white", edgecolor=(0.75, 0.75, 0.75, 1.0)),
    )

    save(fig, out_dir, "axes_grid1_showcase")


# ─── Entry point ─────────────────────────────────────────────────────────────

ALL_PLOTS = [
    basic_line, joins_caps, dashes,
    scatter_basic, scatter_marker_types, scatter_advanced,
    bar_basic_frame, bar_basic_ticks, bar_basic_tick_labels, bar_basic_title,
    bar_basic, bar_horizontal, bar_grouped,
    fill_basic, fill_between, fill_stacked,
    errorbar_basic,
    multi_series_basic, multi_series_color_cycle,
    hist_basic, hist_density, hist_strategies,
    boxplot_basic,
    text_labels_strict, title_strict,
    image_heatmap,
    axes_top_right_inverted,
    axes_control_surface,
    transform_coordinates,
    gridspec_composition,
    figure_labels_composition,
    colorbar_composition,
    annotation_composition,
    patch_showcase,
    mesh_contour_tri,
    plot_variants,
    stat_variants,
    stem_plot,
    specialty_artists,
    units_overview,
    units_dates,
    units_categories,
    units_custom_converter,
    vector_fields,
    polar_axes,
    geo_mollweide_axes,
    unstructured_showcase,
    arrays_showcase,
    axisartist_showcase,
    axes_grid1_showcase,
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default="", help="Directory to write PNG files")
    parser.add_argument(
        "--emit-rng-debug",
        action="store_true",
        help="Emit RNG parity payload as JSON and exit",
    )
    parser.add_argument("--plots", nargs="*", help="Subset of plot names to generate (default: all)")
    args = parser.parse_args()

    if args.emit_rng_debug:
        payload = {
            "normal_data": {
                "hist_basic": _to_list(normal_data(42, 0, 500, 5.0, 1.5)),
                "hist_density": _to_list(normal_data(42, 0, 500, 5.0, 1.5)),
                "hist_strategies_data1": _to_list(normal_data(42, 0, 300, 4.0, 1.0)),
                "hist_strategies_data2": _to_list(normal_data(7, 0, 300, 7.0, 1.2)),
            },
            "uniform_data": {
                "pcg_42_0_1000": pcg_float64_values(42, 0, 1000),
                "pcg_7_0_600": pcg_float64_values(7, 0, 600),
            },
            "histogram_data": {
                "hist_basic": histogram_payload(normal_data(42, 0, 500, 5.0, 1.5), bins="sturges"),
                "hist_density": histogram_payload(normal_data(42, 0, 500, 5.0, 1.5), bins=20, density=True),
                "hist_strategies_data1": histogram_payload(
                    normal_data(42, 0, 300, 4.0, 1.0),
                    bins=15,
                    weights=np.ones(300) / 300,
                ),
                "hist_strategies_data2": histogram_payload(
                    normal_data(7, 0, 300, 7.0, 1.2),
                    bins=15,
                    weights=np.ones(300) / 300,
                ),
            },
        }
        print(json.dumps(payload))
        return

    if not args.output_dir:
        parser.error("--output-dir is required unless --emit-rng-debug is set")

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
