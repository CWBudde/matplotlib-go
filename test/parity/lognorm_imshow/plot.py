from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def _log_norm_fixture_data(rows, cols):
    data = np.zeros((rows, cols))
    for row in range(rows):
        for col in range(cols):
            t = (row * cols + col) / (rows * cols - 1)
            data[row, col] = 10 ** (3 * t)
    return data


def lognorm_imshow(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("LogNorm Imshow")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    im = ax.imshow(
        _log_norm_fixture_data(5, 6),
        cmap="magma",
        norm=mcolors.LogNorm(vmin=1, vmax=1000),
        origin="lower",
        extent=[0, 6, 0, 5],
        aspect="auto",
    )
    ax.set_xlim(0, 6)
    ax.set_ylim(0, 5)
    cbar = fig.colorbar(im, ax=ax)
    cbar.set_label("log value")

    save(fig, out_dir, "lognorm_imshow")


PLOT = lognorm_imshow
