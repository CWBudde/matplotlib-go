package webdemo

import (
	"math"
	"testing"

	"matplotlib-go/core"
)

func TestCatalogStable(t *testing.T) {
	got := Catalog()
	wantIDs := []string{
		"lines",
		"scatter",
		"bars",
		"fills",
		"variants",
		"axes",
		"histogram",
		"statistics",
		"errorbars",
		"units",
		"heatmap",
		"matrix",
		"mesh",
		"vectors",
		"specialty",
		"patches",
		"annotations",
		"composition",
		"polar",
		"subplots",
	}
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

func TestBuildRejectsUnsupportedConfiguredDemo(t *testing.T) {
	original := descriptors
	descriptors = []Descriptor{{
		ID:          "unsupported",
		Title:       "Unsupported",
		Description: "Synthetic test entry.",
	}}
	defer func() {
		descriptors = original
	}()

	if _, _, err := Build("unsupported", 320, 180); err == nil {
		t.Fatal("Build() for unsupported configured demo returned nil error")
	}
}

func TestDefaultDemoIDAndValidDemoID(t *testing.T) {
	if got, want := DefaultDemoID(), "lines"; got != want {
		t.Fatalf("DefaultDemoID() = %q, want %q", got, want)
	}

	for _, descriptor := range Catalog() {
		if !ValidDemoID(descriptor.ID) {
			t.Fatalf("ValidDemoID(%q) = false, want true", descriptor.ID)
		}
	}

	if ValidDemoID("nope") {
		t.Fatal(`ValidDemoID("nope") = true, want false`)
	}
}

func TestBuildEachDemoStructure(t *testing.T) {
	tests := []struct {
		id           string
		width        int
		height       int
		wantAxes     int
		wantTitle    string
		wantXLabel   string
		wantYLabel   string
		wantArtists  int
		wantSizePx   [2]float64
		checkArtists func(t *testing.T, ax *core.Axes)
	}{
		{
			id:          "lines",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Signal Comparison",
			wantXLabel:  "t",
			wantYLabel:  "amplitude",
			wantArtists: 6,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Line2D](ax.Artists); got != 3 {
					t.Fatalf("line artist count = %d, want %d", got, 3)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "scatter",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Scatter Clusters",
			wantXLabel:  "feature x",
			wantYLabel:  "feature y",
			wantArtists: 6,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Scatter2D](ax.Artists); got != 3 {
					t.Fatalf("scatter artist count = %d, want %d", got, 3)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "bars",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Quarterly Revenue",
			wantXLabel:  "quarter",
			wantYLabel:  "EUR million",
			wantArtists: 16,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 1 {
					t.Fatalf("grid artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Bar2D](ax.Artists); got != 2 {
					t.Fatalf("bar artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Text](ax.Artists); got != 12 {
					t.Fatalf("text artist count = %d, want %d", got, 12)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "fills",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Filled Signals",
			wantXLabel:  "t",
			wantYLabel:  "value",
			wantArtists: 6,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Fill2D](ax.Artists); got != 1 {
					t.Fatalf("fill artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Line2D](ax.Artists); got != 2 {
					t.Fatalf("line artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "histogram",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Latency Distribution",
			wantXLabel:  "latency (ms)",
			wantYLabel:  "density",
			wantArtists: 4,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
				hist := firstArtist[*core.Hist2D](t, ax.Artists, "histogram")
				if hist.Norm != core.HistNormDensity {
					t.Fatalf("Hist2D.Norm = %v, want %v", hist.Norm, core.HistNormDensity)
				}
			},
		},
		{
			id:          "errorbars",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Measured Trend With Error Bars",
			wantXLabel:  "sample",
			wantYLabel:  "response",
			wantArtists: 6,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Line2D](ax.Artists); got != 1 {
					t.Fatalf("line artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Scatter2D](ax.Artists); got != 1 {
					t.Fatalf("scatter artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.ErrorBar](ax.Artists); got != 1 {
					t.Fatalf("error bar artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "heatmap",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Heatmap Surface",
			wantXLabel:  "x",
			wantYLabel:  "y",
			wantArtists: 1,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				img := assertArtistType[*core.Image2D](t, ax.Artists[0], "heatmap")
				if got, want := img.Colormap, "inferno"; got != want {
					t.Fatalf("Image2D.Colormap = %q, want %q", got, want)
				}
				if len(img.Data) != 28 || len(img.Data[0]) != 36 {
					t.Fatalf("Image2D.Data size = %dx%d, want 28x36", len(img.Data), len(img.Data[0]))
				}
			},
		},
		{
			id:          "patches",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Patch Showcase",
			wantXLabel:  "x",
			wantYLabel:  "y",
			wantArtists: 6,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Rectangle](ax.Artists); got != 1 {
					t.Fatalf("rectangle artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Circle](ax.Artists); got != 1 {
					t.Fatalf("circle artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Ellipse](ax.Artists); got != 1 {
					t.Fatalf("ellipse artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Polygon](ax.Artists); got != 1 {
					t.Fatalf("polygon artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.FancyArrow](ax.Artists); got != 1 {
					t.Fatalf("arrow artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "polar",
			width:       320,
			height:      180,
			wantAxes:    1,
			wantTitle:   "Polar Wave",
			wantXLabel:  "theta",
			wantYLabel:  "radius",
			wantArtists: 5,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Fill2D](ax.Artists); got != 1 {
					t.Fatalf("fill artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Line2D](ax.Artists); got != 1 {
					t.Fatalf("line artist count = %d, want %d", got, 1)
				}
				if got := countArtists[*core.Legend](ax.Artists); got != 1 {
					t.Fatalf("legend artist count = %d, want %d", got, 1)
				}
			},
		},
		{
			id:          "subplots",
			width:       320,
			height:      180,
			wantAxes:    4,
			wantArtists: 3,
			wantSizePx:  [2]float64{320, 180},
			checkArtists: func(t *testing.T, ax *core.Axes) {
				t.Helper()
				if got := countArtists[*core.Grid](ax.Artists); got != 2 {
					t.Fatalf("grid artist count = %d, want %d", got, 2)
				}
				if got := countArtists[*core.Line2D](ax.Artists); got != 1 {
					t.Fatalf("line artist count = %d, want %d", got, 1)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			fig, descriptor, err := Build(tc.id, tc.width, tc.height)
			if err != nil {
				t.Fatalf("Build(%q) error = %v", tc.id, err)
			}
			if descriptor.ID != tc.id {
				t.Fatalf("descriptor.ID = %q, want %q", descriptor.ID, tc.id)
			}
			if fig == nil {
				t.Fatal("Build() returned nil figure")
			}
			if got, want := len(fig.Children), tc.wantAxes; got != want {
				t.Fatalf("len(fig.Children) = %d, want %d", got, want)
			}
			if got, want := fig.SizePx.X, tc.wantSizePx[0]; got != want {
				t.Fatalf("fig.SizePx.X = %v, want %v", got, want)
			}
			if got, want := fig.SizePx.Y, tc.wantSizePx[1]; got != want {
				t.Fatalf("fig.SizePx.Y = %v, want %v", got, want)
			}

			if tc.id != "subplots" {
				ax := fig.Children[0]
				if got, want := ax.Title, tc.wantTitle; got != want {
					t.Fatalf("ax.Title = %q, want %q", got, want)
				}
				if got, want := ax.XLabel, tc.wantXLabel; got != want {
					t.Fatalf("ax.XLabel = %q, want %q", got, want)
				}
				if got, want := ax.YLabel, tc.wantYLabel; got != want {
					t.Fatalf("ax.YLabel = %q, want %q", got, want)
				}
				if got, want := len(ax.Artists), tc.wantArtists; got != want {
					t.Fatalf("len(ax.Artists) = %d, want %d", got, want)
				}
				tc.checkArtists(t, ax)
				return
			}

			for i, ax := range fig.Children {
				if got, want := len(ax.Artists), tc.wantArtists; got != want {
					t.Fatalf("subplot %d len(ax.Artists) = %d, want %d", i, got, want)
				}
				if got, want := ax.Title, subplotTitle(i); got != want {
					t.Fatalf("subplot %d title = %q, want %q", i, got, want)
				}
				if ax.XLabel != "x" {
					t.Fatalf("subplot %d XLabel = %q, want %q", i, ax.XLabel, "x")
				}
				if ax.YLabel != "y" {
					t.Fatalf("subplot %d YLabel = %q, want %q", i, ax.YLabel, "y")
				}
				tc.checkArtists(t, ax)
			}
		})
	}
}

func TestBuildUsesDefaultDimensions(t *testing.T) {
	fig, descriptor, err := Build("lines", 0, -5)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if descriptor.ID != "lines" {
		t.Fatalf("descriptor.ID = %q, want %q", descriptor.ID, "lines")
	}
	if got, want := fig.SizePx.X, float64(DefaultWidth); got != want {
		t.Fatalf("fig.SizePx.X = %v, want %v", got, want)
	}
	if got, want := fig.SizePx.Y, float64(DefaultHeight); got != want {
		t.Fatalf("fig.SizePx.Y = %v, want %v", got, want)
	}
}

func TestRenderEachDemoProducesImage(t *testing.T) {
	for _, descriptor := range Catalog() {
		t.Run(descriptor.ID, func(t *testing.T) {
			img, gotDescriptor, err := Render(descriptor.ID, 240, 160)
			if err != nil {
				t.Fatalf("Render(%q) error = %v", descriptor.ID, err)
			}
			if gotDescriptor != descriptor {
				t.Fatalf("Render(%q) descriptor = %+v, want %+v", descriptor.ID, gotDescriptor, descriptor)
			}
			if img.Bounds().Dx() != 240 || img.Bounds().Dy() != 160 {
				t.Fatalf("Render(%q) bounds = %v, want 240x160", descriptor.ID, img.Bounds())
			}

			allWhite := true
			for i := 0; i < len(img.Pix); i += 4 {
				if img.Pix[i] != 255 || img.Pix[i+1] != 255 || img.Pix[i+2] != 255 || img.Pix[i+3] != 255 {
					allWhite = false
					break
				}
			}
			if allWhite {
				t.Fatalf("Render(%q) returned an all-white image", descriptor.ID)
			}
		})
	}
}

func TestRenderUsesDefaultDimensions(t *testing.T) {
	img, descriptor, err := Render("lines", 0, -1)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if descriptor.ID != "lines" {
		t.Fatalf("descriptor.ID = %q, want %q", descriptor.ID, "lines")
	}
	if img.Bounds().Dx() != DefaultWidth || img.Bounds().Dy() != DefaultHeight {
		t.Fatalf("Render() bounds = %v, want %dx%d", img.Bounds(), DefaultWidth, DefaultHeight)
	}
}

func BenchmarkRender(b *testing.B) {
	for _, descriptor := range Catalog() {
		b.Run(descriptor.ID, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				img, _, err := Render(descriptor.ID, DefaultWidth, DefaultHeight)
				if err != nil {
					b.Fatalf("Render(%q) error = %v", descriptor.ID, err)
				}
				if img.Bounds().Dx() != DefaultWidth || img.Bounds().Dy() != DefaultHeight {
					b.Fatalf("Render(%q) bounds = %v, want %dx%d", descriptor.ID, img.Bounds(), DefaultWidth, DefaultHeight)
				}
			}
		})
	}
}

func TestDefaultAxesRect(t *testing.T) {
	got := defaultAxesRect()
	if got.Min.X != 0.10 || got.Min.Y != 0.12 || got.Max.X != 0.96 || got.Max.Y != 0.90 {
		t.Fatalf("defaultAxesRect() = %+v, want Min(0.10,0.12) Max(0.96,0.90)", got)
	}
}

func TestLinspace(t *testing.T) {
	if got := linspace(4, 9, 1); len(got) != 1 || got[0] != 4 {
		t.Fatalf("linspace(4, 9, 1) = %v, want [4]", got)
	}

	got := linspace(-1, 1, 5)
	want := []float64{-1, -0.5, 0, 0.5, 1}
	assertFloatSlicesEqual(t, got, want)
}

func TestScatterClusterIsDeterministic(t *testing.T) {
	x1, y1 := scatterCluster(1, 11, -1.2, 0.5, 8)
	x2, y2 := scatterCluster(1, 11, -1.2, 0.5, 8)
	assertFloatSlicesEqual(t, x1, x2)
	assertFloatSlicesEqual(t, y1, y2)

	if len(x1) != 8 || len(y1) != 8 {
		t.Fatalf("scatterCluster lengths = %d, %d, want 8, 8", len(x1), len(y1))
	}
}

func TestDeterministicNormalIsRepeatable(t *testing.T) {
	got := deterministicNormal(6, 47, 8.5)
	if len(got) != 6 {
		t.Fatalf("len(deterministicNormal(...)) = %d, want %d", len(got), 6)
	}
	assertFloatSlicesEqual(t, got, deterministicNormal(6, 47, 8.5))
	if math.Abs(got[0]-46.591536068582016) > 1e-12 {
		t.Fatalf("first sample = %v, want %v", got[0], 46.591536068582016)
	}
}

func TestStrPtr(t *testing.T) {
	if got := strPtr("inferno"); got == nil || *got != "inferno" {
		t.Fatalf("strPtr() = %v, want pointer to %q", got, "inferno")
	}
}

func assertFloatSlicesEqual(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice lengths = %d and %d, want equal", len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("slice[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func assertArtistType[T any](t *testing.T, art core.Artist, label string) T {
	t.Helper()
	typed, ok := art.(T)
	if !ok {
		t.Fatalf("%s type = %T, want %T", label, art, *new(T))
	}
	return typed
}

func firstArtist[T any](t *testing.T, artists []core.Artist, label string) T {
	t.Helper()
	for _, art := range artists {
		if typed, ok := art.(T); ok {
			return typed
		}
	}
	t.Fatalf("did not find %s of type %T", label, *new(T))
	var zero T
	return zero
}

func countArtists[T any](artists []core.Artist) int {
	count := 0
	for _, art := range artists {
		if _, ok := art.(T); ok {
			count++
		}
	}
	return count
}

func subplotTitle(i int) string {
	return []string{"Panel 1", "Panel 2", "Panel 3", "Panel 4"}[i]
}
