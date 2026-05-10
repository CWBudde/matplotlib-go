package test

import "testing"

func TestLargeScatter_Golden(t *testing.T) {
	runGoldenTest(t, "large_scatter")
}

func TestMixedCollection_Golden(t *testing.T) {
	runGoldenTest(t, "mixed_collection")
}

func TestQuadMesh_Golden(t *testing.T) {
	runGoldenTest(t, "quad_mesh")
}

func TestGouraudTriangles_Golden(t *testing.T) {
	runGoldenTest(t, "gouraud_triangles")
}

func TestClipPathBatch_Golden(t *testing.T) {
	runGoldenTest(t, "clip_path_batch")
}
