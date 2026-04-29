package render

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	fontFamilySansSerif = "sans-serif"
	fontFamilySerif     = "serif"
	fontFamilyMonospace = "monospace"
)

// FontStyle describes a font posture in Matplotlib-style font properties.
type FontStyle string

const (
	FontStyleNormal  FontStyle = "normal"
	FontStyleItalic  FontStyle = "italic"
	FontStyleOblique FontStyle = "oblique"
)

// FontProperties is the renderer-facing subset of Matplotlib's FontProperties.
type FontProperties struct {
	Families []string
	Style    FontStyle
	Weight   int
	File     string
}

// FontFace describes a discovered font file.
type FontFace struct {
	Path   string
	Family string
	Style  FontStyle
	Weight int
}

// FontManager resolves FontProperties to concrete font files and caches the
// result. It deliberately keeps discovery conservative: exact family matches
// are preferred and generic families use Matplotlib-like fallback lists.
type FontManager struct {
	mu    sync.RWMutex
	cache map[string]FontFace
	dirs  []string
}

var defaultFontManager = NewFontManager()

// NewFontManager returns an empty font manager with system discovery enabled.
func NewFontManager() *FontManager {
	return &FontManager{cache: map[string]FontFace{}}
}

// DefaultFontManager returns the process-wide font manager.
func DefaultFontManager() *FontManager {
	return defaultFontManager
}

// AddFontDir adds a directory to scan before system fallbacks.
func (m *FontManager) AddFontDir(dir string) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirs = append(m.dirs, dir)
	m.cache = map[string]FontFace{}
}

// ParseFontProperties parses the current rc-style font key into properties.
func ParseFontProperties(fontKey string) FontProperties {
	fontKey = strings.TrimSpace(fontKey)
	props := FontProperties{Style: FontStyleNormal, Weight: 400}
	if fontKey == "" {
		return props
	}
	if pathLooksLikeFontFile(fontKey) {
		props.File = fontKey
		return props
	}
	props.Families = parseFontFamilyList(fontKey)
	return props
}

// FindFontPath resolves a font key to a file path.
func (m *FontManager) FindFontPath(fontKey string) string {
	face, ok := m.FindFont(ParseFontProperties(fontKey))
	if !ok {
		return ""
	}
	return face.Path
}

// FindFont resolves font properties to a concrete font face.
func (m *FontManager) FindFont(props FontProperties) (FontFace, bool) {
	props = normalizeFontProperties(props)
	if props.File != "" {
		if path := existingFontPath(props.File); path != "" {
			return FontFace{Path: path, Family: filepath.Base(path), Style: props.Style, Weight: props.Weight}, true
		}
		return FontFace{}, false
	}
	if len(props.Families) == 0 {
		return FontFace{}, false
	}

	key := fontPropertiesCacheKey(props)
	m.mu.RLock()
	if cached, ok := m.cache[key]; ok {
		m.mu.RUnlock()
		return cached, cached.Path != ""
	}
	dirs := append([]string(nil), m.dirs...)
	m.mu.RUnlock()

	face, ok := findFontFace(props, dirs)
	m.mu.Lock()
	if ok {
		m.cache[key] = face
	} else {
		m.cache[key] = FontFace{}
	}
	m.mu.Unlock()
	return face, ok
}

// CSSFontFamily returns an SVG/CSS font-family list for the requested key.
func CSSFontFamily(fontKey string) string {
	families := ParseFontProperties(fontKey).Families
	if len(families) == 0 {
		return "DejaVu Sans, Arial, sans-serif"
	}
	switch normalizeFontFamilyName(families[0]) {
	case fontFamilySerif, "dejavuserif":
		return "DejaVu Serif, serif"
	case fontFamilyMonospace, "dejavusansmono":
		return "DejaVu Sans Mono, monospace"
	case fontFamilySansSerif, "sansserif", "dejavusans":
		return "DejaVu Sans, Arial, sans-serif"
	default:
		return "DejaVu Sans, Arial, sans-serif"
	}
}

func normalizeFontProperties(props FontProperties) FontProperties {
	if props.Style == "" {
		props.Style = FontStyleNormal
	}
	if props.Weight <= 0 {
		props.Weight = 400
	}
	props.File = strings.TrimSpace(props.File)
	props.Families = normalizeFontFamilies(props.Families)
	return props
}

