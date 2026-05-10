package bar_basic_title

import (
	"image"

	"github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, true, true)
}
