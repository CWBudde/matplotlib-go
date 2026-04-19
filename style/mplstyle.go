package style

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"matplotlib-go/color"
	"matplotlib-go/render"
)

// MPLStyleIssue captures one ignored or unsupported rcParam entry.
type MPLStyleIssue struct {
	Line  int
	Key   string
	Value string
}

// MPLStyleReport describes how a .mplstyle file was applied.
type MPLStyleReport struct {
	Applied     []string
	Unsupported []MPLStyleIssue
}

var supportedMPLStyleKeys = []string{
	"axes.edgecolor",
	"axes.facecolor",
	"axes.labelcolor",
	"axes.linewidth",
	"axes.prop_cycle",
	"figure.dpi",
	"figure.facecolor",
	"font.family",
	"font.size",
	"grid.alpha",
	"grid.color",
	"grid.linewidth",
	"grid.major.color",
	"grid.minor.color",
	"legend.edgecolor",
	"legend.facecolor",
	"legend.labelcolor",
	"lines.color",
	"lines.linewidth",
	"text.color",
	"xtick.color",
	"ytick.color",
}

type mplStyleState struct {
	rc RC

	figureFaceValue string
	figureFaceSet   bool
	textColorValue  string
	textColorSet    bool
	lineColorValue  string
	lineColorSet    bool

	axesFaceValue string
	axesFaceSet   bool
	axesEdgeValue string
	axesEdgeSet   bool

	lineWidthPt      float64
	lineWidthSet     bool
	axisLineWidthPt  float64
	axisLineWidthSet bool
	gridLineWidthPt  float64
	gridLineWidthSet bool

	gridColorValue string
	gridColorSet   bool
	gridMajorValue string
	gridMajorSet   bool
	gridMinorValue string
	gridMinorSet   bool
	gridAlpha      float64
	gridAlphaSet   bool

	legendFaceValue string
	legendFaceSet   bool
	legendEdgeValue string
	legendEdgeSet   bool
	legendTextValue string
	legendTextSet   bool
}

// SupportedMPLStyleKeys returns the subset of rcParams understood by the loader.
func SupportedMPLStyleKeys() []string {
	keys := make([]string, len(supportedMPLStyleKeys))
	copy(keys, supportedMPLStyleKeys)
	return keys
}

// ParseMPLStyle parses a Matplotlib .mplstyle payload into a theme.
//
// Only the rcParams returned by SupportedMPLStyleKeys are applied. Unknown keys
// are reported in the returned report and ignored.
func ParseMPLStyle(name, src string) (Theme, MPLStyleReport, error) {
	report := MPLStyleReport{}
	state := mplStyleState{
		rc: Default,
	}

	lines := strings.Split(src, "\n")
	for i, rawLine := range lines {
		lineNo := i + 1
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := splitMPLStyleLine(rawLine)
		if !ok {
			return Theme{}, report, fmt.Errorf("parse .mplstyle line %d: expected key: value", lineNo)
		}

		normalizedKey := normalizeThemeName(key)
		if err := applyMPLStyleEntry(&state, normalizedKey, value, lineNo, &report); err != nil {
			return Theme{}, report, err
		}
	}

	finalizeMPLStyleState(&state)

	return Theme{
		Name: normalizeMPLStyleName(name),
		RC:   Apply(state.rc),
	}, report, nil
}

// LoadMPLStyleFile loads and parses a Matplotlib .mplstyle file.
func LoadMPLStyleFile(path string) (Theme, MPLStyleReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, MPLStyleReport{}, err
	}

	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return ParseMPLStyle(name, string(data))
}

