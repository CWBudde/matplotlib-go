package core

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

var displayTextCommandReplacer = strings.NewReplacer(
	`\\alpha`, "α",
	`\\beta`, "β",
	`\\gamma`, "γ",
	`\\delta`, "δ",
	`\\epsilon`, "ε",
	`\\eta`, "η",
	`\\theta`, "θ",
	`\\lambda`, "λ",
	`\\mu`, "μ",
	`\\nu`, "ν",
	`\\pi`, "π",
	`\\rho`, "ρ",
	`\\sigma`, "σ",
	`\\tau`, "τ",
	`\\phi`, "φ",
	`\\chi`, "χ",
	`\\psi`, "ψ",
	`\\omega`, "ω",
	`\\Gamma`, "Γ",
	`\\Delta`, "Δ",
	`\\Theta`, "Θ",
	`\\Lambda`, "Λ",
	`\\Pi`, "Π",
	`\\Sigma`, "Σ",
	`\\Phi`, "Φ",
	`\\Psi`, "Ψ",
	`\\Omega`, "Ω",
	`\\pm`, "±",
	`\\mp`, "∓",
	`\\times`, "×",
	`\\cdot`, "·",
	`\\deg`, "°",
	`\\le`, "≤",
	`\\ge`, "≥",
	`\\neq`, "≠",
	`\\approx`, "≈",
	`\\infty`, "∞",
	`\\partial`, "∂",
	`\\nabla`, "∇",
	`\\rightarrow`, "→",
	`\\leftarrow`, "←",
	`\\leftrightarrow`, "↔",
)

var mathTextCommandMap = map[string]string{
	"alpha":          "α",
	"beta":           "β",
	"gamma":          "γ",
	"delta":          "δ",
	"epsilon":        "ε",
	"varepsilon":     "ϵ",
	"eta":            "η",
	"theta":          "θ",
	"vartheta":       "ϑ",
	"lambda":         "λ",
	"mu":             "μ",
	"nu":             "ν",
	"pi":             "π",
	"rho":            "ρ",
	"sigma":          "σ",
	"tau":            "τ",
	"phi":            "φ",
	"varphi":         "φ",
	"chi":            "χ",
	"psi":            "ψ",
	"omega":          "ω",
	"Gamma":          "Γ",
	"Delta":          "Δ",
	"Theta":          "Θ",
	"Lambda":         "Λ",
	"Pi":             "Π",
	"Sigma":          "Σ",
	"Phi":            "Φ",
	"Psi":            "Ψ",
	"Omega":          "Ω",
	"pm":             "±",
	"mp":             "∓",
	"times":          "×",
	"cdot":           "·",
	"div":            "÷",
	"ast":            "∗",
	"circ":           "∘",
	"bullet":         "•",
	"deg":            "°",
	"le":             "≤",
	"leq":            "≤",
	"ge":             "≥",
	"geq":            "≥",
	"ne":             "≠",
	"neq":            "≠",
	"approx":         "≈",
	"equiv":          "≡",
	"propto":         "∝",
	"sim":            "∼",
	"infty":          "∞",
	"partial":        "∂",
	"nabla":          "∇",
	"ell":            "ℓ",
	"sum":            "∑",
	"prod":           "∏",
	"int":            "∫",
	"oint":           "∮",
	"forall":         "∀",
	"exists":         "∃",
	"in":             "∈",
	"notin":          "∉",
	"subset":         "⊂",
	"subseteq":       "⊆",
	"supset":         "⊃",
	"supseteq":       "⊇",
	"cup":            "∪",
	"cap":            "∩",
	"land":           "∧",
	"lor":            "∨",
	"oplus":          "⊕",
	"otimes":         "⊗",
	"to":             "→",
	"rightarrow":     "→",
	"leftarrow":      "←",
	"leftrightarrow": "↔",
	"ldots":          "…",
}

