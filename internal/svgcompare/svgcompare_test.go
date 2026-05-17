package svgcompare

import (
	"strings"
	"testing"
)

func TestEqualIgnoresAttributeOrder(t *testing.T) {
	a := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect x="1" y="2" width="3" height="4" /></svg>`)
	b := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect height="4" width="3" y="2" x="1" /></svg>`)

	want, err := Parse(a)
	if err != nil {
		t.Fatalf("parse a: %v", err)
	}
	got, err := Parse(b)
	if err != nil {
		t.Fatalf("parse b: %v", err)
	}
	if !Equal(want, got) {
		t.Fatalf("attribute-order-only difference should compare equal, got diff: %s", Diff(want, got))
	}
}

func TestEqualIgnoresInsignificantWhitespace(t *testing.T) {
	a := []byte("<svg><g><rect x=\"1\" y=\"2\" /></g></svg>")
	b := []byte("<svg>\n  <g>\n    <rect x=\"1\" y=\"2\" />\n  </g>\n</svg>")

	diff, err := ParseAndDiff(a, b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if diff != "" {
		t.Fatalf("inter-element whitespace should not produce a diff, got: %s", diff)
	}
}

func TestDiffReportsAttributeValueMismatch(t *testing.T) {
	a := []byte(`<svg><rect x="1" y="2" /></svg>`)
	b := []byte(`<svg><rect x="1" y="3" /></svg>`)

	diff, err := ParseAndDiff(a, b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if diff == "" {
		t.Fatal("expected a diff for differing attribute values")
	}
	if !strings.Contains(diff, "/rect") || !strings.Contains(diff, `"y"`) || !strings.Contains(diff, `"2"`) || !strings.Contains(diff, `"3"`) {
		t.Fatalf("diff should locate offending element and values, got: %s", diff)
	}
}

func TestDiffReportsMissingAttribute(t *testing.T) {
	a := []byte(`<svg><rect x="1" y="2" stroke="red" /></svg>`)
	b := []byte(`<svg><rect x="1" y="2" /></svg>`)

	diff, err := ParseAndDiff(a, b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(diff, "missing attribute") || !strings.Contains(diff, "stroke") {
		t.Fatalf("diff should report missing attribute by name, got: %s", diff)
	}
}

func TestDiffReportsUnexpectedAttribute(t *testing.T) {
	a := []byte(`<svg><rect x="1" y="2" /></svg>`)
	b := []byte(`<svg><rect x="1" y="2" fill="blue" /></svg>`)

	diff, err := ParseAndDiff(a, b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(diff, "unexpected attribute") || !strings.Contains(diff, "fill") {
		t.Fatalf("diff should report unexpected attribute by name, got: %s", diff)
	}
}

func TestDiffReportsChildCountMismatch(t *testing.T) {
	a := []byte(`<svg><g><rect /><rect /></g></svg>`)
	b := []byte(`<svg><g><rect /></g></svg>`)

	diff, err := ParseAndDiff(a, b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(diff, "child count differs") || !strings.Contains(diff, "/g") {
		t.Fatalf("diff should report child count mismatch with path, got: %s", diff)
	}
}

func TestDiffReportsTextMismatch(t *testing.T) {
	a := []byte(`<svg><text>hello</text></svg>`)
	b := []byte(`<svg><text>world</text></svg>`)

	diff, err := ParseAndDiff(a, b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(diff, "text differs") {
		t.Fatalf("diff should report text mismatch, got: %s", diff)
	}
}

func TestParseSupportsXlinkHref(t *testing.T) {
	// xlink:href and href on the same element should round-trip distinctly.
	src := []byte(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"><use xlink:href="#a" href="#a" /></svg>`)
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected one child, got %d", len(root.Children))
	}
	use := root.Children[0]
	if use.Name != "use" {
		t.Fatalf("expected <use>, got <%s>", use.Name)
	}
	var keys []string
	for _, a := range use.Attrs {
		keys = append(keys, a.Key)
	}
	wantKeys := []string{"href", "xlink:href"}
	if len(keys) != len(wantKeys) || keys[0] != wantKeys[0] || keys[1] != wantKeys[1] {
		t.Fatalf("expected sorted attr keys %v, got %v", wantKeys, keys)
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	if _, err := Parse([]byte("<svg><g></svg>")); err == nil {
		t.Fatal("unterminated element should produce a parse error")
	}
}