func applyMPLStyleEntry(state *mplStyleState, key, value string, lineNo int, report *MPLStyleReport) error {
	if state == nil || report == nil {
		return errors.New("nil mplstyle state")
	}

	switch key {
	case "figure.dpi":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.DPI = parsed
	case "figure.facecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.figureFaceValue = normalizeMPLValue(value)
		state.figureFaceSet = true
	case "font.family":
		parsed, err := parseMPLFontFamily(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.FontKey = parsed
	case "font.size":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.FontSize = parsed
	case "text.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.textColorValue = normalizeMPLValue(value)
		state.textColorSet = true
	case "lines.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.lineColorValue = normalizeMPLValue(value)
		state.lineColorSet = true
	case "lines.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.lineWidthPt = parsed
		state.lineWidthSet = true
	case "axes.facecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.axesFaceValue = normalizeMPLValue(value)
		state.axesFaceSet = true
	case "axes.edgecolor", "xtick.color", "ytick.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.axesEdgeValue = normalizeMPLValue(value)
		state.axesEdgeSet = true
	case "axes.labelcolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.textColorValue = normalizeMPLValue(value)
		state.textColorSet = true
	case "axes.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.axisLineWidthPt = parsed
		state.axisLineWidthSet = true
	case "axes.prop_cycle":
		parsed, err := parseMPLColorCycle(value, state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.ColorCycle = parsed
	case "grid.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridColorValue = normalizeMPLValue(value)
		state.gridColorSet = true
	case "grid.major.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridMajorValue = normalizeMPLValue(value)
		state.gridMajorSet = true
	case "grid.minor.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridMinorValue = normalizeMPLValue(value)
		state.gridMinorSet = true
	case "grid.alpha":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridAlpha = parsed
		state.gridAlphaSet = true
	case "grid.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridLineWidthPt = parsed
		state.gridLineWidthSet = true
	case "legend.facecolor":
		if err := validateMPLColorValue(value, state.rc, true); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendFaceValue = normalizeMPLValue(value)
		state.legendFaceSet = true
	case "legend.edgecolor":
		if err := validateMPLColorValue(value, state.rc, true); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendEdgeValue = normalizeMPLValue(value)
		state.legendEdgeSet = true
	case "legend.labelcolor":
		if err := validateMPLColorValue(value, state.rc, true); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendTextValue = normalizeMPLValue(value)
		state.legendTextSet = true
	default:
		report.Unsupported = append(report.Unsupported, MPLStyleIssue{
			Line:  lineNo,
			Key:   key,
			Value: strings.TrimSpace(value),
		})
		return nil
	}

	report.Applied = append(report.Applied, key)
	return nil
}

func finalizeMPLStyleState(state *mplStyleState) {
	if state == nil {
		return
	}

	if state.figureFaceSet {
		if parsed, err := parseMPLColor(state.figureFaceValue, state.rc); err == nil {
			state.rc.Background = [4]float64{parsed.R, parsed.G, parsed.B, parsed.A}
		}
	}
	if state.textColorSet {
		if parsed, err := parseMPLColor(state.textColorValue, state.rc); err == nil {
			state.rc.TextColor = [4]float64{parsed.R, parsed.G, parsed.B, parsed.A}
		}
	}
	if state.lineColorSet {
		if parsed, err := parseMPLColor(state.lineColorValue, state.rc); err == nil {
			state.rc.LineColor = [4]float64{parsed.R, parsed.G, parsed.B, parsed.A}
		}
	}
	if state.axesFaceSet {
		if parsed, err := parseMPLColor(state.axesFaceValue, state.rc); err == nil {
			state.rc.AxesBackground = parsed
		}
	}
	if state.axesEdgeSet {
		if parsed, err := parseMPLColor(state.axesEdgeValue, state.rc); err == nil {
			state.rc.AxesEdgeColor = parsed
		}
	}

	if state.lineWidthSet {
		state.rc.LineWidth = mplPointsToPixels(state.lineWidthPt, state.rc.DPI)
	}
	if state.axisLineWidthSet {
		state.rc.AxisLineWidth = mplPointsToPixels(state.axisLineWidthPt, state.rc.DPI)
	}
	if state.gridLineWidthSet {
		width := mplPointsToPixels(state.gridLineWidthPt, state.rc.DPI)
		state.rc.GridLineWidth = width
		state.rc.MinorGridLineWidth = width
	}

	major := state.rc.GridColor
	minor := state.rc.MinorGridColor
	if state.gridColorSet {
		if parsed, err := parseMPLColor(state.gridColorValue, state.rc); err == nil {
			major = parsed
			minor = parsed
		}
	}
	if state.gridMajorSet {
		if parsed, err := parseMPLColor(state.gridMajorValue, state.rc); err == nil {
			major = parsed
		}
	}
	if state.gridMinorSet {
		if parsed, err := parseMPLColor(state.gridMinorValue, state.rc); err == nil {
			minor = parsed
		}
	}
	if state.gridAlphaSet {
		major.A = state.gridAlpha
		minor.A = state.gridAlpha
	}
	state.rc.GridColor = major
	state.rc.MinorGridColor = minor

	if state.legendFaceSet {
		state.rc.LegendBackground = resolveMPLSpecialColor(state.legendFaceValue, state.rc, state.rc.AxesBackground)
	}
	if state.legendEdgeSet {
		state.rc.LegendBorderColor = resolveMPLSpecialColor(state.legendEdgeValue, state.rc, state.rc.AxesEdgeColor)
	}
	if state.legendTextSet {
		state.rc.LegendTextColor = resolveMPLSpecialColor(state.legendTextValue, state.rc, state.rc.DefaultTextColor())
	}
}

func splitMPLStyleLine(raw string) (string, string, bool) {
	noComment := stripMPLStyleComment(raw)
	if strings.TrimSpace(noComment) == "" {
		return "", "", false
	}

	idx := strings.Index(noComment, ":")
	if idx < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(noComment[:idx])
	value := strings.TrimSpace(noComment[idx+1:])
	if key == "" || value == "" {
		return "", "", false
	}
	return key, value, true
}

func stripMPLStyleComment(raw string) string {
	inQuote := rune(0)
	for i, r := range raw {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			inQuote = r
			continue
		}
		if r == '#' {
			if i == 0 {
				return ""
			}
			prev, _ := utf8LastRune(raw[:i])
			if unicode.IsSpace(prev) {
				return strings.TrimRightFunc(raw[:i], unicode.IsSpace)
			}
		}
	}
	return raw
}

func utf8LastRune(s string) (rune, int) {
	return utf8.DecodeLastRuneInString(s)
}

func parseMPLFloat(value string) (float64, error) {
	normalized := normalizeMPLValue(value)
	parsed, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %q", value)
	}
	return parsed, nil
}

