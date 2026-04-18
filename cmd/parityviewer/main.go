package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const goldenUpdateBuildTag = "freetype"

type caseEntry struct {
	Suite       string
	Baseline    string
	Name        string
	RMSE        float64
	AvgDiff     float64
	MaxDiff     uint8
	DiffPixels  int
	TotalPixels int
	DiffRatio   float64
	RefWidth    int
	RefHeight   int
	ActWidth    int
	ActHeight   int
	RefDocB64   string
	ActDocB64   string
	RefWhiteB64 string
	RefDarkB64  string
	RefCheckB64 string
	ActWhiteB64 string
	ActDarkB64  string
	ActCheckB64 string
	RawDiffB64  string
	AmpDiffB64  string
}

type loadResult struct {
	Cases         []caseEntry
	ComparedCount int
	SkippedCount  int
}

type metrics struct {
	RMSE        float64
	AvgDiff     float64
	MaxDiff     uint8
	DiffPixels  int
	TotalPixels int
	DiffRatio   float64
}

func main() {
	port := flag.String("port", envOr("PORT", "8090"), "Port to listen on")
	parityDir := flag.String("parity-dir", "", "Optional parity directory with suite/baseline-* / artifacts")
	baselineDir := flag.String("baseline-dir", filepath.Join("testdata", "matplotlib_ref"), "Baseline PNG directory (used when --parity-dir is not set)")
	artifactDir := flag.String("artifact-dir", filepath.Join("testdata", "golden"), "Artifact PNG directory (used when --parity-dir is not set)")
	nameFilter := flag.String("name-filter", "", "Optional case-name substring filter")
	namePrefix := flag.String("name-prefix", "", "Optional case-name prefix filter")
	printOnly := flag.Bool("print", false, "Print comparison rows and exit")
	flag.Parse()

	root, err := detectRepoRoot()
	if err != nil {
		log.Fatalf("detect repo root: %v", err)
	}

	baseDir := filepath.Clean(*baselineDir)
	if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(root, baseDir)
	}
	artDir := filepath.Clean(*artifactDir)
	if !filepath.IsAbs(artDir) {
		artDir = filepath.Join(root, artDir)
	}
	parDir := filepath.Clean(*parityDir)
	if *parityDir != "" && !filepath.IsAbs(parDir) {
		parDir = filepath.Join(root, parDir)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		result, err := loadCases(*parityDir != "", parDir, baseDir, artDir, *nameFilter, *namePrefix)
		if err != nil {
			http.Error(w, fmt.Sprintf("load parity cases: %v", err), http.StatusInternalServerError)
			return
		}
		renderPage(w, result)
	})
	http.HandleFunc("/rerender", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if r.FormValue("all") == "1" {
			if err := rerenderAllArtifacts(root); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "missing name parameter", http.StatusBadRequest)
			return
		}
		if err := rerenderArtifact(root, name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if *printOnly {
		result, err := loadCases(*parityDir != "", parDir, baseDir, artDir, *nameFilter, *namePrefix)
		if err != nil {
			log.Fatalf("load parity cases: %v", err)
		}
		printCases(os.Stdout, result)
		return
	}

	addr := ":" + *port
	log.Printf("Parity viewer running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func rerenderArtifact(repoRoot, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("missing name")
	}
	return runGoGoldenUpdate(repoRoot, testNameFromCaseName(name))
}

func rerenderAllArtifacts(repoRoot string) error {
	return runGoGoldenUpdate(repoRoot, "")
}

func runGoGoldenUpdate(repoRoot, runPattern string) error {
	cmd := newGoldenUpdateCommand(repoRoot, runPattern)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(out.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("go test failed: %w: %s", err, msg)
	}
	if runPattern != "" {
		outText := out.String()
		if strings.Contains(outText, "testing: warning: no tests to run") {
			return fmt.Errorf("no matching test for case %q", strings.TrimSuffix(strings.TrimPrefix(runPattern, "^"), "$"))
		}
	}
	return nil
}

func newGoldenUpdateCommand(repoRoot, runPattern string) *exec.Cmd {
	args := []string{"test", "-tags", goldenUpdateBuildTag, "-count", "1", "-update-golden"}
	if runPattern != "" {
		args = append(args, "-run", runPattern)
	}
	args = append(args, "./test")

	cmd := exec.Command("go", args...)
	cmd.Dir = repoRoot

	cmd.Env = os.Environ()
	cmd.Env = setEnv(cmd.Env, "CGO_ENABLED", "1")
	if cacheDir, hasCache := os.LookupEnv("GOCACHE"); !hasCache || strings.TrimSpace(cacheDir) == "" {
		cmd.Env = setEnv(cmd.Env, "GOCACHE", "/tmp/mpl-parity-gocache")
	}

	return cmd
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	found := false
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			found = true
		}
	}
	if !found {
		env = append(env, prefix+value)
	}
	return env
}

func testNameFromCaseName(name string) string {
	parts := strings.Split(name, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		p = strings.ToLower(p)
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}
	return "Test" + b.String() + "_Golden"
}

func loadCases(useParity bool, parityDir, baselineDir, artifactDir, nameFilter, namePrefix string) (loadResult, error) {
	if useParity {
		return loadCasesFromParityDir(parityDir, nameFilter, namePrefix)
	}
	return loadCasesFromDirectories(baselineDir, artifactDir, nameFilter, namePrefix)
}

