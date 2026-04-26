package core

import (
	"testing"

	"matplotlib-go/internal/geom"
	"matplotlib-go/transform"
)

func TestDrawContextCoordinateHelpersSupportBlendedNestedProjectionTransforms(t *testing.T) {
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			DataToAxes: transform.ChainSeparable(
				transform.NewScaleTransform(transform.NewLinear(10, 20), transform.NewLinear(-1, 1)),
				transform.NewSeparable(
					transform.OffsetAxis{Delta: 0.1},
					transform.OffsetAxis{Delta: 0.2},
				),
			),
			AxesToPixel: transform.NewDisplayRectTransform(geom.Rect{
				Min: geom.Pt{X: 100, Y: 100},
				Max: geom.Pt{X: 300, Y: 500},
			}),
		},
		FigureRect: geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 800, Y: 600},
		},
		Clip: geom.Rect{
			Min: geom.Pt{X: 100, Y: 100},
			Max: geom.Pt{X: 300, Y: 500},
		},
	}

	data := ctx.TransData().Apply(geom.Pt{X: 15, Y: 0})
	if data.X != 220 || data.Y != 220 {
		t.Fatalf("transData = %+v, want {220 220}", data)
	}

	axes := ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0.5})
	if axes.X != 200 || axes.Y != 300 {
		t.Fatalf("transAxes = %+v, want {200 300}", axes)
	}

	figure := ctx.TransFigure().Apply(geom.Pt{X: 0.25, Y: 0.25})
	if figure.X != 200 || figure.Y != 450 {
		t.Fatalf("transFigure = %+v, want {200 450}", figure)
	}

	blended := ctx.TransformFor(BlendCoords(CoordData, CoordAxes)).Apply(geom.Pt{X: 15, Y: 0.75})
	if blended.X != 220 || blended.Y != 200 {
		t.Fatalf("blended transform = %+v, want {220 200}", blended)
	}
}
