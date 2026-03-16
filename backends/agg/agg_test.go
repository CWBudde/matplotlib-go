package agg

import (
	"math"
	"testing"

	"matplotlib-go/backends"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// TestAggRenderer_Interface verifies that AggRenderer implements render.Renderer.
func TestAggRenderer_Interface(t *testing.T) {
	var _ render.Renderer = (*AggRenderer)(nil)
}

// TestNew tests basic AggRenderer creation.
func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		expectError bool
	}{
		{"valid dimensions", 800, 600, false},
		{"square canvas", 512, 512, false},
		{"zero width", 0, 600, true},
		{"zero height", 800, 0, true},
		{"negative width", -100, 600, true},
		{"negative height", 800, -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := New(tt.width, tt.height)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("New(%d, %d) expected error, got nil", tt.width, tt.height)
				}
				if renderer != nil {
					t.Errorf("New(%d, %d) expected nil renderer on error", tt.width, tt.height)
				}
			} else {
				if err != nil {
					t.Errorf("New(%d, %d) unexpected error: %v", tt.width, tt.height, err)
				}
				if renderer == nil {
					t.Errorf("New(%d, %d) unexpected nil renderer", tt.width, tt.height)
				} else {
					if renderer.width != tt.width {
						t.Errorf("New(%d, %d) width = %d, expected %d", tt.width, tt.height, renderer.width, tt.width)
					}
					if renderer.height != tt.height {
						t.Errorf("New(%d, %d) height = %d, expected %d", tt.width, tt.height, renderer.height, tt.height)
					}
					if renderer.ctx == nil {
						t.Errorf("New(%d, %d) ctx is nil", tt.width, tt.height)
					}
				}
			}
		})
	}
}

// TestBeginEnd tests the Begin/End lifecycle.
func TestBeginEnd(t *testing.T) {
	renderer, err := New(800, 600)
	if err != nil {
		t.Fatalf("New(800, 600) failed: %v", err)
	}

	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 800, Y: 600}}

	// Test successful Begin
	err = renderer.Begin(viewport)
	if err != nil {
		t.Errorf("Begin() failed: %v", err)
	}
	if !renderer.began {
		t.Errorf("Begin() did not set began flag")
	}

	// Test double Begin error
	err = renderer.Begin(viewport)
	if err == nil {
		t.Errorf("Begin() called twice should return error")
	}

	// Test successful End
	err = renderer.End()
	if err != nil {
		t.Errorf("End() failed: %v", err)
	}
	if renderer.began {
		t.Errorf("End() did not clear began flag")
	}

	// Test double End error
	err = renderer.End()
	if err == nil {
		t.Errorf("End() called twice should return error")
	}
}

// TestSaveRestore tests the state stack functionality.
func TestSaveRestore(t *testing.T) {
	renderer, err := New(800, 600)
	if err != nil {
		t.Fatalf("New(800, 600) failed: %v", err)
	}

	// Initial state
	if len(renderer.stateStack) != 0 {
		t.Errorf("Initial state stack should be empty, got %d", len(renderer.stateStack))
	}

	// Save state
	renderer.Save()
	if len(renderer.stateStack) != 1 {
		t.Errorf("After Save(), stack length should be 1, got %d", len(renderer.stateStack))
	}

	// Save again
	renderer.Save()
	if len(renderer.stateStack) != 2 {
		t.Errorf("After second Save(), stack length should be 2, got %d", len(renderer.stateStack))
	}

	// Restore state
	renderer.Restore()
	if len(renderer.stateStack) != 1 {
		t.Errorf("After Restore(), stack length should be 1, got %d", len(renderer.stateStack))
	}

	// Restore again
	renderer.Restore()
	if len(renderer.stateStack) != 0 {
		t.Errorf("After second Restore(), stack length should be 0, got %d", len(renderer.stateStack))
	}

	// Restore from empty stack (should not crash)
	renderer.Restore()
	if len(renderer.stateStack) != 0 {
		t.Errorf("Restore() from empty stack should maintain 0 length, got %d", len(renderer.stateStack))
	}
}

