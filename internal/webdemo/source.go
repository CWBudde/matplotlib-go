package webdemo

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

//go:embed demo.go
var demoSource string

var demoSourceSnippets = parseDemoSourceSnippets()

var demoSourceFunctionByID = map[string]string{
	"lines":       "buildLinesDemo",
	"scatter":     "buildScatterDemo",
	"bars":        "buildBarsDemo",
	"fills":       "buildFillDemo",
	"variants":    "buildPlotVariantsDemo",
	"axes":        "buildAxesDemo",
	"histogram":   "buildHistogramDemo",
	"statistics":  "buildStatisticsDemo",
	"errorbars":   "buildErrorBarsDemo",
	"units":       "buildUnitsDemo",
	"heatmap":     "buildHeatmapDemo",
	"matrix":      "buildMatrixDemo",
	"mesh":        "buildMeshDemo",
	"vectors":     "buildVectorFieldsDemo",
	"specialty":   "buildSpecialtyDemo",
	"patches":     "buildPatchesDemo",
	"annotations": "buildAnnotationsDemo",
	"composition": "buildCompositionDemo",
	"polar":       "buildPolarDemo",
	"projections": "buildProjectionsDemo",
	"subplots":    "buildSubplotsDemo",
	"radialforce": "buildRadialforceDemo",
}

// Source returns the Go source snippet that builds the requested demo.
func Source(id string) (string, Descriptor, error) {
	var descriptor Descriptor
	for _, candidate := range descriptors {
		if candidate.ID == id {
			descriptor = candidate
			break
		}
	}
	if descriptor.ID == "" {
		return "", Descriptor{}, fmt.Errorf("webdemo: unknown demo %q", id)
	}

	funcName, ok := demoSourceFunctionByID[id]
	if !ok {
		return "", Descriptor{}, fmt.Errorf("webdemo: no source mapping for demo %q", id)
	}
	source, ok := demoSourceSnippets[funcName]
	if !ok {
		return "", Descriptor{}, fmt.Errorf("webdemo: source for %s was not found", funcName)
	}
	return source, descriptor, nil
}

func parseDemoSourceSnippets() map[string]string {
	snippets := map[string]string{}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "demo.go", demoSource, parser.ParseComments)
	if err != nil {
		return snippets
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}
		start := fset.Position(fn.Pos()).Offset
		end := fset.Position(fn.End()).Offset
		if start < 0 || end > len(demoSource) || start >= end {
			continue
		}
		snippets[fn.Name.Name] = strings.TrimSpace(demoSource[start:end])
	}
	return snippets
}
