from __future__ import annotations

import importlib
from pathlib import Path

CASE_IDS = [
    line.strip()
    for line in (Path(__file__).with_name("cases.txt")).read_text().splitlines()
    if line.strip()
]


def load_plot(name: str):
    if name not in CASE_IDS:
        raise KeyError(name)
    module = importlib.import_module(f"{__name__}.{name}.plot")
    return module.PLOT


def all_plots():
    return [load_plot(name) for name in CASE_IDS]