var mathTextOperatorMap = map[string]string{
	"arccos": "arccos",
	"arcsin": "arcsin",
	"arctan": "arctan",
	"arg":    "arg",
	"cos":    "cos",
	"cosh":   "cosh",
	"cot":    "cot",
	"coth":   "coth",
	"csc":    "csc",
	"deg":    "deg",
	"det":    "det",
	"dim":    "dim",
	"exp":    "exp",
	"gcd":    "gcd",
	"hom":    "hom",
	"inf":    "inf",
	"ker":    "ker",
	"lg":     "lg",
	"lim":    "lim",
	"liminf": "lim inf",
	"limsup": "lim sup",
	"ln":     "ln",
	"log":    "log",
	"max":    "max",
	"min":    "min",
	"Pr":     "Pr",
	"sec":    "sec",
	"sin":    "sin",
	"sinh":   "sinh",
	"sup":    "sup",
	"tan":    "tan",
	"tanh":   "tanh",
}

var mathTextAccentMarks = map[string]rune{
	"bar":      '\u0305',
	"overline": '\u0305',
	"hat":      '\u0302',
	"tilde":    '\u0303',
	"vec":      '\u20d7',
	"dot":      '\u0307',
}

var mathTextPassthroughCommands = map[string]struct{}{
	"mathrm":       {},
	"mathit":       {},
	"mathbf":       {},
	"mathsf":       {},
	"mathtt":       {},
	"text":         {},
	"operatorname": {},
}

var mathTextEmptyCommands = map[string]struct{}{
	"left":              {},
	"middle":            {},
	"right":             {},
	"limits":            {},
	"nolimits":          {},
	"displaystyle":      {},
	"textstyle":         {},
	"scriptstyle":       {},
	"scriptscriptstyle": {},
}

var mathTextSpacingCommands = map[string]string{
	",":     " ",
	":":     " ",
	";":     " ",
	"quad":  "  ",
	"qquad": "    ",
}

var mathTextDelimiterCommands = map[string]string{
	"langle": "⟨",
	"rangle": "⟩",
	"lbrace": "{",
	"rbrace": "}",
	"lvert":  "|",
	"rvert":  "|",
	"lVert":  "‖",
	"rVert":  "‖",
	"lfloor": "⌊",
	"rfloor": "⌋",
	"lceil":  "⌈",
	"rceil":  "⌉",
}

var superscriptRunes = map[rune]string{
	'0': "⁰",
	'1': "¹",
	'2': "²",
	'3': "³",
	'4': "⁴",
	'5': "⁵",
	'6': "⁶",
	'7': "⁷",
	'8': "⁸",
	'9': "⁹",
	'+': "⁺",
	'-': "⁻",
	'=': "⁼",
	'(': "⁽",
	')': "⁾",
	'n': "ⁿ",
	'i': "ⁱ",
	'a': "ᵃ",
	'b': "ᵇ",
	'c': "ᶜ",
	'd': "ᵈ",
	'e': "ᵉ",
	'f': "ᶠ",
	'g': "ᵍ",
	'h': "ʰ",
	'j': "ʲ",
	'k': "ᵏ",
	'l': "ˡ",
	'm': "ᵐ",
	'o': "ᵒ",
	'p': "ᵖ",
	'r': "ʳ",
	's': "ˢ",
	't': "ᵗ",
	'u': "ᵘ",
	'v': "ᵛ",
	'w': "ʷ",
	'x': "ˣ",
	'y': "ʸ",
	'z': "ᶻ",
}

var subscriptRunes = map[rune]string{
	'0': "₀",
	'1': "₁",
	'2': "₂",
	'3': "₃",
	'4': "₄",
	'5': "₅",
	'6': "₆",
	'7': "₇",
	'8': "₈",
	'9': "₉",
	'+': "₊",
	'-': "₋",
	'=': "₌",
	'(': "₍",
	')': "₎",
	'a': "ₐ",
	'e': "ₑ",
	'h': "ₕ",
	'i': "ᵢ",
	'j': "ⱼ",
	'k': "ₖ",
	'l': "ₗ",
	'm': "ₘ",
	'n': "ₙ",
	'o': "ₒ",
	'p': "ₚ",
	'r': "ᵣ",
	's': "ₛ",
	't': "ₜ",
	'u': "ᵤ",
	'v': "ᵥ",
	'x': "ₓ",
	'β': "ᵦ",
	'γ': "ᵧ",
	'ρ': "ᵨ",
	'φ': "ᵩ",
	'χ': "ᵪ",
}

