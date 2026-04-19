package agg

import (
	"math"

	agglib "github.com/cwbudde/agg_go"
)

type textBackend uint8

const (
	textBackendTrueType textBackend = iota
	textBackendGSV
)

type configuredTextFont struct {
	backend  textBackend
	size     float64
	outline  *agglib.FreeTypeOutlineText
	fallback bool
}

func (f configuredTextFont) Close() {
	if f.outline != nil {
		_ = f.outline.Close()
	}
}

func (r *Renderer) configureTextFont(size float64) configuredTextFont {
	if size <= 0 {
		return configuredTextFont{}
	}
	if outline := r.loadTrueTypeOutline(size); outline != nil {
		return configuredTextFont{
			backend: textBackendTrueType,
			size:    size,
			outline: outline,
		}
	}
	r.fallback = true
	return configuredTextFont{
		backend:  textBackendGSV,
		size:     math.Max(size, 6),
		fallback: true,
	}
}

func (r *Renderer) loadTrueTypeOutline(size float64) *agglib.FreeTypeOutlineText {
	if r.fontPath == "" || size <= 0 {
		return nil
	}

	outline, err := agglib.NewFreeTypeOutlineText()
	if err != nil {
		return nil
	}
	outline.SetResolution(r.resolution)
	outline.SetHinting(true)
	outline.SetFlip(true)
	if err := outline.SetTrueTypeInterpreterVersion(35); err != nil {
		_ = outline.Close()
		return nil
	}
	outline.SetSize(size, 0)
	if err := outline.LoadFont(r.fontPath); err != nil {
		_ = outline.Close()
		return nil
	}
	return outline
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

func appendFreeTypeOutlineText(ctx *aggSurface, text *agglib.FreeTypeOutlineText) bool {
	if ctx == nil || text == nil {
		return false
	}

	ctx.BeginPath()
	text.Rewind(0)

	hasVertices := false
	for {
		x1, y1, cmd := text.Vertex()
		switch {
		case cmd == agglib.PathCmdStop:
			return hasVertices
		case cmd == agglib.PathCmdMoveTo:
			ctx.MoveTo(x1, y1)
			hasVertices = true
		case cmd == agglib.PathCmdLineTo:
			ctx.LineTo(x1, y1)
			hasVertices = true
		case agglib.IsPathCurve3(cmd):
			x2, y2, cmd2 := text.Vertex()
			if !agglib.IsPathCurve3(cmd2) {
				return hasVertices
			}
			ctx.QuadricCurveTo(x1, y1, x2, y2)
			hasVertices = true
		case agglib.IsPathCurve4(cmd):
			x2, y2, cmd2 := text.Vertex()
			x3, y3, cmd3 := text.Vertex()
			if !agglib.IsPathCurve4(cmd2) || !agglib.IsPathCurve4(cmd3) {
				return hasVertices
			}
			ctx.CubicCurveTo(x1, y1, x2, y2, x3, y3)
			hasVertices = true
		case agglib.IsPathEndPoly(cmd):
			if agglib.IsPathClose(cmd) {
				ctx.ClosePath()
			}
		}
	}
}

func drawFreeTypeOutlineText(ctx *aggSurface, text *agglib.FreeTypeOutlineText, x, y float64) bool {
	if ctx == nil || text == nil {
		return false
	}

	// Match KnobMan's low-level AGG path: anchor the baseline on integer pixels
	// before extracting the outline path.
	text.SetStartPoint(float64(int(x)), float64(int(y)))

	return appendFreeTypeOutlineText(ctx, text)
}
