from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _two_slope_fixture_data(rows, cols):
    data = np.zeros((rows, cols))
    for row in range(rows):
        y = row / max(1, rows - 1)
        for col in range(cols):
            x = col / max(1, cols - 1)
            data[row, col] = 6 * x - 3 + 1.5 * math.sin((y - 0.5) * math.pi)
    return data


def twoslope_norm_image(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("TwoSlopeNorm Diverging")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    cmap = mcolors.LinearSegmentedColormap.from_list(
        "diverging fixture",
        [
            (0.0, (0.23, 0.30, 0.75, 1.0)),
            (0.5, (0.86, 0.86, 0.86, 1.0)),
            (1.0, (0.71, 0.02, 0.15, 1.0)),
        ],
    )
    im = ax.imshow(
        _two_slope_fixture_data(5, 7),
        cmap=cmap,
        norm=mcolors.TwoSlopeNorm(vmin=-3, vcenter=0, vmax=6),
        origin="lower",
        extent=[0, 7, 0, 5],
        aspect="auto",
    )
    ax.set_xlim(0, 7)
    ax.set_ylim(0, 5)
    cbar = fig.colorbar(im, ax=ax)
    cbar.set_label("anomaly")

    save(fig, out_dir, "twoslope_norm_image")


PLOT = twoslope_norm_image
