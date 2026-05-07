package core

import "github.com/cwbudde/matplotlib-go/transform"

type axes3DProjection struct {
	name string
}

func newAxes3DProjection() *axes3DProjection {
	return &axes3DProjection{name: "3d"}
}

func (p *axes3DProjection) Name() string {
	if p == nil || p.name == "" {
		return "3d"
	}
	return p.name
}

func (p *axes3DProjection) CloneProjection() Projection {
	if p == nil {
		return newAxes3DProjection()
	}
	clone := *p
	return &clone
}

func (p *axes3DProjection) ConfigureAxes(ax *Axes) {
	if ax == nil {
		return
	}
	ax.XScale = transform.NewLinear(default3DViewMin, default3DViewMax)
	ax.YScale = transform.NewLinear(default3DViewMin, default3DViewMax)
	ax.boxAspect = 1
	ax.XAxis = NewXAxis()
	ax.YAxis = NewYAxis()
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ShowFrame = false
}

func (p *axes3DProjection) DataToAxes(*Axes) transform.T {
	return nil
}
