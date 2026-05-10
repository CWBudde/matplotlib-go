package test

import "testing"

func TestImshowBilinear_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_bilinear")
}

func TestImshowBicubic_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_bicubic")
}
