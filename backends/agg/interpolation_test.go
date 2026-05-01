package agg

import (
	"testing"

	agglib "github.com/cwbudde/agg_go"
)

func TestParseInterpolationName(t *testing.T) {
	cases := []struct {
		name string
		want agglib.ImageFilter
		ok   bool
	}{
		{"", agglib.NoFilter, true},
		{"none", agglib.NoFilter, true},
		{"nearest", agglib.NoFilter, true},
		{"bilinear", agglib.Bilinear, true},
		{"bicubic", agglib.Bicubic, true},
		{"hanning", agglib.Hanning, true},
		{"hermite", agglib.Hermite, true},
		{"quadric", agglib.Quadric, true},
		{"catrom", agglib.Catrom, true},
		{"spline16", agglib.Spline16, true},
		{"spline36", agglib.Spline36, true},
		{"blackman", agglib.Blackman, true},
		{"BILINEAR", agglib.Bilinear, true},     // case-insensitive
		{"  bilinear  ", agglib.Bilinear, true}, // trim whitespace
	}
	for _, c := range cases {
		got, ok := parseInterpolationName(c.name)
		if ok != c.ok {
			t.Errorf("parseInterpolationName(%q) ok = %v, want %v", c.name, ok, c.ok)
		}
		if got != c.want {
			t.Errorf("parseInterpolationName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestParseInterpolationName_Unknown(t *testing.T) {
	got, ok := parseInterpolationName("definitely-not-a-filter")
	if ok {
		t.Fatal("expected ok=false for unknown filter")
	}
	if got != agglib.NoFilter {
		t.Fatalf("fallback = %v, want NoFilter", got)
	}
}

func TestResampleForFilter(t *testing.T) {
	cases := []struct {
		filter agglib.ImageFilter
		want   agglib.ImageResample
	}{
		{agglib.NoFilter, agglib.NoResample},
		{agglib.Bilinear, agglib.ResampleAlways},
		{agglib.Bicubic, agglib.ResampleAlways},
		{agglib.Catrom, agglib.ResampleAlways},
	}
	for _, c := range cases {
		if got := resampleForFilter(c.filter); got != c.want {
			t.Errorf("resampleForFilter(%v) = %v, want %v", c.filter, got, c.want)
		}
	}
}
