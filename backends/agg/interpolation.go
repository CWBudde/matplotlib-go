package agg

import (
	"strings"

	agglib "github.com/cwbudde/agg_go"
)

// parseInterpolationName maps a matplotlib-style interpolation string to an
// agg_go ImageFilter. The bool reports whether the name was recognised; the
// returned filter falls back to NoFilter for unrecognised inputs (caller may
// log a warning or surface the failure as the application sees fit).
//
// Recognised names (case-insensitive, surrounding whitespace trimmed):
//   - ""        — same as "none" / "nearest"; no filtering
//   - "none"    — no filtering (nearest-neighbour)
//   - "nearest" — alias for "none"
//   - "bilinear", "bicubic", "hanning", "hermite", "quadric", "catrom",
//     "spline16", "spline36", "blackman" — direct mappings to agg_go filters.
func parseInterpolationName(name string) (agglib.ImageFilter, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "none", "nearest":
		return agglib.NoFilter, true
	case "bilinear":
		return agglib.Bilinear, true
	case "bicubic":
		return agglib.Bicubic, true
	case "hanning":
		return agglib.Hanning, true
	case "hermite":
		return agglib.Hermite, true
	case "quadric":
		return agglib.Quadric, true
	case "catrom":
		return agglib.Catrom, true
	case "spline16":
		return agglib.Spline16, true
	case "spline36":
		return agglib.Spline36, true
	case "blackman":
		return agglib.Blackman, true
	default:
		return agglib.NoFilter, false
	}
}

// resampleForFilter selects an appropriate ImageResample mode for a given
// filter. NoFilter implies NoResample (nearest-neighbour); every other filter
// implies ResampleAlways so the filter actually fires on transformed images.
func resampleForFilter(filter agglib.ImageFilter) agglib.ImageResample {
	if filter == agglib.NoFilter {
		return agglib.NoResample
	}
	return agglib.ResampleAlways
}