func printCases(w io.Writer, result loadResult) {
	fmt.Fprintf(w, "found=%d skipped=%d\n", result.ComparedCount, result.SkippedCount)
	for _, c := range result.Cases {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\trmse=%.4f\tavg=%.4f\tmax=%d\tdiff_pixels=%d\tdiff_ratio=%.6f\tsize=%dx%d->%dx%d\n",
			c.Suite,
			c.Baseline,
			c.Name,
			c.RMSE,
			c.AvgDiff,
			c.MaxDiff,
			c.DiffPixels,
			c.DiffRatio,
			c.RefWidth,
			c.RefHeight,
			c.ActWidth,
			c.ActHeight,
		)
	}
}

func detectRepoRoot() (string, error) {
	wd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for range 12 {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		next := filepath.Dir(wd)
		if next == wd {
			break
		}
		wd = next
	}
	return "", errors.New("go.mod not found from cwd")
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func loadCasesFromDirectories(baselineDir, artifactDir, nameFilter, namePrefix string) (loadResult, error) {
	filter := strings.ToLower(strings.TrimSpace(nameFilter))
	prefix := strings.ToLower(strings.TrimSpace(namePrefix))
	var result loadResult

	baseline, err := filepath.Glob(filepath.Join(baselineDir, "*.png"))
	if err != nil {
		return loadResult{}, fmt.Errorf("glob baseline %s: %w", baselineDir, err)
	}
	sort.Strings(baseline)

	for _, baselinePath := range baseline {
		name := strings.TrimSuffix(filepath.Base(baselinePath), filepath.Ext(baselinePath))
		artifactPath := filepath.Join(artifactDir, filepath.Base(baselinePath))
		_, statErr := os.Stat(artifactPath)
		if os.IsNotExist(statErr) {
			result.SkippedCount++
			continue
		}
		if statErr != nil {
			return loadResult{}, fmt.Errorf("stat artifact %s: %w", artifactPath, statErr)
		}
		if filter != "" && !strings.Contains(strings.ToLower(name), filter) {
			continue
		}
		if prefix != "" && !strings.HasPrefix(strings.ToLower(name), prefix) {
			continue
		}

		entry, err := buildEntry("plots", filepath.Base(baselineDir), name, baselinePath, artifactPath)
		if err != nil {
			return loadResult{}, fmt.Errorf("build entry for %s: %w", name, err)
		}
		result.Cases = append(result.Cases, entry)
		result.ComparedCount++
	}

	sort.Slice(result.Cases, func(i, j int) bool {
		if result.Cases[i].RMSE == result.Cases[j].RMSE {
			return result.Cases[i].Name < result.Cases[j].Name
		}
		return result.Cases[i].RMSE > result.Cases[j].RMSE
	})

	return result, nil
}

func loadCasesFromParityDir(parityDir, nameFilter, namePrefix string) (loadResult, error) {
	filter := strings.ToLower(strings.TrimSpace(nameFilter))
	prefix := strings.ToLower(strings.TrimSpace(namePrefix))
	var result loadResult

	suiteDirs, err := os.ReadDir(parityDir)
	if err != nil {
		return loadResult{}, fmt.Errorf("read parity dir %s: %w", parityDir, err)
	}

	for _, suiteDir := range suiteDirs {
		if !suiteDir.IsDir() {
			continue
		}
		suiteName := suiteDir.Name()
		suitePath := filepath.Join(parityDir, suiteName)

		children, err := os.ReadDir(suitePath)
		if err != nil {
			return loadResult{}, fmt.Errorf("read suite dir %s: %w", suitePath, err)
		}

		for _, child := range children {
			if !child.IsDir() || !strings.HasPrefix(child.Name(), "baseline-") {
				continue
			}

			baselineName := child.Name()
			baselineDir := filepath.Join(suitePath, baselineName)
			artifactDir := filepath.Join(suitePath, "artifacts")
			artifactBaselineDir := filepath.Join(artifactDir, baselineName)

			files, err := filepath.Glob(filepath.Join(baselineDir, "*.png"))
			if err != nil {
				return loadResult{}, fmt.Errorf("glob baselines in %s: %w", baselineDir, err)
			}
			sort.Strings(files)

			for _, baselinePath := range files {
				name := strings.TrimSuffix(filepath.Base(baselinePath), filepath.Ext(baselinePath))
				artifactPath := filepath.Join(artifactDir, filepath.Base(baselinePath))
				if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
					artifactPath = filepath.Join(artifactBaselineDir, filepath.Base(baselinePath))
					if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
						result.SkippedCount++
						continue
					}
				}
				if filter != "" && !strings.Contains(strings.ToLower(name), filter) {
					continue
				}
				if prefix != "" && !strings.HasPrefix(strings.ToLower(name), prefix) {
					continue
				}
				entry, err := buildEntry(suiteName, baselineName, name, baselinePath, artifactPath)
				if err != nil {
					return loadResult{}, fmt.Errorf("build entry %s/%s/%s: %w", suiteName, baselineName, name, err)
				}
				result.Cases = append(result.Cases, entry)
				result.ComparedCount++
			}
		}
	}

	sort.Slice(result.Cases, func(i, j int) bool {
		if result.Cases[i].RMSE == result.Cases[j].RMSE {
			if result.Cases[i].Suite == result.Cases[j].Suite {
				if result.Cases[i].Baseline == result.Cases[j].Baseline {
					return result.Cases[i].Name < result.Cases[j].Name
				}
				return result.Cases[i].Baseline < result.Cases[j].Baseline
			}
			return result.Cases[i].Suite < result.Cases[j].Suite
		}
		return result.Cases[i].RMSE > result.Cases[j].RMSE
	})

	return result, nil
}

