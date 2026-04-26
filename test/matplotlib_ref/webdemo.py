#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.7", "numpy"]
# ///
"""Generate Matplotlib reference images for the browser demo catalog.

The functions in this file intentionally mirror internal/webdemo/demo.go. They
are not imported by Go tests; they feed the parity viewer with PNG baselines for
the charts currently exposed by the WASM demo.
"""

import argparse
import datetime as dt
import math
import os
import sys

import matplotlib

matplotlib.use("Agg")
import matplotlib.dates as mdates
import matplotlib.pyplot as plt
import numpy as np
from matplotlib.patches import Circle, Ellipse, FancyArrow, Polygon, Rectangle
from matplotlib.sankey import Sankey
from mpl_toolkits.axes_grid1.inset_locator import inset_axes, mark_inset

DPI = 100
DEFAULT_WIDTH = 960
DEFAULT_HEIGHT = 540
MASK64 = (1 << 64) - 1


def color(r, g, b, a=1.0):
    return (r, g, b, a)


def lw(px):
    return px * 72.0 / DPI


def ss(radius_px):
    radius_pt = radius_px * 72.0 / DPI
    return math.pi * radius_pt * radius_pt


def make_fig(width_px, height_px):
    fig = plt.figure(figsize=(width_px / DPI, height_px / DPI), dpi=DPI)
    fig.patch.set_facecolor("white")
    return fig


def rect(min_x=0.10, min_y=0.12, max_x=0.96, max_y=0.90):
    return [min_x, min_y, max_x - min_x, max_y - min_y]


def grid_rects(nrows, ncols, left, right, bottom, top, wspace, hspace, width_ratios=None, height_ratios=None):
    """Return axes rectangles using matplotlib-go's figure-normalized GridSpec math."""
    width_ratios = list(width_ratios or [1.0] * ncols)
    height_ratios = list(height_ratios or [1.0] * nrows)
    inner_w = right - left
    inner_h = top - bottom
    available_w = inner_w - wspace * (ncols - 1)
    available_h = inner_h - hspace * (nrows - 1)
    widths = [available_w * r / sum(width_ratios) for r in width_ratios]
    heights = [available_h * r / sum(height_ratios) for r in height_ratios]

    out = []
    for row in range(nrows):
        row_rects = []
        y_top = top - sum(heights[:row]) - hspace * row
        y0 = y_top - heights[row]
        for col in range(ncols):
            x0 = left + sum(widths[:col]) + wspace * col
            row_rects.append([x0, y0, widths[col], heights[row]])
        out.append(row_rects)
    return out


def span_rect(rects, row, col, row_span, col_span):
    x0 = rects[row][col][0]
    y0 = rects[row + row_span - 1][col][1]
    x1 = rects[row][col + col_span - 1][0] + rects[row][col + col_span - 1][2]
    y1 = rects[row][col][1] + rects[row][col][3]
    return [x0, y0, x1 - x0, y1 - y0]


def linspace(start, stop, n):
    return np.linspace(start, stop, n)


def save(fig, out_dir, name):
    path = os.path.join(out_dir, f"{name}.png")
    fig.savefig(path, dpi=DPI, facecolor="white", bbox_inches=None)
    plt.close(fig)
    print(f"wrote {path}")


def _pcg_step(hi, lo):
    mul_hi = 2549297995355413924
    mul_lo = 4865540595714422341
    inc_hi = 6364136223846793005
    inc_lo = 1442695040888963407
    mul_lo_lo = (lo * mul_lo) & MASK64
    mul_lo_hi = (lo * mul_lo) >> 64
    hi_new = (mul_lo_hi + (hi * mul_lo) + (lo * mul_hi)) & MASK64
    lo_new = mul_lo_lo + inc_lo
    carry = 1 if lo_new > MASK64 else 0
    lo_new &= MASK64
    hi_new = (hi_new + inc_hi + carry) & MASK64
    return hi_new, lo_new


def _pcg_float64(hi, lo):
    hi, lo = _pcg_step(hi, lo)
    cheap_mul = 0xDA942042E4DD58B5
    val = hi ^ (hi >> 32)
    val = (val * cheap_mul) & MASK64
    val ^= val >> 48
    val = (val * (lo | 1)) & MASK64
    return (val & ((1 << 53) - 1)) / float(1 << 53), hi, lo


def normal_sample(state):
    hi, lo = state
    u1, hi, lo = _pcg_float64(hi, lo)
    u2, hi, lo = _pcg_float64(hi, lo)
    if u1 == 0:
        u1 = sys.float_info.min
    return math.sqrt(-2 * math.log(u1)) * math.cos(2 * math.pi * u2), (hi, lo)


def scatter_cluster(seed1, seed2, center_x, center_y, n):
    state = (seed1 & MASK64, seed2 & MASK64)
    x, y = [], []
    for _ in range(n):
        sx, state = normal_sample(state)
        sy, state = normal_sample(state)
        x.append(center_x + 0.65 * sx)
        y.append(center_y + 0.55 * sy)
    return np.array(x), np.array(y)


