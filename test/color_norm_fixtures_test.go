package test

import "testing"

func TestBoundaryNormPColorMesh_Golden(t *testing.T) {
	runGoldenTest(t, "boundarynorm_pcolormesh")
}

func TestLogNormImshow_Golden(t *testing.T) {
	runGoldenTest(t, "lognorm_imshow")
}

func TestTwoSlopeNormImage_Golden(t *testing.T) {
	runGoldenTest(t, "twoslope_norm_image")
}

func TestColorbarExtensions_Golden(t *testing.T) {
	runGoldenTest(t, "colorbar_extensions")
}
