package core

import (
	"math"
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

func TestAxesImage_DefaultOptions(t *testing.T) {
	data := [][]float64{
		{0, 1},
		{2, 3},
	}
	ax := &Axes{}
	img := ax.Image(data)

	if img == nil {
		t.Fatal("expected image artist")
	}
	if img.Colormap != "viridis" {
		t.Fatalf("expected default colormap viridis, got %q", img.Colormap)
	}
	if img.VMin != 0 || img.VMax != 3 {
		t.Fatalf("expected vmin/vmax 0..3, got %v..%v", img.VMin, img.VMax)
	}
	if img.Alpha != 1 {
		t.Fatalf("expected alpha 1, got %v", img.Alpha)
	}
	if img.AngleDeg != 0 {
		t.Fatalf("expected angle 0, got %v", img.AngleDeg)
	}
	if img.Z() >= NewGrid(AxisBottom).Z() {
		t.Fatalf("expected default image z-order below grid, got image=%v grid=%v", img.Z(), NewGrid(AxisBottom).Z())
	}
}

func TestAxesImage_CustomOptions(t *testing.T) {
	data := [][]float64{{1, 2}, {3, 4}}
	cmap := "gray"
	vmin := -5.0
	vmax := 10.0
	alpha := 2.0
	angle := 45.0
	xMin := -1.0
	xMax := 3.0
	yMin := -2.0
	yMax := 4.0
	rotateX := 1.5
	rotateY := 2.5
	ax := &Axes{}
	img := ax.Image(
		data,
		ImageOptions{
			Colormap:        &cmap,
			VMin:            &vmin,
			VMax:            &vmax,
			Alpha:           &alpha,
			Angle:           &angle,
			XMin:            &xMin,
			XMax:            &xMax,
			YMin:            &yMin,
			YMax:            &yMax,
			Origin:          ImageOriginUpper,
			RotationAnchor:  ImageAnchorCustom,
			RotationAnchorX: &rotateX,
			RotationAnchorY: &rotateY,
		},
	)

	if img.Colormap != "gray" {
		t.Fatalf("expected colormap gray, got %q", img.Colormap)
	}
	if img.VMin != -5 || img.VMax != 10 {
		t.Fatalf("expected vmin/vmax -5..10, got %v..%v", img.VMin, img.VMax)
	}
	if img.Alpha != 1 {
		t.Fatalf("expected alpha clamped to 1, got %v", img.Alpha)
	}
	if img.Origin != ImageOriginUpper {
		t.Fatalf("expected origin upper")
	}
	if img.AngleDeg != 45 {
		t.Fatalf("expected angle 45, got %v", img.AngleDeg)
	}
	if img.RotateAt != ImageAnchorCustom {
		t.Fatalf("expected custom anchor, got %v", img.RotateAt)
	}
	if img.RotateX != rotateX || img.RotateY != rotateY {
		t.Fatalf("unexpected rotation anchor: %v,%v", img.RotateX, img.RotateY)
	}
	if img.XMin != -1 || img.XMax != 3 || img.YMin != -2 || img.YMax != 4 {
		t.Fatalf("unexpected extents: %v..%v / %v..%v", img.XMin, img.XMax, img.YMin, img.YMax)
	}
}

func TestImageRasterizeRejectsEmptyInputs(t *testing.T) {
	if _, ok := (&Image2D{}).rasterize(); ok {
		t.Fatal("expected empty image data to fail rasterization")
	}
	if _, ok := (&Image2D{Data: [][]float64{}}).rasterize(); ok {
		t.Fatal("expected zero-row image data to fail rasterization")
	}
	if _, ok := (&Image2D{Data: [][]float64{{}}}).rasterize(); ok {
		t.Fatal("expected zero-column image data to fail rasterization")
	}
}

func TestImageRasterizeUsesOriginAndSkipsNonFinite(t *testing.T) {
	data := [][]float64{
		{0, 1},
		{math.NaN(), 2},
	}
	img := &Image2D{
		Data:     data,
		Colormap: "gray",
		VMin:     0,
		VMax:     2,
		Alpha:    1,
		Origin:   ImageOriginLower,
	}

	rendered, ok := img.rasterize()
	if !ok {
		t.Fatal("expected rasterization to succeed")
	}

	rgbaData, ok := rendered.(*render.ImageData)
	if !ok {
		t.Fatalf("expected render.ImageData, got %T", rendered)
	}
	pix := rgbaData.RGBA()
	if pix == nil {
		t.Fatal("expected RGBA image from rasterizer")
	}

	// Origin lower flips row order: second data row is written at the top.
	bottomLeft := pix.RGBAAt(0, 1)
	if bottomLeft.R != 0 || bottomLeft.G != 0 || bottomLeft.B != 0 || bottomLeft.A != 255 {
		t.Fatalf("expected bottom-left pixel to be black, got %+v", bottomLeft)
	}

	bottomMid := pix.RGBAAt(1, 1)
	if bottomMid.R != 127 || bottomMid.G != 127 || bottomMid.B != 127 || bottomMid.A != 255 {
		t.Fatalf("expected bottom-middle pixel to be mid-gray, got %+v", bottomMid)
	}

	topMid := pix.RGBAAt(0, 0)
	if topMid.A != 0 {
		t.Fatalf("expected non-finite top-left value to be transparent, got alpha %d", topMid.A)
	}

	topRight := pix.RGBAAt(1, 0)
	if topRight.R != 255 || topRight.G != 255 || topRight.B != 255 || topRight.A != 255 {
		t.Fatalf("expected top-right white pixel, got %+v", topRight)
	}
}

func TestImage_DrawAngleZeroCallsImage(t *testing.T) {
	i := &Image2D{
		Data: [][]float64{
			{0, 1},
		},
		Alpha: 1,
		XMax:  1,
		YMax:  1,
	}
	r := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	err := r.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	i.Draw(r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if r.imageCalls != 1 {
		t.Fatalf("expected Image to be called once, got %d", r.imageCalls)
	}
	if r.transformedCalls != 0 {
		t.Fatalf("expected ImageTransformed not to be called, got %d", r.transformedCalls)
	}
}

func TestImage_DrawRotatedCallsImageTransformed(t *testing.T) {
	angle := 30.0
	i := &Image2D{
		Data:     [][]float64{{0, 1}},
		AngleDeg: angle,
		Alpha:    1,
		XMax:     1,
		YMax:     1,
	}
	r := &imageSpyRenderer{}
	ctx := createTestDrawContext()

	err := r.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	i.Draw(r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if r.transformedCalls != 1 {
		t.Fatalf("expected ImageTransformed to be called once, got %d", r.transformedCalls)
	}
	if r.imageCalls != 0 {
		t.Fatalf("expected Image not to be called when transform renderer is available, got %d", r.imageCalls)
	}
}

func TestImage_DrawRotatedFallsBackToImage(t *testing.T) {
	angle := 30.0
	i := &Image2D{
		Data:     [][]float64{{0, 1}},
		AngleDeg: angle,
		Alpha:    1,
		XMax:     1,
		YMax:     1,
	}
	r := &imageSpyNoTransformRenderer{}
	ctx := createTestDrawContext()

	err := r.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	i.Draw(r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("end: %v", err)
	}

	if r.imageCalls != 1 {
		t.Fatalf("expected fallback to Image, got %d", r.imageCalls)
	}
}

func TestImage_DrawNilRendererDoesNothing(t *testing.T) {
	i := &Image2D{
		Data: [][]float64{{1}},
	}
	ctx := createTestDrawContext()
	i.Draw(nil, ctx)
}

type imageSpyRenderer struct {
	imageCalls       int
	transformedCalls int
	lastDst          geom.Rect
	lastTransform    geom.Affine
}

func (r *imageSpyRenderer) Begin(geom.Rect) error { return nil }
func (r *imageSpyRenderer) End() error            { return nil }
func (r *imageSpyRenderer) Save()                 {}
func (r *imageSpyRenderer) Restore()              {}
func (r *imageSpyRenderer) ClipRect(geom.Rect)    {}
func (r *imageSpyRenderer) ClipPath(geom.Path)    {}
func (r *imageSpyRenderer) Path(geom.Path, *render.Paint) {
}
func (r *imageSpyRenderer) Image(_ render.Image, dst geom.Rect) {
	r.imageCalls++
	r.lastDst = dst
}
func (r *imageSpyRenderer) ImageTransformed(_ render.Image, dst geom.Rect, t geom.Affine) {
	r.transformedCalls++
	r.lastDst = dst
	r.lastTransform = t
}
func (r *imageSpyRenderer) GlyphRun(render.GlyphRun, render.Color) {}
func (r *imageSpyRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}

type imageSpyNoTransformRenderer struct {
	imageCalls int
}

func (r *imageSpyNoTransformRenderer) Begin(geom.Rect) error { return nil }
func (r *imageSpyNoTransformRenderer) End() error            { return nil }
func (r *imageSpyNoTransformRenderer) Save()                 {}
func (r *imageSpyNoTransformRenderer) Restore()              {}
func (r *imageSpyNoTransformRenderer) ClipRect(geom.Rect)    {}
func (r *imageSpyNoTransformRenderer) ClipPath(geom.Path)    {}
func (r *imageSpyNoTransformRenderer) Path(geom.Path, *render.Paint) {
}
func (r *imageSpyNoTransformRenderer) Image(_ render.Image, dst geom.Rect) { r.imageCalls++; _ = dst }
func (r *imageSpyNoTransformRenderer) GlyphRun(render.GlyphRun, render.Color) {
}
func (r *imageSpyNoTransformRenderer) MeasureText(string, float64, string) render.TextMetrics {
	return render.TextMetrics{}
}
