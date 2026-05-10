#!/usr/bin/env python3
from __future__ import annotations

import matplotlib.pyplot as plt


def main():
    print("Matplotlib backend information")
    print("==============================")
    print()

    # Matplotlib exposes the active backend rather than a matplotlib-go style registry.
    print(f"active backend: {plt.get_backend()}")
    print("common static formats: png, pdf, svg")
    print()
    print("Recommended Backends by Use Case:")
    print("---------------------------------")
    print("  basic: Agg")
    print("  publication: svg/pdf")
    print("  interactive: QtAgg/TkAgg when installed")
    print("  scientific: Agg for reproducible batch output")


if __name__ == "__main__":
    main()
