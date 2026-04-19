package test

import (
	"os"
	"testing"
)

const optionalVisualTestsEnv = "RUN_OPTIONAL_VISUAL_TESTS"

func requireOptionalVisualTests(t *testing.T) {
	t.Helper()
	if os.Getenv(optionalVisualTestsEnv) == "true" {
		return
	}
	t.Skip("skipping optional visual parity test (set RUN_OPTIONAL_VISUAL_TESTS=true to run)")
}
