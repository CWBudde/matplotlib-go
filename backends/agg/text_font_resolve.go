package agg

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/render"
)

func resolveFontPath(family string) string {
	if strings.TrimSpace(family) == "" {
		return ""
	}
	return render.DefaultFontManager().FindFontPath(family)
}