func normalizeFontFamilies(families []string) []string {
	out := make([]string, 0, len(families))
	seen := map[string]struct{}{}
	for _, family := range families {
		normalized := normalizeFontFamily(family)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeFontFamily(family string) string {
	family = strings.TrimSpace(strings.Trim(family, `"'`))
	switch normalizeFontFamilyName(family) {
	case "":
		return ""
	case "default", "sans", "sansserif":
		return fontFamilySansSerif
	case "serif":
		return fontFamilySerif
	case "mono", "monospace", "monospaced":
		return fontFamilyMonospace
	case "dejavusans":
		return "DejaVu Sans"
	case "dejavuserif":
		return "DejaVu Serif"
	case "dejavusansmono":
		return "DejaVu Sans Mono"
	default:
		return family
	}
}

func normalizeFontFamilyName(family string) string {
	family = strings.ToLower(strings.TrimSpace(strings.Trim(family, `"'`)))
	family = strings.ReplaceAll(family, " ", "")
	family = strings.ReplaceAll(family, "_", "")
	return strings.ReplaceAll(family, "-", "")
}

func parseFontFamilyList(fontKey string) []string {
	fontKey = strings.TrimSpace(fontKey)
	fontKey = strings.TrimPrefix(strings.TrimSuffix(fontKey, "]"), "[")
	fontKey = strings.TrimPrefix(strings.TrimSuffix(fontKey, ")"), "(")
	parts := splitFontFamilies(fontKey)
	if len(parts) == 0 {
		return []string{fontKey}
	}
	return parts
}

func splitFontFamilies(value string) []string {
	var parts []string
	var b strings.Builder
	quote := rune(0)
	for _, r := range value {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
				continue
			}
			b.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
		case r == ',':
			if item := strings.TrimSpace(b.String()); item != "" {
				parts = append(parts, item)
			}
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	if item := strings.TrimSpace(b.String()); item != "" {
		parts = append(parts, item)
	}
	return parts
}

func fontPropertiesCacheKey(props FontProperties) string {
	return strings.Join(props.Families, "\x00") + "|" + string(props.Style) + "|" + strconv.Itoa(props.Weight) + "|" + filepath.Clean(props.File)
}

func findFontFace(props FontProperties, dirs []string) (FontFace, bool) {
	for _, family := range props.Families {
		for _, candidate := range candidateFontFamilies(family) {
			if path := findFontInDirs(candidate, dirs); path != "" {
				return FontFace{Path: path, Family: candidate, Style: props.Style, Weight: props.Weight}, true
			}
			if path := resolveFontWithFCMatchExact(candidate, props); path != "" {
				return FontFace{Path: path, Family: candidate, Style: props.Style, Weight: props.Weight}, true
			}
		}
		for _, path := range fallbackFontPaths(family) {
			if path := existingFontPath(path); path != "" {
				return FontFace{Path: path, Family: family, Style: props.Style, Weight: props.Weight}, true
			}
		}
	}
	return FontFace{}, false
}

func findFontInDirs(family string, dirs []string) string {
	needle := normalizeFontFamilyName(family)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !pathLooksLikeFontFile(name) {
				continue
			}
			base := strings.TrimSuffix(name, filepath.Ext(name))
			if normalizeFontFamilyName(base) == needle {
				return filepath.Join(dir, name)
			}
		}
	}
	return ""
}

func candidateFontFamilies(family string) []string {
	switch normalizeFontFamilyName(family) {
	case fontFamilySansSerif, "sans", "sansserif":
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
	case fontFamilyMonospace, "mono", "monospaced":
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

func resolveFontWithFCMatchExact(family string, props FontProperties) string {
	if _, err := exec.LookPath("fc-match"); err != nil {
		return ""
	}
	for _, pattern := range fcMatchPatterns(family, props) {
		out, err := exec.Command("fc-match", "-f", "%{family}\n%{file}\n", pattern).Output()
		if err != nil {
			continue
		}
		families, path := parseFCMatchOutput(string(out))
		if path == "" || !familyListMatchesRequested(families, family) {
			continue
		}
		if path := existingFontPath(path); path != "" {
			return path
		}
	}
	return ""
}

func fcMatchPatterns(family string, props FontProperties) []string {
	style := "Regular"
	switch props.Style {
	case FontStyleItalic:
		style = "Italic"
	case FontStyleOblique:
		style = "Oblique"
	}
	return []string{
		family + ":style=" + style,
		family + ":style=Book",
		family + ":style=Roman",
		family,
	}
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
	requested = normalizeFontFamilyName(requested)
	for _, family := range families {
		if normalizeFontFamilyName(family) == requested {
			return true
		}
	}
	return false
}

func fallbackFontPaths(family string) []string {
	switch runtime.GOOS {
	case "darwin":
		switch normalizeFontFamilyName(family) {
		case fontFamilySerif:
			return []string{
				"/System/Library/Fonts/Supplemental/Times New Roman.ttf",
				"/System/Library/Fonts/NewYork.ttf",
			}
		case fontFamilyMonospace, "mono":
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
		switch normalizeFontFamilyName(family) {
		case fontFamilySerif:
			return []string{`C:\Windows\Fonts\times.ttf`}
		case fontFamilyMonospace, "mono":
			return []string{`C:\Windows\Fonts\consola.ttf`, `C:\Windows\Fonts\cour.ttf`}
		default:
			return []string{`C:\Windows\Fonts\DejaVuSans.ttf`, `C:\Windows\Fonts\arial.ttf`}
		}
	default:
		switch normalizeFontFamilyName(family) {
		case fontFamilySerif:
			return []string{
				"/usr/share/fonts/truetype/dejavu/DejaVuSerif.ttf",
				"/usr/share/fonts/truetype/liberation2/LiberationSerif-Regular.ttf",
			}
		case fontFamilyMonospace, "mono":
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

func existingFontPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func pathLooksLikeFontFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".ttf" || ext == ".otf" || ext == ".ttc" || ext == ".dfont"
}
