package test

import (
	"image"
	_ "image/png"
	"os"
	"testing"
)

// TestContourLinesParity compares the middle subplot (Tripcolor+Tricontour)
// between the Go golden and the matplotlib reference.
func TestContourLinesParity(t *testing.T) {
	goldenPath := "../testdata/golden/unstructured_showcase.png"
	refPath := "../testdata/matplotlib_ref/unstructured_showcase.png"

	golden := mustLoadImage(t, goldenPath)
	ref := mustLoadImage(t, refPath)

	// Middle subplot: figure fractions 0.37-0.63 (x), 0.16-0.88 (y)
	// On 1320x520: x≈488-832, y≈83-458
	// Scan for dark near-black pixels (contour lines: ~RGB 20,31,46)
	t.Log("=== Go Golden — contour lines region (x=488..832, y=200..400) ===")
	scanDarkRows(t, golden, 488, 832, 200, 400)

	t.Log("=== Matplotlib Ref — contour lines region (x=488..832, y=200..400) ===")
	scanDarkRows(t, ref, 488, 832, 200, 400)
}

func mustLoadImage(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}

func scanDarkRows(t *testing.T, img image.Image, x0, x1, y0, y1 int) {
	t.Helper()
	b := img.Bounds()
	darkPixels := 0
	for y := max2(y0, b.Min.Y); y < min2(y1, b.Max.Y); y++ {
		for x := max2(x0, b.Min.X); x < min2(x1, b.Max.X); x++ {
			r, g, bv, _ := img.At(x, y).RGBA()
			// Dark pixel: all channels < 30% (contour line color is ~(20,31,46))
			if r < 0x5000 && g < 0x5000 && bv < 0x5000 {
				darkPixels++
			}
		}
	}
	t.Logf("  Dark pixels in region: %d", darkPixels)

	// Count rows that have any dark pixel
	rowsWithDark := 0
	for y := max2(y0, b.Min.Y); y < min2(y1, b.Max.Y); y++ {
		for x := max2(x0, b.Min.X); x < min2(x1, b.Max.X); x++ {
			r, g, bv, _ := img.At(x, y).RGBA()
			if r < 0x5000 && g < 0x5000 && bv < 0x5000 {
				rowsWithDark++
				break
			}
		}
	}
	t.Logf("  Rows with dark pixels: %d / %d", rowsWithDark, y1-y0)
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestContourLinesParity2 does a wider scan to find where the dark pixels actually are
func TestContourLinesParity2(t *testing.T) {
	goldenPath := "../testdata/golden/unstructured_showcase.png"
	refPath := "../testdata/matplotlib_ref/unstructured_showcase.png"

	golden := mustLoadImage(t, goldenPath)
	ref := mustLoadImage(t, refPath)

	// Middle subplot full region: x=488..832, y=83..458
	t.Log("=== Full middle subplot ===")
	t.Log("--- Go Golden ---")
	scanDarkRows(t, golden, 488, 832, 83, 458)
	t.Log("--- Matplotlib Ref ---")
	scanDarkRows(t, ref, 488, 832, 83, 458)

	// Bottom half
	t.Log("=== Bottom half y=300..458 ===")
	t.Log("--- Go Golden ---")
	scanDarkRows(t, golden, 488, 832, 300, 458)
	t.Log("--- Matplotlib Ref ---")
	scanDarkRows(t, ref, 488, 832, 300, 458)
}

// TestContourNoLabelsPixels checks dark pixel count without LabelLines
func TestContourNoLabelsPixels(t *testing.T) {
	import_path := "../testdata/golden/unstructured_showcase.png"
	t.Log("Current golden (with inline labels):")
	golden := mustLoadImage(t, import_path)
	scanDarkRows(t, golden, 488, 832, 83, 458)

	import_path_ref := "../testdata/matplotlib_ref/unstructured_showcase.png"
	t.Log("Matplotlib reference (with inline labels):")
	ref := mustLoadImage(t, import_path_ref)
	scanDarkRows(t, ref, 488, 832, 83, 458)
}

// TestContourPixelThresholds compares with different brightness thresholds
func TestContourPixelThresholds(t *testing.T) {
	goldenPath := "../testdata/golden/unstructured_showcase.png"
	refPath := "../testdata/matplotlib_ref/unstructured_showcase.png"

	golden := mustLoadImage(t, goldenPath)
	ref := mustLoadImage(t, refPath)

	for _, threshold := range []uint32{0x3000, 0x5000, 0x7000, 0x9000} {
		goCount := countDarkPixels(golden, 488, 832, 83, 458, threshold)
		refCount := countDarkPixels(ref, 488, 832, 83, 458, threshold)
		t.Logf("Threshold %.0f%%: Go=%d, Ref=%d, ratio=%.2f", float64(threshold)/0xFFFF*100, goCount, refCount, float64(goCount)/float64(refCount))
	}
}

func countDarkPixels(img image.Image, x0, x1, y0, y1 int, threshold uint32) int {
	b := img.Bounds()
	count := 0
	for y := max2(y0, b.Min.Y); y < min2(y1, b.Max.Y); y++ {
		for x := max2(x0, b.Min.X); x < min2(x1, b.Max.X); x++ {
			r, g, bv, _ := img.At(x, y).RGBA()
			if r < threshold && g < threshold && bv < threshold {
				count++
			}
		}
	}
	return count
}

// TestContourHorizontalBands scans horizontal bands to find where dark pixels are
func TestContourHorizontalBands(t *testing.T) {
	goldenPath := "../testdata/golden/unstructured_showcase.png"
	refPath := "../testdata/matplotlib_ref/unstructured_showcase.png"

	golden := mustLoadImage(t, goldenPath)
	ref := mustLoadImage(t, refPath)

	// Scan the middle subplot in 50-row bands, at the strict 19% threshold
	for y := 83; y < 458; y += 50 {
		yEnd := y + 50
		if yEnd > 458 {
			yEnd = 458
		}
		goCount := countDarkPixels(golden, 488, 832, y, yEnd, 0x3000)
		refCount := countDarkPixels(ref, 488, 832, y, yEnd, 0x3000)
		t.Logf("y=%d..%d: Go=%d, Ref=%d, ratio=%.2f", y, yEnd, goCount, refCount, float64(goCount)/float64(refCount+1))
	}
}
