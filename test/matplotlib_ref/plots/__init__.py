from __future__ import annotations

import importlib

PLOT_NAMES = ['basic_line', 'joins_caps', 'dashes', 'scatter_basic', 'scatter_marker_types', 'scatter_advanced', 'bar_basic_frame', 'bar_basic_ticks', 'bar_basic_tick_labels', 'bar_basic_title', 'bar_basic', 'bar_horizontal', 'bar_grouped', 'fill_basic', 'fill_between', 'fill_stacked', 'errorbar_basic', 'multi_series_basic', 'multi_series_color_cycle', 'hist_basic', 'hist_density', 'hist_strategies', 'boxplot_basic', 'text_labels_strict', 'title_strict', 'image_heatmap', 'axes_top_right_inverted', 'axes_control_surface', 'transform_coordinates', 'gridspec_composition', 'figure_labels_composition', 'colorbar_composition', 'annotation_composition', 'patch_showcase', 'mesh_contour_tri', 'plot_variants', 'stat_variants', 'stem_plot', 'specialty_artists', 'units_overview', 'units_dates', 'units_categories', 'units_custom_converter', 'vector_fields', 'polar_axes', 'geo_mollweide_axes', 'unstructured_showcase', 'arrays_showcase', 'axisartist_showcase', 'axes_grid1_showcase']

def load_plot(name: str):
    if name not in PLOT_NAMES:
        raise KeyError(name)
    module = importlib.import_module(f"{__name__}.{name}")
    return module.PLOT

def all_plots():
    return [load_plot(name) for name in PLOT_NAMES]
