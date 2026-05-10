package bar_basic_title

import (
	"github.com/cwbudde/matplotlib-go/examples/parity/internal/common"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, true, true)
}
