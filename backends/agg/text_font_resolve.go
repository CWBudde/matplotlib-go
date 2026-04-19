package agg

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var resolvedFontPathCache sync.Map

const (
	fontFamilySansSerif = "sans-serif"
	fontFamilySerif     = "serif"
	fontFamilyMonospace = "monospace"
)

func resolveFontPath(family string) string {
	normalized := normalizeFontFamily(family)
	if normalized == "" {
		return ""
	}

	if cached, ok := resolvedFontPathCache.Load(normalized); ok {
		if path, ok := cached.(string); ok {
			return path
		}
	}

	path := findFontPath(normalized)
	resolvedFontPathCache.Store(normalized, path)
	return path
}

func normalizeFontFamily(family string) string {
	family = strings.TrimSpace(family)
	switch strings.ToLower(strings.ReplaceAll(family, " ", "")) {
	case "":
		return ""
	case "dejavusans":
		return "DejaVu Sans"
	case "dejavuserif":
		return "DejaVu Serif"
	case "dejavusansmono":
		return "DejaVu Sans Mono"
	case "sansserif":
		return fontFamilySansSerif
	case "serif":
		return fontFamilySerif
	case "monospace", "monospaced":
		return fontFamilyMonospace
	default:
		return family
	}
}

func findFontPath(family string) string {
	if filepath.IsAbs(family) {
		if _, err := os.Stat(family); err == nil {
			return family
		}
		return ""
	}

	for _, candidate := range candidateFontFamilies(family) {
		if path := resolveFontWithFCMatchExact(candidate); path != "" {
			return path
		}
	}

	for _, path := range fallbackFontPaths(family) {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func candidateFontFamilies(family string) []string {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case fontFamilySansSerif:
		// Mirror Matplotlib's default sans-serif preference order.
		return []string{
			"DejaVu Sans",
			"Bitstream Vera Sans",
			"Computer Modern Sans Serif",
			"Lucida Grande",
			"Verdana",
			"Geneva",
			"Lucid",
			"Arial",
			"Helvetica",
			"Avant Garde",
		}
	case fontFamilySerif:
		return []string{
			"DejaVu Serif",
			"Bitstream Vera Serif",
			"Computer Modern Roman",
			"New Century Schoolbook",
			"Century Schoolbook L",
			"Utopia",
			"ITC Bookman",
			"Bookman",
			"Nimbus Roman No9 L",
			"Times New Roman",
			"Times",
			"Palatino",
			"Charter",
		}
	case fontFamilyMonospace:
		return []string{
			"DejaVu Sans Mono",
			"Bitstream Vera Sans Mono",
			"Computer Modern Typewriter",
			"Andale Mono",
			"Nimbus Mono L",
			"Courier New",
			"Courier",
			"Fixed",
			"Terminal",
		}
	default:
		return []string{family}
	}
}

func resolveFontWithFCMatchExact(family string) string {
	if _, err := exec.LookPath("fc-match"); err != nil {
		return ""
	}

	for _, pattern := range []string{
		family + ":style=Regular",
		family + ":style=Book",
		family + ":style=Roman",
		family,
	} {
		out, err := exec.Command("fc-match", "-f", "%{family}\n%{file}\n", pattern).Output()
		if err != nil {
			continue
		}

		families, path := parseFCMatchOutput(string(out))
		if path == "" || !familyListMatchesRequested(families, family) {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func parseFCMatchOutput(out string) ([]string, string) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return nil, ""
	}

	path := strings.TrimSpace(lines[len(lines)-1])
	if len(lines) == 1 {
		return nil, path
	}

	var families []string
	for _, family := range strings.Split(lines[0], ",") {
		name := strings.TrimSpace(family)
		if name != "" {
			families = append(families, name)
		}
	}

	return families, path
}

func familyListMatchesRequested(families []string, requested string) bool {
	for _, family := range families {
		if strings.EqualFold(strings.TrimSpace(family), strings.TrimSpace(requested)) {
			return true
		}
	}
	return false
}

func fallbackFontPaths(family string) []string {
	switch runtime.GOOS {
	case "darwin":
		switch strings.ToLower(strings.TrimSpace(family)) {
		case fontFamilySerif:
			return []string{
				"/System/Library/Fonts/Supplemental/Times New Roman.ttf",
				"/System/Library/Fonts/NewYork.ttf",
			}
		case fontFamilyMonospace:
			return []string{
				"/System/Library/Fonts/Menlo.ttc",
				"/System/Library/Fonts/Courier.dfont",
			}
		default:
			return []string{
				"/System/Library/Fonts/Helvetica.ttc",
				"/System/Library/Fonts/Supplemental/Arial.ttf",
				"/System/Library/Fonts/Supplemental/DejaVuSans.ttf",
			}
		}
	case "windows":
		switch strings.ToLower(strings.TrimSpace(family)) {
		case fontFamilySerif:
			return []string{`C:\Windows\Fonts\times.ttf`}
		case fontFamilyMonospace:
			return []string{`C:\Windows\Fonts\consola.ttf`, `C:\Windows\Fonts\cour.ttf`}
		default:
			return []string{`C:\Windows\Fonts\DejaVuSans.ttf`, `C:\Windows\Fonts\arial.ttf`}
		}
	default:
		switch strings.ToLower(strings.TrimSpace(family)) {
		case fontFamilySerif:
			return []string{
				"/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf",
				"/usr/share/fonts/truetype/liberation2/LiberationSerif-Regular.ttf",
			}
		case fontFamilyMonospace:
			return []string{
				"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
				"/usr/share/fonts/truetype/liberation2/LiberationMono-Regular.ttf",
			}
		default:
			return []string{
				"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
				"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
				"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf",
			}
		}
	}
}
