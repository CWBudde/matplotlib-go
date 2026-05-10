package test

import (
	"testing"
)

func TestUnstructuredShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "unstructured_showcase")
}

func TestArraysShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "arrays_showcase")
}

func TestAxisArtistShowcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axisartist_showcase")
}

func TestAxesGrid1Showcase_Golden(t *testing.T) {
	requireOptionalVisualTests(t)
	runGoldenTest(t, "axes_grid1_showcase")
}
