package test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
)

func TestReferenceCompareCasesMatchExampleCatalog(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range referenceCompareCases {
		if seen[tc.name] {
			t.Fatalf("duplicate reference compare case %q", tc.name)
		}
		seen[tc.name] = true
		if _, ok := examplecatalog.Lookup(tc.name); !ok {
			t.Fatalf("reference compare case %q is missing from example catalog", tc.name)
		}
	}
}
