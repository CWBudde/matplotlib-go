package render

import (
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"matplotlib-go/internal/geom"
)

func TestTextPathBuildsValidGlyphOutlines(t *testing.T) {
	withDejaVuFontManager(t)

	textPath, ok := TextPath("Ag", geom.Pt{X: 10, Y: 20}, 14, "DejaVu Sans")
	if !ok {
		t.Fatal("TextPath returned !ok")
	}
	if len(textPath.C) == 0 || len(textPath.V) == 0 {
		t.Fatalf("expected path commands and vertices, got %+v", textPath)
	}
	if !textPath.Validate() {
		t.Fatalf("invalid path: %+v", textPath)
	}
}

func TestFontManagerResolveTextRunsUsesRequestedFace(t *testing.T) {
	manager := withDejaVuFontManager(t)

	runs, ok := manager.ResolveTextRuns("ABC", "DejaVu Sans")
	if !ok || len(runs) != 1 {
		t.Fatalf("ResolveTextRuns = %+v, %v; want one run", runs, ok)
	}
	if runs[0].Text != "ABC" || runs[0].Face.Path == "" {
		t.Fatalf("unexpected resolved run: %+v", runs[0])
	}
}

func TestTextPathSkipsMissingFontsAndEmptyInputs(t *testing.T) {
	if path, ok := TextPath("", geom.Pt{}, 12, "missing-font"); ok || len(path.C) != 0 {
		t.Fatalf("empty text should not build a path: %+v, %v", path, ok)
	}
	if path, ok := TextPath("A", geom.Pt{}, 0, "missing-font"); ok || len(path.C) != 0 {
		t.Fatalf("zero size should not build a path: %+v, %v", path, ok)
	}
	if path, ok := TextPath("A", geom.Pt{}, 12, filepath.Join(t.TempDir(), "missing.ttf")); ok || len(path.C) != 0 {
		t.Fatalf("missing font should not build a path: %+v, %v", path, ok)
	}
}

func withDejaVuFontManager(t *testing.T) *FontManager {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "DejaVuSans.ttf")
	if err := os.WriteFile(path, dejavusans.TTF, 0o644); err != nil {
		t.Fatalf("write test font: %v", err)
	}

	manager := NewFontManager()
	manager.AddFontDir(dir)
	previous := defaultFontManager
	defaultFontManager = manager
	t.Cleanup(func() { defaultFontManager = previous })
	return manager
}
