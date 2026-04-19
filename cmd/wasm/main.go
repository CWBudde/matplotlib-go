//go:build js && wasm

package main

import (
	"syscall/js"

	"matplotlib-go/internal/webdemo"
)

var callbacks []js.Func

func main() {
	callbacks = append(callbacks,
		js.FuncOf(listDemos),
		js.FuncOf(renderDemo),
		js.FuncOf(defaultDemoID),
	)

	api := js.Global().Get("Object").New()
	api.Set("listDemos", callbacks[0])
	api.Set("renderDemo", callbacks[1])
	api.Set("defaultDemoID", callbacks[2])
	js.Global().Set("matplotlibGoWASM", api)
	js.Global().Get("console").Call("log", "matplotlib-go wasm ready")

	select {}
}

func listDemos(_ js.Value, _ []js.Value) any {
	result := js.Global().Get("Array").New()
	for _, descriptor := range webdemo.Catalog() {
		item := js.Global().Get("Object").New()
		item.Set("id", descriptor.ID)
		item.Set("title", descriptor.Title)
		item.Set("description", descriptor.Description)
		result.Call("push", item)
	}
	return result
}

func renderDemo(_ js.Value, args []js.Value) any {
	id := webdemo.DefaultDemoID()
	width := webdemo.DefaultWidth
	height := webdemo.DefaultHeight

	if len(args) > 0 && args[0].Type() == js.TypeString && webdemo.ValidDemoID(args[0].String()) {
		id = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeNumber {
		width = args[1].Int()
	}
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		height = args[2].Int()
	}

	img, descriptor, err := webdemo.Render(id, width, height)
	if err != nil {
		result := js.Global().Get("Object").New()
		result.Set("error", err.Error())
		return result
	}

	pixels := js.Global().Get("Uint8ClampedArray").New(len(img.Pix))
	js.CopyBytesToJS(pixels, img.Pix)

	result := js.Global().Get("Object").New()
	result.Set("id", descriptor.ID)
	result.Set("title", descriptor.Title)
	result.Set("description", descriptor.Description)
	result.Set("width", img.Bounds().Dx())
	result.Set("height", img.Bounds().Dy())
	result.Set("pixels", pixels)
	return result
}

func defaultDemoID(_ js.Value, _ []js.Value) any {
	return webdemo.DefaultDemoID()
}
