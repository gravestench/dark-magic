package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

func main() {
	output := flag.String("output", "dist/d2.zip", "output ZIP path")
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}
	expandedOutput, err := darkpaths.ExpandHost(*output)
	if err != nil {
		fail(err)
	}
	if err := os.MkdirAll(filepath.Dir(expandedOutput), 0o755); err != nil {
		fail(err)
	}
	file, err := os.Create(expandedOutput)
	if err != nil {
		fail(err)
	}
	if err := content.WriteD2LegacyArchive(file); err != nil {
		_ = file.Close()
		fail(err)
	}
	if err := file.Close(); err != nil {
		fail(err)
	}
}

func fail(err error) { fmt.Fprintf(os.Stderr, "d2legacy-pack: %v\n", err); os.Exit(1) }
