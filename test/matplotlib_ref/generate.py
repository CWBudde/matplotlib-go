#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.7"]
# ///
"""Generate matplotlib reference images for matplotlib-go test comparisons."""

from __future__ import annotations

import argparse
import json
import os
import sys

import numpy as np

try:
    from test.matplotlib_ref.common import histogram_payload, normal_data, pcg_float64_values, _to_list
    from test.matplotlib_ref.plots import PLOT_NAMES, all_plots, load_plot
except ModuleNotFoundError:
    sys.path.append(os.path.dirname(__file__))
    from common import histogram_payload, normal_data, pcg_float64_values, _to_list
    from plots import PLOT_NAMES, all_plots, load_plot


def rng_debug_payload():
    return {
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


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", default="", help="Directory to write PNG files")
    parser.add_argument("--emit-rng-debug", action="store_true", help="Emit RNG parity payload as JSON and exit")
    parser.add_argument("--plots", nargs="*", help="Subset of plot names to generate (default: all)")
    args = parser.parse_args()

    if args.emit_rng_debug:
        print(json.dumps(rng_debug_payload()))
        return

    if not args.output_dir:
        parser.error("--output-dir is required unless --emit-rng-debug is set")

    os.makedirs(args.output_dir, exist_ok=True)

    if args.plots:
        unknown = set(args.plots) - set(PLOT_NAMES)
        if unknown:
            print(f"Unknown plots: {', '.join(sorted(unknown))}", file=sys.stderr)
            print(f"Available: {', '.join(sorted(PLOT_NAMES))}", file=sys.stderr)
            sys.exit(1)
        to_run = [load_plot(name) for name in args.plots]
    else:
        to_run = all_plots()

    print(f"Generating {len(to_run)} matplotlib reference image(s) → {args.output_dir}/")
    for fn in to_run:
        fn(args.output_dir)
    print("Done.")


if __name__ == "__main__":
    main()