func buildEntry(suite, baseline, name, baselinePath, artifactPath string) (caseEntry, error) {
	ref, err := readPNGAsRGBA(baselinePath)
	if err != nil {
		return caseEntry{}, fmt.Errorf("read baseline: %w", err)
	}
	act, err := readPNGAsRGBA(artifactPath)
	if err != nil {
		return caseEntry{}, fmt.Errorf("read artifact: %w", err)
	}

	rawDiff := rawDiffImage(ref, act)
	ampDiff := amplifiedDiffImage(ref, act)
	stats := compareImages(ref, act)

	refDocB64, err := pngToBase64(compositeOverSolid(ref, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	actDocB64, err := pngToBase64(compositeOverSolid(act, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}

	refWhite, err := pngToBase64(compositeOverSolid(ref, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	actWhite, err := pngToBase64(compositeOverSolid(act, color.RGBA{R: 255, G: 255, B: 255, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	refDark, err := pngToBase64(compositeOverSolid(ref, color.RGBA{R: 12, G: 16, B: 22, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	actDark, err := pngToBase64(compositeOverSolid(act, color.RGBA{R: 12, G: 16, B: 22, A: 255}))
	if err != nil {
		return caseEntry{}, err
	}
	refCheck, err := pngToBase64(compositeOverCheckerboard(ref))
	if err != nil {
		return caseEntry{}, err
	}
	actCheck, err := pngToBase64(compositeOverCheckerboard(act))
	if err != nil {
		return caseEntry{}, err
	}
	rawDiffB64, err := pngToBase64(rawDiff)
	if err != nil {
		return caseEntry{}, err
	}
	ampDiffB64, err := pngToBase64(ampDiff)
	if err != nil {
		return caseEntry{}, err
	}

	return caseEntry{
		Suite:       suite,
		Baseline:    baseline,
		Name:        name,
		RMSE:        stats.RMSE,
		AvgDiff:     stats.AvgDiff,
		MaxDiff:     stats.MaxDiff,
		DiffPixels:  stats.DiffPixels,
		TotalPixels: stats.TotalPixels,
		DiffRatio:   stats.DiffRatio,
		RefWidth:    ref.Bounds().Dx(),
		RefHeight:   ref.Bounds().Dy(),
		ActWidth:    act.Bounds().Dx(),
		ActHeight:   act.Bounds().Dy(),
		RefDocB64:   refDocB64,
		RefWhiteB64: refWhite,
		RefDarkB64:  refDark,
		RefCheckB64: refCheck,
		ActDocB64:   actDocB64,
		ActWhiteB64: actWhite,
		ActDarkB64:  actDark,
		ActCheckB64: actCheck,
		RawDiffB64:  rawDiffB64,
		AmpDiffB64:  ampDiffB64,
	}, nil
}

func compareImages(ref, act *image.RGBA) metrics {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	totalPixels := bounds.Dx() * bounds.Dy()
	if totalPixels <= 0 {
		return metrics{}
	}

	var (
		sumSq      float64
		totalDiff  float64
		diffPixels int
		maxDiff    uint8
	)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rr, rg, rb, ra := rgbaAt(ref, x, y)
			ar, ag, ab, aa := rgbaAt(act, x, y)

			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)

			pixelMax := max4(dr, dg, db, da)
			if pixelMax > maxDiff {
				maxDiff = pixelMax
			}
			if pixelMax != 0 {
				diffPixels++
			}

			totalDiff += float64(dr) + float64(dg) + float64(db) + float64(da)
			sumSq += sqDiff(rr, ar) + sqDiff(rg, ag) + sqDiff(rb, ab) + sqDiff(ra, aa)
		}
	}

	return metrics{
		RMSE:        math.Sqrt(sumSq / float64(totalPixels*4)),
		AvgDiff:     totalDiff / float64(totalPixels*4),
		MaxDiff:     maxDiff,
		DiffPixels:  diffPixels,
		TotalPixels: totalPixels,
		DiffRatio:   float64(diffPixels) / float64(totalPixels),
	}
}

func rawDiffImage(ref, act *image.RGBA) *image.RGBA {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rr, rg, rb, ra := rgbaAt(ref, bounds.Min.X+x, bounds.Min.Y+y)
			ar, ag, ab, aa := rgbaAt(act, bounds.Min.X+x, bounds.Min.Y+y)
			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)
			if dr == 0 && dg == 0 && db == 0 && da == 0 {
				out.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				continue
			}
			out.SetRGBA(x, y, color.RGBA{R: dr, G: dg, B: db, A: clampAlpha(da)})
		}
	}
	return out
}

func amplifiedDiffImage(ref, act *image.RGBA) *image.RGBA {
	bounds := unionBounds(ref.Bounds(), act.Bounds())
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			rr, rg, rb, ra := rgbaAt(ref, bounds.Min.X+x, bounds.Min.Y+y)
			ar, ag, ab, aa := rgbaAt(act, bounds.Min.X+x, bounds.Min.Y+y)
			dr := absDiff8(rr, ar)
			dg := absDiff8(rg, ag)
			db := absDiff8(rb, ab)
			da := absDiff8(ra, aa)
			pixelMax := max4(dr, dg, db, da)
			if pixelMax == 0 {
				out.SetRGBA(x, y, color.RGBA{R: 0, G: 0xaa, B: 0, A: 255})
				continue
			}
			intensity := uint8(255)
			if pixelMax < 255 {
				intensity = uint8((float64(pixelMax) / 255.0) * 255.0)
			}
			out.SetRGBA(x, y, color.RGBA{R: intensity, G: 0, B: 0, A: 255})
		}
	}
	return out
}

func rgbaAt(img *image.RGBA, x, y int) (uint8, uint8, uint8, uint8) {
	if img == nil || !image.Pt(x, y).In(img.Bounds()) {
		return 0, 0, 0, 0
	}
	i := img.PixOffset(x, y)
	return img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]
}

func unionBounds(a, b image.Rectangle) image.Rectangle {
	minX := min(a.Min.X, b.Min.X)
	minY := min(a.Min.Y, b.Min.Y)
	maxX := max(a.Max.X, b.Max.X)
	maxY := max(a.Max.Y, b.Max.Y)
	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}
	return image.Rect(minX, minY, maxX, maxY)
}

func max4(a, b, c, d uint8) uint8 {
	if a < b {
		a = b
	}
	if a < c {
		a = c
	}
	if a < d {
		a = d
	}
	return a
}

func absDiff8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func sqDiff(a, b uint8) float64 {
	d := float64(int(a) - int(b))
	return d * d
}

func pngToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func readPNGAsRGBA(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(img.Bounds())
	for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
		for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba, nil
}

func clampAlpha(a uint8) uint8 {
	if a == 0 {
		return 255
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func compositeOverSolid(src *image.RGBA, bg color.RGBA) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := rgbaAt(src, x, y)
			out.SetRGBA(x-bounds.Min.X, y-bounds.Min.Y, compositePixel(r, g, b, a, bg))
		}
	}
	return out
}

