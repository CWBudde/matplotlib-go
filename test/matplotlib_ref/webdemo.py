#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = ["matplotlib>=3.7", "numpy"]
# ///
"""Generate Matplotlib reference images for the browser demo catalog."""

from __future__ import annotations

import argparse
import os
import sys

try:
    from test.matplotlib_ref.webdemo_common import DEFAULT_HEIGHT, DEFAULT_WIDTH
    from test.matplotlib_ref.webdemos import DEMO_NAMES, all_demos, load_demo
except ModuleNotFoundError:
    sys.path.append(os.path.dirname(__file__))
    from webdemo_common import DEFAULT_HEIGHT, DEFAULT_WIDTH
    from webdemos import DEMO_NAMES, all_demos, load_demo


def requested_demo_names(raw_items):
    requested = []
    for item in raw_items or []:
        requested.extend(part.strip() for part in item.split(",") if part.strip())
    if not requested or requested == ["all"]:
        return list(DEMO_NAMES)
    return requested


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", required=True, help="Directory to write PNG files")
    parser.add_argument("--width", type=int, default=DEFAULT_WIDTH, help="Rendered width in pixels")
    parser.add_argument("--height", type=int, default=DEFAULT_HEIGHT, help="Rendered height in pixels")
    parser.add_argument("--plots", nargs="*", default=None, help="Subset of web demo IDs to generate")
    parser.add_argument("--list", action="store_true", help="List available demo IDs and exit")
    args = parser.parse_args()

    if args.list:
        for name in DEMO_NAMES:
            print(name)
        return

    names = requested_demo_names(args.plots)
    unknown = sorted(set(names) - set(DEMO_NAMES))
    if unknown:
        parser.error(f"unknown web demo IDs: {', '.join(unknown)}")

    os.makedirs(args.output_dir, exist_ok=True)
    for name in names:
        load_demo(name)(args.output_dir, args.width, args.height)


if __name__ == "__main__":
    main()
