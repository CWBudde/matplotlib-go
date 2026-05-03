//go:build freetype

package agg

/*
#cgo pkg-config: freetype2
#include <stdlib.h>
#include <ft2build.h>
#include FT_FREETYPE_H

static FT_Int32 mpl_go_force_autohint_load_flags(void) {
	return FT_LOAD_DEFAULT | FT_LOAD_FORCE_AUTOHINT;
}

static int mpl_go_pixel_mode_gray(void) {
	return FT_PIXEL_MODE_GRAY;
}

static int mpl_go_pixel_mode_mono(void) {
	return FT_PIXEL_MODE_MONO;
}

static int mpl_go_has_kerning(FT_Face face) {
	return FT_HAS_KERNING(face);
}

static void mpl_go_freetype_version(FT_Library library, FT_Int *major, FT_Int *minor, FT_Int *patch) {
	FT_Library_Version(library, major, minor, patch);
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"unsafe"

	agglib "github.com/cwbudde/agg_go"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) drawNativeFreetypeText(text string, face render.FontFace, origin geom.Pt, size float64, textColor render.Color) bool {
	if r.ctx == nil || text == "" || face.Path == "" || size <= 0 {
		return false
	}

	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return false
	}
	defer C.FT_Done_FreeType(library)

	path := C.CString(face.Path)
	defer C.free(unsafe.Pointer(path))

	var ftFace C.FT_Face
	if C.FT_New_Face(library, path, 0, &ftFace) != 0 {
		return false
	}
	defer C.FT_Done_Face(ftFace)

	dpi := r.resolution
	if dpi == 0 {
		dpi = 72
	}
	charSize := C.FT_F26Dot6(math.Round(size * 64.0))
	if C.FT_Set_Char_Size(ftFace, 0, charSize, C.FT_UInt(dpi), C.FT_UInt(dpi)) != 0 {
		return false
	}

	var matrix C.FT_Matrix
	matrix.xx = 0x10000
	matrix.yy = 0x10000

	penX := C.FT_Pos(0)
	baselineY := float64(r.height) - origin.Y
	loadFlags := C.mpl_go_force_autohint_load_flags()
	paint := renderColorToRGBA(textColor)
	uniform := image.NewUniform(paint)
	var previousGlyph C.FT_UInt
	drewGlyph := false

	for _, rr := range text {
		glyphIndex := C.FT_Get_Char_Index(ftFace, C.FT_ULong(rr))
		if glyphIndex == 0 {
			return false
		}
		if previousGlyph != 0 && C.mpl_go_has_kerning(ftFace) != 0 {
			var kerning C.FT_Vector
			if C.FT_Get_Kerning(ftFace, previousGlyph, glyphIndex, C.FT_KERNING_DEFAULT, &kerning) == 0 {
				penX += kerning.x
			}
		}
		delta := C.FT_Vector{
			x: C.FT_Pos(math.Round((origin.X + float64(penX)/64.0) * 64.0)),
			y: C.FT_Pos(math.Round(baselineY * 64.0)),
		}
		C.FT_Set_Transform(ftFace, &matrix, &delta)
		if C.FT_Load_Glyph(ftFace, glyphIndex, loadFlags) != 0 {
			return false
		}

		slot := ftFace.glyph
		if C.FT_Render_Glyph(slot, C.FT_RENDER_MODE_NORMAL) != 0 {
			return false
		}
		bitmap := slot.bitmap
		width := int(bitmap.width)
		height := int(bitmap.rows)
		if width > 0 && height > 0 && bitmap.buffer != nil {
			mask, ok := freetypeBitmapMask(bitmap)
			if !ok {
				return false
			}
			src := image.NewRGBA(mask.Bounds())
			draw.DrawMask(src, src.Bounds(), uniform, image.Point{}, mask, image.Point{}, draw.Over)
			img, err := agglib.NewImageFromStandardImage(src)
			if err != nil {
				return false
			}
			x := float64(slot.bitmap_left)
			y := float64(r.height - int(slot.bitmap_top))
			if err := r.ctx.DrawImageScaled(img, x, y, float64(width), float64(height)); err != nil {
				return false
			}
			drewGlyph = true
		}

		penX += slot.advance.x
		previousGlyph = glyphIndex
	}

	return drewGlyph
}

func nativeFreetypeVersion() string {
	var library C.FT_Library
	if C.FT_Init_FreeType(&library) != 0 {
		return ""
	}
	defer C.FT_Done_FreeType(library)

	var major, minor, patch C.FT_Int
	C.mpl_go_freetype_version(library, &major, &minor, &patch)
	return fmt.Sprintf("%d.%d.%d", int(major), int(minor), int(patch))
}

func freetypeBitmapMask(bitmap C.FT_Bitmap) (*image.Alpha, bool) {
	width := int(bitmap.width)
	height := int(bitmap.rows)
	pitch := int(bitmap.pitch)
	if width <= 0 || height <= 0 || pitch == 0 || bitmap.buffer == nil {
		return nil, false
	}

	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	bufferLen := absInt(pitch) * height
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bitmap.buffer)), bufferLen)
	pixelMode := int(bitmap.pixel_mode)

	for row := 0; row < height; row++ {
		srcRow := row * pitch
		if pitch < 0 {
			srcRow = (height - 1 - row) * -pitch
		}
		dstRow := row * mask.Stride
		switch pixelMode {
		case int(C.mpl_go_pixel_mode_gray()):
			copy(mask.Pix[dstRow:dstRow+width], buffer[srcRow:srcRow+width])
		case int(C.mpl_go_pixel_mode_mono()):
			for col := 0; col < width; col++ {
				byteIndex := srcRow + col/8
				bit := byte(1 << uint(7-col%8))
				if buffer[byteIndex]&bit != 0 {
					mask.Pix[dstRow+col] = 0xff
				}
			}
		default:
			return nil, false
		}
	}
	return mask, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
