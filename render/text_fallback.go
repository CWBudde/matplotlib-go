package render

import "unicode/utf8"

// FontRun is a contiguous text run resolved to one font face.
type FontRun struct {
	Text string
	Face FontFace
}

// ResolveTextRuns resolves text into contiguous font runs. The first resolved
// face is always the requested face; generic families are used only when that
// face cannot represent a rune.
func (m *FontManager) ResolveTextRuns(text string, fontKey string) ([]FontRun, bool) {
	if m == nil || text == "" {
		return nil, false
	}

	primary, ok := m.FindFont(ParseFontProperties(fontKey))
	if !ok && fontKey == "" {
		primary, ok = m.FindFont(ParseFontProperties("DejaVu Sans"))
	}
	if !ok {
		return nil, false
	}

	fallbacks := m.fallbackFaces(primary)
	var runs []FontRun
	var current FontFace
	var currentText []rune

	flush := func() {
		if len(currentText) == 0 {
			return
		}
		runs = append(runs, FontRun{
			Text: string(currentText),
			Face: current,
		})
		currentText = currentText[:0]
	}

	for len(text) > 0 {
		r, n := utf8.DecodeRuneInString(text)
		text = text[n:]

		face := primary
		if !fontFaceSupportsRune(primary, r) {
			face = firstFaceSupportingRune(fallbacks, r)
			if face.Path == "" {
				face = primary
			}
		}

		if current.Path != "" && face.Path != current.Path {
			flush()
		}
		current = face
		currentText = append(currentText, r)
	}
	flush()

	return runs, len(runs) > 0
}

func (m *FontManager) fallbackFaces(primary FontFace) []FontFace {
	var faces []FontFace
	seen := map[string]struct{}{primary.Path: {}}
	for _, family := range []string{fontFamilySansSerif, fontFamilySerif, fontFamilyMonospace} {
		face, ok := m.FindFont(FontProperties{Families: []string{family}})
		if !ok || face.Path == "" {
			continue
		}
		if _, ok := seen[face.Path]; ok {
			continue
		}
		seen[face.Path] = struct{}{}
		faces = append(faces, face)
	}
	return faces
}

func firstFaceSupportingRune(faces []FontFace, r rune) FontFace {
	for _, face := range faces {
		if fontFaceSupportsRune(face, r) {
			return face
		}
	}
	return FontFace{}
}
