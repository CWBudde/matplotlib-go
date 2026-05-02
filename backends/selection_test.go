package backends

import (
	"errors"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func withDefaultRegistry(t *testing.T, reg *Registry) {
	t.Helper()
	previous := DefaultRegistry
	DefaultRegistry = reg
	t.Cleanup(func() {
		DefaultRegistry = previous
	})
}

func testRendererFactory(renderer render.Renderer, err error) Factory {
	return func(Config) (render.Renderer, error) {
		return renderer, err
	}
}

func captureConfigFactory(renderer render.Renderer, captured *Config) Factory {
	return func(config Config) (render.Renderer, error) {
		*captured = config
		return renderer, nil
	}
}

func TestHasAllCapabilities(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("capable"), &BackendInfo{
		Name:         "Capable",
		Available:    true,
		Capabilities: []Capability{AntiAliasing, TextShaping, FontHinting},
		Factory:      testRendererFactory(&render.NullRenderer{}, nil),
	})
	withDefaultRegistry(t, reg)

	if !hasAllCapabilities(Backend("capable"), []Capability{TextShaping, FontHinting}) {
		t.Fatal("expected backend to satisfy all required capabilities")
	}
	if hasAllCapabilities(Backend("capable"), []Capability{TextShaping, GPUAccel}) {
		t.Fatal("expected backend to fail missing capability check")
	}
}

func TestResolveBackend(t *testing.T) {
	reg := NewRegistry()
	reg.Register(GoBasic, &BackendInfo{
		Name:         "GoBasic",
		Available:    true,
		Capabilities: []Capability{AntiAliasing},
		Factory:      testRendererFactory(&render.NullRenderer{}, nil),
	})
	reg.Register(Backend("textual"), &BackendInfo{
		Name:         "Textual",
		Available:    true,
		Capabilities: []Capability{AntiAliasing, TextShaping, FontHinting},
		Factory:      testRendererFactory(&render.NullRenderer{}, nil),
	})
	reg.Register(Backend("offline"), &BackendInfo{
		Name:      "Offline",
		Available: false,
		Factory:   testRendererFactory(&render.NullRenderer{}, nil),
	})
	withDefaultRegistry(t, reg)

	tests := []struct {
		name     string
		choice   string
		required []Capability
		want     Backend
		wantErr  string
	}{
		{name: "empty uses best backend", choice: "", want: GoBasic},
		{name: "auto uses best backend", choice: " auto ", want: GoBasic},
		{name: "default uses best backend", choice: "DEFAULT", want: GoBasic},
		{name: "explicit backend normalizes case", choice: "TeXtUaL", want: Backend("textual")},
		{name: "explicit backend checks required capability", choice: "textual", required: TextCapabilities, want: Backend("textual")},
		{name: "unknown backend errors", choice: "missing", wantErr: "unknown backend"},
		{name: "unavailable backend errors", choice: "offline", wantErr: "is not available"},
		{name: "missing required capability errors", choice: "gobasic", required: TextCapabilities, wantErr: "does not satisfy required capabilities"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveBackend(tc.choice, tc.required)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("ResolveBackend(%q) error = %v, want substring %q", tc.choice, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveBackend(%q) failed: %v", tc.choice, err)
			}
			if got != tc.want {
				t.Fatalf("ResolveBackend(%q) = %q, want %q", tc.choice, got, tc.want)
			}
		})
	}
}

func TestNewRendererAndEnv(t *testing.T) {
	expected := &render.NullRenderer{}
	factoryErr := errors.New("factory failed")
	reg := NewRegistry()
	reg.Register(Backend("textual"), &BackendInfo{
		Name:         "Textual",
		Available:    true,
		Capabilities: []Capability{TextShaping, FontHinting},
		Factory:      testRendererFactory(expected, nil),
	})
	reg.Register(Backend("broken"), &BackendInfo{
		Name:      "Broken",
		Available: true,
		Factory:   testRendererFactory(nil, factoryErr),
	})
	withDefaultRegistry(t, reg)

	renderer, backend, err := NewRenderer("textual", SimpleConfig(320, 200, render.Color{A: 1}), TextCapabilities)
	if err != nil {
		t.Fatalf("NewRenderer failed: %v", err)
	}
	if backend != Backend("textual") {
		t.Fatalf("NewRenderer backend = %q, want %q", backend, Backend("textual"))
	}
	if renderer != expected {
		t.Fatal("NewRenderer returned unexpected renderer instance")
	}

	t.Setenv("MATPLOTLIB_BACKEND", "textual")
	envRenderer, envBackend, err := NewRendererFromEnv(SimpleConfig(100, 50, render.Color{A: 1}), TextCapabilities)
	if err != nil {
		t.Fatalf("NewRendererFromEnv failed: %v", err)
	}
	if envBackend != Backend("textual") {
		t.Fatalf("NewRendererFromEnv backend = %q, want %q", envBackend, Backend("textual"))
	}
	if envRenderer != expected {
		t.Fatal("NewRendererFromEnv returned unexpected renderer instance")
	}

	if _, _, err := NewRenderer("broken", Config{}, nil); err == nil || !strings.Contains(err.Error(), factoryErr.Error()) {
		t.Fatalf("NewRenderer should surface factory errors, got %v", err)
	}
}

