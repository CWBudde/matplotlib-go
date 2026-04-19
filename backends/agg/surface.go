package agg

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"sync"

	agglib "github.com/cwbudde/agg_go"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var sfntFontCache sync.Map

type sfntFontResource struct {
	font *sfnt.Font
	data []byte
}

type fontHeightMetrics struct {
	ascent  float64
	descent float64
	lineGap float64
}

// aggSurface owns the explicit AGG image buffer and the renderer attached to it.
// This keeps backend construction local instead of routing through agglib.Context.
type aggSurface struct {
	image          *agglib.Image
	textContext    *agglib.Context
	painter        *agglib.Agg2D
	textFontPath   string
	textSize       float64
	textResolution uint
	textHinting    bool
	textForceAuto  bool
	textRoundOff   bool
	textReady      bool
}

func newAggSurface(w, h int) *aggSurface {
	stride := w * 4
	img := agglib.NewImage(make([]uint8, h*stride), w, h, stride)
	textContext := agglib.NewContextForImage(img)
	painter := textContext.GetAgg2D()
	painter.FillColor(agglib.Black)
	painter.LineColor(agglib.Black)
	painter.LineWidth(1.0)
	painter.ClipBox(0, 0, float64(w), float64(h))

	return &aggSurface{
		image:         img,
		textContext:   textContext,
		painter:       painter,
		textHinting:   true,
		textForceAuto: false,
		textRoundOff:  false,
	}
}

func (s *aggSurface) Clear(c agglib.Color) {
	s.painter.ClearAll(c)
}

func (s *aggSurface) SetResolution(dpi uint) {
	if dpi == 0 {
		return
	}
	if s.textResolution != dpi {
		s.textReady = false
	}
	s.textResolution = dpi
}

func (s *aggSurface) PushTransform() {
	s.painter.PushTransform()
}

func (s *aggSurface) PopTransform() bool {
	return s.painter.PopTransform()
}

func (s *aggSurface) Translate(tx, ty float64) {
	s.painter.Translate(tx, ty)
}

func (s *aggSurface) Rotate(angle float64) {
	s.painter.Rotate(angle)
}

func (s *aggSurface) ClipBox(x1, y1, x2, y2 float64) {
	s.painter.ClipBox(x1, y1, x2, y2)
}

func (s *aggSurface) BeginPath() {
	s.painter.ResetPath()
}

func (s *aggSurface) MoveTo(x, y float64) {
	s.painter.MoveTo(x, y)
}

func (s *aggSurface) LineTo(x, y float64) {
	s.painter.LineTo(x, y)
}

func (s *aggSurface) QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64) {
	s.painter.QuadricCurveTo(xCtrl, yCtrl, xTo, yTo)
}

func (s *aggSurface) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	s.painter.CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

func (s *aggSurface) ClosePath() {
	s.painter.ClosePolygon()
}

func (s *aggSurface) Fill() {
	s.painter.DrawPath(agglib.FillOnly)
}

func (s *aggSurface) Stroke() {
	s.painter.DrawPath(agglib.StrokeOnly)
}

func (s *aggSurface) SetFillColor(c agglib.Color) {
	s.painter.FillColor(c)
}

func (s *aggSurface) SetStrokeColor(c agglib.Color) {
	s.painter.LineColor(c)
}

func (s *aggSurface) SetStrokeWidth(width float64) {
	s.painter.LineWidth(width)
}

func (s *aggSurface) SetLineJoin(join agglib.LineJoin) {
	s.painter.LineJoin(join)
}

func (s *aggSurface) SetLineCap(cap agglib.LineCap) {
	s.painter.LineCap(cap)
}

func (s *aggSurface) SetMiterLimit(limit float64) {
	s.painter.MiterLimit(limit)
}

func (s *aggSurface) ClearDashes() {
	s.painter.RemoveAllDashes()
}

func (s *aggSurface) SetDashPattern(pattern []float64) {
	s.ClearDashes()
	for i := 0; i < len(pattern)-1; i += 2 {
		gapLen := pattern[i]
		if i+1 < len(pattern) {
			gapLen = pattern[i+1]
		}
		s.painter.AddDash(pattern[i], gapLen)
	}
}

func (s *aggSurface) GetImage() *agglib.Image {
	return s.image
}

