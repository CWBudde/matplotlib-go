// Package pdf implements a native PDF renderer backend for matplotlib-go.
//
// The backend emits a single-page PDF document with deterministic object
// numbering and trailer offsets so byte-for-byte reproducibility is the
// default. The initial implementation focuses on the core renderer contract:
// paths with stroke/fill/clip, raster images, and text-as-path output. Real
// embedded font subsetting, native hatch patterns, and tiling pattern fills
// are tracked as follow-up work in PLAN.md Phase 1.1.
//
// PDF-specific output options are carried by render.PDFOptions and can be
// passed through core.SavePDF, core.SaveFig, or the backends.Registry save
// dispatch once the option pipeline is wired up. The default writes an empty
// /Info dictionary and uses text-as-path output for typography, which avoids
// the need for an embedded font program in the first cut.
package pdf