func parseMPLFontFamily(value string) (string, error) {
	normalized := normalizeMPLValue(value)
	if normalized == "" {
		return "", errors.New("empty font family")
	}
	if strings.HasPrefix(normalized, "[") && strings.HasSuffix(normalized, "]") {
		items := splitOutsideQuotes(normalized[1:len(normalized)-1], ',')
		for _, item := range items {
			candidate := normalizeMPLValue(item)
			if candidate != "" {
				return candidate, nil
			}
		}
		return "", errors.New("empty font family list")
	}
	items := splitOutsideQuotes(normalized, ',')
	if len(items) > 0 {
		first := normalizeMPLValue(items[0])
		if first != "" {
			return first, nil
		}
	}
	return normalized, nil
}

func parseMPLColor(value string, rc RC) (render.Color, error) {
	normalized := normalizeMPLValue(value)
	if normalized == "" {
		return render.Color{}, errors.New("empty color")
	}

	switch strings.ToLower(normalized) {
	case "none":
		return render.Color{A: 0}, nil
	case "inherit":
		return render.Color{}, errors.New(`special value "inherit" requires contextual handling`)
	}

	if strings.HasPrefix(normalized, "(") && strings.HasSuffix(normalized, ")") {
		return parseMPLColorTuple(normalized)
	}

	if looksLikeMPLHexColor(normalized) {
		return parseMPLHexColor(normalized)
	}

	if grayscale, err := strconv.ParseFloat(normalized, 64); err == nil {
		return render.Color{R: grayscale, G: grayscale, B: grayscale, A: 1}, nil
	}

	if strings.HasPrefix(strings.ToUpper(normalized), "C") && len(normalized) > 1 {
		idx, err := strconv.Atoi(normalized[1:])
		if err == nil {
			palette := rc.Palette()
			if len(palette) == 0 {
				return render.Color{}, fmt.Errorf("color cycle index %q out of range", normalized)
			}
			return palette[idx%len(palette)], nil
		}
	}

	if parsed, ok := mplNamedColors[strings.ToLower(normalized)]; ok {
		return parsed, nil
	}

	return parseMPLHexColor(normalized)
}