func compositeOverCheckerboard(src *image.RGBA) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	light := color.RGBA{R: 216, G: 222, B: 233, A: 255}
	dark := color.RGBA{R: 195, G: 202, B: 214, A: 255}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			bg := light
			if ((x-bounds.Min.X)/6+(y-bounds.Min.Y)/6)%2 != 0 {
				bg = dark
			}
			r, g, b, a := rgbaAt(src, x, y)
			out.SetRGBA(x-bounds.Min.X, y-bounds.Min.Y, compositePixel(r, g, b, a, bg))
		}
	}
	return out
}

func compositePixel(r, g, b, a uint8, bg color.RGBA) color.RGBA {
	srcA := int(a)
	invA := 255 - srcA
	return color.RGBA{
		R: uint8((int(r)*srcA + int(bg.R)*invA) / 255),
		G: uint8((int(g)*srcA + int(bg.G)*invA) / 255),
		B: uint8((int(b)*srcA + int(bg.B)*invA) / 255),
		A: 255,
	}
}

func renderPage(w io.Writer, result loadResult) {
	fmt.Fprint(w, pageHeader)
	fmt.Fprintf(
		w,
		`<div class="header-meta">%d comparisons loaded`,
		result.ComparedCount,
	)
	if result.SkippedCount > 0 {
		fmt.Fprintf(w, `, %d baselines skipped because no matching artifact exists`, result.SkippedCount)
	}
	fmt.Fprint(w, `.</div></div>`)

	fmt.Fprint(w, `<div class="container" id="cards-container">`)
	for i := range result.Cases {
		renderCard(w, &result.Cases[i])
	}
	if len(result.Cases) == 0 {
		fmt.Fprint(w, `<div class="empty-state">No parity comparisons found. Point --baseline-dir/--artifact-dir at existing PNG sets.</div>`)
	}
	fmt.Fprint(w, pageFooter)
}

