package test

import (
	"image"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	arraysshowcase "github.com/cwbudde/matplotlib-go/examples/arrays/showcase"
	axesgrid1showcase "github.com/cwbudde/matplotlib-go/examples/axes_grid1/showcase"
	axisartistshowcase "github.com/cwbudde/matplotlib-go/examples/axisartist/showcase"
	unstructuredshowcase "github.com/cwbudde/matplotlib-go/examples/unstructured/showcase"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestUnstructuredShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "unstructured_showcase", renderUnstructuredShowcase)
}

func TestArraysShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "arrays_showcase", renderArraysShowcase)
}

func TestAxisArtistShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axisartist_showcase", renderAxisArtistShowcase)
}

func TestAxesGrid1Showcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_grid1_showcase", renderAxesGrid1Showcase)
}

func renderUnstructuredShowcase() image.Image {
	fig := unstructuredshowcase.UnstructuredShowcase()
	r, err := agg.New(unstructuredshowcase.Width, unstructuredshowcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(unstructuredshowcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderArraysShowcase() image.Image {
	fig := arraysshowcase.ArraysShowcase()
	r, err := agg.New(arraysshowcase.Width, arraysshowcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(arraysshowcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderAxisArtistShowcase() image.Image {
	fig := axisartistshowcase.AxisArtistShowcase()
	r, err := agg.New(axisartistshowcase.Width, axisartistshowcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(axisartistshowcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func renderAxesGrid1Showcase() image.Image {
	fig := axesgrid1showcase.AxesGrid1Showcase()
	r, err := agg.New(axesgrid1showcase.Width, axesgrid1showcase.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(axesgrid1showcase.DPI)
	core.DrawFigure(fig, r)
	return r.GetImage()
}