func looksLikeMPLHexColor(value string) bool {
	normalized := strings.TrimPrefix(value, "#")
	switch len(normalized) {
	case 3, 4, 6, 8:
	default:
		return false
	}
	for _, r := range normalized {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return false
		}
	}
	return true
}

func parseMPLColorTuple(value string) (render.Color, error) {
	parts := splitOutsideQuotes(strings.TrimSpace(value[1:len(value)-1]), ',')
	if len(parts) != 3 && len(parts) != 4 {
		return render.Color{}, fmt.Errorf("expected RGB or RGBA tuple, got %q", value)
	}

	channels := [4]float64{0, 0, 0, 1}
	for i, part := range parts {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return render.Color{}, fmt.Errorf("invalid tuple component %q", part)
		}
		channels[i] = parsed
	}

	return render.Color{R: channels[0], G: channels[1], B: channels[2], A: channels[3]}, nil
}

func parseMPLHexColor(value string) (render.Color, error) {
	normalized := strings.TrimPrefix(value, "#")
	switch len(normalized) {
	case 3:
		normalized = strings.Repeat(string(normalized[0]), 2) +
			strings.Repeat(string(normalized[1]), 2) +
			strings.Repeat(string(normalized[2]), 2)
	case 4:
		normalized = strings.Repeat(string(normalized[0]), 2) +
			strings.Repeat(string(normalized[1]), 2) +
			strings.Repeat(string(normalized[2]), 2) +
			strings.Repeat(string(normalized[3]), 2)
	case 6, 8:
		// already normalized
	default:
		return render.Color{}, fmt.Errorf("unsupported color %q", value)
	}

	parseByte := func(part string) (float64, error) {
		n, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return 0, err
		}
		return float64(n) / 255.0, nil
	}

	r, err := parseByte(normalized[0:2])
	if err != nil {
		return render.Color{}, fmt.Errorf("invalid color %q", value)
	}
	g, err := parseByte(normalized[2:4])
	if err != nil {
		return render.Color{}, fmt.Errorf("invalid color %q", value)
	}
	b, err := parseByte(normalized[4:6])
	if err != nil {
		return render.Color{}, fmt.Errorf("invalid color %q", value)
	}
	a := 1.0
	if len(normalized) == 8 {
		a, err = parseByte(normalized[6:8])
		if err != nil {
			return render.Color{}, fmt.Errorf("invalid color %q", value)
		}
	}

	return render.Color{R: r, G: g, B: b, A: a}, nil
}

func parseMPLColorCycle(value string, rc RC) (color.Palette, error) {
	normalized := strings.TrimSpace(value)
	lower := strings.ToLower(normalized)
	if !strings.HasPrefix(lower, "cycler(") || !strings.HasSuffix(normalized, ")") {
		return nil, fmt.Errorf("unsupported cycler syntax %q", value)
	}

	inner := strings.TrimSpace(normalized[len("cycler(") : len(normalized)-1])
	if inner == "" {
		return nil, errors.New("empty cycler")
	}

	var rawList string
	switch {
	case strings.HasPrefix(inner, "'color'"), strings.HasPrefix(inner, `"color"`):
		commaIdx := strings.Index(inner, ",")
		if commaIdx < 0 {
			return nil, fmt.Errorf("unsupported cycler syntax %q", value)
		}
		rawList = strings.TrimSpace(inner[commaIdx+1:])
	case strings.HasPrefix(strings.ToLower(inner), "color"):
		eqIdx := strings.Index(inner, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("unsupported cycler syntax %q", value)
		}
		rawList = strings.TrimSpace(inner[eqIdx+1:])
	default:
		return nil, fmt.Errorf("unsupported cycler key in %q", value)
	}

	if !strings.HasPrefix(rawList, "[") || !strings.HasSuffix(rawList, "]") {
		return nil, fmt.Errorf("unsupported color list %q", rawList)
	}

	items := splitOutsideQuotes(rawList[1:len(rawList)-1], ',')
	if len(items) == 0 {
		return nil, errors.New("empty color cycle")
	}

	palette := make(color.Palette, 0, len(items))
	for _, item := range items {
		parsed, err := parseMPLColor(item, rc)
		if err != nil {
			return nil, err
		}
		palette = append(palette, parsed)
	}
	return palette, nil
}