func renderCard(w io.Writer, entry *caseEntry) {
	if entry == nil {
		return
	}

	fmt.Fprintf(
		w,
		`<div class="card" data-name="%s" data-suite="%s" data-baseline="%s" data-rmse="%.4f" data-avg-diff="%.4f" data-max-diff="%d" data-diff-pixels="%d" data-diff-ratio="%.6f">`,
		htmlEscape(entry.Name),
		htmlEscape(entry.Suite),
		htmlEscape(entry.Baseline),
		entry.RMSE,
		entry.AvgDiff,
		entry.MaxDiff,
		entry.DiffPixels,
		entry.DiffRatio,
	)
	fmt.Fprint(w, `<div class="card-header">`)
	fmt.Fprint(w, `<span class="badge badge-neutral sort-metric-badge" style="display:none"></span>`)
	fmt.Fprintf(w, `<span class="card-title">%s</span>`, htmlEscape(entry.Name))
	fmt.Fprintf(
		w,
		`<button class="rerender-btn" data-suite="%s" data-name="%s" type="button">Re-render Artifact</button>`,
		htmlEscape(entry.Suite),
		htmlEscape(entry.Name),
	)
	fmt.Fprint(w, `<div class="right-badges">`)
	fmt.Fprintf(w, `<span class="badge badge-neutral">%s</span>`, htmlEscape(entry.Suite))
	fmt.Fprintf(w, `<span class="badge badge-neutral">%s</span>`, htmlEscape(entry.Baseline))
	fmt.Fprintf(w, `<span class="badge %s">RMSE %.2f</span>`, badgeClassRMSE(entry.RMSE), entry.RMSE)
	fmt.Fprintf(w, `<span class="badge %s">avg %.2f</span>`, badgeClassAvgDiff(entry.AvgDiff), entry.AvgDiff)
	fmt.Fprintf(w, `<span class="badge %s">max %d</span>`, badgeClassMaxDiff(entry.MaxDiff), entry.MaxDiff)
	fmt.Fprintf(w, `<span class="badge %s">diff %.2f%%</span>`, badgeClassDiffRatio(entry.DiffRatio), entry.DiffRatio*100)
	fmt.Fprint(w, `</div></div>`)

	fmt.Fprint(w, `<div class="card-body"><div class="card-meta">`)
	fmt.Fprintf(w, `size baseline %dx%d, artifact %dx%d`, entry.RefWidth, entry.RefHeight, entry.ActWidth, entry.ActHeight)
	fmt.Fprint(w, `</div><div class="img-grid">`)

	fmt.Fprint(w, `<div class="img-col col-ref">`)
	fmt.Fprint(w, `<label>Baseline</label>`)
	fmt.Fprintf(
		w,
		`<img class="parity-image matte-target" src="data:image/png;base64,%s" alt="baseline" data-matte-document="%s" data-matte-white="%s" data-matte-dark="%s" data-matte-checkerboard="%s">`,
		entry.RefDocB64,
		entry.RefDocB64,
		entry.RefWhiteB64,
		entry.RefDarkB64,
		entry.RefCheckB64,
	)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-artifact">`)
	fmt.Fprint(w, `<label>Artifact</label>`)
	fmt.Fprintf(
		w,
		`<img class="parity-image matte-target" src="data:image/png;base64,%s" alt="artifact" data-matte-document="%s" data-matte-white="%s" data-matte-dark="%s" data-matte-checkerboard="%s">`,
		entry.ActDocB64,
		entry.ActDocB64,
		entry.ActWhiteB64,
		entry.ActDarkB64,
		entry.ActCheckB64,
	)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-overlay">`)
	fmt.Fprint(w, `<label>Overlay</label>`)
	fmt.Fprint(w, `<div class="slider-wrap">`)
	fmt.Fprintf(
		w,
		`<img class="base matte-target" src="data:image/png;base64,%s" alt="base" data-matte-document="%s" data-matte-white="%s" data-matte-dark="%s" data-matte-checkerboard="%s">`,
		entry.RefDocB64,
		entry.RefDocB64,
		entry.RefWhiteB64,
		entry.RefDarkB64,
		entry.RefCheckB64,
	)
	fmt.Fprintf(
		w,
		`<div class="slider-overlay"><img class="matte-target" src="data:image/png;base64,%s" alt="overlay" data-matte-document="%s" data-matte-white="%s" data-matte-dark="%s" data-matte-checkerboard="%s"></div>`,
		entry.ActDocB64,
		entry.ActDocB64,
		entry.ActWhiteB64,
		entry.ActDarkB64,
		entry.ActCheckB64,
	)
	fmt.Fprint(w, `<div class="slider-divider"></div></div></div>`)

	fmt.Fprint(w, `<div class="img-col col-amp">`)
	fmt.Fprint(w, `<label>Diff amplified</label>`)
	fmt.Fprintf(w, `<img class="parity-image" src="data:image/png;base64,%s" alt="amplified-diff">`, entry.AmpDiffB64)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `<div class="img-col col-raw">`)
	fmt.Fprint(w, `<label>Diff raw</label>`)
	fmt.Fprintf(w, `<img class="parity-image" src="data:image/png;base64,%s" alt="raw-diff">`, entry.RawDiffB64)
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `</div></div></div>`)
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}

const (
	badgeClassOK   = "badge-ok"
	badgeClassWarn = "badge-warn"
	badgeClassBad  = "badge-bad"
)