func (s *aggSurface) SetImageFilter(filter agglib.ImageFilter) {
	s.painter.ImageFilter(filter)
}

func (s *aggSurface) GetImageFilter() agglib.ImageFilter {
	return s.painter.GetImageFilter()
}

func (s *aggSurface) SetImageResample(resample agglib.ImageResample) {
	s.painter.ImageResample(resample)
}

func (s *aggSurface) GetImageResample() agglib.ImageResample {
	return s.painter.GetImageResample()
}

func (s *aggSurface) DrawImageScaled(img *agglib.Image, x, y, width, height float64) error {
	if img == nil {
		return errors.New("image is nil")
	}

	return s.painter.TransformImageSimple(img, x, y, x+width, y+height)
}

func (s *aggSurface) DrawImageTransformed(img *agglib.Image, transform *agglib.Transformations) error {
	if img == nil {
		return errors.New("image is nil")
	}
	if transform == nil {
		return s.DrawImageScaled(img, 0, 0, float64(img.Width()), float64(img.Height()))
	}

	w, h := float64(img.Width()), float64(img.Height())
	corners := [][2]float64{
		{0, 0},
		{w, 0},
		{w, h},
		{0, h},
	}

	for i, corner := range corners {
		x, y := transform.Transform(corner[0], corner[1])
		corners[i] = [2]float64{x, y}
	}

	parallelogram := []float64{
		corners[0][0], corners[0][1],
		corners[1][0], corners[1][1],
		corners[3][0], corners[3][1],
	}

	return s.painter.TransformImageParallelogram(img, 0, 0, img.Width(), img.Height(), parallelogram)
}

func (s *aggSurface) ConfigureTextFont(fontPath string, size float64, resolution uint) error {
	if fontPath == "" {
		return errors.New("font path is empty")
	}
	if size <= 0 {
		return errors.New("font size must be positive")
	}
	if resolution == 0 {
		resolution = 72
	}
	if s.textReady && s.textFontPath == fontPath && s.textSize == size && s.textResolution == resolution {
		return nil
	}
	if s.textContext == nil {
		return errors.New("text context is unavailable")
	}

	s.textContext.SetResolution(resolution)
	s.textContext.FlipText(true)
	s.textContext.TextHints(s.textHinting)
	s.textContext.TextForceAutohint(s.textForceAuto)
	if err := s.textContext.Font(fontPath, size, false, false, agglib.RasterFontCache, 0); err != nil {
		s.textReady = false
		return err
	}

	s.textFontPath = fontPath
	s.textSize = size
	s.textResolution = resolution
	s.textReady = true
	return nil
}

func (s *aggSurface) TextWidth(text string) float64 {
	if s.textContext == nil {
		return 0
	}
	return s.textContext.GetTextWidth(text)
}

func (s *aggSurface) TextMetrics(text string) (width, ascent, descent float64) {
	if s.textContext == nil || text == "" {
		return 0, 0, 0
	}

	width = s.textContext.GetTextWidth(text)
	if metrics, ok := s.fontHeightMetrics(); ok {
		if width <= 0 {
			_, _, width, _ = s.textContext.GetTextBounds(text)
		}
		return width, metrics.ascent, metrics.descent
	}

	x, y, width, height := s.textContext.GetTextBounds(text)
	maxY := y + height
	if y < 0 {
		ascent = -y
	}
	if maxY > 0 {
		descent = maxY
	}
	if width <= 0 {
		width = s.textContext.GetTextWidth(text)
	}
	if ascent <= 0 && descent <= 0 {
		ascent = s.TextAscent()
		descent = s.TextDescent()
	}
	_ = x
	return width, ascent, descent
}

