//go:build js && wasm

package main

import (
	"syscall/js"

	plotcanvas "matplotlib-go/canvas"
	wasmcanvas "matplotlib-go/canvas/wasm"
	"matplotlib-go/internal/webdemo"
)

var callbacks []js.Func
var currentManager plotcanvas.FigureManager

func main() {
	callbacks = append(callbacks,
		js.FuncOf(listDemos),
		js.FuncOf(mountDemo),
		js.FuncOf(resizeDemo),
		js.FuncOf(unmountDemo),
		js.FuncOf(defaultDemoID),
	)

	api := js.Global().Get("Object").New()
	api.Set("listDemos", callbacks[0])
	api.Set("mountDemo", callbacks[1])
	api.Set("resizeDemo", callbacks[2])
	api.Set("unmountDemo", callbacks[3])
	api.Set("defaultDemoID", callbacks[4])
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

func mountDemo(_ js.Value, args []js.Value) any {
	canvasID := "plotCanvas"
	id := webdemo.DefaultDemoID()
	width := webdemo.DefaultWidth
	height := webdemo.DefaultHeight

	if len(args) > 0 && args[0].Type() == js.TypeString && args[0].String() != "" {
		canvasID = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeString && webdemo.ValidDemoID(args[1].String()) {
		id = args[1].String()
	}
	if len(args) > 2 && args[2].Type() == js.TypeNumber {
		width = args[2].Int()
	}
	if len(args) > 3 && args[3].Type() == js.TypeNumber {
		height = args[3].Int()
	}

	return loadDemo(canvasID, id, width, height)
}

func resizeDemo(_ js.Value, args []js.Value) any {
	result := js.Global().Get("Object").New()
	if currentManager == nil {
		result.Set("error", "no mounted demo")
		return result
	}

	width := webdemo.DefaultWidth
	height := webdemo.DefaultHeight
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		width = args[0].Int()
	}
	if len(args) > 1 && args[1].Type() == js.TypeNumber {
		height = args[1].Int()
	}
	if err := currentManager.Canvas().Resize(width, height); err != nil {
		result.Set("error", err.Error())
		return result
	}
	result.Set("width", width)
	result.Set("height", height)
	return result
}

func unmountDemo(_ js.Value, _ []js.Value) any {
	if currentManager != nil {
		_ = currentManager.Close()
	}
	currentManager = nil
	return nil
}

func defaultDemoID(_ js.Value, _ []js.Value) any {
	return webdemo.DefaultDemoID()
}

func loadDemo(canvasID, id string, width, height int) any {
	result := js.Global().Get("Object").New()

	fig, descriptor, err := webdemo.Build(id, width, height)
	if err != nil {
		result.Set("error", err.Error())
		return result
	}
	if currentManager != nil {
		_ = currentManager.Close()
	}

	manager, err := wasmcanvas.NewGoBasicManager(canvasID, fig)
	if err != nil {
		result.Set("error", err.Error())
		return result
	}
	currentManager = manager
	if err := manager.Canvas().Resize(width, height); err != nil {
		result.Set("error", err.Error())
		return result
	}

	result.Set("id", descriptor.ID)
	result.Set("title", descriptor.Title)
	result.Set("description", descriptor.Description)
	result.Set("width", width)
	result.Set("height", height)
	return result
}
