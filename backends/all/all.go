// Package all imports the built-in rendering backends for side effects.
package all

import (
	_ "matplotlib-go/backends/agg"
	_ "matplotlib-go/backends/gobasic"
	_ "matplotlib-go/backends/skia"
)
