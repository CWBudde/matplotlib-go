# Matplotlib Agg Text Parity Reference

This document freezes the reference model for text parity work.
It is based on the local upstream checkout in `/tmp/matplotlib` and should be
treated as the canonical source map for backend text rendering work in this
repository.

## Primary Parity Probes

These fixtures are the text-rendering canaries and should stay in the loop for
every meaningful backend text change:

- `bar_basic_tick_labels`
- `text_labels_strict`
- `title_strict`

Relevant files:

- [test/golden_test.go](/mnt/projekte/Code/matplotlib-go/test/golden_test.go)
- [test/matplotlib_ref/generate.py](/mnt/projekte/Code/matplotlib-go/test/matplotlib_ref/generate.py)
- [test/text_strict_test.go](/mnt/projekte/Code/matplotlib-go/test/text_strict_test.go)
- [test/reference_compare_test.go](/mnt/projekte/Code/matplotlib-go/test/reference_compare_test.go)

## Matplotlib Agg Text Pipeline

### 1. Entry Point: `RendererAgg.draw_text`

Matplotlib’s normal text path starts in
[/tmp/matplotlib/lib/matplotlib/backends/backend_agg.py:237](/tmp/matplotlib/lib/matplotlib/backends/backend_agg.py:237).

For non-math text it:

1. prepares an FT2Font object
2. shapes the string via `font._layout(...)`
3. passes the shaped glyph stream into `_draw_text_glyphs_and_boxes(...)`

This is not an outline-fill path. Normal text is rendered as bitmap glyphs.

### 2. Glyph Placement And Device-Space Snapping

Bitmap glyph rendering happens in
[/tmp/matplotlib/lib/matplotlib/backends/backend_agg.py:175](/tmp/matplotlib/lib/matplotlib/backends/backend_agg.py:175).

Important details:

- `font.set_size(size, self.dpi)` is called for each font/size run.
- The transform includes the horizontal `1 / hinting_factor` adjustment.
- The text origin is rounded in device space with `round(0x40 * ...)`.
- Each glyph is rendered individually with `_render_glyph(...)`.
- Each glyph bitmap is then drawn individually with `draw_text_image(...)`.

This means Matplotlib does not build one OR-combined run mask for normal text.

### 3. FT2Font Font Size Setup

Font sizing is set in
[/tmp/matplotlib/src/ft2font.cpp:273](/tmp/matplotlib/src/ft2font.cpp:273).

Key details:

- `FT_Set_Char_Size(face, ptsize * 64, 0, dpi * hinting_factor, dpi)`
- An X transform of `1 / hinting_factor` is then applied

So the hinting factor affects glyph rasterization without changing the final
logical width in the obvious way. This is part of Matplotlib’s Agg text look.

### 4. Text Layout And Run Metrics

Text layout is built in
[/tmp/matplotlib/lib/matplotlib/text.py:416](/tmp/matplotlib/lib/matplotlib/text.py:416).

Important details:

- `_get_layout()` does not use only raw glyph ink bounds.
- It combines:
  - actual measured run metrics `(w, h, d)`
  - font-wide minimum ascent/descent
  - line gap
- Font-wide metrics come from:
  - OS/2 `sTypoAscender` / `sTypoDescender` / `sTypoLineGap`, or
  - `hhea` `ascent` / `descent` / `lineGap`
- This is what drives alignments such as:
  - `top`
  - `baseline`
  - `center_baseline`

This is the key reason a pure ink-bounds model does not match Matplotlib’s
vertical placement.

### 5. Tick Label Placement

Tick label transforms are sourced from:

- x tick labels:
  [/tmp/matplotlib/lib/matplotlib/axes/\_base.py:999](/tmp/matplotlib/lib/matplotlib/axes/_base.py:999)
- y tick labels:
  [/tmp/matplotlib/lib/matplotlib/axes/\_base.py:1081](/tmp/matplotlib/lib/matplotlib/axes/_base.py:1081)
