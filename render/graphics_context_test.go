package render

import "testing"

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
