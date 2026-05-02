package render

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestGraphicsContextEffectivePaintAppliesAlpha(t *testing.T) {
	gc := NewGraphicsContext()
	gc.Alpha = 0.25
	gc.Paint = Paint{
		Stroke: Color{R: 1, A: 0.8},
		Fill:   Color{B: 1, A: 0.4},
		Dashes: []float64{1, 2},
	}

	paint := gc.EffectivePaint()
	if paint.Stroke.A != 0.2 || paint.Fill.A != 0.1 {
		t.Fatalf("effective alpha stroke=%v fill=%v, want 0.2/0.1", paint.Stroke.A, paint.Fill.A)
	}
	paint.Dashes[0] = 99
	if gc.Paint.Dashes[0] == 99 {
		t.Fatal("EffectivePaint reused mutable dash backing storage")
	}
}

func TestGraphicsContextEffectivePaintCarriesRendererState(t *testing.T) {
	gc := NewGraphicsContext().
		WithAntialias(AntialiasOff).
		WithSnap(SnapAuto).
		WithHatch("/", Color{R: 1, A: 0.5}, 2).
		WithHatchSpacing(12).
		WithSketch(SketchParams{Scale: 1, Length: 2, Randomness: 3}).
		WithForcedAlpha(0.25).
		WithClipPathTransform(geomIdentity())
	gc.Paint = Paint{
		Stroke: Color{A: 1},
		Fill:   Color{A: 1},
	}

	paint := gc.EffectivePaint()
	if paint.Antialias != AntialiasOff || paint.Snap != SnapAuto {
		t.Fatalf("effective paint lost antialias/snap state: %+v", paint)
	}
	if paint.Hatch != "/" || paint.HatchColor.R != 1 || paint.HatchLineWidth != 2 || paint.HatchSpacing != 12 {
		t.Fatalf("effective paint lost hatch state: %+v", paint)
	}
	if paint.Sketch != (SketchParams{Scale: 1, Length: 2, Randomness: 3}) {
		t.Fatalf("effective paint lost sketch state: %+v", paint.Sketch)
	}
	if !paint.ForceAlpha || paint.Stroke.A != 0.25 || paint.Fill.A != 0.25 {
		t.Fatalf("effective paint did not force alpha: %+v", paint)
	}
	if !paint.HasClipPathTrans {
		t.Fatalf("effective paint lost clip path transform: %+v", paint)
	}
}

func geomIdentity() geom.Affine {
	return geom.Affine{A: 1, D: 1}
}
