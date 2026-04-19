// Package render exposes the Renderer interface and no-op backend.
//
// Colors passed through this package use normalized sRGBA component values in
// the range [0..1] unless a specific backend documents a different contract.
//
// Phase A: scaffold only. Renderer verbs and NullRenderer are introduced in
// Phase C per TASKS.md.
package render
