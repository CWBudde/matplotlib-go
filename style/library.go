package style

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StyleLibraryEntry describes one .mplstyle file loaded from a style library.
type StyleLibraryEntry struct {
	Name   string
	Path   string
	Report MPLStyleReport
}

// StyleLibraryIssue describes a style-library file or directory that could not
// be loaded. Missing default search-path directories are ignored.
type StyleLibraryIssue struct {
	Path string
	Err  error
}

// StyleLibraryReport summarizes a style-library discovery pass.
type StyleLibraryReport struct {
	Paths   []string
	Loaded  []StyleLibraryEntry
	Skipped []StyleLibraryIssue
}

// StyleLibraryError reports one or more failed style-library paths.
type StyleLibraryError []StyleLibraryIssue

func (e StyleLibraryError) Error() string {
	switch len(e) {
	case 0:
		return "style: no style-library errors"
	case 1:
		return fmt.Sprintf("style: load %s: %v", e[0].Path, e[0].Err)
	default:
		return fmt.Sprintf("style: %d style-library paths could not be loaded", len(e))
	}
}

// Unwrap returns the underlying path errors for errors.Is/errors.As.
func (e StyleLibraryError) Unwrap() []error {
	errs := make([]error, 0, len(e))
	for _, issue := range e {
		if issue.Err != nil {
			errs = append(errs, issue.Err)
		}
	}
	return errs
}

// DefaultStyleLibrarySearchPaths returns the directories scanned by
// LoadStyleLibrary when no explicit paths are provided.
func DefaultStyleLibrarySearchPaths() []string {
	paths := make([]string, 0, 6)
	seen := make(map[string]struct{})
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}
	addList := func(value string) {
		for _, path := range filepath.SplitList(value) {
			add(path)
		}
	}

	addList(os.Getenv("MATPLOTLIB_GO_STYLELIB"))
	addList(os.Getenv("MPLSTYLEPATH"))
	if configDir := strings.TrimSpace(os.Getenv("MPLCONFIGDIR")); configDir != "" {
		add(filepath.Join(configDir, "stylelib"))
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		add(filepath.Join(xdg, "matplotlib", "stylelib"))
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		add(filepath.Join(home, ".config", "matplotlib", "stylelib"))
		add(filepath.Join(home, ".matplotlib", "stylelib"))
	}

	return paths
}

// DiscoverStyleLibrary scans .mplstyle files from style-library directories.
//
// When paths is empty, DefaultStyleLibrarySearchPaths is used and missing
// directories are ignored. Explicit paths may name either directories or single
// .mplstyle files.
func DiscoverStyleLibrary(paths ...string) (map[string]Theme, StyleLibraryReport, error) {
	explicit := len(paths) > 0
	if !explicit {
		paths = DefaultStyleLibrarySearchPaths()
	}

	paths = uniqueStyleLibraryPaths(paths)
	report := StyleLibraryReport{Paths: append([]string(nil), paths...)}
	themes := make(map[string]Theme)

	for _, path := range paths {
		discoverStyleLibraryPath(path, explicit, themes, &report)
	}

	if len(report.Skipped) > 0 {
		return themes, report, StyleLibraryError(report.Skipped)
	}
	return themes, report, nil
}

// LoadStyleLibrary discovers .mplstyle files and registers each loaded theme.
//
// Later files with the same normalized name replace earlier files, matching the
// existing RegisterTheme behavior and allowing user style directories to
// override lower-priority paths.
func LoadStyleLibrary(paths ...string) (StyleLibraryReport, error) {
	themes, report, err := DiscoverStyleLibrary(paths...)
	for _, entry := range report.Loaded {
		RegisterTheme(themes[entry.Name])
	}
	return report, err
}

func uniqueStyleLibraryPaths(paths []string) []string {
	unique := make([]string, 0, len(paths))
	seen := make(map[string]struct{})
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		unique = append(unique, clean)
	}
	return unique
}

func discoverStyleLibraryPath(path string, explicit bool, themes map[string]Theme, report *StyleLibraryReport) {
	info, err := os.Stat(path)
	if err != nil {
		if explicit || !errors.Is(err, os.ErrNotExist) {
			report.Skipped = append(report.Skipped, StyleLibraryIssue{Path: path, Err: err})
		}
		return
	}

	if !info.IsDir() {
		loadStyleLibraryFile(path, themes, report)
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		report.Skipped = append(report.Skipped, StyleLibraryIssue{Path: path, Err: err})
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !isMPLStylePath(entry.Name()) {
			continue
		}
		loadStyleLibraryFile(filepath.Join(path, entry.Name()), themes, report)
	}
}

func loadStyleLibraryFile(path string, themes map[string]Theme, report *StyleLibraryReport) {
	if !isMPLStylePath(path) {
		report.Skipped = append(report.Skipped, StyleLibraryIssue{
			Path: path,
			Err:  fmt.Errorf("style: expected .mplstyle file"),
		})
		return
	}

	theme, mplReport, err := LoadMPLStyleFile(path)
	if err != nil {
		report.Skipped = append(report.Skipped, StyleLibraryIssue{Path: path, Err: err})
		return
	}
	themes[theme.Name] = theme
	report.Loaded = append(report.Loaded, StyleLibraryEntry{
		Name:   theme.Name,
		Path:   path,
		Report: mplReport,
	})
}

func isMPLStylePath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".mplstyle")
}
