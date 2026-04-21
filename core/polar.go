package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/transform"
)

const (
	polarCircleSegments   = 128
	polarArcMinSegments   = 12
	polarRadialLabelAngle = math.Pi / 8
)

type polarDataTransform struct {
	theta transform.Scale
	r     transform.Scale
}

func (t polarDataTransform) Apply(p geom.Pt) geom.Pt {
	uTheta := 0.0
	if t.theta != nil {
		uTheta = t.theta.Fwd(p.X)
	}
	uRadius := 0.0
	if t.r != nil {
		uRadius = t.r.Fwd(p.Y)
	}

	angle := 2 * math.Pi * uTheta
	return geom.Pt{
		X: 0.5 + 0.5*uRadius*math.Cos(angle),
		Y: 0.5 + 0.5*uRadius*math.Sin(angle),
	}
}

func (t polarDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	dx := p.X - 0.5
	dy := p.Y - 0.5
	uRadius := 2 * math.Hypot(dx, dy)
	angle := math.Atan2(dy, dx)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	uTheta := angle / (2 * math.Pi)

	theta := 0.0
	if t.theta != nil {
		var ok bool
		theta, ok = t.theta.Inv(uTheta)
		if !ok {
			return geom.Pt{}, false
		}
	}

	radius := 0.0
	if t.r != nil {
		var ok bool
		radius, ok = t.r.Inv(uRadius)
		if !ok {
			return geom.Pt{}, false
		}
	}

	return geom.Pt{X: theta, Y: radius}, true
}

func polarCenterAndRadius(clip geom.Rect) (geom.Pt, float64) {
	size := math.Min(clip.W(), clip.H())
	return geom.Pt{
		X: clip.Min.X + clip.W()/2,
		Y: clip.Min.Y + clip.H()/2,
	}, size / 2
}

func polarAxesBackgroundPath(clip geom.Rect) geom.Path {
	center, radius := polarCenterAndRadius(clip)
	return polarCirclePath(center, radius)
}

func polarCirclePath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}
	return polarArcPath(center, radius, 0, 2*math.Pi, polarCircleSegments, true)
}

func polarArcPath(center geom.Pt, radius, start, end float64, segments int, closePath bool) geom.Path {
	if radius <= 0 || segments <= 0 {
		return geom.Path{}
	}
	span := end - start
	if closePath && math.Abs(span) >= 2*math.Pi {
		span = 2 * math.Pi
	}

	path := geom.Path{}
	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		angle := start + span*t
		pt := polarPixelPoint(center, radius, angle)
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	if closePath {
		path.Close()
	}
	return path
}

func polarPixelPoint(center geom.Pt, radius, angle float64) geom.Pt {
	return geom.Pt{
		X: center.X + radius*math.Cos(angle),
		Y: center.Y - radius*math.Sin(angle),
	}
}

func polarAngleForTheta(scale transform.Scale, theta float64) float64 {
	if scale == nil {
		return 0
	}
	return 2 * math.Pi * scale.Fwd(theta)
}

func polarThetaTicks(axis *Axis, scale transform.Scale, minor bool) []float64 {
	if axis == nil || scale == nil {
		return nil
	}
	domainMin, domainMax := scale.Domain()

	var locator Locator
	if minor {
		locator = axis.MinorLocator
	} else {
		locator = axis.Locator
	}
	if locator == nil {
		return nil
	}

	count := axis.majorTickTargetCount()
	if minor {
		count = axis.minorTickTargetCount()
	}
	ticks := visibleTicks(locator.Ticks(domainMin, domainMax, count), domainMin, domainMax)
	if len(ticks) == 0 {
		return nil
	}

	span := math.Abs(domainMax - domainMin)
	if approx(span, 2*math.Pi, 1e-9*math.Max(1, span)) && len(ticks) > 1 {
		last := ticks[len(ticks)-1]
		if approx(last, domainMax, 1e-9*math.Max(1, span)) {
			ticks = ticks[:len(ticks)-1]
		}
	}
	return ticks
}

func polarRadialTicks(axis *Axis, scale transform.Scale, minor bool) []float64 {
	if axis == nil || scale == nil {
		return nil
	}
	domainMin, domainMax := scale.Domain()

	var locator Locator
	if minor {
		locator = axis.MinorLocator
	} else {
		locator = axis.Locator
	}
	if locator == nil {
		return nil
	}

	count := axis.majorTickTargetCount()
	if minor {
		count = axis.minorTickTargetCount()
	}
	ticks := visibleTicks(locator.Ticks(domainMin, domainMax, count), domainMin, domainMax)
	out := make([]float64, 0, len(ticks))
	for _, tick := range ticks {
		u := scale.Fwd(tick)
		if u <= 0 || u > 1 {
			continue
		}
		out = append(out, tick)
	}
	return out
}

func polarTickPaint(color render.Color, width float64, dashes []float64) render.Paint {
	return render.Paint{
		LineWidth: width,
		Stroke:    color,
		LineCap:   render.CapButt,
		LineJoin:  render.JoinMiter,
		Dashes:    dashes,
	}
}

func polarTickLabelAlignments(angle float64) (TextAlign, textLayoutVerticalAlign) {
	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}

	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	hAlign := TextAlignCenter
	switch {
	case cosA > 0.25:
		hAlign = TextAlignLeft
	case cosA < -0.25:
		hAlign = TextAlignRight
	}

	vAlign := textLayoutVAlignCenter
	switch {
	case sinA > 0.25:
		vAlign = textLayoutVAlignBottom
	case sinA < -0.25:
		vAlign = textLayoutVAlignTop
	}

	return hAlign, vAlign
}

func polarAxesContainsDisplayPoint(clip geom.Rect, p geom.Pt) bool {
	center, radius := polarCenterAndRadius(clip)
	return math.Hypot(p.X-center.X, p.Y-center.Y) <= radius
}