type mathTextParser struct {
	input []rune
	pos   int
}

type displayTextSegment struct {
	text   string
	isMath bool
}

func normalizeDisplayText(text string) string {
	if text == "" {
		return ""
	}

	runes := []rune(text)
	var out strings.Builder
	var segment strings.Builder
	inMath := false

	flushPlain := func() {
		if segment.Len() == 0 {
			return
		}
		out.WriteString(displayTextCommandReplacer.Replace(segment.String()))
		segment.Reset()
	}

	flushMath := func() {
		out.WriteString(normalizeMathText(segment.String()))
		segment.Reset()
	}

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) && runes[i+1] == '$' {
			segment.WriteRune('$')
			i++
			continue
		}
		if runes[i] == '$' {
			if inMath {
				flushMath()
			} else {
				flushPlain()
			}
			inMath = !inMath
			continue
		}
		segment.WriteRune(runes[i])
	}

	if inMath {
		out.WriteRune('$')
		out.WriteString(displayTextCommandReplacer.Replace(segment.String()))
		return out.String()
	}

	flushPlain()
	return out.String()
}

func normalizeMathText(text string) string {
	text = strings.ReplaceAll(text, `\\`, `\`)
	parser := mathTextParser{input: []rune(text)}
	return parser.parseUntil(0)
}

func splitDisplayTextSegments(text string) ([]displayTextSegment, bool, bool) {
	if text == "" {
		return nil, false, true
	}

	runes := []rune(text)
	segments := []displayTextSegment{}
	var segment strings.Builder
	inMath := false
	hasMath := false

	flush := func() {
		if segment.Len() == 0 {
			return
		}
		segments = append(segments, displayTextSegment{
			text:   segment.String(),
			isMath: inMath,
		})
		segment.Reset()
	}

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) && runes[i+1] == '$' {
			segment.WriteRune('$')
			i++
			continue
		}
		if runes[i] == '$' {
			flush()
			inMath = !inMath
			if inMath {
				hasMath = true
			}
			continue
		}
		segment.WriteRune(runes[i])
	}

	if inMath {
		return nil, false, false
	}

	flush()
	return segments, hasMath, true
}

func fullMathExpression(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) < 2 || runes[0] != '$' || runes[len(runes)-1] != '$' {
		return "", false
	}
	for i := 1; i < len(runes)-1; i++ {
		if runes[i] == '$' && runes[i-1] != '\\' {
			return "", false
		}
	}
	expr := strings.TrimSpace(string(runes[1 : len(runes)-1]))
	if expr == "" {
		return "", false
	}
	return expr, true
}

func displayTextIsEmpty(text string) bool {
	return normalizeDisplayText(text) == ""
}

func drawDisplayText(textRen render.TextDrawer, text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if ren, ok := textRen.(render.Renderer); ok {
		if layout, ok := layoutDisplayText(ren, text, size, fontKey); ok {
			drawMathTextLayout(ren, textRen, layout, origin, textColor, fontKey)
			return
		}
	}
	display := normalizeDisplayText(text)
	if display == "" {
		return
	}
	primeTextFont(textRen, display, size, fontKey)
	textRen.DrawText(display, origin, size, textColor)
}

func drawDisplayTextRotated(textRen render.RotatedTextDrawer, text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if expr, ok := fullMathExpression(text); ok {
		if ren, ok := textRen.(render.Renderer); ok {
			if layout, ok := LayoutMathText(ren, expr, size, fontKey); ok {
				if drawMathTextLayoutRotated(ren, layout, anchor, angle, textColor, fontKey) {
					return
				}
			}
		}
	}
	display := normalizeDisplayText(text)
	if display == "" {
		return
	}
	primeTextFont(textRen, display, size, fontKey)
	textRen.DrawTextRotated(display, anchor, size, angle, textColor)
}

func drawDisplayTextVertical(textRen render.VerticalTextDrawer, text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
	if expr, ok := fullMathExpression(text); ok {
		if ren, ok := textRen.(render.Renderer); ok {
			if layout, ok := LayoutMathText(ren, expr, size, fontKey); ok {
				if drawMathTextLayoutVertical(ren, layout, center, textColor, fontKey) {
					return
				}
			}
		}
	}
	display := normalizeDisplayText(text)
	if display == "" {
		return
	}
	primeTextFont(textRen, display, size, fontKey)
	textRen.DrawTextVertical(display, center, size, textColor)
}

func primeTextFont(textRen render.TextDrawer, sample string, size float64, fontKey string) {
	if fontKey == "" {
		return
	}
	if ren, ok := textRen.(render.Renderer); ok {
		_ = ren.MeasureText(sample, size, fontKey)
	}
}

func drawMathTextLayout(r render.Renderer, textRen render.TextDrawer, layout MathTextLayout, origin geom.Pt, textColor render.Color, fontKey string) {
	for _, rule := range layout.Rules {
		rect := geom.Rect{
			Min: geom.Pt{X: origin.X + rule.Rect.Min.X, Y: origin.Y + rule.Rect.Min.Y},
			Max: geom.Pt{X: origin.X + rule.Rect.Max.X, Y: origin.Y + rule.Rect.Max.Y},
		}
		r.Path(pixelRectPath(rect), &render.Paint{Fill: textColor})
	}
	for _, run := range layout.Runs {
		runFontKey := resolveRunFontKey(run, fontKey)
		primeTextFont(textRen, run.Text, run.FontSize, runFontKey)
		textRen.DrawText(run.Text, geom.Pt{X: origin.X + run.Offset.X, Y: origin.Y + run.Offset.Y}, run.FontSize, textColor)
	}
}

func drawMathTextLayoutRotated(r render.Renderer, layout MathTextLayout, anchor geom.Pt, angle float64, textColor render.Color, fontKey string) bool {
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	origin := geom.Pt{
		X: anchor.X - layout.Width/2,
		Y: anchor.Y - layout.Descent,
	}
	return drawMathTextLayoutPathTransformed(r, layout, origin, anchor, angle, textColor, fontKey)
}

func drawMathTextLayoutVertical(r render.Renderer, layout MathTextLayout, center geom.Pt, textColor render.Color, fontKey string) bool {
	origin := alignedSingleLineOrigin(center, singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{
			Width:   layout.Width,
			Ascent:  layout.Ascent,
			Descent: layout.Descent,
			Height:  layout.Height,
		},
	}, TextAlignCenter, textLayoutVAlignCenter)
	return drawMathTextLayoutPathTransformed(r, layout, origin, center, math.Pi/2, textColor, fontKey)
}

func drawMathTextLayoutPathTransformed(r render.Renderer, layout MathTextLayout, origin geom.Pt, pivot geom.Pt, angle float64, textColor render.Color, fontKey string) bool {
	paths, ok := mathTextLayoutPaths(r, layout, origin, fontKey)
	if !ok {
		return false
	}
	if angle == 0 {
		for _, path := range paths {
			r.Path(path, &render.Paint{Fill: textColor})
		}
		return true
	}

	cos := math.Cos(angle)
	sin := math.Sin(angle)
	affine := translateAffine(pivot).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffine(geom.Pt{X: -pivot.X, Y: -pivot.Y}))
	for _, path := range paths {
		r.Path(applyAffinePath(path, affine), &render.Paint{Fill: textColor})
	}
	return true
}

func mathTextLayoutPaths(r render.Renderer, layout MathTextLayout, origin geom.Pt, fontKey string) ([]geom.Path, bool) {
	paths := make([]geom.Path, 0, len(layout.Rules)+len(layout.Runs))
	for _, rule := range layout.Rules {
		rect := geom.Rect{
			Min: geom.Pt{X: origin.X + rule.Rect.Min.X, Y: origin.Y + rule.Rect.Min.Y},
			Max: geom.Pt{X: origin.X + rule.Rect.Max.X, Y: origin.Y + rule.Rect.Max.Y},
		}
		paths = append(paths, pixelRectPath(rect))
	}
	for _, run := range layout.Runs {
		runFontKey := resolveRunFontKey(run, fontKey)
		runPath, ok := mathTextRunPath(r, run.Text, geom.Pt{X: origin.X + run.Offset.X, Y: origin.Y + run.Offset.Y}, run.FontSize, runFontKey)
		if !ok {
			return nil, false
		}
		paths = append(paths, runPath)
	}
	return paths, true
}

func mathTextRunPath(r render.Renderer, text string, origin geom.Pt, size float64, fontKey string) (geom.Path, bool) {
	if pather, ok := r.(render.TextPather); ok {
		if path, ok := pather.TextPath(text, origin, size, fontKey); ok {
			return path, true
		}
	}
	return render.TextPath(text, origin, size, fontKey)
}

func resolveRunFontKey(run MathTextLayoutRun, fallback string) string {
	if strings.TrimSpace(run.FontKey) != "" {
		return run.FontKey
	}
	return fallback
}

func layoutDisplayText(r render.Renderer, text string, size float64, fontKey string) (MathTextLayout, bool) {
	if expr, ok := fullMathExpression(text); ok {
		return LayoutMathText(r, expr, size, fontKey)
	}

	segments, hasMath, ok := splitDisplayTextSegments(text)
	if !ok || !hasMath {
		return MathTextLayout{}, false
	}

	var out mathLayoutBox
	x := 0.0
	for _, segment := range segments {
		var child mathLayoutBox
		if segment.isMath {
			layout, ok := LayoutMathText(r, segment.text, size, fontKey)
			if !ok {
				return MathTextLayout{}, false
			}
			child = mathLayoutBox{
				runs:    append([]MathTextLayoutRun(nil), layout.Runs...),
				rules:   append([]MathTextLayoutRule(nil), layout.Rules...),
				Width:   layout.Width,
				Ascent:  layout.Ascent,
				Descent: layout.Descent,
			}
		} else {
			child = layoutMathTextRun(r, displayTextCommandReplacer.Replace(segment.text), size, fontKey)
		}
		if child.Width <= 0 && len(child.runs) == 0 && len(child.rules) == 0 {
			continue
		}
		out.appendTranslated(child, x, 0)
		x += child.Width
		out.Ascent = maxFloat64(out.Ascent, child.Ascent)
		out.Descent = maxFloat64(out.Descent, child.Descent)
	}

	if x <= 0 && len(out.runs) == 0 && len(out.rules) == 0 {
		return MathTextLayout{}, false
	}

	return MathTextLayout{
		Runs:    out.runs,
		Rules:   out.rules,
		Width:   x,
		Ascent:  out.Ascent,
		Descent: out.Descent,
		Height:  out.Ascent + out.Descent,
	}, true
}

func (p *mathTextParser) parseUntil(stop rune) string {
	var out strings.Builder
	for p.pos < len(p.input) {
		r := p.input[p.pos]
		if stop != 0 && r == stop {
			break
		}
		switch r {
		case '{':
			p.pos++
			out.WriteString(p.parseUntil('}'))
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
		case '}':
			if stop == 0 {
				p.pos++
				continue
			}
			return out.String()
		case '^':
			p.pos++
			out.WriteString(convertMathScript(p.parseArgument(), superscriptRunes, "^"))
		case '_':
			p.pos++
			out.WriteString(convertMathScript(p.parseArgument(), subscriptRunes, "_"))
		case '\\':
			out.WriteString(p.parseCommand())
		case '~':
			out.WriteRune(' ')
			p.pos++
		default:
			out.WriteRune(r)
			p.pos++
		}
	}
	return out.String()
}

func (p *mathTextParser) parseArgument() string {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return ""
	}

	switch p.input[p.pos] {
	case '{':
		p.pos++
		arg := p.parseUntil('}')
		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
		}
		return arg
	case '\\':
		return p.parseCommand()
	default:
		r := p.input[p.pos]
		p.pos++
		return string(r)
	}
}

func (p *mathTextParser) parseCommand() string {
	p.pos++
	if p.pos >= len(p.input) {
		return `\`
	}

	r := p.input[p.pos]
	if !unicode.IsLetter(r) {
		p.pos++
		switch r {
		case ',', ';', ':', ' ':
			return " "
		case '!':
			return ""
		default:
			return string(r)
		}
	}

	start := p.pos
	for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
		p.pos++
	}
	name := string(p.input[start:p.pos])

	if mapped, ok := mathTextCommandMap[name]; ok {
		return mapped
	}
	if spacing, ok := mathTextSpacingCommands[name]; ok {
		return spacing
	}
	if delim, ok := mathTextDelimiterCommands[name]; ok {
		return delim
	}
	if op, ok := mathTextOperatorMap[name]; ok {
		return op
	}
	if _, ok := mathTextPassthroughCommands[name]; ok {
		return p.parseArgument()
	}
	if mark, ok := mathTextAccentMarks[name]; ok {
		return applyMathAccent(p.parseArgument(), mark)
	}
	if name == "begin" {
		return p.parseEnvironment()
	}
	if _, ok := mathTextEmptyCommands[name]; ok {
		p.skipSpaces()
		return ""
	}

	switch name {
	case "frac":
		return formatMathFraction(p.parseArgument(), p.parseArgument())
	case "sqrt":
		index := ""
		if p.pos < len(p.input) && p.input[p.pos] == '[' {
			p.pos++
			index = p.parseUntil(']')
			if p.pos < len(p.input) && p.input[p.pos] == ']' {
				p.pos++
			}
		}
		arg := p.parseArgument()
		if arg == "" {
			return "√"
		}
		if index != "" {
			return "√[" + index + "]" + groupMathAtom(arg)
		}
		return "√" + groupMathAtom(arg)
	default:
		return `\` + name
	}
}

func (p *mathTextParser) parseEnvironment() string {
	name := p.parseBraceText()
	left, right, ok := matrixEnvironmentDelimiters(name)
	if !ok {
		return ""
	}
	if name == "array" && p.pos < len(p.input) && p.input[p.pos] == '{' {
		_ = p.parseBraceText()
	}

	rows := p.parseMatrixRows(name)
	var out strings.Builder
	if left != "" {
		out.WriteString(left)
	}
	for i, row := range rows {
		if i > 0 {
			out.WriteString("; ")
		}
		out.WriteString(strings.Join(row, " "))
	}
	if right != "" {
		out.WriteString(right)
	}
	return out.String()
}

func (p *mathTextParser) parseMatrixRows(envName string) [][]string {
	rows := [][]string{}
	for {
		if p.startsEnvironmentEnd(envName) {
			p.consumeEnvironmentEnd(envName)
			break
		}
		row := []string{}
		for {
			cell := strings.TrimSpace(p.parseMatrixCell(envName))
			row = append(row, cell)
			if p.startsEnvironmentEnd(envName) {
				p.consumeEnvironmentEnd(envName)
				rows = append(rows, row)
				return rows
			}
			if p.consumeMatrixRowSeparator() {
				rows = append(rows, row)
				break
			}
			if p.pos < len(p.input) && p.input[p.pos] == '&' {
				p.pos++
				continue
			}
			rows = append(rows, row)
			return rows
		}
	}
	return rows
}

func (p *mathTextParser) parseMatrixCell(envName string) string {
	var out strings.Builder
	for p.pos < len(p.input) {
		if p.startsEnvironmentEnd(envName) || p.pos < len(p.input) && p.input[p.pos] == '&' || p.startsMatrixRowSeparator() {
			break
		}
		switch p.input[p.pos] {
		case '{':
			p.pos++
			out.WriteString(p.parseUntil('}'))
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
		case '^':
			p.pos++
			out.WriteString(convertMathScript(p.parseArgument(), superscriptRunes, "^"))
		case '_':
			p.pos++
			out.WriteString(convertMathScript(p.parseArgument(), subscriptRunes, "_"))
		case '\\':
			out.WriteString(p.parseCommand())
		case '~':
			out.WriteRune(' ')
			p.pos++
		default:
			out.WriteRune(p.input[p.pos])
			p.pos++
		}
	}
	return out.String()
}

func (p *mathTextParser) parseBraceText() string {
	p.skipSpace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return ""
	}
	p.pos++
	start := p.pos
	depth := 1
	for p.pos < len(p.input) {
		switch p.input[p.pos] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				text := string(p.input[start:p.pos])
				p.pos++
				return text
			}
		}
		p.pos++
	}
	return string(p.input[start:])
}

func (p *mathTextParser) startsNamedCommand(name string) bool {
	if p.pos >= len(p.input)-len(name) || p.input[p.pos] != '\\' {
		return false
	}
	i := p.pos + 1
	for _, want := range name {
		if i >= len(p.input) || p.input[i] != want {
			return false
		}
		i++
	}
	return i >= len(p.input) || !unicode.IsLetter(p.input[i])
}

func (p *mathTextParser) consumeNamedCommand(name string) bool {
	if !p.startsNamedCommand(name) {
		return false
	}
	p.pos += 1 + len([]rune(name))
	return true
}

func (p *mathTextParser) startsEnvironmentEnd(name string) bool {
	save := p.pos
	if !p.consumeNamedCommand("end") {
		p.pos = save
		return false
	}
	text := p.parseBraceText()
	p.pos = save
	return text == name
}

func (p *mathTextParser) consumeEnvironmentEnd(name string) bool {
	save := p.pos
	if !p.consumeNamedCommand("end") {
		p.pos = save
		return false
	}
	if p.parseBraceText() != name {
		p.pos = save
		return false
	}
	return true
}

func (p *mathTextParser) startsMatrixRowSeparator() bool {
	if p.pos+1 >= len(p.input) || p.input[p.pos] != '\\' {
		return false
	}
	if p.input[p.pos+1] == '\\' {
		return true
	}
	if !unicode.IsLetter(p.input[p.pos+1]) {
		return false
	}
	i := p.pos + 1
	for i < len(p.input) && unicode.IsLetter(p.input[i]) {
		i++
	}
	return string(p.input[p.pos+1:i]) == "cr"
}

func (p *mathTextParser) consumeMatrixRowSeparator() bool {
	if !p.startsMatrixRowSeparator() {
		return false
	}
	if p.input[p.pos+1] == '\\' {
		p.pos += 2
	} else {
		p.pos++
		for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
			p.pos++
		}
	}
	return true
}

func (p *mathTextParser) skipSpaces() {
	for p.pos < len(p.input) && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}

func applyMathAccent(text string, mark rune) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var out strings.Builder
	for _, r := range text {
		out.WriteRune(r)
		if !unicode.IsSpace(r) {
			out.WriteRune(mark)
		}
	}
	return out.String()
}

func (p *mathTextParser) skipSpace() {
	for p.pos < len(p.input) && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}

func convertMathScript(text string, table map[rune]string, marker string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	var out strings.Builder
	for _, r := range trimmed {
		mapped, ok := table[r]
		if !ok {
			if utf8.RuneCountInString(trimmed) <= 1 {
				return marker + trimmed
			}
			return marker + "(" + trimmed + ")"
		}
		out.WriteString(mapped)
	}
	return out.String()
}

func formatMathFraction(num, den string) string {
	if num == "" {
		num = "?"
	}
	if den == "" {
		den = "?"
	}
	return groupMathAtom(num) + "⁄" + groupMathAtom(den)
}

func groupMathAtom(text string) string {
	if !needsMathGrouping(text) {
		return text
	}
	return "(" + text + ")"
}

func needsMathGrouping(text string) bool {
	if utf8.RuneCountInString(text) <= 1 {
		return false
	}
	for _, r := range text {
		if unicode.IsSpace(r) || strings.ContainsRune("+-=⁄<>", r) {
			return true
		}
	}
	return false
}
