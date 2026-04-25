package core

import "matplotlib-go/internal/geom"

// AnchoredBoxLocator resolves a display-space box inside a clip rect.
type AnchoredBoxLocator interface {
	Rect(clip geom.Rect, width, height float64) geom.Rect
}

// AnchoredBoxLocatorFunc adapts a function into an AnchoredBoxLocator.
type AnchoredBoxLocatorFunc func(clip geom.Rect, width, height float64) geom.Rect

func (f AnchoredBoxLocatorFunc) Rect(clip geom.Rect, width, height float64) geom.Rect {
	if f == nil {
		return geom.Rect{}
	}
	return f(clip, width, height)
}

// BoxHorizontalAlign controls horizontal box anchoring around a locator point.
type BoxHorizontalAlign uint8

const (
	BoxAlignLeft BoxHorizontalAlign = iota
	BoxAlignCenter
	BoxAlignRight
)

// BoxVerticalAlign controls vertical box anchoring around a locator point.
type BoxVerticalAlign uint8

const (
	BoxAlignTop BoxVerticalAlign = iota
	BoxAlignMiddle
	BoxAlignBottom
)

// RelativeAnchoredBoxLocator places a box relative to a clip rect using
// normalized coordinates plus optional pixel offsets.
type RelativeAnchoredBoxLocator struct {
	X       float64
	Y       float64
	OffsetX float64
	OffsetY float64
	HAlign  BoxHorizontalAlign
	VAlign  BoxVerticalAlign
}

func (l RelativeAnchoredBoxLocator) Rect(clip geom.Rect, width, height float64) geom.Rect {
	anchor := geom.Pt{
		X: clip.Min.X + clip.W()*l.X + l.OffsetX,
		Y: clip.Min.Y + clip.H()*l.Y + l.OffsetY,
	}

	minX := anchor.X
	switch l.HAlign {
	case BoxAlignCenter:
		minX -= width / 2
	case BoxAlignRight:
		minX -= width
	}

	minY := anchor.Y
	switch l.VAlign {
	case BoxAlignMiddle:
		minY -= height / 2
	case BoxAlignBottom:
		minY -= height
	}

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: minX + width, Y: minY + height},
	}
}

// NewAnchoredOffsetLocator returns a locator anchored to one of the standard
// legend corners plus explicit pixel offsets.
func NewAnchoredOffsetLocator(location LegendLocation, inset, dx, dy float64) AnchoredBoxLocator {
	return AnchoredBoxLocatorFunc(func(clip geom.Rect, width, height float64) geom.Rect {
		box := anchoredBoxRect(clip, width, height, location, inset)
		box.Min.X += dx
		box.Max.X += dx
		box.Min.Y += dy
		box.Max.Y += dy
		return box
	})
}

func resolveAnchoredBoxRect(locator AnchoredBoxLocator, clip geom.Rect, width, height float64, location LegendLocation, inset float64) geom.Rect {
	if locator != nil {
		rect := locator.Rect(clip, width, height)
		if rect.Max.X > rect.Min.X && rect.Max.Y > rect.Min.Y {
			return rect
		}
	}
	return anchoredBoxRect(clip, width, height, location, inset)
}