func (s *aggSurface) fontHeightMetrics() (fontHeightMetrics, bool) {
	if s.textFontPath == "" || s.textSize <= 0 {
		return fontHeightMetrics{}, false
	}
	resolution := s.textResolution
	if resolution == 0 {
		resolution = 72
	}

	resource, err := loadSFNTFont(s.textFontPath)
	if err != nil {
		return fontHeightMetrics{}, false
	}
	scale, ok := sfntMetricScale(resource.data, s.textSize, float64(resolution))
	if !ok {
		ppem := fixed.Int26_6(math.Round(s.textSize * float64(resolution) * 64.0 / 72.0))
		if ppem <= 0 {
			return fontHeightMetrics{}, false
		}

		metrics, err := resource.font.Metrics(nil, ppem, xfont.HintingNone)
		if err != nil {
			return fontHeightMetrics{}, false
		}
		return fontHeightMetrics{
			ascent:  float64(metrics.Ascent) / 64.0,
			descent: float64(metrics.Descent) / 64.0,
		}, metrics.Ascent > 0 || metrics.Descent > 0
	}

	if ascent, descent, lineGap, ok := sfntTableHeightMetrics(resource.data, scale); ok {
		return fontHeightMetrics{
			ascent:  ascent,
			descent: descent,
			lineGap: lineGap,
		}, true
	}

	return fontHeightMetrics{}, false
}

func loadSFNTFont(path string) (*sfntFontResource, error) {
	if cached, ok := sfntFontCache.Load(path); ok {
		return cached.(*sfntFontResource), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fontData, err := sfnt.Parse(data)
	if err != nil {
		return nil, err
	}

	actual, _ := sfntFontCache.LoadOrStore(path, &sfntFontResource{font: fontData, data: data})
	return actual.(*sfntFontResource), nil
}

func sfntMetricScale(data []byte, size, dpi float64) (float64, bool) {
	head := sfntTableData(data, "head")
	if len(head) < 20 {
		return 0, false
	}
	unitsPerEm := binary.BigEndian.Uint16(head[18:20])
	if unitsPerEm == 0 {
		return 0, false
	}
	return size * dpi / 72.0 / float64(unitsPerEm), true
}

func sfntTableHeightMetrics(data []byte, scale float64) (ascent, descent, lineGap float64, ok bool) {
	if table := sfntTableData(data, "OS/2"); len(table) >= 74 {
		return scale * float64(sfntInt16(table[68:70])),
			scale * float64(-sfntInt16(table[70:72])),
			scale * float64(sfntInt16(table[72:74])),
			true
	}
	if table := sfntTableData(data, "hhea"); len(table) >= 10 {
		return scale * float64(sfntInt16(table[4:6])),
			scale * float64(-sfntInt16(table[6:8])),
			scale * float64(sfntInt16(table[8:10])),
			true
	}
	return 0, 0, 0, false
}

func sfntTableData(data []byte, tag string) []byte {
	if len(data) < 12 || len(tag) != 4 {
		return nil
	}
	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	dirOffset := 12
	for i := 0; i < numTables; i++ {
		entryOffset := dirOffset + 16*i
		if entryOffset+16 > len(data) {
			return nil
		}
		if string(data[entryOffset:entryOffset+4]) != tag {
			continue
		}
		offset := int(binary.BigEndian.Uint32(data[entryOffset+8 : entryOffset+12]))
		length := int(binary.BigEndian.Uint32(data[entryOffset+12 : entryOffset+16]))
		if offset < 0 || length < 0 || offset+length > len(data) {
			return nil
		}
		return data[offset : offset+length]
	}
	return nil
}

func sfntInt16(b []byte) int16 {
	return int16(binary.BigEndian.Uint16(b))
}

func (s *aggSurface) TextBounds(text string) (x, y, width, height float64) {
	if s.textContext == nil || text == "" {
		return 0, 0, 0, 0
	}
	return s.textContext.GetTextBounds(text)
}

func (s *aggSurface) TextAscent() float64 {
	if s.textContext == nil {
		return 0
	}
	ascent := s.textContext.GetAscender()
	if ascent <= 0 {
		return s.textContext.FontHeight()
	}
	return ascent
}

func (s *aggSurface) TextDescent() float64 {
	if s.textContext == nil {
		return 0
	}
	desc := -s.textContext.GetDescender()
	if desc < 0 {
		return 0
	}
	return desc
}

func (s *aggSurface) DrawText(text string, x, y float64) {
	if s.painter == nil {
		return
	}
	// Agg2D text rendering can pick up prior stroke state from path drawing.
	// Reset to neutral text-safe defaults before issuing the text draw.
	s.painter.LineWidth(1.0)
	s.painter.LineCap(agglib.CapButt)
	s.painter.LineJoin(agglib.JoinMiter)
	s.painter.ResetPath()
	s.painter.Text(x, y, text, s.textRoundOff, 0, 0)
}
