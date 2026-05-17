// Package svgcompare provides structural parsing, normalization, and diffing
// for SVG documents produced by the matplotlib-go SVG backend. It exists so
// golden tests can compare SVG output by structure rather than by raw bytes:
// two SVGs that differ only in attribute order or insignificant whitespace are
// reported as equivalent.
//
// The package intentionally does not try to be a general-purpose SVG diff
// tool; it targets the specific output shape produced by backends/svg and is
// kept dependency-free so it can ship as an internal helper.
package svgcompare

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// Node is a parsed, normalized SVG element. Attributes are stored in sorted
// key order so two structurally-equivalent SVGs share identical Node trees.
// Text holds element character data with leading/trailing whitespace trimmed
// (and empty strings dropped). Children are kept in document order because
// document order is semantically meaningful in SVG (later elements paint over
// earlier ones).
type Node struct {
	Name     string
	Attrs    []Attr
	Text     string
	Children []*Node
}

// Attr is one (key, value) attribute on a Node.
type Attr struct {
	Key   string
	Value string
}

// Parse returns the root Node parsed from raw SVG source. Comments,
// processing instructions, and XML namespace prefixes are preserved as part of
// the attribute set (e.g. "xmlns:xlink") so namespace drift surfaces in diffs.
func Parse(src []byte) (*Node, error) {
	decoder := xml.NewDecoder(strings.NewReader(string(src)))
	var stack []*Node
	var root *Node

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("svgcompare: parse: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			node := &Node{
				Name:  qname(t.Name),
				Attrs: normalizeAttrs(t.Attr),
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else {
				root = node
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("svgcompare: parse: unexpected </%s>", qname(t.Name))
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			trimmed := strings.TrimSpace(string(t))
			if trimmed == "" {
				continue
			}
			top := stack[len(stack)-1]
			if top.Text == "" {
				top.Text = trimmed
			} else {
				top.Text += " " + trimmed
			}
		}
	}

	if len(stack) != 0 {
		return nil, fmt.Errorf("svgcompare: parse: unterminated element <%s>", stack[len(stack)-1].Name)
	}
	if root == nil {
		return nil, fmt.Errorf("svgcompare: parse: no root element")
	}
	return root, nil
}

func qname(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	// SVG output uses xmlns="..." so the decoder fills Space with the URI of
	// the default namespace, which would otherwise drown the comparison in
	// long URI prefixes. Drop the URI for unprefixed names; we only care that
	// the local name matches.
	return name.Local
}

func normalizeAttrs(in []xml.Attr) []Attr {
	out := make([]Attr, 0, len(in))
	for _, a := range in {
		key := a.Name.Local
		if a.Name.Space != "" {
			// Prefix attributes from a namespace (e.g. xlink:href) keep their
			// prefix-like form so they remain distinguishable from unprefixed
			// versions; we approximate the prefix from the URI's last colon
			// segment when the decoder has stripped it.
			key = namespacePrefix(a.Name.Space) + ":" + a.Name.Local
		}
		out = append(out, Attr{Key: key, Value: a.Value})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

func namespacePrefix(uri string) string {
	switch uri {
	case "http://www.w3.org/1999/xlink":
		return "xlink"
	case "http://www.w3.org/2000/svg":
		return ""
	case "http://www.w3.org/XML/1998/namespace":
		return "xml"
	}
	// Fall back to the last path segment so unknown namespaces still produce
	// a stable, comparable prefix rather than a giant URI.
	if i := strings.LastIndexAny(uri, "/#"); i >= 0 && i < len(uri)-1 {
		return uri[i+1:]
	}
	return uri
}

// Equal reports whether two parsed SVG trees are structurally equal: same
// element names, same children in the same order, same attribute keys with
// the same values, and same trimmed text content per element.
func Equal(a, b *Node) bool {
	return diffNode(a, b, "") == ""
}

// Diff returns a human-readable description of the first structural
// difference between expected and actual, or the empty string if they match.
// The output is formatted as "path: reason" with an XPath-like location so
// failing tests can jump straight to the offending element.
func Diff(expected, actual *Node) string {
	return diffNode(expected, actual, "")
}

func diffNode(want, got *Node, path string) string {
	if want == nil && got == nil {
		return ""
	}
	if want == nil {
		return fmt.Sprintf("%s: unexpected element <%s>", path, got.Name)
	}
	if got == nil {
		return fmt.Sprintf("%s: missing element <%s>", path, want.Name)
	}
	loc := path + "/" + want.Name
	if want.Name != got.Name {
		return fmt.Sprintf("%s: element name differs: want <%s>, got <%s>", path, want.Name, got.Name)
	}
	if msg := diffAttrs(want.Attrs, got.Attrs, loc); msg != "" {
		return msg
	}
	if want.Text != got.Text {
		return fmt.Sprintf("%s: text differs:\n  want: %q\n  got:  %q", loc, want.Text, got.Text)
	}
	if len(want.Children) != len(got.Children) {
		return fmt.Sprintf("%s: child count differs: want %d, got %d", loc, len(want.Children), len(got.Children))
	}
	for i := range want.Children {
		childPath := fmt.Sprintf("%s[%d]", loc, i)
		if msg := diffNode(want.Children[i], got.Children[i], childPath); msg != "" {
			return msg
		}
	}
	return ""
}

func diffAttrs(want, got []Attr, loc string) string {
	wantMap := make(map[string]string, len(want))
	for _, a := range want {
		wantMap[a.Key] = a.Value
	}
	gotMap := make(map[string]string, len(got))
	for _, a := range got {
		gotMap[a.Key] = a.Value
	}
	for _, a := range want {
		v, ok := gotMap[a.Key]
		if !ok {
			return fmt.Sprintf("%s: missing attribute %q (want %q)", loc, a.Key, a.Value)
		}
		if v != a.Value {
			return fmt.Sprintf("%s: attribute %q differs: want %q, got %q", loc, a.Key, a.Value, v)
		}
	}
	for _, a := range got {
		if _, ok := wantMap[a.Key]; !ok {
			return fmt.Sprintf("%s: unexpected attribute %q (=%q)", loc, a.Key, a.Value)
		}
	}
	return ""
}

// ParseAndDiff is a convenience wrapper that parses both inputs and returns
// the structural diff, or the empty string when they match.
func ParseAndDiff(expected, actual []byte) (string, error) {
	want, err := Parse(expected)
	if err != nil {
		return "", fmt.Errorf("parse expected: %w", err)
	}
	got, err := Parse(actual)
	if err != nil {
		return "", fmt.Errorf("parse actual: %w", err)
	}
	return Diff(want, got), nil
}
