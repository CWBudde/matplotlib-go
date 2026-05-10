package test

import "testing"

func TestImshowClipped_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_clipped")
}

func TestImshowTransformed_Golden(t *testing.T) {
	runGoldenTest(t, "imshow_transformed")
}

func TestImageAlpha_Golden(t *testing.T) {
	runGoldenTest(t, "image_alpha")
}

func TestMatshowBasic_Golden(t *testing.T) {
	runGoldenTest(t, "matshow_basic")
}

func TestSpyMarker_Golden(t *testing.T) {
	runGoldenTest(t, "spy_marker")
}

func TestSpyImage_Golden(t *testing.T) {
	runGoldenTest(t, "spy_image")
}
