package skia

import "testing"

func TestBackendStrategyDocumentsBuildAndDependencyPolicy(t *testing.T) {
	strategy := BackendStrategy()

	if strategy.BuildTag != "skia" {
		t.Fatalf("BuildTag = %q, want skia", strategy.BuildTag)
	}
	if strategy.Binding != BindingExternalCAPI {
		t.Fatalf("Binding = %q, want %q", strategy.Binding, BindingExternalCAPI)
	}
	if strategy.DefaultMode != ModeCPU {
		t.Fatalf("DefaultMode = %q, want %q", strategy.DefaultMode, ModeCPU)
	}
	if strategy.CPUStatus != StatusImplemented {
		t.Fatalf("CPUStatus = %q, want %q", strategy.CPUStatus, StatusImplemented)
	}
	if strategy.GPUStatus != StatusDeferred {
		t.Fatalf("GPUStatus = %q, want %q", strategy.GPUStatus, StatusDeferred)
	}
	if strategy.CIDefault != CIDefaultStub {
		t.Fatalf("CIDefault = %q, want %q", strategy.CIDefault, CIDefaultStub)
	}
	if len(strategy.RequiredLibraries) == 0 {
		t.Fatal("strategy should name required external libraries")
	}
}
