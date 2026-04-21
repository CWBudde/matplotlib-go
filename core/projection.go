package core

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"matplotlib-go/transform"
)

// ProjectionFactory constructs a fresh projection instance.
type ProjectionFactory func() Projection

// Projection customizes how an axes maps data into axes coordinates.
type Projection interface {
	Name() string
	ConfigureAxes(ax *Axes)
	DataToAxes(ax *Axes) transform.T
}

// ProjectionCloner preserves projection-local state when axes create derived
// contexts or overlay axes.
type ProjectionCloner interface {
	CloneProjection() Projection
}

var (
	projectionRegistryMu sync.RWMutex
	projectionRegistry   = map[string]ProjectionFactory{}
)

func init() {
	mustRegisterProjection("rectilinear", func() Projection { return rectilinearProjection{} })
	mustRegisterProjection("polar", func() Projection { return newPolarProjection() })
}

// RegisterProjection installs a named axes projection.
func RegisterProjection(name string, factory ProjectionFactory) error {
	key := normalizeProjectionName(name)
	if key == "" {
		return fmt.Errorf("projection name must not be empty")
	}
	if factory == nil {
		return fmt.Errorf("projection %q has nil factory", key)
	}

	projectionRegistryMu.Lock()
	defer projectionRegistryMu.Unlock()

	if _, exists := projectionRegistry[key]; exists {
		return fmt.Errorf("projection %q already registered", key)
	}
	projectionRegistry[key] = factory
	return nil
}

func mustRegisterProjection(name string, factory ProjectionFactory) {
	if err := RegisterProjection(name, factory); err != nil {
		panic(err)
	}
}

func normalizeProjectionName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func lookupProjection(name string) (Projection, error) {
	key := normalizeProjectionName(name)
	if key == "" {
		key = "rectilinear"
	}

	projectionRegistryMu.RLock()
	factory := projectionRegistry[key]
	projectionRegistryMu.RUnlock()

	if factory == nil {
		return nil, fmt.Errorf("unknown projection %q", name)
	}
	return factory(), nil
}

func cloneProjection(proj Projection) Projection {
	if proj == nil {
		clone, _ := lookupProjection("rectilinear")
		return clone
	}
	if cloner, ok := proj.(ProjectionCloner); ok {
		return cloner.CloneProjection()
	}
	clone, err := lookupProjection(proj.Name())
	if err != nil {
		return proj
	}
	return clone
}

func isPolarProjection(proj Projection) bool {
	return proj != nil && normalizeProjectionName(proj.Name()) == "polar"
}

type rectilinearProjection struct{}

func (rectilinearProjection) Name() string { return "rectilinear" }

func (rectilinearProjection) ConfigureAxes(*Axes) {}

func (rectilinearProjection) DataToAxes(ax *Axes) transform.T {
	if ax == nil {
		return nil
	}
	return transform.NewScaleTransform(ax.effectiveXScale(), ax.effectiveYScale())
}

type polarProjection struct {
	thetaOffset      float64
	thetaDirection   float64
	radialLabelAngle float64
}

func newPolarProjection() *polarProjection {
	return &polarProjection{
		thetaDirection:   1,
		radialLabelAngle: defaultPolarRadialLabelAngle,
	}
}

func (*polarProjection) Name() string { return "polar" }

func (p *polarProjection) CloneProjection() Projection {
	if p == nil {
		return newPolarProjection()
	}
	clone := *p
	return &clone
}

func (p *polarProjection) ConfigureAxes(ax *Axes) {
	if ax == nil {
		return
	}

	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(0, 1)
	ax.XAxis = NewXAxis()
	ax.YAxis = NewYAxis()
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ShowFrame = false

	ax.XAxis.Locator = MultipleLocator{Base: math.Pi / 4}
	ax.XAxis.MinorLocator = MultipleLocator{Base: math.Pi / 12}
	ax.XAxis.Formatter = FuncFormatter(formatPolarThetaLabel)
	ax.XAxis.ShowSpine = true
	ax.XAxis.ShowTicks = true
	ax.XAxis.ShowLabels = true

	ax.YAxis.Locator = LinearLocator{}
	ax.YAxis.MinorLocator = MinorLinearLocator{N: 2}
	ax.YAxis.Formatter = ScalarFormatter{Prec: 3}
	ax.YAxis.ShowSpine = true
	ax.YAxis.ShowTicks = true
	ax.YAxis.ShowLabels = true
}

func (p *polarProjection) DataToAxes(ax *Axes) transform.T {
	if ax == nil {
		return nil
	}
	return polarDataTransform{
		theta:          ax.effectiveXScale(),
		r:              ax.effectiveYScale(),
		thetaOffset:    p.thetaOffset,
		thetaDirection: p.thetaDirection,
	}
}

func formatPolarThetaLabel(theta float64) string {
	deg := math.Mod(theta*180/math.Pi, 360)
	if deg < 0 {
		deg += 360
	}
	if approx(deg, math.Round(deg), 1e-9) {
		return fmt.Sprintf("%.0f deg", math.Round(deg))
	}
	return fmt.Sprintf("%.1f deg", deg)
}
