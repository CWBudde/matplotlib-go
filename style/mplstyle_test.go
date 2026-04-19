package style

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMPLStyleSubset(t *testing.T) {
	src := `
font.size: 10.0
font.family: "DejaVu Sans"
lines.linewidth: 2.0
lines.color: C1
text.color: "#333333"
axes.facecolor: E5E5E5
axes.edgecolor: white
axes.linewidth: 1.0
axes.labelcolor: 555555
axes.prop_cycle: cycler('color', ['E24A33', '348ABD', '988ED5'])
xtick.color: 555555
ytick.color: 555555
grid.color: white
grid.alpha: 0.75
grid.linewidth: 0.5
legend.facecolor: inherit
legend.edgecolor: 0.5
legend.labelcolor: black
figure.facecolor: white
patch.facecolor: 348ABD
`

	theme, report, err := ParseMPLStyle("GGPlot.mplstyle", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if theme.Name != "ggplot" {
		t.Fatalf("theme name = %q, want ggplot", theme.Name)
	}
	if len(report.Applied) != 18 {
		t.Fatalf("applied count = %d, want 18", len(report.Applied))
	}
	if len(report.Unsupported) != 1 || report.Unsupported[0].Key != "patch.facecolor" {
		t.Fatalf("unexpected unsupported report: %+v", report.Unsupported)
	}

	if theme.RC.FontKey != "DejaVu Sans" || theme.RC.FontSize != 10 {
		t.Fatalf("unexpected font settings: %+v", theme.RC)
	}
	if got, want := theme.RC.LineWidth, 2.0*100.0/72.0; !almostEqual(got, want) {
		t.Fatalf("line width = %v, want %v", got, want)
	}
	if got, want := theme.RC.AxisLineWidth, 1.0*100.0/72.0; !almostEqual(got, want) {
		t.Fatalf("axis line width = %v, want %v", got, want)
	}
	if got, want := theme.RC.GridLineWidth, 0.5*100.0/72.0; !almostEqual(got, want) {
		t.Fatalf("grid line width = %v, want %v", got, want)
	}
	if got, want := theme.RC.MinorGridLineWidth, theme.RC.GridLineWidth; !almostEqual(got, want) {
		t.Fatalf("minor grid line width = %v, want %v", got, want)
	}
	if got := theme.RC.AxesBackground; !almostEqual(got.R, 0xE5/255.0) || !almostEqual(got.G, 0xE5/255.0) || !almostEqual(got.B, 0xE5/255.0) {
		t.Fatalf("axes background = %+v", got)
	}
	if got := theme.RC.AxesEdgeColor; !almostEqual(got.R, 0x55/255.0) || !almostEqual(got.G, 0x55/255.0) || !almostEqual(got.B, 0x55/255.0) {
		t.Fatalf("axes edge color = %+v", got)
	}
	if got := theme.RC.DefaultTextColor(); !almostEqual(got.R, 0x55/255.0) || !almostEqual(got.G, 0x55/255.0) || !almostEqual(got.B, 0x55/255.0) {
		t.Fatalf("text color = %+v", got)
	}
	if got := theme.RC.GridColor; !almostEqual(got.A, 0.75) {
		t.Fatalf("grid alpha = %v, want 0.75", got.A)
	}
	if got, want := theme.RC.LegendBackground, theme.RC.AxesBackground; got != want {
		t.Fatalf("legend background = %+v, want inherit %+v", got, want)
	}
	if got := theme.RC.LegendBorderColor; !almostEqual(got.R, 0.5) || !almostEqual(got.G, 0.5) || !almostEqual(got.B, 0.5) {
		t.Fatalf("legend border color = %+v", got)
	}
	if got, want := theme.RC.Palette()[0], mustParseTestColor(t, "E24A33"); got != want {
		t.Fatalf("palette[0] = %+v, want %+v", got, want)
	}
	if got, want := theme.RC.DefaultLineColor(), theme.RC.Palette()[1]; got != want {
		t.Fatalf("line color = %+v, want %+v", got, want)
	}
}

func TestParseMPLStyleCyclerKeywordForm(t *testing.T) {
	src := ` + "`" + `
axes.prop_cycle: cycler(color=['003FFF', '03ED3A'])
` + "`" + `

	theme, report, err := ParseMPLStyle("custom", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported entries: %+v", report.Unsupported)
	}
	if got, want := theme.RC.Palette()[0], mustParseTestColor(t, "003FFF"); got != want {
		t.Fatalf("palette[0] = %+v, want %+v", got, want)
	}
	if got, want := theme.RC.Palette()[1], mustParseTestColor(t, "03ED3A"); got != want {
		t.Fatalf("palette[1] = %+v, want %+v", got, want)
	}
}

func TestParseMPLStyleInvalidValue(t *testing.T) {
	_, _, err := ParseMPLStyle("broken", "lines.linewidth: nope\n")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadMPLStyleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dark_background.mplstyle")
	if err := os.WriteFile(path, []byte("figure.facecolor: black\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	theme, report, err := LoadMPLStyleFile(path)
	if err != nil {
		t.Fatalf("LoadMPLStyleFile() error = %v", err)
	}
	if theme.Name != "dark_background" {
		t.Fatalf("theme name = %q, want dark_background", theme.Name)
	}
	if len(report.Applied) != 1 {
		t.Fatalf("applied count = %d, want 1", len(report.Applied))
	}
	if got := theme.RC.FigureBackground(); got.R != 0 || got.G != 0 || got.B != 0 || got.A != 1 {
		t.Fatalf("figure background = %+v, want black", got)
	}
}

func TestSupportedMPLStyleKeysSorted(t *testing.T) {
	keys := SupportedMPLStyleKeys()
	if len(keys) == 0 {
		t.Fatal("expected supported keys")
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("supported keys not sorted: %v", keys)
		}
	}
}

func mustParseTestColor(t *testing.T, value string) render.Color {
	t.Helper()
	parsed, err := parseMPLColor(value, Default)
	if err != nil {
		t.Fatalf("parseMPLColor(%q) error = %v", value, err)
	}
	return parsed
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
