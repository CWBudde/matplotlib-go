from __future__ import annotations

import importlib

DEMO_SPECS = [('lines', 'demo_lines'), ('scatter', 'demo_scatter'), ('bars', 'demo_bars'), ('fills', 'demo_fills'), ('variants', 'demo_variants'), ('axes', 'demo_axes'), ('histogram', 'demo_histogram'), ('statistics', 'demo_statistics'), ('errorbars', 'demo_errorbars'), ('units', 'demo_units'), ('heatmap', 'demo_heatmap'), ('matrix', 'demo_matrix'), ('mesh', 'demo_mesh'), ('vectors', 'demo_vectors'), ('specialty', 'demo_specialty'), ('patches', 'demo_patches'), ('annotations', 'demo_annotations'), ('composition', 'demo_composition'), ('polar', 'demo_polar'), ('projections', 'demo_projections'), ('subplots', 'demo_subplots')]
DEMO_NAMES = [name for name, _ in DEMO_SPECS]

def load_demo(name: str):
    for demo_name, func_name in DEMO_SPECS:
        if demo_name == name:
            module = importlib.import_module(f"{__name__}.{demo_name}")
            return getattr(module, "DEMO")
    raise KeyError(name)

def all_demos():
    return [(name, load_demo(name)) for name in DEMO_NAMES]
