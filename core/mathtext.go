package core

import (
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
	if expr, ok := fullMathExpression(text); ok {
		if ren, ok := textRen.(render.Renderer); ok {
			if layout, ok := LayoutMathText(ren, expr, size, fontKey); ok {
				drawMathTextLayout(ren, textRen, layout, origin, textColor, fontKey)
				return
			}
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
	display := normalizeDisplayText(text)
	if display == "" {
		return
	}
	primeTextFont(textRen, display, size, fontKey)
	textRen.DrawTextRotated(display, anchor, size, angle, textColor)
}

func drawDisplayTextVertical(textRen render.VerticalTextDrawer, text string, center geom.Pt, size float64, textColor render.Color, fontKey string) {
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
		primeTextFont(textRen, run.Text, run.FontSize, fontKey)
		textRen.DrawText(run.Text, geom.Pt{X: origin.X + run.Offset.X, Y: origin.Y + run.Offset.Y}, run.FontSize, textColor)
	}
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
	if op, ok := mathTextOperatorMap[name]; ok {
		return op
	}
	if _, ok := mathTextPassthroughCommands[name]; ok {
		return p.parseArgument()
	}
	if mark, ok := mathTextAccentMarks[name]; ok {
		return applyMathAccent(p.parseArgument(), mark)
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
	case "left", "right":
		return ""
	default:
		return `\` + name
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
