# Phase 0 Architecture Alignment

This document turns the Phase 0 parity gaps into concrete contracts for this
Go port. It complements `docs/architecture-comparison-with-matplotlib.md`,
which describes the upstream differences at a higher level.

## 1. Artist state, callbacks, stale state, and lifecycle

Current baseline:

- `core.Artist` remains a small rendering interface: `Draw`, `Z`, and `Bounds`.
- Mutable artist behavior stays on concrete artist structs instead of a shared
  Matplotlib-style base class.
- Draw traversal is owned by `core.DrawFigure`: clipped data artists first,
  overlays second, then axes/frame/ticks/text outside the axes clip.

Port decision:

- Do not add a large base artist class only for parity.
- Add callback, stale, picker, and animation hooks as narrow optional
  interfaces when a feature needs them.
- Treat stale propagation as a scheduling signal for canvases and managers, not
  as a required part of every artist mutation.
- Keep overlay rendering explicit through `core.OverlayArtist` so annotations
  and labels can intentionally bypass the axes clip.

Parity-critical lifecycle boundaries:

- Figure layout is prepared before artist draw traversal.
- Axes-local data artists render under the axes clip and projection frame clip.
- Overlay artists render after clipped artists and before axes decorations.
- Axes decorations render outside the clip so spines, outward ticks, and labels
  can straddle the axes boundary.
- Interactive mutation should invalidate transforms or canvas scheduling state,
  but should not draw directly from property setters.

## 2. Transform and coordinate-system parity

Current baseline:

- `transform.T` remains the core transform contract.
- `core.DrawContext` exposes Matplotlib-style `TransData`, `TransAxes`,
  `TransFigure`, and blended coordinate resolution.
- The default model is explicit value composition rather than an always-stateful
  transform graph.

Added contract:

- `transform.TransformNode` is the opt-in compatibility layer for dependency
  invalidation.
- `transform.Invalidation` preserves the affine vs non-affine split needed for
  projection-heavy or layout-heavy transform chains.
- `transform.CachedTransform` provides a minimal cached `transform.T` wrapper
  for shared dynamic compositions.

Port decision:

- Use transform nodes only where shared dynamic dependencies matter.
- Keep simple transforms as plain values to avoid global invalidation machinery
  in the common static rendering path.

## 3. Renderer and backend runtime parity

Current baseline:

- `render.Renderer` stays compact and backend-neutral.
- Optional backend features are represented by Go interfaces such as
  `render.TextDrawer`, `render.PNGExporter`, and `render.SVGExporter`.
- `backends.Registry` is the single registry for backend factories,
  capabilities, managers, canvases, and save-format handlers.

Contract:

- A backend capability declaration must match the renderer's optional
  interfaces when the capability has a runtime interface check.
- Save dispatch belongs to backend format handlers, not ad hoc path-extension
  switches scattered through call sites.
- Backend contract tests should cover capability declarations, save routing,
  text support, clipping support, and export behavior.

## 4. Canvas, events, managers, and state API

Current baseline:

- `canvas.FigureCanvas` owns drawing, resizing, event connection, disconnection,
  and close lifecycle.
- `canvas.FigureManager` owns presentation, host title, close behavior, and
  tools.
- `canvas.Dispatcher` owns connection IDs and handler lifecycle.

Added contract:

- First-class event wrappers map to Matplotlib event families:
  `DrawEvent`, `ResizeEvent`, `CloseEvent`, `MouseEvent`, `KeyEvent`, and
  `PickEvent`.
- The generic `canvas.Event` remains the normalized payload used by handlers.
- Backend-specific native payloads belong in `Event.Native`.

Port decision:

- Backends should emit normalized events through `Dispatcher`.
- Interactive backends may add timers and redraw queues, but `draw_idle`-style
  scheduling should be manager/canvas behavior rather than artist behavior.
- Pick behavior should be layered on top of normalized mouse events and artist
  hit-testing, not embedded in the base event dispatcher.

## 5. Style/RC and pyplot parity

Current baseline:

- `style.RC` is a concrete Go value type.
- Matplotlib `.mplstyle` support is intentionally limited to
  `style.SupportedMPLStyleKeys`.
- `pyplot` provides a focused stateful facade over figures, axes, managers, and
  common plot helpers.

Expansion plan:

- Add rc keys by extending `supportedMPLStyleKeys`, parser validation, and
  option precedence tests together.
- Prefer explicit Go fields for stable rc concepts over untyped global maps.
- Track unsupported `.mplstyle` keys through `MPLStyleReport` instead of
  silently accepting them.
- Group pyplot parity into buckets: figure/axes lifecycle, axes properties,
  plot wrappers, save/output dispatch, context helpers, and interactive
  manager hooks.

## 6. Architecture-first test strategy

Phase 0 architecture tests should pin contracts before feature work expands:

- Transform semantics: invalidation propagation, affine/non-affine stale flags,
  and cached transform rebuild behavior.
- Backend behavior: capability declarations, runtime optional-interface checks,
  and save-format routing.
- Event lifecycle: canonical event types, typed event wrappers, connection IDs,
  disconnect behavior, resize/draw/close dispatch.
- RC behavior: supported key list stability, parser reporting, and explicit
  precedence when rc params are applied over defaults.

Review gate for future feature phases:

- New feature work that touches rendering, transforms, events, or rc parsing
  should either reuse these contracts or add a matching architecture test.
