package backends

import (
	"strings"
	"testing"

	"matplotlib-go/render"
)

func TestRegistry(t *testing.T) {
	// Test basic registry operations
	reg := NewRegistry()

	// Test empty registry
	if len(reg.Available()) != 0 {
		t.Error("New registry should be empty")
	}

	// Test registration
	testBackend := Backend("test")
	reg.Register(testBackend, &BackendInfo{
		Name:         "Test Backend",
		Description:  "Test backend for unit tests",
		Capabilities: []Capability{AntiAliasing},
		Factory: func(config Config) (render.Renderer, error) {
			return &render.NullRenderer{}, nil
		},
		Available: true,
	})

	if len(reg.Available()) != 1 {
		t.Error("Registry should have one backend")
	}

	// Test retrieval
	info, ok := reg.Get(testBackend)
	if !ok {
		t.Error("Should find registered backend")
	}
	if info.Name != "Test Backend" {
		t.Error("Backend info should match")
	}

	// Test capability checking
	if !reg.HasCapability(testBackend, AntiAliasing) {
		t.Error("Backend should have AntiAliasing capability")
	}
	if reg.HasCapability(testBackend, GPUAccel) {
		t.Error("Backend should not have GPUAccel capability")
	}
}

func TestBackendSelection(t *testing.T) {
	// Create a test backend for this test
	testBackend := Backend("test")
	Register(testBackend, &BackendInfo{
		Name:         "Test Backend",
		Description:  "Test backend for unit tests",
		Capabilities: []Capability{AntiAliasing},
		Factory: func(config Config) (render.Renderer, error) {
			return &render.NullRenderer{}, nil
		},
		Available: true,
	})

	for _, backend := range DefaultRegistry.Available() {
		backend := backend
		t.Run(string(backend), func(t *testing.T) {
			config := TestDefaultConfig(100, 100)
			renderer, err := DefaultRegistry.Create(backend, config)
			if err != nil {
				t.Fatalf("should create renderer for %s: %v", backend, err)
			}
			if renderer == nil {
				t.Fatalf("renderer for %s should not be nil", backend)
			}
		})
	}
}

func TestCapabilityMatrix(t *testing.T) {
	matrix := CapabilityMatrix()
	if matrix == "" {
		t.Error("Capability matrix should not be empty")
	}

	// Should contain header
	if !strings.Contains(matrix, "Backend") {
		t.Error("Matrix should contain Backend header")
	}
}

func TestRecommendedBackends(t *testing.T) {
	// Test known use cases
	useCases := []string{"basic", "publication", "interactive", "scientific"}

	for _, useCase := range useCases {
		backend, err := GetRecommendedBackend(useCase)
		if err != nil {
			// It's OK if no backend satisfies requirements
			continue
		}

		if backend == "" {
			t.Errorf("GetRecommendedBackend should return non-empty backend for %s", useCase)
		}
	}

	// Test unknown use case
	_, err := GetRecommendedBackend("unknown")
	if err == nil {
		t.Error("Should return error for unknown use case")
	}
}
