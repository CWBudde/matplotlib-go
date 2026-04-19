# Phase X.5 Text Parity Validation

This document freezes the validation and release discipline for AGG text-parity
work. The goal is to make every meaningful backend text change go through the
same small, repeatable loop instead of relying on memory.

## Focused Regressions

These regressions should stay in place and remain part of the normal text
parity loop.

### Backend correctness

- duplicated glyph / space replay:
  - [backends/agg/agg_freetype_test.go](/mnt/projekte/Code/matplotlib-go/backends/agg/agg_freetype_test.go)
  - `TestTrailingSpaceDoesNotRenderDuplicateGlyph`
  - `TestInternalSpaceDoesNotReplayPreviousGlyph`
- stable font line metrics vs string-specific ink:
  - [backends/agg/agg_freetype_test.go](/mnt/projekte/Code/matplotlib-go/backends/agg/agg_freetype_test.go)
  - `TestMeasureTextUsesStableFontLineMetrics`

### Core placement

- tick-label anchor placement:
  - [core/axis_test.go](/mnt/projekte/Code/matplotlib-go/core/axis_test.go)
  - `TestTickLabelPositionUsesBoundsForBottomXAxis`
  - `TestTickLabelPositionUsesBoundsForLeftYAxis`
  - `TestTickLabelPositionUsesFontHeightMetricsForBottomXAxis`
  - `TestTickLabelPositionUsesBottomAlignmentForTopXAxis`
  - `TestTickLabelPositionUsesCenterBaselineForRightYAxis`
- ylabel anchor integration:
  - [core/artist_test.go](/mnt/projekte/Code/matplotlib-go/core/artist_test.go)
  - `TestDrawAxesLabels_YLabelUsesTickBoundsAndLabelPad`

### Strict parity probes

- broad parity probes:
  - `bar_basic_tick_labels`
  - `bar_basic_title`
  - `hist_strategies`
- strict text probes:
  - `text_labels_strict`
  - `title_strict`

Relevant files:

- [test/reference_compare_test.go](/mnt/projekte/Code/matplotlib-go/test/reference_compare_test.go)
- [test/text_strict_test.go](/mnt/projekte/Code/matplotlib-go/test/text_strict_test.go)
- [test/matplotlib_ref_test.go](/mnt/projekte/Code/matplotlib-go/test/matplotlib_ref_test.go)
- [test/golden_test.go](/mnt/projekte/Code/matplotlib-go/test/golden_test.go)

## Repeatable Validation Loop

### 1. Backend correctness

Run:

```bash
go test -tags freetype ./backends/agg -run "TestUsesDejaVuSansWithoutFallback|TestRasterTextWidthTracksRendererDPI|TestMeasureTextUsesStableFontLineMetrics|TestTrailingSpaceDoesNotRenderDuplicateGlyph|TestInternalSpaceDoesNotReplayPreviousGlyph" -count=1 -v
```

### 2. Core placement

Run:

```bash
go test ./core -run "TestTitleFontSizeUsesTitleOnlyCompensation|TestDrawAxesLabels_YLabelUsesTickBoundsAndLabelPad|TestTickLabelPositionUsesBoundsForBottomXAxis|TestTickLabelPositionUsesBoundsForLeftYAxis|TestTickLabelPositionUsesFontHeightMetricsForBottomXAxis|TestTickLabelPositionUsesBottomAlignmentForTopXAxis|TestTickLabelPositionUsesCenterBaselineForRightYAxis|TestAlignedTextOrigin|TestAxesTextDrawsNormalizedContent|TestAnnotationDrawOverlayRendersArrowAndText|TestAxesTextSupportsAxesAndBlendedCoordinates" -count=1 -v
```

### 3. Canary parity probes

Run:

```bash
go test -tags freetype ./test -run "TestMpl_BarBasicTickLabels|TestMpl_BarBasicTitle|TestMpl_HistStrategies|TestTextLabelsStrict_MatplotlibRef|TestTitleStrict_MatplotlibRef" -count=1 -v
```

Current known state:

- `bar_basic_tick_labels`, `bar_basic_title`, `hist_strategies`, and `text_labels_strict` should pass comfortably.
- `TestTitleStrict_MatplotlibRef` is currently diagnostic: it is expected to remain slightly below threshold until the remaining backend title-rendering gap is fixed.

### 4. Refresh only affected goldens

Run:

```bash
go test -tags freetype ./test -run "TestBarBasicTickLabels_Golden|TestBarBasicTitle_Golden|TestHistStrategies_Golden|TestTextLabelsStrict_Golden|TestTitleStrict_Golden" -count=1 -update-golden -v
```

### 5. Compare committed goldens against Matplotlib refs

Run:

```bash
go test -tags freetype ./test -run "TestReferenceImages_GoldenVsMatplotlibRef/bar_basic_tick_labels|TestReferenceImages_GoldenVsMatplotlibRef/bar_basic_title|TestReferenceImages_GoldenVsMatplotlibRef/hist_strategies|TestTextLabelsStrict_MatplotlibRef|TestTitleStrict_MatplotlibRef" -count=1 -v
```

## Visual Inspection

After any meaningful backend text change, inspect these five cases visually:

- `bar_basic_tick_labels`
- `bar_basic_title`
- `hist_strategies`
- `text_labels_strict`
- `title_strict`

Preferred workflow:

1. regenerate only the affected goldens
2. run the Matplotlib comparison tests so fresh artifact images are written to
   `testdata/_artifacts/`
3. open the parity viewer or inspect the generated `_artifacts` PNGs directly

Useful command:

```bash
just parity-viewer FILTER="bar_basic_tick_labels|bar_basic_title|hist_strategies|text_labels_strict|title_strict"
```

## `agg_go` Release Flow

When the fix belongs in `../agg_go`, do not leave the repo on an untracked local
backend state.

Required sequence:

1. make the `../agg_go` change and add backend-local regression coverage there
2. commit the backend change cleanly
3. create and push a new `agg_go` tag
4. update `matplotlib-go` to that tag in `go.mod`
5. rerun the Phase X.5 validation loop above
6. regenerate only the affected goldens

Do not:

- keep `matplotlib-go` depending on an unversioned local backend state longer
  than necessary
- regenerate unrelated goldens for a backend-text-only change
