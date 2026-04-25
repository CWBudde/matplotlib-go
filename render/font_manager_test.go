package render

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFontPropertiesHandlesListsAndPaths(t *testing.T) {
	props := ParseFontProperties(`["DejaVu Sans", "Arial", sans-serif]`)
	if len(props.Families) != 3 {
		t.Fatalf("family count = %d, want 3 (%+v)", len(props.Families), props)
	}
	if props.Families[0] != "DejaVu Sans" || props.Families[2] != "sans-serif" {
		t.Fatalf("families = %#v", props.Families)
	}

	path := filepath.Join(t.TempDir(), "ExampleFont.ttf")
	if err := os.WriteFile(path, []byte("not a real font"), 0o644); err != nil {
		t.Fatalf("write font placeholder: %v", err)
	}
	props = ParseFontProperties(path)
	if props.File != path || len(props.Families) != 0 {
		t.Fatalf("path properties = %+v", props)
	}
}

func TestFontManagerResolvesDirectFontFileAndCaches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ExampleFont.ttf")
	if err := os.WriteFile(path, []byte("not a real font"), 0o644); err != nil {
		t.Fatalf("write font placeholder: %v", err)
	}

	manager := NewFontManager()
	face, ok := manager.FindFont(FontProperties{File: path})
	if !ok || face.Path != path {
		t.Fatalf("FindFont direct path = %+v, %v; want %q", face, ok, path)
	}

	if got := manager.FindFontPath(path); got != path {
		t.Fatalf("FindFontPath direct path = %q, want %q", got, path)
	}
}

func TestFontManagerScansAddedDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Example Sans.ttf")
	if err := os.WriteFile(path, []byte("not a real font"), 0o644); err != nil {
		t.Fatalf("write font placeholder: %v", err)
	}

	manager := NewFontManager()
	manager.AddFontDir(dir)
	face, ok := manager.FindFont(FontProperties{Families: []string{"Example Sans"}})
	if !ok || face.Path != path {
		t.Fatalf("FindFont scanned path = %+v, %v; want %q", face, ok, path)
	}
}

func TestCSSFontFamilyVariants(t *testing.T) {
	tests := map[string]string{
		"serif":       "DejaVu Serif, serif",
		"sans-serif":  "DejaVu Sans, Arial, sans-serif",
		"monospace":   "DejaVu Sans Mono, monospace",
		"mono_space":  "DejaVu Sans Mono, monospace",
		"custom-font": "DejaVu Sans, Arial, sans-serif",
	}

	for key, want := range tests {
		if got := CSSFontFamily(key); got != want {
			t.Fatalf("CSSFontFamily(%q) = %q, want %q", key, got, want)
		}
	}
}
