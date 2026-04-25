package agg

import (
	"strings"

	"matplotlib-go/render"
)

func resolveFontPath(family string) string {
	if strings.TrimSpace(family) == "" {
		return ""
	}
	return render.DefaultFontManager().FindFontPath(family)
}
