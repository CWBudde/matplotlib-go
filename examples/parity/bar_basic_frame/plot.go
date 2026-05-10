package bar_basic_frame

import (
	"github.com/cwbudde/matplotlib-go/examples/parity/internal/common"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(false, false, false)
}