def deterministic_normal(n, mean, sigma):
    state = (42, 7)
    out = []
    for _ in range(n):
        sample, state = normal_sample(state)
        out.append(mean + sigma * sample)
    return np.array(out)


def vector_grid(x, y):
    u = np.zeros((len(y), len(x)))
    v = np.zeros((len(y), len(x)))
    for yi, yv in enumerate(y):
        for xi, xv in enumerate(x):
            u[yi, xi] = 1.0 + 0.12 * math.cos(yv * 0.7)
            v[yi, xi] = 0.35 * math.sin((xv - 3) * 0.8) - 0.10 * (yv - 2.5)
    return u, v


def demo_lines(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Signal Comparison")
    ax.set_xlabel("t")
    ax.set_ylabel("amplitude")
    ax.grid(True)
    x = linspace(0, 12, 160)
    ax.plot(x, np.sin(x), color=color(0.16, 0.42, 0.82), linewidth=lw(3.0), label="sin(t)")
    ax.plot(x, 0.7 * np.cos(0.7 * x + 0.3), color=color(0.91, 0.45, 0.16), linewidth=lw(2.2), dashes=[10, 6], label="0.7 cos(0.7t + 0.3)")
    ax.plot(x, np.sin(1.6 * x) * np.exp(-x / 11), color=color(0.13, 0.62, 0.38), linewidth=lw(2.2), label="damped")
    ax.set_xlim(0, 12)
    ax.set_ylim(-1.4, 1.4)
    ax.legend()
    save(fig, out_dir, "lines")


def demo_scatter(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Scatter Clusters")
    ax.set_xlabel("feature x")
    ax.set_ylabel("feature y")
    ax.grid(True)
    edge = color(1, 1, 1, 0.8)
    specs = [
        (scatter_cluster(1, 11, -1.2, 0.5, 64), "D", 10, color(0.13, 0.49, 0.92), "cluster a"),
        (scatter_cluster(2, 22, 1.0, 1.4, 64), "^", 12, color(0.93, 0.39, 0.26), "cluster b"),
        (scatter_cluster(3, 33, 2.4, -0.8, 64), "s", 11, color(0.24, 0.72, 0.42), "cluster c"),
    ]
    for (x, y), marker, size, c, label in specs:
        ax.scatter(x, y, marker=marker, s=ss(size), color=c, edgecolors=edge, linewidths=lw(1.25), alpha=0.8, label=label)
    ax.set_xlim(-3.2, 4.2)
    ax.set_ylim(-3.0, 3.4)
    ax.legend()
    save(fig, out_dir, "scatter")


def demo_bars(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Quarterly Revenue")
    ax.set_xlabel("quarter")
    ax.set_ylabel("EUR million")
    ax.grid(True, axis="y")
    x_a = np.array([-0.18, 0.82, 1.82, 2.82])
    x_b = np.array([0.18, 1.18, 2.18, 3.18])
    a = np.array([18, 24, 29, 34])
    b = np.array([14, 20, 27, 31])
    edge = color(0.18, 0.18, 0.22, 0.7)
    bars_a = ax.bar(x_a, a, width=0.34, color=color(0.16, 0.42, 0.82, 0.9), edgecolor=edge, linewidth=lw(1), label="Product A")
    bars_b = ax.bar(x_b, b, width=0.34, color=color(0.91, 0.45, 0.16, 0.9), edgecolor=edge, linewidth=lw(1), label="Product B")
    ax.bar_label(bars_a, labels=["18", "24", "29", "34"])
    ax.bar_label(bars_b, labels=["14", "20", "27", "31"])
    for i, label in enumerate(["Q1", "Q2", "Q3", "Q4"]):
        ax.text(i, -2.5, label, ha="center")
    ax.set_xlim(-0.75, 3.75)
    ax.set_ylim(-4, 38)
    ax.legend()
    save(fig, out_dir, "bars")


def demo_fills(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Filled Signals")
    ax.set_xlabel("t")
    ax.set_ylabel("value")
    ax.grid(True)
    x = linspace(0, 2 * math.pi, 180)
    upper = 0.85 * np.sin(x) + 0.22 * np.cos(2 * x - 0.4)
    lower = -0.45 * np.cos(x - 0.2) - 0.18 * np.sin(2.4 * x)
    ax.fill_between(x, upper, lower, color=color(0.22, 0.60, 0.88), edgecolor=color(0.09, 0.30, 0.48), linewidth=lw(1.1), alpha=0.30, label="band")
    ax.plot(x, upper, color=color(0.10, 0.24, 0.62), linewidth=lw(2.2), label="upper")
    ax.plot(x, lower, color=color(0.86, 0.34, 0.18), linewidth=lw(2.2), dashes=[9, 5], label="lower")
    ax.set_xlim(0, 2 * math.pi)
    ax.set_ylim(-1.25, 1.25)
    ax.legend()
    save(fig, out_dir, "fills")


def demo_histogram(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Latency Distribution")
    ax.set_xlabel("latency (ms)")
    ax.set_ylabel("density")
    ax.grid(True)
    ax.hist(deterministic_normal(400, 47.0, 8.5), bins=24, density=True, color=color(0.42, 0.23, 0.77, 0.7), edgecolor=color(0.17, 0.12, 0.33), linewidth=lw(0.8), label="requests")
    ax.margins(0.05)
    ax.legend()
    save(fig, out_dir, "histogram")


def demo_errorbars(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Measured Trend With Error Bars")
    ax.set_xlabel("sample")
    ax.set_ylabel("response")
    ax.grid(True)
    x = np.array([1, 2, 3, 4, 5, 6])
    y = np.array([1.8, 2.5, 2.2, 3.1, 2.8, 3.7])
    ax.plot(x, y, color=color(0.12, 0.47, 0.71), linewidth=lw(2.1), label="trend")
    ax.scatter(x, y, color=color(0.17, 0.63, 0.17), s=ss(5), label="samples")
    ax.errorbar(x, y, xerr=[0.20, 0.25, 0.15, 0.22, 0.30, 0.18], yerr=[0.28, 0.20, 0.35, 0.24, 0.30, 0.22], fmt="none", ecolor=color(0.10, 0.12, 0.16), linewidth=lw(1.2), capsize=7, label="1sigma")
    ax.set_xlim(0.4, 6.6)
    ax.set_ylim(1.0, 4.3)
    ax.legend()
    save(fig, out_dir, "errorbars")


def heatmap_data(rows=28, cols=36):
    data = np.zeros((rows, cols))
    for row in range(rows):
        y = -1 + 2 * row / float(rows - 1)
        for col in range(cols):
            x = -1 + 2 * col / float(cols - 1)
            r1 = math.hypot(x + 0.35, y - 0.15)
            r2 = math.hypot(x - 0.4, y + 0.2)
            data[row, col] = math.sin(8 * r1) / (1 + 3 * r1) + 0.8 * math.cos(7 * r2)
    return data


def demo_heatmap(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Heatmap Surface")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    data = heatmap_data()
    ax.imshow(data, origin="upper", cmap="inferno", extent=[0, data.shape[1], data.shape[0], 0], aspect="auto")
    ax.set_xlim(0, data.shape[1])
    ax.set_ylim(0, data.shape[0])
    save(fig, out_dir, "heatmap")


def demo_patches(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Patch Showcase")
    ax.set_xlabel("x")
    ax.set_ylabel("y")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 4)
    patches = [
        Rectangle((0.5, 0.6), 1.5, 1.0, facecolor=color(0.95, 0.70, 0.23, 0.86), edgecolor=color(0.48, 0.27, 0.08), linewidth=lw(1.1), label="rectangle"),
        Circle((2.8, 1.25), 0.56, facecolor=color(0.22, 0.57, 0.82, 0.82), edgecolor=color(0.11, 0.29, 0.44), linewidth=lw(1.0), label="circle"),
        Ellipse((4.9, 2.85), 1.4, 0.92, angle=24, facecolor=color(0.23, 0.72, 0.51, 0.80), edgecolor=color(0.10, 0.36, 0.24), linewidth=lw(1.0), label="ellipse"),
        Polygon([(1.6, 3.0), (2.2, 2.1), (0.9, 2.3)], facecolor=color(0.84, 0.34, 0.34, 0.82), edgecolor=color(0.48, 0.14, 0.14), linewidth=lw(1.0), label="polygon"),
    ]
    for patch in patches:
        ax.add_patch(patch)
    ax.add_patch(FancyArrow(3.4, 0.8, 1.4, 1.1, width=0.16, head_width=0.48, head_length=0.42, facecolor=color(0.91, 0.42, 0.22, 0.88), edgecolor=color(0.58, 0.22, 0.10), linewidth=lw(1.0), label="arrow"))
    ax.legend()
    save(fig, out_dir, "patches")


def demo_polar(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect(0.12, 0.10, 0.88, 0.88), projection="polar")
    ax.set_title("Polar Wave")
    ax.set_xlabel("theta")
    ax.set_ylabel("radius")
    theta = linspace(0, 2 * math.pi, 720)
    radius = 0.55 + 0.28 * np.cos(4 * theta) + 0.08 * np.sin(9 * theta)
    ax.set_ylim(0, 1.1)
    ax.grid(color=color(0.80, 0.82, 0.86), linewidth=lw(0.9))
    ax.fill(theta, radius, color=color(0.36, 0.56, 0.92), edgecolor=color(0.14, 0.25, 0.52), linewidth=lw(1.0), alpha=0.24, label="filled area")
    ax.plot(theta, radius, color=color(0.16, 0.33, 0.73), linewidth=lw(2.2), label="r(theta)")
    ax.legend()
    save(fig, out_dir, "polar")


def demo_phase7(out_dir, width, height):
    fig = make_fig(width, height)
    geo = fig.add_axes(rect(0.06, 0.16, 0.48, 0.84), projection="mollweide")
    geo.set_title("Mollweide Projection")
    geo.set_xlabel("longitude")
    geo.set_ylabel("latitude")
    geo.grid(color=color(0.78, 0.80, 0.84), linewidth=lw(0.8))
    lon = linspace(-math.pi, math.pi, 241)
    geo.plot(lon, 0.35 * np.sin(3 * lon), color=color(0.14, 0.34, 0.70), linewidth=lw(2.0))
    ax = fig.add_axes(rect(0.57, 0.16, 0.96, 0.84))
    ax.set_title("Zoomed Inset")
    ax.set_xlabel("x")
    ax.set_ylabel("response")
    ax.set_xlim(0, 10)
    ax.set_ylim(-1.2, 1.2)
    ax.grid(True)
    x = linspace(0, 10, 320)
    y = np.sin(x) * (0.75 + 0.20 * np.cos(2 * x))
    ax.plot(x, y, color=color(0.12, 0.36, 0.72), linewidth=lw(2.0))
    ins = inset_axes(ax, width="43%", height="40%", loc="upper right", borderpad=0.5)
    ins.set_title("detail")
    ins.plot(x, y, color=color(0.12, 0.36, 0.72), linewidth=lw(1.6))
    ins.set_xlim(2, 4)
    ins.set_ylim(-0.2, 1.05)
    ins.grid(True)
    mark_inset(ax, ins, loc1=2, loc2=4, fc="none", ec="0.5")
    save(fig, out_dir, "phase7")


def demo_subplots(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.92, 0.08, 0.12)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    palette = [color(0.16, 0.42, 0.82), color(0.91, 0.45, 0.16), color(0.24, 0.72, 0.42), color(0.80, 0.20, 0.42)]
    x = linspace(0, 10, 120)
    for idx, ax in enumerate([axs[0][0], axs[0][1], axs[1][0], axs[1][1]]):
        row, col = divmod(idx, 2)
        ax.set_title(f"Panel {idx + 1}")
        ax.set_xlabel("x")
        ax.set_ylabel("y")
        ax.grid(True)
        y = np.sin((col + 1) * x + idx * 0.4) * np.exp(-0.12 * (row + 1) * x / 10)
        ax.plot(x, y, color=palette[idx], linewidth=lw(2.4))
    axs[0][0].set_xlim(0, 10)
    axs[0][0].set_ylim(-1.25, 1.25)
    save(fig, out_dir, "subplots")


def demo_variants(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.11, 0.91, 0.10, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("Step + Stairs")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5.2)
    ax.grid(True, axis="y")
    ax.step([0.6, 1.4, 2.2, 3.0, 3.8, 4.6, 5.4], [1.1, 2.5, 1.7, 3.4, 2.9, 4.1, 3.6], where="post", color=color(0.15, 0.39, 0.78), linewidth=lw(2), label="step")
    ax.stairs([0.9, 1.7, 1.4, 2.6, 1.8, 2.2], [0.4, 1.1, 2.0, 2.9, 3.7, 4.6, 5.5], baseline=0.35, fill=True, color=color(0.91, 0.49, 0.20, 0.72), edgecolor=color(0.58, 0.26, 0.08), linewidth=lw(1.5), label="stairs")
    ax.legend()
    ax = axs[0][1]
    ax.set_title("FillBetweenX + Refs")
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 6)
    ax.grid(True, axis="x")
    ax.fill_betweenx([0.4, 1.2, 2.0, 2.8, 3.6, 4.4, 5.2], [1.3, 2.1, 1.7, 2.8, 2.2, 3.1, 2.6], [3.4, 4.1, 4.8, 5.1, 5.6, 6.0, 6.3], color=color(0.24, 0.68, 0.54, 0.72), edgecolor=color(0.12, 0.38, 0.28), linewidth=lw(1.2))
    ax.axvspan(2.2, 3.1, color=color(0.92, 0.75, 0.18), alpha=0.20)
    ax.axhline(4.0, color=color(0.52, 0.18, 0.18), linewidth=lw(1.2), dashes=[4, 3])
    ax.axline((0.9, 0.3), (6.4, 5.6), color=color(0.22, 0.22, 0.22), linewidth=lw(1.1))
    ax = axs[1][0]
    ax.set_title("Broken BarH")
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 4.4)
    ax.grid(True, axis="x")
    ax.broken_barh([(0.8, 1.6), (3.1, 2.2), (6.5, 1.3)], (0.7, 0.9), facecolors=color(0.21, 0.51, 0.76))
    ax.broken_barh([(1.6, 1.0), (4.0, 1.4), (7.1, 1.7)], (2.1, 0.9), facecolors=color(0.86, 0.38, 0.16))
    for x, y, txt in [(1.6, 1.15, "prep"), (4.2, 1.15, "run"), (7.15, 1.15, "cool"), (2.1, 2.55, "IO"), (4.7, 2.55, "fit"), (7.95, 2.55, "ship")]:
        ax.text(x, y, txt, ha="center", va="center", color="white", fontsize=10)
    ax = axs[1][1]
    ax.set_title("Stacked Bars")
    ax.set_xlim(0.4, 4.6)
    ax.set_ylim(0, 7.6)
    ax.grid(True, axis="y")
    x = np.array([1, 2, 3, 4])
    a = np.array([1.4, 2.2, 1.8, 2.5])
    b = np.array([2.1, 1.6, 2.4, 1.7])
    bottom = ax.bar(x, a, color=color(0.16, 0.59, 0.49))
    top = ax.bar(x, b, bottom=a, color=color(0.88, 0.47, 0.16))
    ax.bar_label(bottom, labels=["A1", "A2", "A3", "A4"], label_type="center", color="white", fontsize=10)
    ax.bar_label(top, fmt="%.1f", color=color(0.20, 0.20, 0.20))
    save(fig, out_dir, "variants")


def demo_axes(out_dir, width, height):
    fig = make_fig(width, height)
    left = fig.add_axes(rect(0.08, 0.14, 0.42, 0.86))
    left.set_title("Top/Right + Equal Aspect")
    left.set_xlabel("top x")
    left.set_ylabel("right y")
    left.set_xlim(-1, 5)
    left.set_ylim(-1, 5)
    left.grid(True)
    left.xaxis.set_label_position("top")
    left.xaxis.tick_top()
    left.yaxis.set_label_position("right")
    left.yaxis.tick_right()
    left.set_aspect("equal", adjustable="box")
    left.minorticks_on()
    left.plot([-0.5, 0.8, 2.2, 4.2], [-0.2, 1.0, 2.1, 4.4], color=color(0.10, 0.32, 0.76), linewidth=lw(2))
    left.scatter([0, 1.5, 3.5, 4.5], [0, 1.8, 3.2, 4.6], color=color(0.92, 0.48, 0.20, 0.92), edgecolor=color(0.52, 0.22, 0.08), linewidth=lw(1), s=ss(8))
    right = fig.add_axes(rect(0.56, 0.14, 0.94, 0.86))
    right.set_title("Log, Twin, Secondary")
    right.set_xlabel("seconds")
    right.set_ylabel("count")
    right.set_xlim(0, 10)
    right.set_ylim(1, 100)
    right.set_yscale("log")
    right.grid(True)
    right.plot([0, 2, 4, 6, 8, 10], [2, 6, 9, 18, 40, 82], color=color(0.12, 0.45, 0.72), linewidth=lw(2), label="log series")
    twin = right.twinx()
    twin.set_ylim(0, 100)
    twin.plot([0, 2, 4, 6, 8, 10], [10, 22, 38, 58, 81, 96], color=color(0.80, 0.22, 0.22), linewidth=lw(1.8), label="twin")
    right.secondary_xaxis("top", functions=(lambda x: x * 10, lambda x: x / 10))
    right.legend()
    save(fig, out_dir, "axes")


def demo_statistics(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.91, 0.10, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    data = [[1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2], [1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0], [2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6]]
    ax = axs[0][0]
    ax.set_title("Box + Violin")
    ax.set_xlim(0.4, 3.6)
    ax.set_ylim(0.6, 5.4)
    ax.grid(True, axis="y")
    ax.boxplot(data, positions=[1, 2, 3], widths=0.42, patch_artist=True, boxprops={"facecolor": color(0.39, 0.62, 0.84, 0.38)})
    ax.violinplot(data, showmeans=True, showmedians=True)
    ax = axs[0][1]
    ax.set_title("ECDF")
    vals = np.array([1.2, 1.8, 2.0, 2.0, 3.1, 3.7, 4.3, 5.0, 5.8, 6.6, 7.0])
    ax.step(np.sort(vals), np.arange(1, len(vals) + 1) / len(vals), where="post", color=color(0.18, 0.36, 0.75), linewidth=lw(2))
    ax.set_xlim(0, 8)
    ax.set_ylim(0, 1.05)
    ax.grid(True, axis="y")
    ax = axs[1][0]
    ax.set_title("StackPlot")
    ax.stackplot([0, 1, 2, 3, 4, 5], [[1.0, 1.4, 1.3, 1.8, 1.6, 2.0], [0.8, 1.1, 1.4, 1.2, 1.6, 1.8], [0.5, 0.8, 1.0, 1.4, 1.1, 1.5]], colors=[color(0.20, 0.55, 0.75, 0.76), color(0.90, 0.48, 0.18, 0.76), color(0.35, 0.66, 0.42, 0.76)])
    ax.set_xlim(0, 5)
    ax.set_ylim(0, 7)
    ax.grid(True, axis="y")
    ax = axs[1][1]
    ax.set_title("Cumulative Multi-Hist")
    ax.hist([[0.3, 0.8, 1.2, 1.7, 2.6, 3.4, 4.1, 5.2], [0.5, 1.1, 1.9, 2.3, 2.8, 3.0, 3.7, 4.5, 5.0], [1.0, 1.6, 2.2, 2.9, 3.5, 4.2, 4.8, 5.4]], bins=[0, 1, 2, 3, 4, 5, 6], stacked=True, color=[color(0.22, 0.55, 0.70, 0.8), color(0.86, 0.42, 0.19, 0.8), color(0.36, 0.62, 0.36, 0.8)])
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 6)
    ax.grid(True, axis="y")
    save(fig, out_dir, "statistics")


def demo_units(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(1, 3, 0.06, 0.98, 0.17, 0.86, 0.10, 0.08)
    axs = [fig.add_axes(rects[0][col]) for col in range(3)]
    ax = axs[0]
    ax.set_title("Dates")
    ax.set_ylabel("requests")
    ax.grid(True, axis="y")
    dates = [dt.datetime(2024, 1, d) for d in [1, 3, 7, 10]]
    ax.plot(dates, [12, 18, 9, 21], color=color(0.12, 0.47, 0.71), linewidth=lw(2))
    ax.xaxis.set_major_formatter(mdates.DateFormatter("%Y-%m-%d"))
    ax.tick_params(axis="x", rotation=30)
    ax = axs[1]
    ax.set_title("Categories")
    ax.set_ylabel("count")
    ax.grid(True, axis="y")
    ax.bar(["draft", "review", "ship", "watch"], [3, 8, 6, 4], color=color(1.0, 0.50, 0.05), edgecolor=color(0.60, 0.30, 0.03), linewidth=lw(1))
    ax = axs[2]
    ax.set_title("Categorical Y")
    ax.set_xlabel("hours")
    ax.grid(True, axis="x")
    ax.barh(["north", "south", "east"], [4, 7, 5], color=color(0.17, 0.63, 0.17), edgecolor=color(0.09, 0.36, 0.09), linewidth=lw(1))
    save(fig, out_dir, "units")


def demo_matrix(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(1, 3, 0.07, 0.92, 0.14, 0.86, 0.10, 0.06)
    axs = [fig.add_axes(rects[0][col]) for col in range(3)]
    axs[0].set_title("MatShow")
    axs[0].imshow([[0.1, 0.5, 0.9], [0.7, 0.3, 0.2], [0.4, 0.8, 0.6]], cmap="viridis", origin="upper")
    axs[1].set_title("Spy")
    axs[1].spy([[1, 0, 0, 2, 0], [0, 0, 3, 0, 0], [4, 0, 0, 0, 5], [0, 6, 0, 0, 0]], color=color(0.13, 0.43, 0.72), markersize=8)
    axs[2].set_title("Annotated Heatmap")
    data = np.array([[0.1, 0.7, 0.4], [0.9, 0.2, 0.5], [0.3, 0.8, 0.6]])
    im = axs[2].imshow(data, cmap="magma", origin="upper")
    for y in range(3):
        for x in range(3):
            axs[2].text(x, y, f"{data[y, x]:.1f}", ha="center", va="center", color="white" if data[y, x] > 0.5 else color(0.05, 0.05, 0.05), fontsize=9)
    fig.colorbar(im, ax=axs[2], label="value")
    save(fig, out_dir, "matrix")


def demo_mesh(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.91, 0.12, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("PColorMesh")
    ax.pcolormesh([0, 1, 2, 3, 4], [0, 1, 2, 3], [[0.2, 0.6, 0.3, 0.9], [0.4, 0.8, 0.5, 0.7], [0.1, 0.3, 0.9, 0.6]], edgecolors=color(0.95, 0.95, 0.95), linewidth=lw(0.8))
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    ax = axs[0][1]
    ax.set_title("Contour + Contourf")
    z = np.array([[0, 0.4, 0.8, 0.4, 0], [0.2, 0.8, 1.3, 0.8, 0.2], [0.3, 1.0, 1.7, 1.0, 0.3], [0.2, 0.8, 1.3, 0.8, 0.2], [0, 0.4, 0.8, 0.4, 0]])
    ax.contourf(z, levels=[0.2, 0.6, 1.0, 1.4, 1.8])
    ax.contour(z, levels=[0.4, 0.8, 1.2, 1.6], colors=[color(0.18, 0.18, 0.18)])
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 4)
    ax = axs[1][0]
    ax.set_title("Hist2D")
    ax.hist2d([0.4, 0.7, 1.1, 1.4, 1.8, 2.1, 2.3, 2.6, 2.9, 3.2, 3.4, 3.6], [0.6, 1.0, 1.2, 1.6, 1.4, 2.0, 2.3, 2.1, 2.8, 3.0, 3.2, 3.4], bins=[[0, 1, 2, 3, 4], [0, 1, 2, 3, 4]])
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 4)
    ax = axs[1][1]
    ax.set_title("Triangulation")
    tx = [0.4, 1.6, 3.0, 0.8, 2.1, 3.5]
    ty = [0.5, 0.4, 0.7, 2.2, 2.8, 2.1]
    triangles = [[0, 1, 3], [1, 4, 3], [1, 2, 4], [2, 5, 4]]
    vals = [0.2, 0.8, 1.0, 1.5, 1.1, 0.6]
    ax.tripcolor(tx, ty, triangles, vals)
    ax.triplot(tx, ty, triangles, color=color(0.15, 0.15, 0.15), linewidth=lw(1))
    ax.tricontour(tx, ty, triangles, vals, levels=[0.7, 1.1], colors=[color(0.98, 0.98, 0.98)])
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 4)
    save(fig, out_dir, "mesh")


def demo_vectors(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 2, 0.08, 0.97, 0.10, 0.91, 0.10, 0.16)
    axs = [[fig.add_axes(rects[row][col]) for col in range(2)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("Quiver + Key")
    qx, qy, qu, qv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.8 + col
            y = 0.8 + row * 0.95
            qx.append(x)
            qy.append(y)
            qu.append(0.55 + 0.08 * math.sin(y * 0.9))
            qv.append(0.22 * math.cos(x * 0.8))
    q = ax.quiver(qx, qy, qu, qv, color=color(0.14, 0.42, 0.73), scale=10, units="dots", width=2.2)
    ax.quiverkey(q, 0.78, 0.12, 0.5, "0.5", coordinates="axes", labelpos="E")
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax.grid(True)
    ax = axs[0][1]
    ax.set_title("Barbs")
    bx, by, bu, bv = [], [], [], []
    for row in range(4):
        for col in range(5):
            x = 0.9 + col * 0.95
            y = 0.8 + row * 0.95
            bx.append(x)
            by.append(y)
            bu.append(14 + 5 * math.sin(y * 0.8))
            bv.append(8 * math.cos(x * 0.7))
    ax.barbs(bx, by, bu, bv, barbcolor=color(0.47, 0.23, 0.12), flagcolor=color(0.86, 0.52, 0.24), length=6)
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax.grid(True)
    ax = axs[1][0]
    ax.set_title("Streamplot")
    sx = np.array([0, 1, 2, 3, 4, 5, 6])
    sy = np.array([0, 1, 2, 3, 4, 5])
    su, sv = vector_grid(sx, sy)
    ax.streamplot(sx, sy, su, sv, start_points=np.array([[0.4, 0.8], [0.4, 2.2], [0.4, 3.6]]), color=color(0.13, 0.53, 0.39), integration_direction="forward", broken_streamlines=False)
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    ax = axs[1][1]
    ax.set_title("Quiver Grid XY")
    xg = np.array([0.8, 1.8, 2.8, 3.8, 4.8])
    yg = np.array([0.8, 1.8, 2.8, 3.8])
    xx, yy = np.meshgrid(xg, yg)
    ax.quiver(xx, yy, -(yy - 2.4) * 0.35, (xx - 2.8) * 0.35, color=color(0.74, 0.23, 0.27), pivot="middle", angles="xy", scale=9, units="dots", width=1.9)
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    save(fig, out_dir, "vectors")


def demo_specialty(out_dir, width, height):
    fig = make_fig(width, height)
    rects = grid_rects(2, 3, 0.07, 0.98, 0.09, 0.91, 0.10, 0.14)
    axs = [[fig.add_axes(rects[row][col]) for col in range(3)] for row in range(2)]
    ax = axs[0][0]
    ax.set_title("Eventplot")
    ax.eventplot([[0.8, 1.4, 3.1, 4.6, 7.3], [1.2, 2.9, 4.0, 6.4, 8.6], [0.5, 2.2, 5.4, 6.8, 9.1]], lineoffsets=[1, 2, 3], linelengths=[0.6, 0.7, 0.8], colors=[color(0.18, 0.44, 0.74), color(0.84, 0.38, 0.16), color(0.20, 0.63, 0.42)])
    ax.set_xlim(0, 10)
    ax.set_ylim(0.4, 3.6)
    ax = axs[0][1]
    ax.set_title("Hexbin")
    ax.hexbin([0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88], [0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58], C=[1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1], gridsize=7, reduce_C_function=np.mean)
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax = axs[0][2]
    ax.set_title("Pie")
    ax.pie([28, 22, 18, 32], labels=["Core", "I/O", "Render", "Docs"], autopct="%.0f%%", startangle=90, labeldistance=1.08, explode=[0, 0.04, 0, 0.02])
    ax = axs[1][0]
    ax.set_title("Stem")
    markerline, stemlines, baseline = ax.stem([1, 2, 3, 4, 5, 6, 7], [0.9, 2.2, 1.6, 3.3, 2.4, 3.7, 2.1], basefmt=" ")
    plt.setp(stemlines, color=color(0.15, 0.42, 0.73))
    plt.setp(markerline, color=color(0.15, 0.42, 0.73), markersize=7)
    ax.axhline(0.3, color="0.4")
    ax.set_xlim(0.5, 7.5)
    ax.set_ylim(-0.2, 4.2)
    ax.grid(True, axis="y")
    ax = axs[1][1]
    ax.set_title("Table")
    ax.axis("off")
    ax.table(cellText=[["Latency", "18ms", "14ms"], ["Throughput", "220/s", "265/s"]], rowLabels=["A", "B"], colLabels=["Metric", "Q1", "Q2"], bbox=[0.04, 0.18, 0.92, 0.64])
    ax = axs[1][2]
    ax.set_title("Sankey")
    ax.axis("off")
    Sankey(ax=ax, scale=0.16, offset=0.2).add(flows=[-2, 3, 1.5], labels=["Waste", "CPU", "Cache"], orientations=[-1, 1, 1]).finish()
    save(fig, out_dir, "specialty")


def demo_annotations(out_dir, width, height):
    fig = make_fig(width, height)
    ax = fig.add_axes(rect())
    ax.set_title("Coordinate Text + Arrow Annotation")
    ax.set_xlabel("x")
    ax.set_ylabel("response")
    ax.grid(True)
    x = linspace(0, 8, 120)
    y = np.sin(x) * np.exp(-x / 8)
    peak = int(np.argmax(y))
    ax.plot(x, y, color=color(0.13, 0.43, 0.72), linewidth=lw(2.2), label="signal")
    ax.annotate("peak", xy=(x[peak], y[peak]), xytext=(44, -34), textcoords="offset points", arrowprops={"arrowstyle": "->"})
    ax.text(0.03, 0.94, "axes coords", transform=ax.transAxes, ha="left", va="top", fontsize=11)
    fig.text(0.50, 0.10, "figure coords", ha="center", va="bottom", fontsize=11)
    ax.text(6.0, -0.55, "data coords", fontsize=11, color=color(0.56, 0.22, 0.18))
    ax.text(0.98, 0.02, "anchored\ntext box", transform=ax.transAxes, ha="right", va="bottom", bbox={"facecolor": "white", "edgecolor": "0.7"})
    ax.set_xlim(0, 8)
    ax.set_ylim(-0.8, 1.1)
    ax.legend()
    save(fig, out_dir, "annotations")


def demo_composition(out_dir, width, height):
    fig = make_fig(width, height)
    fig.suptitle("GridSpec, Figure Labels, Legend, Colorbar")
    fig.supxlabel("shared figure x")
    fig.supylabel("shared figure y")
    rects = grid_rects(2, 3, 0.08, 0.92, 0.14, 0.86, 0.08, 0.12, width_ratios=[1.3, 1, 0.9])
    left = fig.add_axes(span_rect(rects, 0, 0, 2, 1))
    left.set_title("spanning axes")
    left.grid(True, axis="y")
    left.plot([0, 1, 2, 3, 4], [1.0, 1.6, 1.2, 2.2, 1.8], color=color(0.16, 0.42, 0.82), linewidth=lw(2), label="left")
    left.scatter([0, 1, 2, 3, 4], [1.0, 1.6, 1.2, 2.2, 1.8], color=color(0.91, 0.45, 0.16), s=ss(6), label="points")
    top = fig.add_axes(rects[0][1], sharex=left)
    top.set_title("shared x")
    top.plot([0, 1, 2, 3, 4], [2, 1, 2.4, 1.7, 2.8], color=color(0.23, 0.62, 0.34), linewidth=lw(1.8), label="top")
    bottom = fig.add_axes(rects[1][1])
    bottom.set_title("anchored")
    bottom.grid(True)
    bottom.plot([0, 1, 2, 3, 4], [3.0, 2.6, 2.9, 2.1, 1.9], color=color(0.69, 0.27, 0.67), linewidth=lw(1.8), label="bottom")
    bottom.text(0.02, 0.98, "axes note", transform=bottom.transAxes, ha="left", va="top", bbox={"facecolor": "white", "edgecolor": "0.7"})
    heat = fig.add_axes(span_rect(rects, 0, 2, 2, 1))
    heat.set_title("colorbar")
    im = heat.imshow([[0.2, 0.5, 0.7], [0.9, 0.4, 0.1], [0.3, 0.8, 0.6]], cmap="inferno", origin="upper")
    fig.colorbar(im, ax=heat, label="intensity")
    fig.legend()
    fig.text(0.98, 0.02, "figure note", ha="right", va="bottom", bbox={"facecolor": "white", "edgecolor": "0.7"})
    save(fig, out_dir, "composition")


DEMOS = {
    "lines": demo_lines,
    "scatter": demo_scatter,
    "bars": demo_bars,
    "fills": demo_fills,
    "variants": demo_variants,
    "axes": demo_axes,
    "histogram": demo_histogram,
    "statistics": demo_statistics,
    "errorbars": demo_errorbars,
    "units": demo_units,
    "heatmap": demo_heatmap,
    "matrix": demo_matrix,
    "mesh": demo_mesh,
    "vectors": demo_vectors,
    "specialty": demo_specialty,
    "patches": demo_patches,
    "annotations": demo_annotations,
    "composition": demo_composition,
    "polar": demo_polar,
    "phase7": demo_phase7,
    "subplots": demo_subplots,
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True, help="Directory to write PNG files")
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH, help="Rendered width in pixels")
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT, help="Rendered height in pixels")
    parser.add_argument("--plots", nargs="*", default=None, help="Subset of web demo IDs to generate")
    parser.add_argument("--list", action="store_true", help="List available demo IDs and exit")
    args = parser.parse_args()

    if args.list:
        for name in DEMOS:
            print(name)
        return

    requested = []
    for item in args.plots or []:
        requested.extend(part.strip() for part in item.split(",") if part.strip())

    names = list(DEMOS)
    if requested and requested != ["all"]:
        unknown = sorted(set(requested) - set(DEMOS))
        if unknown:
            parser.error(f"unknown web demo IDs: {', '.join(unknown)}")
        names = requested

    os.makedirs(args.output_dir, exist_ok=True)
    for name in names:
        DEMOS[name](args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
