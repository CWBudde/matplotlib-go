//go:build !js || !wasm

package main

import "fmt"

func main() {
	fmt.Println("cmd/wasm is only supported for GOOS=js GOARCH=wasm")
}
