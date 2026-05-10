package test

import "testing"

func TestPColorFlat_Golden(t *testing.T) {
	runGoldenTest(t, "pcolor_flat")
}

func TestPColorMeshNearest_Golden(t *testing.T) {
	runGoldenTest(t, "pcolormesh_nearest")
}

func TestPColorMeshGouraud_Golden(t *testing.T) {
	runGoldenTest(t, "pcolormesh_gouraud")
}

func TestPColorMeshMasked_Golden(t *testing.T) {
	runGoldenTest(t, "pcolormesh_masked")
}

func TestHist2DWeightedDensity_Golden(t *testing.T) {
	runGoldenTest(t, "hist2d_weighted_density")
}
