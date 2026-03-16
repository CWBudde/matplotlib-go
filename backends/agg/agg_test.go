package agg

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

var white = render.Color{R: 1, G: 1, B: 1, A: 1}

// mustNew creates a renderer or fails the test.
func mustNew(t *testing.T, w, h int) *Renderer {
	t.Helper()
	r, err := New(w, h, white)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestNew(t *testing.T) {
	r := mustNew(t, 200, 100)
	if r.width != 200 || r.height != 100 {
		t.Errorf("unexpected dimensions: %dx%d", r.width, r.height)
	}
}

func TestNewInvalidDimensions(t *testing.T) {
	cases := []struct {
		name string
		w, h int
	}{
		{"zero width", 0, 100},
		{"zero height", 100, 0},
		{"negative width", -1, 100},
		{"negative height", 100, -1},
		{"both zero", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.w, tc.h, white)
			if err == nil {
				t.Errorf("New(%d, %d) should return error", tc.w, tc.h)
			}
		})
	}
}

func TestBeginEnd(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}

	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if err := r.Begin(viewport); err == nil {
		t.Fatal("double Begin should fail")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if err := r.End(); err == nil {
		t.Fatal("End before Begin should fail")
	}
}

func TestSaveRestore(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	r.Save()
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 50, Y: 50}})
	if r.clipRect == nil {
		t.Fatal("clip should be set after ClipRect")
	}
	r.Restore()
	if r.clipRect != nil {
		t.Fatal("clip should be nil after Restore")
	}
	_ = r.End()
}

func TestSaveRestoreUnderflow(t *testing.T) {
	r := mustNew(t, 100, 100)
	// Restore on empty stack should not panic
	r.Restore()
}

func TestPathLine(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 90})

	paint := &render.Paint{
		LineWidth: 2.0,
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
	}

	// Should not panic
	r.Path(p, paint)
	_ = r.End()
}

func TestPathFill(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 90})
	p.LineTo(geom.Pt{X: 10, Y: 90})
	p.Close()

	paint := &render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	r.Path(p, paint)
	_ = r.End()

	// Verify something was drawn (pixel at center should be red)
	img := r.GetImage()
	c := img.RGBAAt(50, 50)
	if c.R < 200 {
		t.Errorf("center pixel should be red, got R=%d", c.R)
	}
}

func TestMeasureText(t *testing.T) {
	r := mustNew(t, 100, 100)
	m := r.MeasureText("Hello", 12.0, "")
	if m.W <= 0 || m.H <= 0 {
		t.Errorf("text metrics should be positive: W=%f H=%f", m.W, m.H)
	}

	empty := r.MeasureText("", 12.0, "")
	if empty.W != 0 || empty.H != 0 {
		t.Errorf("empty text should have zero metrics")
	}
}

func TestGetImage(t *testing.T) {
	r := mustNew(t, 200, 150)
	img := r.GetImage()
	if img == nil {
		t.Fatal("GetImage returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("unexpected image dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestRendererInterface(t *testing.T) {
	var _ render.Renderer = (*Renderer)(nil)
}

func TestQuantize(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0, 0},
		{1.0, 1.0},
		{0.1234567890123, 0.123457},           // rounded to grid
		{-3.14159265, -3.141593},               // negative
		{1e-7, 0},                              // below grid, rounds to 0
		{0.0000005, 0.000001},                  // half grid rounds up
		{100.123456789, 100.123457},            // large value
	}
	for _, tc := range cases {
		got := quantize(tc.in)
		if math.Abs(got-tc.want) > quantizationGrid/2 {
			t.Errorf("quantize(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestQuantizePt(t *testing.T) {
	pt := geom.Pt{X: 1.23456789, Y: 9.87654321}
	q := quantizePt(pt)
	if math.Abs(q.X-1.234568) > quantizationGrid {
		t.Errorf("X not quantized: %v", q.X)
	}
	if math.Abs(q.Y-9.876543) > quantizationGrid {
		t.Errorf("Y not quantized: %v", q.Y)
	}
}

func TestQuantizeIdempotent(t *testing.T) {
	v := 3.141592653589793
	q1 := quantize(v)
	q2 := quantize(q1)
	if q1 != q2 {
		t.Errorf("quantize not idempotent: %v != %v", q1, q2)
	}
}
