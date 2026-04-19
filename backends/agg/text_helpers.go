package agg

import (
	"math"

	agglib "github.com/cwbudde/agg_go"
)

type textBackend uint8

const (
	textBackendTrueType textBackend = iota
	textBackendRaster
	textBackendGSV
)

type configuredTextFont struct {
	backend  textBackend
	fontPath string
	size     float64
	fallback bool
}

func (r *Renderer) configureTextFont(size float64, fontKey string) configuredTextFont {
	if size <= 0 {
		return configuredTextFont{}
	}
	if fontPath := r.resolveTextFontPath(fontKey); fontPath != "" && r.ctx.ConfigureTextFont(fontPath, size, r.resolution) == nil {
		return configuredTextFont{
			backend:  textBackendTrueType,
			fontPath: fontPath,
			size:     size,
		}
	}
	if fontPath := r.resolveTextFontPath(fontKey); fontPath != "" {
		if _, ok := r.measureRasterText("M", fontPath, size); ok {
			return configuredTextFont{
				backend:  textBackendRaster,
				fontPath: fontPath,
				size:     size,
			}
		}
	}
	r.fallback = true
	return configuredTextFont{
		backend:  textBackendGSV,
		size:     math.Max(size, 6),
		fallback: true,
	}
}

func measureLocalGSVTextWidth(text string, size float64) float64 {
	gsv := agglib.NewGSVText()
	gsv.SetFlip(true)
	gsv.SetSize(size, 0)
	return gsv.MeasureText(text)
}

func appendLocalGSVText(ctx *aggSurface, x, y, size float64, text string) bool {
	if ctx == nil || text == "" {
		return false
	}

	gsv := agglib.NewGSVText()
	gsv.SetFlip(true)
	gsv.SetSize(size, 0)
	gsv.SetText(text)
	gsv.SetStartPoint(float64(int(x)), float64(int(y)))

	ctx.BeginPath()
	gsv.Rewind(0)

	hasVertices := false
	for {
		vx, vy, cmd := gsv.Vertex()
		switch cmd {
		case agglib.GSVPathCmdStop:
			return hasVertices
		case agglib.GSVPathCmdMoveTo:
			ctx.MoveTo(vx, vy)
			hasVertices = true
		case agglib.GSVPathCmdLineTo:
			ctx.LineTo(vx, vy)
			hasVertices = true
		}
	}
}

func (r *Renderer) resolveTextFontPath(fontKey string) string {
	if resolved := resolveFontPath(fontKey); resolved != "" {
		return resolved
	}
	return r.fontPath
}