- tick label artist setup:
  [/tmp/matplotlib/lib/matplotlib/axis.py:392](/tmp/matplotlib/lib/matplotlib/axis.py:392)
  and
  [/tmp/matplotlib/lib/matplotlib/axis.py:452](/tmp/matplotlib/lib/matplotlib/axis.py:452)

Matplotlib semantics:

- bottom x tick labels: transformed downward by pad, `va="top"`, `ha=xtick.alignment`
- left y tick labels: transformed left by pad, `va=ytick.alignment`, `ha="right"`

Default y tick alignment is `center_baseline`, not plain `center`.

### 6. Glyph Run Bounding Box Packing

FT2Font builds the run bounding box in
[/tmp/matplotlib/src/ft2font.cpp:467](/tmp/matplotlib/src/ft2font.cpp:467)
and packs glyph bitmaps into the run image in
[/tmp/matplotlib/src/ft2font.cpp:708](/tmp/matplotlib/src/ft2font.cpp:708).

Important details:

- glyph positions come from shaped `x_offset`, `y_offset`, `x_advance`, `y_advance`
- the run bbox is built in subpixel space
- bitmap placement uses:
  - `bitmap.left`
  - `bitmap.top`
  - `bbox.xMin`
  - `bbox.yMax`
  - the `+1` row adjustment

## Upstream Defaults That Matter

### Text Hinting

From
[/tmp/matplotlib/lib/matplotlib/backends/backend_agg.py:43](/tmp/matplotlib/lib/matplotlib/backends/backend_agg.py:43)
and
[/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:309](/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:309):

- `text.hinting = default`
- `text.hinting_factor = 1`

`font_manager.py` passes `text.hinting_factor` into FT2Font in
[/tmp/matplotlib/lib/matplotlib/font_manager.py:1819](/tmp/matplotlib/lib/matplotlib/font_manager.py:1819).

### Tick Defaults

From
[/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:511](/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:511)
through
[/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:549](/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:549):

- `xtick.major.size = 3.5`
- `xtick.major.pad = 3.5`
- `xtick.alignment = center`
- `ytick.major.size = 3.5`
- `ytick.major.pad = 3.5`
- `ytick.alignment = center_baseline`

### Title Defaults

From
[/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:397](/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:397)
through
[/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:402](/tmp/matplotlib/lib/matplotlib/mpl-data/matplotlibrc:402):

- `axes.titlesize = large`
- `axes.titleweight = normal`
- `axes.titlepad = 6.0`

## Local Files That Correspond To This Pipeline

These are the main local files that should be compared against the upstream
pipeline above when working on text parity:

- AGG renderer integration:
  - [backends/agg/agg.go](/mnt/projekte/Code/matplotlib-go/backends/agg/agg.go)
  - [backends/agg/surface.go](/mnt/projekte/Code/matplotlib-go/backends/agg/surface.go)
- AGG text backend:
  - [../agg_go/internal/agg2d/text.go](/mnt/projekte/Code/agg_go/internal/agg2d/text.go)
  - [../agg_go/internal/font/freetype/engine.go](/mnt/projekte/Code/agg_go/internal/font/freetype/engine.go)
  - [../agg_go/internal/font/freetype/shaping_harfbuzz.go](/mnt/projekte/Code/agg_go/internal/font/freetype/shaping_harfbuzz.go)
- Axes/tick layout:
  - [core/axis.go](/mnt/projekte/Code/matplotlib-go/core/axis.go)
  - [core/artist.go](/mnt/projekte/Code/matplotlib-go/core/artist.go)

## Guardrails For The Next Steps

- Treat the upstream behavior above as the reference model, not previous tuning
  experiments in this repository.
- Prefer backend fixes over local empirical offsets.
- When a fix changes text output, inspect at least:
  - `bar_basic_tick_labels`
  - `text_labels_strict`
  - `title_strict`
