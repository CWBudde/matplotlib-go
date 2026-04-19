package webdemo

import "testing"

func TestCatalogStable(t *testing.T) {
	got := Catalog()
	wantIDs := []string{"lines", "scatter", "bars", "histogram", "heatmap", "subplots"}
	if len(got) != len(wantIDs) {
		t.Fatalf("Catalog() len = %d, want %d", len(got), len(wantIDs))
	}

	for i, want := range wantIDs {
		if got[i].ID != want {
			t.Fatalf("Catalog()[%d].ID = %q, want %q", i, got[i].ID, want)
		}
		if got[i].Title == "" {
			t.Fatalf("Catalog()[%d].Title is empty", i)
		}
		if got[i].Description == "" {
			t.Fatalf("Catalog()[%d].Description is empty", i)
		}
	}
}

func TestRenderProducesImage(t *testing.T) {
	img, descriptor, err := Render("lines", 320, 180)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if descriptor.ID != "lines" {
		t.Fatalf("Render() descriptor ID = %q, want %q", descriptor.ID, "lines")
	}
	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 180 {
		t.Fatalf("Render() bounds = %v, want 320x180", img.Bounds())
	}
	if len(img.Pix) != 320*180*4 {
		t.Fatalf("Render() pixel buffer length = %d, want %d", len(img.Pix), 320*180*4)
	}

	allWhite := true
	for i := 0; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 255 || img.Pix[i+1] != 255 || img.Pix[i+2] != 255 || img.Pix[i+3] != 255 {
			allWhite = false
			break
		}
	}
	if allWhite {
		t.Fatal("Render() returned an all-white image")
	}
}

func TestBuildRejectsUnknownDemo(t *testing.T) {
	if _, _, err := Build("nope", 0, 0); err == nil {
		t.Fatal("Build() for unknown demo returned nil error")
	}
}
