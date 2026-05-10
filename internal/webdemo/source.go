package webdemo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
)

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

	c, ok := examplecatalog.LookupWebDemo(id)
	if !ok {
		return "", Descriptor{}, fmt.Errorf("webdemo: no parity source mapping for demo %q", id)
	}
	source, err := readRepoFile(c.GoPath)
	if err != nil {
		return "", Descriptor{}, err
	}
	return "// " + c.GoPath + "\n" + strings.TrimSpace(source), descriptor, nil
}

func readRepoFile(rel string) (string, error) {
	path := filepath.Clean(rel)
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRootFromWorkingDir(), path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("webdemo: read %s: %w", rel, err)
	}
	return string(data), nil
}

func repoRootFromWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "."
		}
		wd = parent
	}
}
