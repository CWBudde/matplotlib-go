package test

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/imagecmp"
)

const (
	textStrictTolerance   = 1
	textStrictMinPSNR     = 48.0
	textStrictMaxMeanAbs  = 1.25
	titleStrictTolerance  = 1
	titleStrictMinPSNR    = 46.5
	titleStrictMaxMeanAbs = 2.00
)

func TestTextLabelsStrict_MatplotlibRef(t *testing.T) {
	requireOptionalVisualTests(t)
	got := renderTextLabelsStrict()
	wantPath := filepath.Join("..", "testdata", "matplotlib_ref", "text_labels_strict.png")
	want, err := imagecmp.LoadPNG(wantPath)
	if err != nil {
		t.Fatalf("load matplotlib reference %s: %v", wantPath, err)
	}

	artifactsDir := filepath.Join("..", "testdata", "_artifacts", "text_labels_strict")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("create artifacts directory %s: %v", artifactsDir, err)
	}
	savePNGOrFail(t, got, filepath.Join(artifactsDir, "rendered.png"))
	savePNGOrFail(t, want, filepath.Join(artifactsDir, "matplotlib_ref.png"))

	diff, err := imagecmp.ComparePNG(got, want, textStrictTolerance)
	if err != nil {
		t.Fatalf("compare rendered text labels against matplotlib ref: %v", err)
	}
	if diff.PSNR < textStrictMinPSNR || diff.MeanAbs > textStrictMaxMeanAbs {
		if err := imagecmp.SaveDiffImage(got, want, textStrictTolerance, filepath.Join(artifactsDir, "diff.png")); err != nil {
			t.Fatalf("save diff image: %v", err)
		}
		t.Fatalf(
			"text labels mismatch: MaxDiff=%d, MeanAbs=%.2f (max %.2f), PSNR=%.2fdB (min %.2f)",
			diff.MaxDiff,
			diff.MeanAbs,
			textStrictMaxMeanAbs,
			diff.PSNR,
			textStrictMinPSNR,
		)
	}
}

func TestTitleStrict_MatplotlibRef(t *testing.T) {
	requireOptionalVisualTests(t)
	got := renderTitleStrict()
	wantPath := filepath.Join("..", "testdata", "matplotlib_ref", "title_strict.png")
	want, err := imagecmp.LoadPNG(wantPath)
	if err != nil {
		t.Fatalf("load matplotlib reference %s: %v", wantPath, err)
	}

	artifactsDir := filepath.Join("..", "testdata", "_artifacts", "title_strict")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("create artifacts directory %s: %v", artifactsDir, err)
	}
	savePNGOrFail(t, got, filepath.Join(artifactsDir, "rendered.png"))
	savePNGOrFail(t, want, filepath.Join(artifactsDir, "matplotlib_ref.png"))

	diff, err := imagecmp.ComparePNG(got, want, titleStrictTolerance)
	if err != nil {
		t.Fatalf("compare rendered title against matplotlib ref: %v", err)
	}
	if diff.PSNR < titleStrictMinPSNR || diff.MeanAbs > titleStrictMaxMeanAbs {
		if err := imagecmp.SaveDiffImage(got, want, titleStrictTolerance, filepath.Join(artifactsDir, "diff.png")); err != nil {
			t.Fatalf("save diff image: %v", err)
		}
		t.Fatalf(
			"title mismatch: MaxDiff=%d, MeanAbs=%.2f (max %.2f), PSNR=%.2fdB (min %.2f)",
			diff.MaxDiff,
			diff.MeanAbs,
			titleStrictMaxMeanAbs,
			diff.PSNR,
			titleStrictMinPSNR,
		)
	}
}

var _ image.Image
