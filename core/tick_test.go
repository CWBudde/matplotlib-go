package core

import (
	"math"
	"math/rand"
	"strings"
	"testing"
)

func strictlyIncreasing(xs []float64) bool {
	for i := 1; i < len(xs); i++ {
		if !(xs[i] > xs[i-1]) {
			return false
		}
	}
	return true
}

func TestLinearLocator_BasicRanges(t *testing.T) {
	cases := [][2]float64{{-1, 1}, {0, 1e-9}, {1, 1e6}, {-1e6, -1}, {2, 2}}
	targets := []int{3, 5, 7}
	for _, c := range cases {
		for _, n := range targets {
			ticks := (LinearLocator{}).Ticks(c[0], c[1], n)
			if len(ticks) == 0 {
				t.Fatalf("no ticks for range %+v", c)
			}
			if !strictlyIncreasing(ticks) {
				t.Fatalf("ticks not strictly increasing: %+v", ticks)
			}
			minVal, maxVal := c[0], c[1]
			if minVal > maxVal {
				minVal, maxVal = maxVal, minVal
			}
			if ticks[0] > minVal+1e-12 {
				t.Fatalf("first tick %v > min %v", ticks[0], minVal)
			}
			if ticks[len(ticks)-1] < maxVal-1e-12 {
				t.Fatalf("last tick %v < max %v", ticks[len(ticks)-1], maxVal)
			}
			// Do not assert exact count band here; coverage and monotonicity suffice.
		}
	}
}

func TestLinearLocator_Property(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		a := r.Float64()*2e6 - 1e6
		b := a + (r.Float64()*2e6 + 1e-9)
		n := 2 + int(r.Float64()*8)
		ticks := (LinearLocator{}).Ticks(a, b, n)
		if !strictlyIncreasing(ticks) {
			t.Fatalf("not increasing: %+v", ticks)
		}
		// Coverage
		minVal, maxVal := a, b
		if minVal > maxVal {
			minVal, maxVal = maxVal, minVal
		}
		if ticks[0] > minVal+1e-9 {
			t.Fatalf("first > min: %v > %v", ticks[0], minVal)
		}
		if ticks[len(ticks)-1] < maxVal-1e-9 {
			t.Fatalf("last < max: %v < %v", ticks[len(ticks)-1], maxVal)
		}
	}
}

func TestLinearLocator_HistogramStyleRange(t *testing.T) {
	ticks := (LinearLocator{}).Ticks(0, 0.196, 6)
	want := []float64{0, 0.025, 0.05, 0.075, 0.1, 0.125, 0.15, 0.175, 0.2}
	if len(ticks) != len(want) {
		t.Fatalf("tick count mismatch: got %v want %v", ticks, want)
	}
	for i := range want {
		if math.Abs(ticks[i]-want[i]) > 1e-12 {
			t.Fatalf("tick %d mismatch: got %.17g want %.17g", i, ticks[i], want[i])
		}
	}
}

func TestLogLocator_MajorsMonotone(t *testing.T) {
	bases := []float64{2, 10}
	for _, b := range bases {
		l := LogLocator{Base: b}
		ticks := l.Ticks(1, 1e6, 0)
		if len(ticks) == 0 {
			t.Fatalf("no log ticks for base %v", b)
		}
		if !strictlyIncreasing(ticks) {
			t.Fatalf("log ticks not increasing: %+v", ticks)
		}
		// All ticks should be within [min,max]
		if ticks[0] < 1-1e-12 || ticks[len(ticks)-1] > 1e6+1e-12 {
			t.Fatalf("log ticks out of range: first=%v last=%v", ticks[0], ticks[len(ticks)-1])
		}
	}
}

func TestLogLocator_MinorsBetweenMajors(t *testing.T) {
	l := LogLocator{Base: 10, Minor: true}
	ticks := l.Ticks(1, 1e3, 0)
	if !strictlyIncreasing(ticks) {
		t.Fatalf("log ticks not increasing: %+v", ticks)
	}
	// Must contain the canonical set within [1,1000]
	want := []float64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000}
	// Build a map for quick lookup with tolerance
	has := func(v float64) bool {
		for _, t := range ticks {
			if math.Abs(t-v) <= 1e-12 {
				return true
			}
		}
		return false
	}
	for _, v := range want {
		if !has(v) {
			t.Fatalf("missing expected tick %v in %+v", v, ticks)
		}
	}
}

func TestMinorLinearLocator_SubdividesMajors(t *testing.T) {
	minors := (MinorLinearLocator{N: 5}).Ticks(0, 10, 0)
	if len(minors) == 0 {
		t.Fatal("expected minor ticks")
	}
	// Minor ticks should not coincide with major ticks
	majors := (LinearLocator{}).Ticks(0, 10, 6)
	majorSet := map[float64]bool{}
	for _, m := range majors {
		majorSet[m] = true
	}
	for _, v := range minors {
		for mj := range majorSet {
			if math.Abs(v-mj) < 1e-10 {
				t.Errorf("minor tick %v coincides with major tick %v", v, mj)
			}
		}
	}
	// Should be strictly increasing
	if !strictlyIncreasing(minors) {
		t.Errorf("minor ticks not strictly increasing: %v", minors)
	}
}

func TestMinorLinearLocator_DefaultN(t *testing.T) {
	// N=0 should default to 5
	m0 := (MinorLinearLocator{N: 0}).Ticks(0, 10, 0)
	m5 := (MinorLinearLocator{N: 5}).Ticks(0, 10, 0)
	if len(m0) != len(m5) {
		t.Errorf("N=0 should default to N=5: got %d vs %d ticks", len(m0), len(m5))
	}
}

func TestScalarFormatter_TrimAndScientific(t *testing.T) {
	f := ScalarFormatter{Prec: 6}
	if got := f.Format(1.0); got != "1" {
		t.Fatalf("Format(1.0)=%q", got)
	}
	if got := f.Format(1.230000); got != "1.23" {
		t.Fatalf("trim zeros: %q", got)
	}
	if got := f.Format(1234567); !strings.Contains(got, "e") {
		t.Fatalf("expected scientific for large: %q", got)
	}
	if got := f.Format(0.0000123); !strings.Contains(got, "e") {
		t.Fatalf("expected scientific for small: %q", got)
	}
}

func TestFormatScalarTickLabel_UsesStepPrecision(t *testing.T) {
	f := ScalarFormatter{Prec: 3}
	cases := []struct {
		value float64
		step  float64
		want  string
	}{
		{value: 0, step: 0.025, want: "0.000"},
		{value: 0.05, step: 0.025, want: "0.050"},
		{value: 2, step: 2, want: "2"},
		{value: 0.2, step: 0.2, want: "0.2"},
	}

	for _, tc := range cases {
		if got := formatScalarTickLabel(f, tc.value, tc.step); got != tc.want {
			t.Fatalf("formatScalarTickLabel(%v, %v)=%q want %q", tc.value, tc.step, got, tc.want)
		}
	}
}
