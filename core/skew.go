package core

import (
	"fmt"
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/transform"
)

const defaultSkewXAngleDeg = 30.0

type skewXProjection struct {
	angleDeg float64
}

func newSkewXProjection() *skewXProjection {
	return &skewXProjection{angleDeg: defaultSkewXAngleDeg}
}

func skewXProjectionForAxes(ax *Axes) (*skewXProjection, bool) {
	if ax == nil {
		return nil, false
	}
	proj, ok := ax.projection.(*skewXProjection)
	return proj, ok && proj != nil
}

func (p *skewXProjection) Name() string {
	return "skewx"
}

func (p *skewXProjection) CloneProjection() Projection {
	if p == nil {
		return newSkewXProjection()
	}
	clone := *p
	return &clone
}

func (p *skewXProjection) ConfigureAxes(ax *Axes) {
	if ax == nil {
		return
	}

	ax.XScale = transform.NewLinear(-40, 50)
	ax.YScale = transform.NewLog(1050, 100, 10)
	ax.XAxis = NewXAxis()
	ax.YAxis = NewYAxis()
	ax.XAxisTop = cloneAxisForSide(ax.XAxis, AxisTop)
	ax.YAxisRight = nil
	ax.ShowFrame = true

	ax.XAxis.Locator = MultipleLocator{Base: 10}
	ax.XAxis.MinorLocator = MultipleLocator{Base: 5}
	ax.XAxis.Formatter = ScalarFormatter{Prec: 0}
	ax.XAxisTop.Locator = ax.XAxis.Locator
	ax.XAxisTop.MinorLocator = ax.XAxis.MinorLocator
	ax.XAxisTop.Formatter = ax.XAxis.Formatter

	pressureTicks := []float64{100, 200, 300, 500, 700, 850, 1000}
	ax.YAxis.Locator = FixedLocator{TicksList: pressureTicks}
	ax.YAxis.MinorLocator = LogLocator{Base: 10, Minor: true, Subs: []float64{2, 3, 4, 5, 6, 7, 8, 9}}
	ax.YAxis.Formatter = FuncFormatter(func(v float64) string {
		return fmt.Sprintf("%.0f", v)
	})
}

func (p *skewXProjection) DataToAxes(ax *Axes) transform.T {
	if ax == nil {
		return nil
	}
	angle := defaultSkewXAngleDeg
	if p != nil {
		angle = p.angleDeg
	}
	return skewXDataTransform{
		x:      ax.effectiveXScale(),
		y:      ax.effectiveYScale(),
		factor: math.Tan(angle * math.Pi / 180),
	}
}

type skewXDataTransform struct {
	x      transform.Scale
	y      transform.Scale
	factor float64
}

func (t skewXDataTransform) Apply(p geom.Pt) geom.Pt {
	u := p.X
	if t.x != nil {
		u = t.x.Fwd(p.X)
	}
	v := p.Y
	if t.y != nil {
		v = t.y.Fwd(p.Y)
	}
	return geom.Pt{
		X: u + t.factor*(v-0.5),
		Y: v,
	}
}

func (t skewXDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	v := p.Y
	y := v
	if t.y != nil {
		var ok bool
		y, ok = t.y.Inv(v)
		if !ok {
			return geom.Pt{}, false
		}
	}

	u := p.X - t.factor*(v-0.5)
	x := u
	if t.x != nil {
		var ok bool
		x, ok = t.x.Inv(u)
		if !ok {
			return geom.Pt{}, false
		}
	}

	return geom.Pt{X: x, Y: y}, true
}
