package agg

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type texManagerConfig struct {
	CacheDir      string
	LaTeXCommand  string
	DVIPNGCommand string
}

type texManager struct {
	mu            sync.Mutex
	cacheDir      string
	latexCommand  string
	dvipngCommand string
}

type texRenderResult struct {
	PNGPath string
	Image   *image.RGBA
	Metrics render.TextMetrics
}

func newTeXManager(cfg texManagerConfig) *texManager {
	if cfg.CacheDir == "" {
		cfg.CacheDir = defaultTeXCacheDir()
	}
	if cfg.LaTeXCommand == "" {
		cfg.LaTeXCommand = "latex"
	}
	if cfg.DVIPNGCommand == "" {
		cfg.DVIPNGCommand = "dvipng"
	}
	return &texManager{
		cacheDir:      cfg.CacheDir,
		latexCommand:  cfg.LaTeXCommand,
		dvipngCommand: cfg.DVIPNGCommand,
	}
}

func defaultTeXCacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil && dir != "" {
		return filepath.Join(dir, "matplotlib-go", "tex.cache")
	}
	return filepath.Join(os.TempDir(), "matplotlib-go", "tex.cache")
}

func (m *texManager) render(text string, size float64, dpi uint, fontKey string) (texRenderResult, error) {
	if m == nil {
		return texRenderResult{}, errors.New("agg: TeX manager is unavailable")
	}
	if strings.TrimSpace(text) == "" || size <= 0 {
		return texRenderResult{}, nil
	}
	if dpi == 0 {
		dpi = 72
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	pngPath, err := m.makePNG(text, size, dpi, fontKey)
	if err != nil {
		return texRenderResult{}, err
	}
	img, err := decodePNG(pngPath)
	if err != nil {
		return texRenderResult{}, err
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	return texRenderResult{
		PNGPath: pngPath,
		Image:   img,
		Metrics: render.TextMetrics{
			W:       float64(w),
			H:       float64(h),
			Ascent:  float64(h),
			Descent: 0,
		},
	}, nil
}

func (m *texManager) makeDVI(text string, size float64, fontKey string) (string, error) {
	source := texSource(text, size, fontKey)
	base, err := m.basePath(source, 0)
	if err != nil {
		return "", err
	}
	dviPath := base + ".dvi"
	if fileExists(dviPath) {
		return dviPath, nil
	}

	tmp, err := os.MkdirTemp(filepath.Dir(dviPath), "tex-build-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	texPath := filepath.Join(tmp, "file.tex")
	if err := os.WriteFile(texPath, []byte(source), 0o644); err != nil {
		return "", err
	}
	if err := runTeXCommand([]string{m.latexCommand, "-interaction=nonstopmode", "-halt-on-error", "-no-shell-escape", "file.tex"}, text, tmp); err != nil {
		return "", err
	}
	tmpDVI := filepath.Join(tmp, "file.dvi")
	if !fileExists(tmpDVI) {
		return "", fmt.Errorf("agg: latex did not produce %s for TeX string %q", tmpDVI, text)
	}
	if err := os.Rename(tmpDVI, dviPath); err != nil {
		return "", err
	}
	_ = os.Rename(texPath, base+".tex")
	return dviPath, nil
}

func (m *texManager) makePNG(text string, size float64, dpi uint, fontKey string) (string, error) {
	source := texSource(text, size, fontKey)
	base, err := m.basePath(source, dpi)
	if err != nil {
		return "", err
	}
	pngPath := base + ".png"
	if fileExists(pngPath) {
		return pngPath, nil
	}

	dviPath, err := m.makeDVI(text, size, fontKey)
	if err != nil {
		return "", err
	}

	tmp, err := os.MkdirTemp(filepath.Dir(pngPath), "tex-raster-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	cmd := []string{
		m.dvipngCommand,
		"-bg", "Transparent",
		"-D", strconv.FormatUint(uint64(dpi), 10),
		"-T", "tight",
		"-o", "file.png",
		dviPath,
	}
	if err := runTeXCommand(cmd, text, tmp); err != nil {
		return "", err
	}
	tmpPNG := filepath.Join(tmp, "file.png")
	if !fileExists(tmpPNG) {
		return "", fmt.Errorf("agg: dvipng did not produce %s for TeX string %q", tmpPNG, text)
	}
	if err := os.Rename(tmpPNG, pngPath); err != nil {
		return "", err
	}
	return pngPath, nil
}

func (m *texManager) basePath(source string, dpi uint) (string, error) {
	key := source + "\ndpi=" + strconv.FormatUint(uint64(dpi), 10)
	sum := sha256.Sum256([]byte(key))
	hash := hex.EncodeToString(sum[:])
	dir := filepath.Join(m.cacheDir, hash[:2], hash[2:4])
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, hash), nil
}

func runTeXCommand(command []string, text, cwd string) error {
	if len(command) == 0 || command[0] == "" {
		return errors.New("agg: empty TeX command")
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("agg: failed to process TeX string %q because %s could not be found", text, command[0])
	}
	return fmt.Errorf("agg: %s failed while processing TeX string %q\ncommand: %s\noutput:\n%s", filepath.Base(command[0]), text, strings.Join(command, " "), strings.TrimSpace(string(out)))
}

func texSource(text string, size float64, fontKey string) string {
	baselineSkip := size * 1.25
	fontCommand := texFontCommand(fontKey)
	return strings.Join([]string{
		`\RequirePackage{fix-cm}`,
		`\documentclass{article}`,
		`\newcommand{\mathdefault}[1]{#1}`,
		`\usepackage[utf8]{inputenc}`,
		`\DeclareUnicodeCharacter{2212}{\ensuremath{-}}`,
		`\usepackage[papersize=72in, margin=1in]{geometry}`,
		usePackageIfNotLoaded("underscore", "strings"),
		usePackageIfNotLoaded("textcomp", ""),
		`\pagestyle{empty}`,
		`\begin{document}`,
		fmt.Sprintf(`\fontsize{%s}{%s}%%`, formatTeXFloat(size), formatTeXFloat(baselineSkip)),
		`\ifdefined\psfrag\else\hbox{}\fi%`,
		fmt.Sprintf(`{%s %s}%%`, fontCommand, text),
		`\end{document}`,
		"",
	}, "\n")
}

func usePackageIfNotLoaded(pkg, option string) string {
	if option != "" {
		option = "[" + option + "]"
	}
	return fmt.Sprintf(`\makeatletter\@ifpackageloaded{%s}{}{\usepackage%s{%s}}\makeatother`, pkg, option, pkg)
}

func texFontCommand(fontKey string) string {
	key := strings.ToLower(fontKey)
	switch {
	case strings.Contains(key, "mono"), strings.Contains(key, "courier"), strings.Contains(key, "typewriter"):
		return `\ttfamily`
	case strings.Contains(key, "sans"), strings.Contains(key, "helvetica"), strings.Contains(key, "arial"):
		return `\sffamily`
	default:
		return `\rmfamily`
	}
}

func formatTeXFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func decodePNG(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)
	return dst, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// LastTeXError returns the most recent TeX pipeline error recorded by MeasureTeX
// or DrawTeX. A nil value means the last TeX operation succeeded.
func (r *Renderer) LastTeXError() error {
	if r == nil {
		return nil
	}
	return r.texErr
}

// MeasureTeX measures a TeX string by rendering it through the external
// latex+dvipng cache and using the resulting tight PNG dimensions.
func (r *Renderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}
	return result.Metrics, true
}

// DrawTeX draws a TeX string through the external latex+dvipng cache.
func (r *Renderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	r.drawTeXImage(result, geom.Pt{X: origin.X, Y: origin.Y - result.Metrics.Ascent}, textColor)
	return true
}

// DrawTeXRotated draws a TeX string rotated around the Matplotlib-style text
// rotation anchor.
func (r *Renderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}

	metrics := result.Metrics
	bounds := render.TextBounds{X: 0, Y: -metrics.Ascent, W: metrics.W, H: metrics.H}
	origin := rotatedTextOrigin(anchor, metrics, bounds, true)
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	r.drawTeXImageRotated(result, topLeft, anchor, angle, textColor)
	return true
}

