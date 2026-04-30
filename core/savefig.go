package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"matplotlib-go/render"
)

// supportedSaveExtensions is the registry of extensions handled by SaveFig.
//
// Adding a new exporter (e.g. PDF, PostScript) means appending to this map and
// implementing the corresponding render-side capability interface.
var supportedSaveExtensions = map[string]func(*Figure, render.Renderer, string) error{
	".png": SavePNG,
	".svg": SaveSVG,
}

// SaveFig draws the figure and writes it to path using the appropriate exporter
// inferred from the file extension. Currently supports .png and .svg.
//
// The renderer must implement the corresponding capability interface
// (render.PNGExporter for .png, render.SVGExporter for .svg).
func SaveFig(fig *Figure, r render.Renderer, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("savefig: path %q has no extension; supported: %s", path, supportedExtensionsList())
	}
	handler, ok := supportedSaveExtensions[ext]
	if !ok {
		return fmt.Errorf("savefig: unsupported extension %q for %q; supported: %s", ext, path, supportedExtensionsList())
	}
	return handler(fig, r, path)
}

func supportedExtensionsList() string {
	keys := make([]string, 0, len(supportedSaveExtensions))
	for k := range supportedSaveExtensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
