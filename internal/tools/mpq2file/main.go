// Command mpq2file extracts one asset from a directory, ZIP, or MPQ source.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/pkg/paths"
)

func main() {
	sourcePath := flag.String("source", "", "directory, ZIP, or MPQ containing the asset")
	assetPath := flag.String("asset", "", "asset path inside the source")
	outputPath := flag.String("output", "", "output file; omit to write bytes to stdout")
	flag.Parse()
	if *sourcePath == "" || *assetPath == "" {
		fmt.Fprintln(os.Stderr, "usage: mpq2file -source <directory-or-archive> -asset <asset> [-output file]")
		os.Exit(2)
	}
	source, err := content.OpenSource(*sourcePath)
	if err != nil {
		fatal(err)
	}
	defer content.Close(source)
	data, err := fs.ReadFile(source, *assetPath)
	if err != nil {
		fatal(err)
	}
	if *outputPath == "" {
		if _, err := os.Stdout.Write(data); err != nil {
			fatal(err)
		}
		return
	}
	expandedOutput, err := darkpaths.ExpandHost(*outputPath)
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(expandedOutput, data, 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mpq2file:", err)
	os.Exit(1)
}
