package main

import (
	"slices"
	"testing"
)

func TestNewGoldenUpdateCommandIncludesFreetypeTag(t *testing.T) {
	t.Setenv("GOCACHE", "")

	cmd := newGoldenUpdateCommand("/tmp/repo", "^TestCase$")

	if cmd.Dir != "/tmp/repo" {
		t.Fatalf("Dir = %q, want %q", cmd.Dir, "/tmp/repo")
	}
	if !slices.Contains(cmd.Args, "-tags") {
		t.Fatalf("Args missing -tags: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, goldenUpdateBuildTag) {
		t.Fatalf("Args missing %q tag: %v", goldenUpdateBuildTag, cmd.Args)
	}
	if !slices.Contains(cmd.Args, "-update-golden") {
		t.Fatalf("Args missing -update-golden: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, "./test") {
		t.Fatalf("Args missing ./test package: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, "^TestCase$") {
		t.Fatalf("Args missing run pattern: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Env, "CGO_ENABLED=1") {
		t.Fatalf("Env missing CGO_ENABLED=1: %v", cmd.Env)
	}
	if !slices.Contains(cmd.Env, "GOCACHE=/tmp/mpl-parity-gocache") {
		t.Fatalf("Env missing fallback GOCACHE: %v", cmd.Env)
	}
}

