package agg

import (
	"errors"

	agglib "github.com/cwbudde/agg_go"
)

// aggSurface owns the explicit AGG image buffer and the renderer attached to it.
// This keeps backend construction local instead of routing through agglib.Context.
type aggSurface struct {
	image          *agglib.Image
	textContext    *agglib.Context
	painter        *agglib.Agg2D
	textFontPath   string
	textSize       float64
	textResolution uint
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
		image:       img,
		textContext: textContext,
		painter:     painter,
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
	s.textContext.TextHints(true)
	s.textContext.TextForceAutohint(true)
	s.textContext.TextHintingFactor(8)
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
	s.painter.Text(x, y, text, false, 0, 0)
}