func splitOutsideQuotes(value string, sep rune) []string {
	parts := make([]string, 0, 4)
	var current strings.Builder
	inQuote := rune(0)
	for _, r := range value {
		if inQuote != 0 {
			current.WriteRune(r)
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			inQuote = r
			current.WriteRune(r)
			continue
		}
		if r == sep {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	parts = append(parts, strings.TrimSpace(current.String()))
	return parts
}

func normalizeMPLValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		if (trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') ||
			(trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"') {
			return strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		}
	}
	return trimmed
}

func normalizeMPLStyleName(name string) string {
	normalized := normalizeThemeName(strings.TrimSuffix(name, ".mplstyle"))
	if normalized == "" {
		return "custom"
	}
	return normalized
}

func validateMPLColorValue(value string, rc RC, allowInherit bool) error {
	normalized := normalizeMPLValue(value)
	if allowInherit && strings.EqualFold(normalized, "inherit") {
		return nil
	}
	_, err := parseMPLColor(normalized, rc)
	return err
}

func resolveMPLSpecialColor(value string, rc RC, inherited render.Color) render.Color {
	switch strings.ToLower(value) {
	case "", "inherit":
		return inherited
	default:
		parsed, err := parseMPLColor(value, rc)
		if err != nil {
			return inherited
		}
		return parsed
	}
}

func mplPointsToPixels(points, dpi float64) float64 {
	if dpi <= 0 {
		dpi = Default.DPI
	}
	if dpi <= 0 {
		dpi = 72
	}
	return points * dpi / 72.0
}

var mplNamedColors = func() map[string]render.Color {
	return map[string]render.Color{
		"b":       {R: 0, G: 0, B: 1, A: 1},
		"g":       {R: 0, G: 0.5, B: 0, A: 1},
		"r":       {R: 1, G: 0, B: 0, A: 1},
		"c":       {R: 0, G: 0.75, B: 0.75, A: 1},
		"m":       {R: 0.75, G: 0, B: 0.75, A: 1},
		"y":       {R: 0.75, G: 0.75, B: 0, A: 1},
		"k":       {R: 0, G: 0, B: 0, A: 1},
		"w":       {R: 1, G: 1, B: 1, A: 1},
		"black":   {R: 0, G: 0, B: 0, A: 1},
		"white":   {R: 1, G: 1, B: 1, A: 1},
		"red":     {R: 1, G: 0, B: 0, A: 1},
		"green":   {R: 0, G: 0.5, B: 0, A: 1},
		"blue":    {R: 0, G: 0, B: 1, A: 1},
		"cyan":    {R: 0, G: 1, B: 1, A: 1},
		"magenta": {R: 1, G: 0, B: 1, A: 1},
		"yellow":  {R: 1, G: 1, B: 0, A: 1},
		"grey":    {R: 0.5, G: 0.5, B: 0.5, A: 1},
		"gray":    {R: 0.5, G: 0.5, B: 0.5, A: 1},
		"orange":  {R: 1, G: 0.647, B: 0, A: 1},
		"purple":  {R: 0.5, G: 0, B: 0.5, A: 1},
		"brown":   {R: 0.647, G: 0.165, B: 0.165, A: 1},
		"pink":    {R: 1, G: 0.753, B: 0.796, A: 1},
	}
}()