func badgeClassRMSE(v float64) string {
	if v <= 5 {
		return badgeClassOK
	}
	if v <= 20 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassAvgDiff(v float64) string {
	if v <= 2 {
		return badgeClassOK
	}
	if v <= 8 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassMaxDiff(v uint8) string {
	if v <= 10 {
		return badgeClassOK
	}
	if v <= 40 {
		return badgeClassWarn
	}
	return badgeClassBad
}

func badgeClassDiffRatio(v float64) string {
	if v <= 0.01 {
		return badgeClassOK
	}
	if v <= 0.05 {
		return badgeClassWarn
	}
	return badgeClassBad
}

const pageHeader = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Matplotlib-Go Parity Viewer</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #101216; color: #d7dce4; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 13px; }
.sticky-header { position: sticky; top: 0; z-index: 100; background: rgba(16,18,22,0.96); backdrop-filter: blur(10px); border-bottom: 1px solid #2d3440; padding: 10px 12px; }
.controls { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.controls h1 { font-size: 15px; color: #f4f7fb; margin-right: 8px; }
.controls input, .controls select { background: #171b22; color: #d7dce4; border: 1px solid #334155; padding: 5px 8px; font-family: inherit; font-size: 12px; border-radius: 6px; }
.controls label { display: flex; align-items: center; gap: 5px; font-size: 12px; color: #aeb8c8; }
.header-meta { margin-top: 8px; color: #93a1b5; font-size: 12px; }
#summary { margin-left: auto; color: #93a1b5; font-size: 12px; }
.container { padding: 12px; }
.card { background: #141922; border: 1px solid #293241; margin-bottom: 10px; border-radius: 10px; overflow: hidden; box-shadow: 0 8px 24px rgba(0,0,0,0.2); }
.card-header { padding: 9px 12px; cursor: pointer; display: flex; align-items: center; gap: 8px; background: #181f2a; user-select: none; }
.card-header:hover { background: #1d2531; }
.card-body { display: none; padding: 12px; }
.card.open .card-body { display: block; }
.card-title { font-size: 13px; color: #f4f7fb; flex: 1; }
.card-meta { color: #93a1b5; margin-bottom: 10px; }
.right-badges { display: flex; gap: 6px; align-items: center; flex-wrap: wrap; }
.badge { padding: 2px 7px; border-radius: 999px; font-size: 11px; font-weight: bold; }
.badge-neutral { background: #202734; color: #d7dce4; border: 1px solid #425168; }
.badge-ok { background: #163323; color: #7cf0a5; border: 1px solid #2e7250; }
.badge-warn { background: #3a2d13; color: #ffc66d; border: 1px solid #6f5521; }
.badge-bad { background: #3b1619; color: #ff8a8f; border: 1px solid #7c2d35; }
.img-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 10px; align-items: start; overflow-x: auto; }
.img-col { display: flex; flex-direction: column; gap: 6px; min-width: 0; overflow: auto; }
.img-col label { font-size: 11px; color: #93a1b5; text-align: center; }
.parity-image { display: block; image-rendering: auto; width: 100%; height: auto; border-radius: 6px; background-color: #0c1016; background-image: none; max-width: 100%; }
.resample-pixelated .parity-image { image-rendering: pixelated; }
.container.matte-document .parity-image, .container.matte-document .slider-wrap { background-color: #fff; }
.container.matte-white .parity-image, .container.matte-white .slider-wrap { background-color: #ffffff; }
.container.matte-dark .parity-image, .container.matte-dark .slider-wrap { background-color: #0c1016; }
.container.matte-checkerboard .parity-image, .container.matte-checkerboard .slider-wrap {
  background-color: #d8dee9;
  background-image:
    linear-gradient(45deg, #c3cad6 25%, transparent 25%, transparent 75%, #c3cad6 75%, #c3cad6),
    linear-gradient(45deg, #c3cad6 25%, transparent 25%, transparent 75%, #c3cad6 75%, #c3cad6);
  background-position: 0 0, 6px 6px;
  background-size: 12px 12px;
}
.original-size .img-grid { grid-template-columns: repeat(5, max-content); }
.original-size .img-col { min-width: max-content; }
.original-size .parity-image, .original-size .slider-wrap { align-self: flex-start; }
.original-size .parity-image { width: auto; height: auto; max-width: none; }
.col-raw, .col-amp { display: none; }
.slider-wrap { position: relative; overflow: hidden; width: 100%; cursor: col-resize; border-radius: 6px; background-color: #0c1016; background-image: none; }
.slider-wrap img.base { display: block; image-rendering: auto; width: 100%; height: auto; }
.resample-pixelated .slider-wrap img.base { image-rendering: pixelated; }
.original-size .slider-wrap { width: auto; }
.original-size .slider-wrap img.base { width: auto; height: auto; max-width: none; }
.slider-overlay { position: absolute; top: 0; left: 0; height: 100%; overflow: hidden; width: 50%; }
.slider-overlay img { display: block; position: absolute; top: 0; left: 0; image-rendering: auto; width: 200%; }
.resample-pixelated .slider-overlay img { image-rendering: pixelated; }
.original-size .slider-overlay img { width: auto !important; }
.slider-divider { position: absolute; top: 0; left: 50%; height: 100%; width: 3px; background: #f8fafc; cursor: col-resize; transform: translateX(-50%); }
.slider-divider::before {
  content: ''; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);
  width: 18px; height: 18px; border-radius: 50%; background: #f8fafc; border: 2px solid #111827;
}
.empty-state { border: 1px dashed #425168; border-radius: 10px; padding: 24px; color: #93a1b5; background: #141922; }
button.rerender-btn { background: #0f172a; color: #e2e8f0; border: 1px solid #334155; border-radius: 6px; padding: 4px 8px; cursor: pointer; font-family: inherit; font-size: 11px; }
button.rerender-btn:hover { background: #273244; }
button.rerender-btn:disabled { opacity: 0.6; cursor: wait; }
code { color: #f4f7fb; }
@media (max-width: 1400px) { .img-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 800px) { .img-grid { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<div class="sticky-header">
  <div class="controls">
    <h1>Matplotlib-Go Parity Viewer</h1>
    <input type="text" id="search" placeholder="Search case…" oninput="filterCards()" style="width:180px">
    <select id="sort-select" onchange="sortCards()">
      <option value="rmse-desc">Sort: RMSE ↓</option>
      <option value="rmse-asc">Sort: RMSE ↑</option>
      <option value="diff-pixels-desc">Sort: Different pixels ↓</option>
      <option value="diff-pixels-asc">Sort: Different pixels ↑</option>
      <option value="diff-ratio-desc">Sort: Diff ratio ↓</option>
      <option value="diff-ratio-asc">Sort: Diff ratio ↑</option>
      <option value="avg-diff-desc">Sort: Avg diff ↓</option>
      <option value="avg-diff-asc">Sort: Avg diff ↑</option>
      <option value="name-asc">Sort: Name ↑</option>
    </select>
    <select id="diff-mode" onchange="setDiffMode(this.value)">
      <option value="amp">Diff: amplified</option>
      <option value="raw">Diff: raw</option>
      <option value="both">Diff: both</option>
    </select>
    <select id="matte-mode" onchange="setMatteMode(this.value)">
      <option value="document" selected>Matte: document bg</option>
      <option value="white">Matte: white</option>
      <option value="dark">Matte: dark</option>
      <option value="checkerboard">Matte: checkerboard</option>
    </select>
    <select id="resample-mode" onchange="setResampleMode(this.value)">
      <option value="smooth">Scaling: smooth</option>
      <option value="pixelated">Scaling: pixelated</option>
    </select>
    <button id="rerender-all-btn" class="rerender-btn" type="button">Re-render Artifacts</button>
    <label><input type="checkbox" id="original-size" onchange="setOriginalSize(this.checked)"> Original size</label>
    <span id="summary"></span>
  </div>
`

const pageFooter = `</div>
<script>
(function() {
  function metric(card, attr) {
    return parseFloat(card.dataset[attr] || 0);
  }

  function filterCards() {
    var q = document.getElementById('search').value.toLowerCase();
    document.querySelectorAll('.card').forEach(function(card) {
      var name = (card.dataset.name || '').toLowerCase();
      card.style.display = name.indexOf(q) >= 0 ? '' : 'none';
    });
    updateSummary();
  }

  function sortCards() {
    var mode = document.getElementById('sort-select').value;
    var container = document.getElementById('cards-container');
    var cards = Array.from(container.querySelectorAll('.card'));
    cards.sort(function(a, b) {
      if (mode === 'rmse-desc') return metric(b, 'rmse') - metric(a, 'rmse');
      if (mode === 'rmse-asc') return metric(a, 'rmse') - metric(b, 'rmse');
      if (mode === 'diff-pixels-desc') return metric(b, 'diffPixels') - metric(a, 'diffPixels');
      if (mode === 'diff-pixels-asc') return metric(a, 'diffPixels') - metric(b, 'diffPixels');
      if (mode === 'diff-ratio-desc') return metric(b, 'diffRatio') - metric(a, 'diffRatio');
      if (mode === 'diff-ratio-asc') return metric(a, 'diffRatio') - metric(b, 'diffRatio');
      if (mode === 'avg-diff-desc') return metric(b, 'avgDiff') - metric(a, 'avgDiff');
      if (mode === 'avg-diff-asc') return metric(a, 'avgDiff') - metric(b, 'avgDiff');
      return (a.dataset.name || '').localeCompare(b.dataset.name || '');
    });
    cards.forEach(function(card) { container.appendChild(card); });
    updateSortMetricBadges(mode);
  }

  function badgeColorClass(attr, value, diffRatio) {
    if (attr === 'rmse') {
      if (value <= 5) return 'badge-ok';
      if (value <= 20) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'avgDiff') {
      if (value <= 2) return 'badge-ok';
      if (value <= 8) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'maxDiff') {
      if (value <= 10) return 'badge-ok';
      if (value <= 40) return 'badge-warn';
      return 'badge-bad';
    }
    if (attr === 'diffPixels' || attr === 'diffRatio') {
      var ratio = parseFloat(diffRatio || 0);
      if (ratio <= 0.01) return 'badge-ok';
      if (ratio <= 0.05) return 'badge-warn';
      return 'badge-bad';
    }
    return 'badge-neutral';
  }

  function updateSortMetricBadges(mode) {
    var label = '';
    var attr = '';
    var format = function(v) { return String(v); };
    if (mode === 'rmse-desc' || mode === 'rmse-asc') {
      label = 'RMSE'; attr = 'rmse'; format = function(v) { return Number(v).toFixed(2); };
    } else if (mode === 'diff-pixels-desc' || mode === 'diff-pixels-asc') {
      label = 'diff px'; attr = 'diffPixels'; format = function(v) { return String(Math.round(Number(v))); };
    } else if (mode === 'diff-ratio-desc' || mode === 'diff-ratio-asc') {
      label = 'diff %'; attr = 'diffRatio'; format = function(v) { return (Number(v) * 100).toFixed(2); };
    } else if (mode === 'avg-diff-desc' || mode === 'avg-diff-asc') {
      label = 'avg'; attr = 'avgDiff'; format = function(v) { return Number(v).toFixed(2); };
    }
    document.querySelectorAll('.card').forEach(function(card) {
      var badge = card.querySelector('.sort-metric-badge');
      if (!badge) return;
      if (!attr) {
        badge.style.display = 'none';
        return;
      }
      var value = parseFloat(card.dataset[attr] || 0);
      badge.className = 'badge ' + badgeColorClass(attr, value, card.dataset.diffRatio) + ' sort-metric-badge';
      badge.textContent = label + ' ' + format(value);
      badge.style.display = '';
    });
  }

  function setDiffMode(mode) {
    document.querySelectorAll('.col-amp').forEach(function(el) {
      el.style.display = (mode === 'amp' || mode === 'both') ? 'flex' : 'none';
    });
    document.querySelectorAll('.col-raw').forEach(function(el) {
      el.style.display = (mode === 'raw' || mode === 'both') ? 'flex' : 'none';
    });
  }

  function updateSummary() {
    var all = document.querySelectorAll('.card');
    var visible = Array.from(all).filter(function(card) { return card.style.display !== 'none'; });
    document.getElementById('summary').textContent = visible.length + ' / ' + all.length + ' cases';
  }

  function setMatteMode(mode) {
    var dataKey = 'matte' + mode.charAt(0).toUpperCase() + mode.slice(1);
    document.querySelectorAll('.matte-target').forEach(function(img) {
      var b64 = img.dataset[dataKey];
      if (!b64) return;
      img.src = 'data:image/png;base64,' + b64;
    });
  }

  function setResampleMode(mode) {
    var container = document.getElementById('cards-container');
    container.classList.remove('resample-smooth', 'resample-pixelated');
    container.classList.add(mode === 'pixelated' ? 'resample-pixelated' : 'resample-smooth');
  }

  function setOriginalSize(on) {
    var container = document.getElementById('cards-container');
    if (on) container.classList.add('original-size'); else container.classList.remove('original-size');
  }

  function setRerenderButtonsDisabled(disabled) {
    document.querySelectorAll('.rerender-btn').forEach(function(button) {
      button.disabled = disabled;
    });
  }

  function rerenderArtifact(name) {
    return fetch('/rerender', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
      body: new URLSearchParams({ name: name }).toString()
    }).then(function(response) {
      if (!response.ok) {
        return response.text().then(function(text) {
          throw new Error(text || 'rerender failed');
        });
      }
    });
  }

  document.querySelectorAll('.card-header').forEach(function(header) {
    header.addEventListener('click', function() {
      header.closest('.card').classList.toggle('open');
    });
  });

  document.querySelectorAll('.rerender-btn').forEach(function(button) {
    if (button.id === 'rerender-all-btn') {
      return;
    }
    button.addEventListener('click', function(event) {
      event.stopPropagation();
      var name = button.dataset.name || '';
      if (!name) return;
      setRerenderButtonsDisabled(true);
      rerenderArtifact(name).then(function() {
        window.location.reload();
      }).catch(function(err) {
        window.alert(err.message);
        setRerenderButtonsDisabled(false);
      });
    });
  });

  var bulkButton = document.getElementById('rerender-all-btn');
  bulkButton.addEventListener('click', function() {
    setRerenderButtonsDisabled(true);
    var cards = Array.from(document.querySelectorAll('.card'));
    var chain = Promise.resolve();
    cards.forEach(function(card) {
      var name = card.dataset.name || '';
      if (!name) return;
      chain = chain.then(function() { return rerenderArtifact(name); });
    });
    chain.then(function() {
      window.location.reload();
    }).catch(function(err) {
      window.alert(err.message);
      setRerenderButtonsDisabled(false);
    });
  });

  document.querySelectorAll('.slider-wrap').forEach(function(wrap) {
    var divider = wrap.querySelector('.slider-divider');
    var overlay = wrap.querySelector('.slider-overlay');
    var dragging = false;
    function setPos(x) {
      var rect = wrap.getBoundingClientRect();
      var pct = Math.max(0, Math.min(1, (x - rect.left) / rect.width));
      overlay.style.width = (pct * 100) + '%';
      divider.style.left = (pct * 100) + '%';
      if (pct > 0) {
        overlay.querySelector('img').style.width = (100 / pct) + '%';
      } else {
        overlay.querySelector('img').style.width = '100%';
      }
    }
    divider.addEventListener('mousedown', function(e) {
      dragging = true;
      e.preventDefault();
    });
    document.addEventListener('mousemove', function(e) { if (dragging) { setPos(e.clientX); } });
    document.addEventListener('mouseup', function() { dragging = false; });
    wrap.addEventListener('click', function(e) { setPos(e.clientX); });
    setPos(wrap.getBoundingClientRect().width / 2);
  });

  window.filterCards = filterCards;
  window.sortCards = sortCards;
  window.setDiffMode = setDiffMode;
  window.setMatteMode = setMatteMode;
  window.setOriginalSize = setOriginalSize;
  window.setResampleMode = setResampleMode;

  sortCards();
  setDiffMode(document.getElementById('diff-mode').value);
  setMatteMode(document.getElementById('matte-mode').value);
  setResampleMode(document.getElementById('resample-mode').value);
  filterCards();
})();
</script>
</body>
</html>
`
