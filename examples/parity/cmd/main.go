package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cwbudde/matplotlib-go/examples/parity"
)

func main() {
	id := flag.String("id", "", "Parity example ID to render")
	outputDir := flag.String("output-dir", ".", "Directory to write PNG files")
	list := flag.Bool("list", false, "List available parity example IDs and exit")
	all := flag.Bool("all", false, "Render every parity example")
	flag.Parse()

	if *list {
		for _, c := range parity.Cases() {
			fmt.Println(c.ID)
		}
		return
	}

	if *all {
		for _, c := range parity.Cases() {
			writeCase(c.ID, *outputDir)
		}
		return
	}

	if strings.TrimSpace(*id) == "" {
		exitf("missing --id, or use --all/--list")
	}
	writeCase(*id, *outputDir)
}

func writeCase(id, outputDir string) {
	path, err := parity.RenderToFile(id, outputDir)
	if err != nil {
		exitf("%v", err)
	}
	fmt.Printf("wrote %s\n", path)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
