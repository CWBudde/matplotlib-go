package backends_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/agg" // register AGG backend
	_ "github.com/cwbudde/matplotlib-go/backends/svg" // register SVG backend
)

func TestRegistry_SaveViaExtension_PNG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.png")

	r, err := backends.Create(backends.AGG, backends.Config{Width: 100, Height: 80, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := backends.DefaultRegistry.SaveViaExtension(backends.AGG, r, path); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("expected non-empty PNG file at %s, got 0 bytes", path)
	}
}

func TestRegistry_SaveViaExtension_SVG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.svg")

	r, err := backends.Create(backends.SVG, backends.Config{Width: 100, Height: 80, DPI: 72})
	if err != nil {
		t.Fatalf("Create SVG: %v", err)
	}
	if err := backends.DefaultRegistry.SaveViaExtension(backends.SVG, r, path); err != nil {
		t.Fatalf("SaveViaExtension: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("expected non-empty SVG file at %s, got 0 bytes", path)
	}
}
