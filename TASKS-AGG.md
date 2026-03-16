# AGG Backend Integration Plan for Matplotlib-Go

This document outlines the plan for integrating Anti-Grain Geometry (AGG) as a high-quality rendering backend for matplotlib-go, providing superior anti-aliased rendering with sub-pixel accuracy.

## Overview

AGG will complement the existing GoBasic backend by offering:
- **High-quality anti-aliasing**: True anti-aliased rendering vs GoBasic's basic rasterization
- **Sub-pixel accuracy**: Critical for scientific visualization precision
- **Advanced path operations**: Complex clipping and gradient fills
- **Professional rendering**: Publication-quality output

## Architecture Integration

### Backend Registry Integration
- Add `AGG` constant to `backends/registry.go`
- Register AGG with superior capabilities:
  - `AntiAliasing`: True
  - `SubPixel`: True  
  - `GradientFill`: True (if AGG port supports)
  - `PathClip`: True
  - `TextShaping`: True

### Renderer Interface Implementation
AGG backend must implement all methods from `render.Renderer`:

```go
type Renderer interface {
    Begin(viewport geom.Rect) error
    End() error
    Save()
    Restore()
    ClipRect(r geom.Rect)
    ClipPath(p geom.Path)
    Path(p geom.Path, paint *Paint)
    Image(img Image, dst geom.Rect)
    GlyphRun(run GlyphRun, color Color)
    MeasureText(text string, size float64, fontKey string) TextMetrics
}
```

## Implementation Phases

### ✅ Phase 1: Foundation (Week 1)
- [x] Create `backends/agg/` directory structure
- [ ] Add AGG dependency: `github.com/MeKo-Christian/agg_go`
- [ ] Implement minimal `AggRenderer` struct
- [ ] Register in backend registry
- [ ] Basic unit tests

### Phase 2: Core Rendering (Week 2-3)
- [ ] Implement `Begin/End` with AGG context management
- [ ] Path rendering with AGG rasterizer
- [ ] Stroke operations (width, joins, caps, dashes)  
- [ ] Fill operations with anti-aliasing
- [ ] Basic clipping support
- [ ] State management (Save/Restore)

### Phase 3: Advanced Features (Week 4-5)
- [ ] Complex path clipping
- [ ] Image rendering with filtering
- [ ] Text rendering integration
- [ ] Performance optimizations
- [ ] Memory management

### Phase 4: Testing & Polish (Week 6-7)
- [ ] Comprehensive golden image tests
- [ ] Performance benchmarks vs GoBasic
- [ ] Visual quality comparisons
- [ ] Documentation and examples
- [ ] Integration testing

## File Structure

```
backends/agg/
├── doc.go           # Package documentation
├── init.go          # Backend registration
├── agg.go           # Main AggRenderer implementation  
├── agg_test.go      # Unit tests
├── path.go          # Path conversion utilities
├── state.go         # State management helpers
└── text.go          # Text rendering helpers
```

## Technical Considerations

### Coordinate Systems
- **matplotlib-go**: Top-left origin, Y increases downward
- **AGG**: Configurable, typically bottom-left origin
- Need coordinate transformation utilities

### Color Space Conversion
```go
// matplotlib-go: linear RGBA [0..1]
type Color struct{ R, G, B, A float64 }

// AGG: typically 8-bit RGBA [0..255] 
// Need conversion functions maintaining precision
```

### Quantization for Determinism
Follow GoBasic's approach:
```go
const quantizationEpsilon = 1e-6
func quantize(v float64) float64 {
    return math.Round(v/quantizationEpsilon) * quantizationEpsilon
}
```

### Path Conversion
Convert `geom.Path` to AGG path format:
```go
// matplotlib-go path commands
type Cmd uint8
const (
    MoveTo Cmd = iota
    LineTo
    QuadTo  
    CubicTo
    ClosePath
)

// Need mapping to AGG path operations
```

## Dependencies