func TestDefaultBackendAndSimpleConfig(t *testing.T) {
	if got := DefaultBackend(); got != GoBasic {
		t.Fatalf("DefaultBackend() = %q, want %q", got, GoBasic)
	}

	bg := render.Color{R: 0.25, G: 0.5, B: 0.75, A: 1}
	cfg := SimpleConfig(640, 480, bg)
	if cfg.Width != 640 || cfg.Height != 480 {
		t.Fatalf("SimpleConfig dimensions = %dx%d, want 640x480", cfg.Width, cfg.Height)
	}
	if cfg.Background != bg {
		t.Fatalf("SimpleConfig background = %+v, want %+v", cfg.Background, bg)
	}
	if cfg.DPI != 72.0 {
		t.Fatalf("SimpleConfig DPI = %v, want 72", cfg.DPI)
	}
}

func TestCreateNormalizesBackgroundDefaults(t *testing.T) {
	reg := NewRegistry()
	var captured Config
	reg.Register(Backend("capture"), &BackendInfo{
		Name:      "Capture",
		Available: true,
		Factory:   captureConfigFactory(&render.NullRenderer{}, &captured),
	})

	if _, err := reg.Create(Backend("capture"), Config{Width: 10, Height: 5}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if got, want := captured.Background, (render.Color{R: 1, G: 1, B: 1, A: 1}); got != want {
		t.Fatalf("zero background normalized to %+v, want %+v", got, want)
	}
	if captured.DPI != 72 {
		t.Fatalf("zero DPI normalized to %v, want 72", captured.DPI)
	}

	if _, err := reg.Create(Backend("capture"), Config{
		Width:      10,
		Height:     5,
		Background: render.Color{R: 1, G: 1, B: 1},
	}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if got, want := captured.Background, (render.Color{R: 1, G: 1, B: 1, A: 1}); got != want {
		t.Fatalf("white background without alpha normalized to %+v, want %+v", got, want)
	}

	explicit := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.5}
	if _, err := reg.Create(Backend("capture"), Config{
		Width:      10,
		Height:     5,
		Background: explicit,
		DPI:        144,
	}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if captured.Background != explicit {
		t.Fatalf("explicit background = %+v, want %+v", captured.Background, explicit)
	}
	if captured.DPI != 144 {
		t.Fatalf("explicit DPI = %v, want 144", captured.DPI)
	}

	transparent := render.Color{R: 1, G: 1, B: 1, A: 0}
	if _, err := reg.Create(Backend("capture"), Config{
		Width:       10,
		Height:      5,
		Background:  transparent,
		Transparent: true,
	}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if captured.Background != transparent {
		t.Fatalf("transparent background = %+v, want %+v", captured.Background, transparent)
	}
}

func TestBackendTestSuiteRunAll(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Backend("suite"), &BackendInfo{
		Name:      "Suite",
		Available: true,
		Factory:   testRendererFactory(&render.NullRenderer{}, nil),
	})
	withDefaultRegistry(t, reg)

	suite := NewTestSuite(Backend("suite"), TestDefaultConfig(96, 64))
	if suite.backend != Backend("suite") {
		t.Fatalf("suite backend = %q, want %q", suite.backend, Backend("suite"))
	}
	if suite.config.Width != 96 || suite.config.Height != 64 {
		t.Fatalf("suite config = %+v, want width=96 height=64", suite.config)
	}

	suite.RunAll(t)
}
