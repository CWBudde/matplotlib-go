package agg

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

func resolveFontPath(family string) string {
	face, ok := resolveFontFace(family)
	if !ok {
		return ""
	}
	return face.Path
}

func resolveFontFace(family string) (render.FontFace, bool) {
	if strings.TrimSpace(family) == "" {
		family = "DejaVu Sans"
	}
	return render.DefaultFontManager().FindFont(render.ParseFontProperties(family))
}

func fontReference(face render.FontFace) string {
	if face.Path != "" {
		return face.Path
	}
	return face.Family
}
