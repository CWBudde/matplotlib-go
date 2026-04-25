package core

import (
	"strings"
	"unicode"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// MathTextLayoutRun is one text draw in a laid-out MathText expression.
type MathTextLayoutRun struct {
	Text     string
	Offset   geom.Pt
	FontSize float64
}

// MathTextLayoutRule is a filled rule, such as a fraction bar or root vinculum.
type MathTextLayoutRule struct {
	Rect geom.Rect
}

// MathTextLayout is a lightweight layout tree flattened into draw runs and
// rules. Offsets and rectangles are relative to the expression baseline.
type MathTextLayout struct {
	Runs    []MathTextLayoutRun
	Rules   []MathTextLayoutRule
	Width   float64
	Ascent  float64
	Descent float64
	Height  float64
}

type mathLayoutKind uint8

const (
	mathLayoutList mathLayoutKind = iota
	mathLayoutText
	mathLayoutScript
	mathLayoutFrac
	mathLayoutSqrt
)

type mathLayoutNode struct {
	kind     mathLayoutKind
	text     string
	children []mathLayoutNode
	base     *mathLayoutNode
	super    *mathLayoutNode
	sub      *mathLayoutNode
	num      *mathLayoutNode
	den      *mathLayoutNode
	radicand *mathLayoutNode
	index    *mathLayoutNode
}

type mathLayoutParser struct {
	input []rune
	pos   int
}

// LayoutMathText parses and lays out one MathText expression without requiring
// dollar delimiters. It supports the same fallback command set as displayed
// text normalization plus baseline-shifted scripts, stacked fractions, and
// square-root vincula.
func LayoutMathText(r render.Renderer, expr string, size float64, fontKey string) (MathTextLayout, bool) {
	if r == nil || strings.TrimSpace(expr) == "" || size <= 0 {
		return MathTextLayout{}, false
	}
	expr = strings.ReplaceAll(expr, `\\`, `\`)
	parser := mathLayoutParser{input: []rune(expr)}
	node := parser.parseUntil(0)
	box := layoutMathNode(r, node, size, fontKey)
	if box.Width <= 0 && len(box.runs) == 0 && len(box.rules) == 0 {
		return MathTextLayout{}, false
	}
	return MathTextLayout{
		Runs:    box.runs,
		Rules:   box.rules,
		Width:   box.Width,
		Ascent:  box.Ascent,
		Descent: box.Descent,
		Height:  box.Ascent + box.Descent,
	}, true
}

func (p *mathLayoutParser) parseUntil(stop rune) mathLayoutNode {
	var children []mathLayoutNode
	appendText := func(text string) {
		if text == "" {
			return
		}
		n := len(children)
		if n > 0 && children[n-1].kind == mathLayoutText {
			children[n-1].text += text
			return
		}
		children = append(children, mathLayoutNode{kind: mathLayoutText, text: text})
	}

	for p.pos < len(p.input) {
		r := p.input[p.pos]
		if stop != 0 && r == stop {
			break
		}
		switch r {
		case '{':
			p.pos++
			children = append(children, p.parseUntil('}').children...)
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
		case '}':
			if stop == 0 {
				p.pos++
				continue
			}
			return mathLayoutNode{kind: mathLayoutList, children: children}
		case '^', '_':
			p.pos++
			children = attachMathScript(children, r, p.parseArgumentNode())
		case '\\':
			node := p.parseCommandNode()
			if node.kind == mathLayoutText {
				appendText(node.text)
			} else if !node.isEmpty() {
				children = append(children, node)
			}
		case '~':
			appendText(" ")
			p.pos++
		default:
			appendText(string(r))
			p.pos++
		}
	}
	return mathLayoutNode{kind: mathLayoutList, children: children}
}

func (p *mathLayoutParser) parseArgumentNode() mathLayoutNode {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return mathLayoutNode{kind: mathLayoutText}
	}
	switch p.input[p.pos] {
	case '{':
		p.pos++
		node := p.parseUntil('}')
		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
		}
		return node
	case '\\':
		return p.parseCommandNode()
	default:
		r := p.input[p.pos]
		p.pos++
		return mathLayoutNode{kind: mathLayoutText, text: string(r)}
	}
}

func (p *mathLayoutParser) parseCommandNode() mathLayoutNode {
	p.pos++
	if p.pos >= len(p.input) {
		return mathLayoutNode{kind: mathLayoutText, text: `\`}
	}
	r := p.input[p.pos]
	if !unicode.IsLetter(r) {
		p.pos++
		switch r {
		case ',', ';', ':', ' ':
			return mathLayoutNode{kind: mathLayoutText, text: " "}
		case '!':
			return mathLayoutNode{kind: mathLayoutText}
		default:
			return mathLayoutNode{kind: mathLayoutText, text: string(r)}
		}
	}

	start := p.pos
	for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
		p.pos++
	}
	name := string(p.input[start:p.pos])

	if mapped, ok := mathTextCommandMap[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: mapped}
	}
	if op, ok := mathTextOperatorMap[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: op}
	}
	if _, ok := mathTextPassthroughCommands[name]; ok {
		return p.parseArgumentNode()
	}
	if mark, ok := mathTextAccentMarks[name]; ok {
		return mathLayoutNode{kind: mathLayoutText, text: applyMathAccent(nodePlainText(p.parseArgumentNode()), mark)}
	}

	switch name {
	case "frac":
		num := p.parseArgumentNode()
		den := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutFrac, num: &num, den: &den}
	case "sqrt":
		var index *mathLayoutNode
		if p.pos < len(p.input) && p.input[p.pos] == '[' {
			p.pos++
			idx := p.parseUntil(']')
			if p.pos < len(p.input) && p.input[p.pos] == ']' {
				p.pos++
			}
			index = &idx
		}
		radicand := p.parseArgumentNode()
		return mathLayoutNode{kind: mathLayoutSqrt, radicand: &radicand, index: index}
	case "left", "right":
		return mathLayoutNode{kind: mathLayoutText}
	default:
		return mathLayoutNode{kind: mathLayoutText, text: `\` + name}
	}
}

func (p *mathLayoutParser) skipSpace() {
	for p.pos < len(p.input) && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}

func attachMathScript(children []mathLayoutNode, marker rune, script mathLayoutNode) []mathLayoutNode {
	if len(children) == 0 {
		return append(children, mathLayoutNode{kind: mathLayoutText, text: string(marker) + nodePlainText(script)})
	}
	last := children[len(children)-1]
	if last.kind != mathLayoutScript {
		base := last
		last = mathLayoutNode{kind: mathLayoutScript, base: &base}
	}
	if marker == '^' {
		last.super = &script
	} else {
		last.sub = &script
	}
	children[len(children)-1] = last
	return children
}

func (n mathLayoutNode) isEmpty() bool {
	return n.kind == mathLayoutText && n.text == ""
}

func nodePlainText(n mathLayoutNode) string {
	switch n.kind {
	case mathLayoutText:
		return n.text
	case mathLayoutList:
		var out strings.Builder
		for _, child := range n.children {
			out.WriteString(nodePlainText(child))
		}
		return out.String()
	case mathLayoutScript:
		base := nodePlainText(pointerNode(n.base))
		if n.sub != nil {
			base += "_" + nodePlainText(*n.sub)
		}
		if n.super != nil {
			base += "^" + nodePlainText(*n.super)
		}
		return base
	case mathLayoutFrac:
		return formatMathFraction(nodePlainText(pointerNode(n.num)), nodePlainText(pointerNode(n.den)))
	case mathLayoutSqrt:
		return "√" + groupMathAtom(nodePlainText(pointerNode(n.radicand)))
	default:
		return ""
	}
}

func pointerNode(n *mathLayoutNode) mathLayoutNode {
	if n == nil {
		return mathLayoutNode{kind: mathLayoutText}
	}
	return *n
}

type mathLayoutBox struct {
	runs    []MathTextLayoutRun
	rules   []MathTextLayoutRule
	Width   float64
	Ascent  float64
	Descent float64
}

func layoutMathNode(r render.Renderer, n mathLayoutNode, size float64, fontKey string) mathLayoutBox {
	switch n.kind {
	case mathLayoutText:
		return layoutMathTextRun(r, n.text, size, fontKey)
	case mathLayoutList:
		return layoutMathList(r, n.children, size, fontKey)
	case mathLayoutScript:
		return layoutMathScript(r, n, size, fontKey)
	case mathLayoutFrac:
		return layoutMathFrac(r, pointerNode(n.num), pointerNode(n.den), size, fontKey)
	case mathLayoutSqrt:
		return layoutMathSqrt(r, pointerNode(n.radicand), n.index, size, fontKey)
	default:
		return mathLayoutBox{}
	}
}

func layoutMathTextRun(r render.Renderer, text string, size float64, fontKey string) mathLayoutBox {
	if text == "" {
		return mathLayoutBox{}
	}
	metrics := r.MeasureText(text, size, fontKey)
	if metrics.W <= 0 {
		metrics.W = float64(len([]rune(text))) * size * 0.5
	}
	if metrics.Ascent <= 0 && metrics.Descent <= 0 {
		metrics.Ascent = size * 0.8
		metrics.Descent = size * 0.2
	}
	return mathLayoutBox{
		runs:    []MathTextLayoutRun{{Text: text, FontSize: size}},
		Width:   metrics.W,
		Ascent:  metrics.Ascent,
		Descent: metrics.Descent,
	}
}

func layoutMathList(r render.Renderer, children []mathLayoutNode, size float64, fontKey string) mathLayoutBox {
	var out mathLayoutBox
	x := 0.0
	for _, child := range children {
		box := layoutMathNode(r, child, size, fontKey)
		out.appendTranslated(box, x, 0)
		x += box.Width
		if box.Ascent > out.Ascent {
			out.Ascent = box.Ascent
		}
		if box.Descent > out.Descent {
			out.Descent = box.Descent
		}
	}
	out.Width = x
	return out
}

func layoutMathScript(r render.Renderer, n mathLayoutNode, size float64, fontKey string) mathLayoutBox {
	base := layoutMathNode(r, pointerNode(n.base), size, fontKey)
	scriptSize := size * 0.7
	x := base.Width
	var out mathLayoutBox
	out.appendTranslated(base, 0, 0)
	out.Width = base.Width
	out.Ascent = base.Ascent
	out.Descent = base.Descent

	scriptWidth := 0.0
	if n.super != nil {
		super := layoutMathNode(r, *n.super, scriptSize, fontKey)
		y := -maxFloat64(base.Ascent*0.55, scriptSize*0.35)
		out.appendTranslated(super, x, y)
		scriptWidth = maxFloat64(scriptWidth, super.Width)
		out.Ascent = maxFloat64(out.Ascent, -y+super.Ascent)
		out.Descent = maxFloat64(out.Descent, y+super.Descent)
	}
	if n.sub != nil {
		sub := layoutMathNode(r, *n.sub, scriptSize, fontKey)
		y := maxFloat64(base.Descent*0.70, scriptSize*0.25)
		out.appendTranslated(sub, x, y)
		scriptWidth = maxFloat64(scriptWidth, sub.Width)
		out.Ascent = maxFloat64(out.Ascent, -y+sub.Ascent)
		out.Descent = maxFloat64(out.Descent, y+sub.Descent)
	}
	out.Width = base.Width + scriptWidth
	return out
}

func layoutMathFrac(r render.Renderer, num, den mathLayoutNode, size float64, fontKey string) mathLayoutBox {
	childSize := size * 0.75
	numBox := layoutMathNode(r, num, childSize, fontKey)
	denBox := layoutMathNode(r, den, childSize, fontKey)
	padding := size * 0.18
	gap := size * 0.14
	ruleThickness := maxFloat64(size*0.045, 0.5)
	width := maxFloat64(numBox.Width, denBox.Width) + 2*padding
	numX := (width - numBox.Width) / 2
	denX := (width - denBox.Width) / 2
	numY := -(gap + ruleThickness/2 + numBox.Descent)
	denY := gap + ruleThickness/2 + denBox.Ascent

	out := mathLayoutBox{
		Width:   width,
		Ascent:  -numY + numBox.Ascent,
		Descent: denY + denBox.Descent,
		rules: []MathTextLayoutRule{{
			Rect: geom.Rect{
				Min: geom.Pt{X: 0, Y: -ruleThickness / 2},
				Max: geom.Pt{X: width, Y: ruleThickness / 2},
			},
		}},
	}
	out.appendTranslated(numBox, numX, numY)
	out.appendTranslated(denBox, denX, denY)
	return out
}

func layoutMathSqrt(r render.Renderer, radicand mathLayoutNode, index *mathLayoutNode, size float64, fontKey string) mathLayoutBox {
	root := layoutMathTextRun(r, "√", size, fontKey)
	radicandBox := layoutMathNode(r, radicand, size, fontKey)
	padding := size * 0.08
	ruleThickness := maxFloat64(size*0.04, 0.5)
	ruleY := -radicandBox.Ascent - padding

	var out mathLayoutBox
	out.appendTranslated(root, 0, 0)
	out.appendTranslated(radicandBox, root.Width, 0)
	out.Width = root.Width + radicandBox.Width
	out.Ascent = maxFloat64(root.Ascent, -ruleY+ruleThickness)
	out.Descent = maxFloat64(root.Descent, radicandBox.Descent)
	out.rules = append(out.rules, MathTextLayoutRule{
		Rect: geom.Rect{
			Min: geom.Pt{X: root.Width, Y: ruleY},
			Max: geom.Pt{X: out.Width, Y: ruleY + ruleThickness},
		},
	})

	if index != nil {
		indexBox := layoutMathNode(r, *index, size*0.55, fontKey)
		x := -indexBox.Width * 0.65
		y := -root.Ascent * 0.55
		out.appendTranslated(indexBox, x, y)
		out.Ascent = maxFloat64(out.Ascent, -y+indexBox.Ascent)
	}
	return out
}

func (b *mathLayoutBox) appendTranslated(child mathLayoutBox, dx, dy float64) {
	for _, run := range child.runs {
		run.Offset.X += dx
		run.Offset.Y += dy
		b.runs = append(b.runs, run)
	}
	for _, rule := range child.rules {
		rule.Rect.Min.X += dx
		rule.Rect.Max.X += dx
		rule.Rect.Min.Y += dy
		rule.Rect.Max.Y += dy
		b.rules = append(b.rules, rule)
	}
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
