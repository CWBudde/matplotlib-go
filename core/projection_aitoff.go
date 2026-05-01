package core

import (
	"math"

	"matplotlib-go/internal/geom"
)

// aitoffDataTransform implements the Aitoff projection.
//
// Reference: third_party/matplotlib/lib/matplotlib/projections/geo.py:255.
//
// Forward:
//
//	α = arccos(cos(lat) cos(lon/2))
//	sinc(α) = sin(α) / α   (1 at α=0)
//	x = cos(lat) sin(lon/2) / sinc(α)
//	y = sin(lat) / sinc(α)
//
// Inverse is intentionally unsupported, matching matplotlib's AitoffAxes
// (the upstream implementation also returns no inverse).
type aitoffDataTransform struct{}

func (aitoffDataTransform) Apply(p geom.Pt) geom.Pt {
	lon := clamp(p.X, -math.Pi, math.Pi)
	lat := clamp(p.Y, -math.Pi/2, math.Pi/2)
	half := lon / 2
	alpha := math.Acos(clamp(math.Cos(lat)*math.Cos(half), -1, 1))
	sinc := 1.0
	if math.Abs(alpha) > 1e-12 {
		sinc = math.Sin(alpha) / alpha
	}
	x := math.Cos(lat) * math.Sin(half) / sinc
	y := math.Sin(lat) / sinc
	// Aitoff's natural extents: x in [-π, π], y in [-π/2, π/2].
	return geom.Pt{
		X: 0.5 + x/(2*math.Pi),
		Y: 0.5 + y/math.Pi,
	}
}

func (aitoffDataTransform) Invert(geom.Pt) (geom.Pt, bool) {
	return geom.Pt{}, false
}

func newAitoffProjection() *geoProjection {
	return &geoProjection{name: "aitoff", transform: aitoffDataTransform{}}
}
