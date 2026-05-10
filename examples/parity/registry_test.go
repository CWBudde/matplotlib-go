package parity

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogHasCanonicalSourcePaths(t *testing.T) {
	cases := Cases()
	if len(cases) < 80 {
		t.Fatalf("len(Cases()) = %d, want at least 80 parity examples", len(cases))
	}

	for _, c := range cases {
		if c.ID == "" {
			t.Fatal("case has empty ID")
		}
		if c.GoSourcePath == "" {
			t.Fatalf("%s has empty GoSourcePath", c.ID)
		}
		if c.PythonSourcePath == "" {
			t.Fatalf("%s has empty PythonSourcePath", c.ID)
		}
		requireParityFile(t, c.GoSourcePath)
		requireParityFile(t, c.PythonSourcePath)
	}
}

func TestRenderKnownParityExample(t *testing.T) {
	img, c, err := Render("lognorm_imshow")
	if err != nil {
		t.Fatalf("Render(lognorm_imshow) error = %v", err)
	}
	if c.ID != "lognorm_imshow" {
		t.Fatalf("case ID = %q, want lognorm_imshow", c.ID)
	}
	assertNonBlankImage(t, img)
}

func TestRenderRejectsUnknownParityExample(t *testing.T) {
	if _, _, err := Render("nope"); err == nil {
		t.Fatal("Render(nope) returned nil error")
	}
}

func requireParityFile(t *testing.T, rel string) {
	t.Helper()
	path := filepath.Join(repoRootForParityTest(t), rel)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing %s: %v", rel, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", rel)
	}
}

func assertNonBlankImage(t *testing.T, img image.Image) {
	t.Helper()
	if img == nil {
		t.Fatal("image is nil")
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		t.Fatalf("image bounds are empty: %v", bounds)
	}
	first := img.At(bounds.Min.X, bounds.Min.Y)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.At(x, y) != first {
				return
			}
		}
	}
	t.Fatal("image is blank")
}

func repoRootForParityTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}