func (r *Renderer) renderTeX(text string, size float64, fontKey string) (texRenderResult, bool) {
	if r == nil || text == "" || size <= 0 {
		return texRenderResult{}, false
	}
	if r.texManager == nil {
		r.texManager = newTeXManager(texManagerConfig{})
	}
	result, err := r.texManager.render(text, size, r.resolution, fontKey)
	if err != nil {
		r.texErr = err
		return texRenderResult{}, false
	}
	r.texErr = nil
	return result, true
}

func (r *Renderer) drawTeXImage(result texRenderResult, topLeft geom.Pt, textColor render.Color) {
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return
	}
	r.Image(render.NewImageData(img), geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	})
}

func (r *Renderer) drawTeXImageRotated(result texRenderResult, topLeft, anchor geom.Pt, angle float64, textColor render.Color) {
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return
	}
	cos := math.Cos(-angle)
	sin := math.Sin(-angle)
	affine := translateAffineAgg(anchor).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffineAgg(geom.Pt{X: -anchor.X, Y: -anchor.Y})).
		Mul(translateAffineAgg(topLeft))
	r.ImageTransformed(render.NewImageData(img), geom.Rect{}, affine)
}

func colorizeTeXImage(src *image.RGBA, c render.Color) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	r := uint8(clamp01(c.R)*255 + 0.5)
	g := uint8(clamp01(c.G)*255 + 0.5)
	b := uint8(clamp01(c.B)*255 + 0.5)
	alphaScale := clamp01(c.A)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a16 := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			a := uint8(float64(a16>>8)*alphaScale + 0.5)
			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}

func translateAffineAgg(p geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: p.X, F: p.Y}
}