// TestClipRect tests rectangular clipping.
func TestClipRect(t *testing.T) {
	renderer, err := New(800, 600)
	if err != nil {
		t.Fatalf("New(800, 600) failed: %v", err)
	}

	rect := geom.Rect{Min: geom.Pt{X: 10.5, Y: 20.7}, Max: geom.Pt{X: 100.3, Y: 200.9}}
	renderer.ClipRect(rect)

	if len(renderer.clipStack) != 1 {
		t.Errorf("ClipRect() should add one clip region, got %d", len(renderer.clipStack))
	}

	clip := renderer.clipStack[0]
	if clip.rect == nil {
		t.Errorf("ClipRect() should set rect field")
	}
	if clip.path != nil {
		t.Errorf("ClipRect() should not set path field")
	}

	// Check quantization
	expected := quantize(10.5)
	if clip.rect.Min.X != expected {
		t.Errorf("ClipRect() Min.X quantization: got %f, expected %f", clip.rect.Min.X, expected)
	}
}

// TestClipPath tests path-based clipping.
func TestClipPath(t *testing.T) {
	renderer, err := New(800, 600)
	if err != nil {
		t.Fatalf("New(800, 600) failed: %v", err)
	}

	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 0.1, Y: 0.2}, {X: 100.3, Y: 200.4}},
	}
	renderer.ClipPath(path)

	if len(renderer.clipStack) != 1 {
		t.Errorf("ClipPath() should add one clip region, got %d", len(renderer.clipStack))
	}

	clip := renderer.clipStack[0]
	if clip.path == nil {
		t.Errorf("ClipPath() should set path field")
	}
	if clip.rect != nil {
		t.Errorf("ClipPath() should not set rect field")
	}

	// Check quantization of vertices
	expected := quantize(0.1)
	if clip.path.V[0].X != expected {
		t.Errorf("ClipPath() vertex quantization: got %f, expected %f", clip.path.V[0].X, expected)
	}
}

// TestQuantization tests coordinate quantization functions.
func TestQuantization(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{0.0000005, 0.000001}, // Rounds to nearest epsilon
		{0.0000015, 0.000002}, // Rounds to epsilon
		{1.0000005, 1.000001}, // Rounds to nearest epsilon
		{1.0000015, 1.000002}, // Rounds up
	}

	for _, tt := range tests {
		result := quantize(tt.input)
		if result != tt.expected {
			t.Errorf("quantize(%f) = %f, expected %f", tt.input, result, tt.expected)
		}
	}
}

// TestQuantizePt tests point quantization.
func TestQuantizePt(t *testing.T) {
	input := geom.Pt{X: 1.0000015, Y: 2.0000008}
	expected := geom.Pt{X: 1.000002, Y: 2.000001}
	
	result := quantizePt(input)
	
	// Use tolerance-based comparison for floating point values
	tolerance := 1e-9
	if math.Abs(result.X - expected.X) > tolerance || math.Abs(result.Y - expected.Y) > tolerance {
		t.Errorf("quantizePt(%v) = %v, expected %v", input, result, expected)
	}
}

// TestBackendRegistration tests that the AGG backend is properly registered.
func TestBackendRegistration(t *testing.T) {
	// Test that AGG backend is available
	available := backends.Available()
	found := false
	for _, backend := range available {
		if backend == backends.AGG {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AGG backend not found in available backends: %v", available)
	}

	// Test backend info
	info, ok := backends.DefaultRegistry.Get(backends.AGG)
	if !ok {
		t.Errorf("AGG backend info not found in registry")
	} else {
		if !info.Available {
			t.Errorf("AGG backend should be marked as available")
		}
		if info.Name == "" {
			t.Errorf("AGG backend should have a name")
		}
		if len(info.Capabilities) == 0 {
			t.Errorf("AGG backend should have capabilities")
		}
	}

	// Test backend creation
	config := backends.SimpleConfig(800, 600, render.Color{R: 1, G: 1, B: 1, A: 1})
	renderer, err := backends.Create(backends.AGG, config)
	if err != nil {
		t.Errorf("Failed to create AGG backend: %v", err)
	}
	if renderer == nil {
		t.Errorf("AGG backend creation returned nil renderer")
	}
}

// TestCapabilities tests that AGG backend has expected capabilities.
func TestCapabilities(t *testing.T) {
	expectedCaps := []backends.Capability{
		backends.AntiAliasing,
		backends.SubPixel,
		backends.PathClip,
	}

	for _, cap := range expectedCaps {
		if !backends.HasCapability(backends.AGG, cap) {
			t.Errorf("AGG backend should have capability %s", cap)
		}
	}
}