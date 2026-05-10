package bar_basic_tick_labels

import (
	"github.com/cwbudde/matplotlib-go/examples/parity/internal/common"
	"image"
)

func Render() image.Image {
	return common.RenderBarBasicScaffold(true, true, false)
}
