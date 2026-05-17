package svg

import (
	"math"
	"testing"
)

func TestShortFloat(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{math.Copysign(0, -1), "0"},
		{1, "1"},
		{-1, "-1"},
		{1.5, "1.5"},
		{-1.5, "-1.5"},
		{0.1, "0.1"},
		{1.234567, "1.234567"},
		{1.2345678, "1.234568"},
		{1.5000001, "1.5"},
		{123456.789, "123456.789"},
		{0.0000001, "0"},
		{math.NaN(), "0"},
		{math.Inf(1), "0"},
		{math.Inf(-1), "0"},
	}

	for _, tc := range tests {
		if got := shortFloat(tc.in); got != tc.want {
			t.Errorf("shortFloat(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRotateTransformUsesShortFloat(t *testing.T) {
	got := rotateTransform(-90, 20, 30)
	want := "rotate(-90 20 30)"
	if got != want {
		t.Errorf("rotateTransform(-90, 20, 30) = %q, want %q", got, want)
	}

	got = rotateTransform(45.5, 1.5, 2.25)
	want = "rotate(45.5 1.5 2.25)"
	if got != want {
		t.Errorf("rotateTransform(45.5, 1.5, 2.25) = %q, want %q", got, want)
	}
}
