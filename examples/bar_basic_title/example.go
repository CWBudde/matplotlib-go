package bar_basic_title

import (
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, true, true)
}
