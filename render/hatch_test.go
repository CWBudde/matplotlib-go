package render

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

type hatchRecordingRenderer struct {
	NullRenderer
	paths []geom.Path
}

func (r *hatchRecordingRenderer) Path(p geom.Path, paint *Paint) {
	r.paths = append(r.paths, p)
}

func TestDrawHatchFallbackClipsToPolygon(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 10, Y: 0})
	clip.LineTo(geom.Pt{X: 0, Y: 10})
	clip.Close()

	r := &hatchRecordingRenderer{}
	ok := DrawHatchFallback(r, clip, Paint{
		Hatch:          "|",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   5,
	})
	if !ok {
		t.Fatal("DrawHatchFallback returned false")
	}
	if len(r.paths) == 0 {
		t.Fatal("expected clipped hatch paths")
	}

	for _, path := range r.paths {
		for _, pt := range path.V {
			if pt.X < -1e-9 || pt.Y < -1e-9 || pt.X+pt.Y > 10+1e-9 {
				t.Fatalf("hatch point %+v escaped triangular clip path in %+v", pt, path.V)
			}
		}
	}
}

func TestDrawHatchFallbackRepeatedPatternTightensSpacing(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 12, Y: 0})
	clip.LineTo(geom.Pt{X: 12, Y: 10})
	clip.LineTo(geom.Pt{X: 0, Y: 10})
	clip.Close()

	single := &hatchRecordingRenderer{}
	if !DrawHatchFallback(single, clip, Paint{
		Hatch:          "|",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   8,
	}) {
		t.Fatal("single hatch fallback returned false")
	}

	repeated := &hatchRecordingRenderer{}
	if !DrawHatchFallback(repeated, clip, Paint{
		Hatch:          "||",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   8,
	}) {
		t.Fatal("repeated hatch fallback returned false")
	}

	if got, want := hatchSegmentCount(repeated.paths), hatchSegmentCount(single.paths); got <= want {
		t.Fatalf("repeated hatch segment count = %d, want more than %d", got, want)
	}
}

func hatchSegmentCount(paths []geom.Path) int {
	count := 0
	for _, path := range paths {
		for _, cmd := range path.C {
			if cmd == geom.LineTo {
				count++
			}
		}
	}
	return count
}
