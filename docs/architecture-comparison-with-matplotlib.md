# Matplotlib-go vs. Upstream Matplotlib Architecture

## Scope

Source snapshot (local clone): `third_party/matplotlib` (downloaded and checked in as an ignored third-party reference tree).

This document highlights architectural differences between:
- the Go port in this repository (`core`, `render`, `transform`, `canvas`, `backends`, `style`, `pyplot`, etc.),
- upstream Matplotlib (`third_party/matplotlib/lib/matplotlib`).

For the concrete port-side contracts that follow from these differences, see
`docs/phase0-architecture-alignment.md`.

## 1) Package footprint and decomposition

Upstream Matplotlib is a full scientific plotting framework with broad domain modules: projections, widgets, animation, testing, UI backends, tooling, docs/test infra, and package-level sub-systems.

The port is intentionally smaller and render-focused:
- drawing primitives and artist orchestration in `core`,
- backend abstraction in `render` and `backends`,
- transform math in `transform`,
- style/runtime wrappers in `style`,
- pyplot-like façade in `pyplot`,
- optional event/runtime layer in `canvas`.

This is visible in layout breadth:
- Upstream: `lib/matplotlib` includes `axes`, `backends`, `animation`, `text`, `tri`, `projections`, `toolkits`, `sphinxext`, `testing`, `widgets`, etc.
- Port: primary feature set is contained in a compact package set with explicit feature boundaries.

## 2) Artist/figure model

Upstream relies on a large mutable class hierarchy (`Artist` base class and subclasses in Python OOP style) with implicit mutation behavior, generated property methods, stale flags, callbacks, pickability, animation hooks, and decorators for rasterization.

Port adopts a lean interface-first model:
- `core/artist.go` defines small interfaces (`Artist`, `OverlayArtist`) and a `DrawContext`.
- `core/figure.go`, `core/axes.go`, and peers provide explicit struct fields and methods.

Result: easier static reasoning and deterministic construction in Go, but less dynamic introspection/state machinery than upstream’s object model.

## 3) Rendering contract

Upstream separates renderer, graphics context, and canvas responsibilities through:
- `backend_bases.RendererBase`,
- `GraphicsContextBase`,
- backend-specific `Renderer` implementations with optional fast paths for path collections, markers, quad meshes, image paths, text, and event integration.

Port uses:
- a compact `render.Renderer` interface (`render/render.go`) plus capability interfaces (`TextDrawer`, `PNGExporter`, `SVGExporter`, etc.),
- concrete backend packages under `backends` (`agg`, `gobasic`).

Result: less method surface and lower backend complexity in the port; easier portability, but with narrower backend feature parity and fewer specialized draw fast paths.

## 4) Transform architecture

Upstream’s transform system is graph-based and stateful:
- `lib/matplotlib/transforms.py` has `TransformNode` with parent/child invalidation, frozen copies, affine/non-affine split, and deferred recomputation.

Port uses explicit values and direct composition:
- `transform/transform.go` and `transform/graph.go` provide `T`, `Separable`, `AffineT`, scale types (`Linear`, `Log`), and deterministic composition without transform-node invalidation.

Result: simpler API and lower runtime metadata overhead; less support for dynamic dependency chains and shared transform caching.

## 5) Backend and canvas/runtime integration

Upstream backend architecture is event-centric and format-aware:
- backend format registration (`backend_bases` `_default_backends`),
- `FigureCanvasBase`, `FigureManagerBase`, rich event model (`KeyEvent`, `MouseEvent`, etc.),
- interactive toolbar/tool manager integration, output routing to dedicated format backends.

Port provides:
- backend selection via `backends/registry.go`,
- minimal event model in `canvas/canvas.go`,
- headless canvas and manager orchestration in `backends/runtime.go`,
- renderer feature checks in save/draw paths.

Result: functional runtime and event dispatch, but reduced interactive/UI/back-end ecosystem relative to full Matplotlib.

## 6) Style/rc configuration

Upstream has exhaustive `rcParams` schema, validators and coercion logic in `lib/matplotlib/rcsetup.py`, with global config layers, contexts, and extensive deprecation/validation behavior.

Port keeps style explicit and lightweight:
- `style/style.go` exposes an `RC` struct with concrete fields and options,
- `style/runtime.go` parses and applies a subset of Matplotlib-style keys with stackable context behavior.

Result: configuration is easier to reason about in Go, but coverage is intentionally reduced.

## 7) API surface and stateful convenience layer

Upstream `pyplot` is a very large, reflection-heavy wrapper layer over objects and backend state, including many aliases, compatibility branches, and generated docs/tooling support.

Port’s `pyplot/pyplot.go` provides a deliberately narrower MATLAB-like API set for parity-oriented workflows:
- figure/axes creation and common plotting helpers,
- manager lifecycle (`Show`, `Pause`, `Save`),
- focused command surface for generated reference parity rather than 1:1 API completeness.

## 8) Testing and validation model

Upstream tests span Python unit tests, image comparison, animation, interactive behavior, and backend coverage at scale.

Port emphasizes deterministic rendering parity in Go tests (e.g. PNG golden/reference checks in `test/`, golden fixtures in `testdata/`, and dedicated parity docs in `docs/`).

## 9) Practical takeaway

The two codebases intentionally diverge in ambition:
- upstream targets a broad general-purpose plotting ecosystem,
- this project targets a curated, deterministic Go rendering subset with explicit extension points.

That difference shows up mostly in architecture (class graph vs interfaces, transform graph vs explicit combinators, backend/event richness vs minimal runtime abstraction), not as a direct line-for-line port.
