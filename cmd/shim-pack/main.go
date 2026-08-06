package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
)

func main() {
	output := flag.String("output", "dist/darkmagic.zip", "output ZIP path")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fail(err)
	}
	file, err := os.Create(*output)
	if err != nil {
		fail(err)
	}
	if err := content.WriteShimArchive(file); err != nil {
		_ = file.Close()
		fail(err)
	}
	if err := file.Close(); err != nil {
		fail(err)
	}
}

func fail(err error) { fmt.Fprintf(os.Stderr, "shim-pack: %v\n", err); os.Exit(1) }
