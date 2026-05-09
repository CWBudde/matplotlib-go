from __future__ import annotations

from pathlib import Path
import sys

try:
    from test.matplotlib_ref.common import *  # noqa: F401,F403
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parents[1]))
    from common import *  # noqa: F401,F403


def boundarynorm_pcolormesh(out_dir):
    fig, ax = plt.subplots(figsize=(640 / DPI, 360 / DPI), dpi=DPI)
    ax.set_position([0.12, 0.16, 0.78, 0.72])
    ax.set_title("BoundaryNorm PColorMesh")
    ax.set_xlabel("x")
    ax.set_ylabel("y")

    data = np.array([
        [0.2, 0.8, 1.2, 1.8],
        [2.2, 2.8, 3.2, 3.8],
        [0.5, 1.5, 2.5, 3.5],
    ])
    norm = mcolors.BoundaryNorm([0, 1, 2, 3, 4], ncolors=256)
    mesh = ax.pcolormesh([0, 1, 2, 3, 4], [0, 1, 2, 3], data, cmap="viridis", norm=norm, shading="flat")
    ax.set_xlim(0, 4)
    ax.set_ylim(0, 3)
    cbar = fig.colorbar(mesh, ax=ax)
    cbar.set_label("band")

    save(fig, out_dir, "boundarynorm_pcolormesh")


PLOT = boundarynorm_pcolormesh