### AGG Go Port
- Repository: `https://github.com/MeKo-Christian/agg_go`
- Status: "In development; most internals exist; examples and APIs stabilizing"
- Key packages:
  - `basics` - Core AGG functionality
  - `pixfmt` - Pixel format handling
  - `rasterizer` - Path rasterization
  - `renderer` - High-level rendering API
  - `transform` - Geometric transformations

### Integration Requirements
- Go modules: Add to `go.mod`
- Build tags: Optional conditional compilation
- Platform considerations: AGG is cross-platform

## Testing Strategy

### Unit Tests
- Interface compliance: `var _ render.Renderer = (*AggRenderer)(nil)`
- State management: Save/Restore stack correctness
- Path conversion: geom.Path → AGG path equivalence
- Color conversion: Precision and consistency

### Golden Image Tests
Create AGG variants of existing golden tests:
```
testdata/golden/
├── basic_line.png           # GoBasic reference
├── basic_line_agg.png       # AGG variant
├── scatter_basic.png        # GoBasic reference  
├── scatter_basic_agg.png    # AGG variant
└── ...
```

### Visual Quality Tests
- Anti-aliasing verification
- Sub-pixel precision validation
- Performance benchmarking
- Memory leak detection

## Performance Targets

| Metric | GoBasic | AGG Target |
|--------|---------|------------|
| Simple line plot | 1.0x | 1.5-2.0x |
| Complex scatter | 1.0x | 1.5-2.5x |  
| Anti-aliasing quality | Basic | High |
| Memory usage | Low | Moderate |

## Configuration Options

```go
type AggConfig struct {
    AntiAliasing    bool     // Enable/disable AA
    SubPixelPrec    bool     // Sub-pixel precision
    GammaCorrection float64  // Gamma value for AA
    FilterType      string   // Image scaling filter
    ThreadCount     int      // Parallel rendering threads
}
```

## Backend Selection Logic

```go
// Automatic selection based on use case
backends.GetRecommendedBackend("scientific")  // → AGG
backends.GetRecommendedBackend("basic")       // → GoBasic  
backends.GetRecommendedBackend("publication") // → AGG

// Manual selection
backends.Create(backends.AGG, config)
backends.Create(backends.GoBasic, config)
```

## Risk Mitigation

### AGG Port Stability
- **Risk**: AGG Go port is "in development"
- **Mitigation**: Pin to specific commit, contribute fixes upstream
- **Fallback**: Graceful degradation to GoBasic

### Performance Concerns  
- **Risk**: AGG slower than GoBasic for simple plots
- **Mitigation**: Profile and optimize hot paths
- **Fallback**: Smart backend selection based on complexity

### Memory Usage
- **Risk**: AGG internal buffers use significant memory
- **Mitigation**: Buffer pooling, lifecycle management
- **Monitoring**: Memory usage tests in CI

## Success Metrics

### Functional
- [ ] All existing golden tests pass with AGG backend
- [ ] No visual regressions in plot output
- [ ] Full `render.Renderer` interface implementation
- [ ] Cross-platform consistency

### Quality  
- [ ] Measurable anti-aliasing improvement
- [ ] Sub-pixel precision validation
- [ ] Professional publication-quality output
- [ ] Zero visual artifacts

### Performance
- [ ] AGG rendering within 2.5x of GoBasic performance
- [ ] No memory leaks in long-running applications  
- [ ] Reasonable memory overhead (< 2x GoBasic)
- [ ] Deterministic output across platforms

### Integration
- [ ] Seamless backend switching
- [ ] Proper error handling and fallbacks
- [ ] Clear documentation and examples
- [ ] CI/CD integration

## Future Enhancements

### Phase 5: Advanced Graphics (Future)
- [ ] Gradient fills and patterns
- [ ] Complex path effects
- [ ] Advanced text features (kerning, ligatures)
- [ ] GPU acceleration exploration

### Phase 6: Export Formats (Future)  
- [ ] Vector export (SVG, PDF)
- [ ] High-DPI display support
- [ ] Print-quality output formats
- [ ] Interactive rendering features

## Conclusion

AGG integration will position matplotlib-go as a professional-grade plotting library suitable for scientific computing, publications, and high-quality visualizations while maintaining the simplicity and performance of the existing GoBasic backend for everyday use.