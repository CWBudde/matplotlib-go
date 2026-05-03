//go:build !freetype

package agg

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) drawNativeFreetypeText(_ string, _ render.FontFace, _ geom.Pt, _ float64, _ render.Color) bool {
	return false
}

func nativeFreetypeVersion() string {
	return ""
}
