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
import json
import math
import os
import sys

import numpy as np
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


# ─── Entry point ─────────────────────────────────────────────────────────────

ALL_PLOTS = [
    basic_line, joins_caps, dashes,
    scatter_basic, scatter_marker_types, scatter_advanced,
    bar_basic, bar_horizontal, bar_grouped,
    fill_basic, fill_between, fill_stacked,
    errorbar_basic,
    multi_series_basic, multi_series_color_cycle,
    hist_basic, hist_density, hist_strategies,
    boxplot_basic,
    text_labels_strict,
    image_heatmap,
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
