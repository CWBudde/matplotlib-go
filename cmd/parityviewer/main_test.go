package main

import (
	"slices"
	"strings"
	"testing"
)

func TestNewGoldenUpdateCommandIncludesFreetypeTag(t *testing.T) {
	t.Setenv("GOCACHE", "")

	cmd := newGoldenUpdateCommand("/tmp/repo", "^TestCase$")

	if cmd.Dir != "/tmp/repo" {
		t.Fatalf("Dir = %q, want %q", cmd.Dir, "/tmp/repo")
	}
	if !slices.Contains(cmd.Args, "-tags") {
		t.Fatalf("Args missing -tags: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, goldenUpdateBuildTag) {
		t.Fatalf("Args missing %q tag: %v", goldenUpdateBuildTag, cmd.Args)
	}
	if !slices.Contains(cmd.Args, "-update-golden") {
		t.Fatalf("Args missing -update-golden: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, "./test") {
		t.Fatalf("Args missing ./test package: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Args, "^TestCase$") {
		t.Fatalf("Args missing run pattern: %v", cmd.Args)
	}
	if !slices.Contains(cmd.Env, "CGO_ENABLED=1") {
		t.Fatalf("Env missing CGO_ENABLED=1: %v", cmd.Env)
	}
	if !slices.Contains(cmd.Env, "GOCACHE=/tmp/mpl-parity-gocache") {
		t.Fatalf("Env missing fallback GOCACHE: %v", cmd.Env)
	}
}

func TestPageFooterPersistsViewerStateAcrossRerender(t *testing.T) {
	requiredSnippets := []string{
		"var viewerStateStorageKey = 'mpl-parity-viewer-state-v1';",
		"var viewerStateControlIDs = ['search', 'sort-select', 'diff-mode', 'matte-mode', 'resample-mode', 'original-size'];",
		"state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);",
		"window.sessionStorage.setItem(viewerStateStorageKey, JSON.stringify(state));",
		"restoreViewerState();",
		"bindViewerStatePersistence();",
		"setOriginalSize(document.getElementById('original-size').checked);",
		"saveViewerState();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing snippet %q", snippet)
		}
	}
}

func TestPageFooterInitializesSliderAtCenteredPercentage(t *testing.T) {
	if !strings.Contains(pageFooter, "function applyPos(pct)") {
		t.Fatalf("pageFooter missing applyPos helper")
	}
	if !strings.Contains(pageFooter, "applyPos(0.5);") {
		t.Fatalf("pageFooter missing centered slider initialization")
	}
	if strings.Contains(pageFooter, "setPos(wrap.getBoundingClientRect().width / 2);") {
		t.Fatalf("pageFooter still initializes slider with width-based clientX")
	}
}

func TestPageFooterPersistsOpenCardsAcrossReload(t *testing.T) {
	requiredSnippets := []string{
		"function cardStateKey(card) {",
		"state.openCards = Array.from(document.querySelectorAll('.card.open')).map(cardStateKey);",
		"var openCards = Array.isArray(state.openCards) ? state.openCards : [];",
		"card.classList.toggle('open', openCardSet.has(cardStateKey(card)));",
		"header.closest('.card').classList.toggle('open');",
		"saveViewerState();",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(pageFooter, snippet) {
			t.Fatalf("pageFooter missing snippet %q", snippet)
		}
	}
}
