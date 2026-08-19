// Command d2legacy_pack writes the embedded d2legacy content as a deterministic distribution archive.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

const (
	defaultOutputPath = "dist/d2legacy.zip"
	usageExitCode     = 2
	failureExitCode   = 1
)

// main limits command startup to flag validation and translating operational failures into process exits.
func main() {
	outputPath := flag.String("output", defaultOutputPath, "output ZIP path")

	flag.Parse()

	rejectPositionalArguments()

	if err := writeArchiveFile(*outputPath); err != nil {
		fail(err)
	}
}

// rejectPositionalArguments preserves the command's flag-only interface and its usage exit contract.
func rejectPositionalArguments() {
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(usageExitCode)
	}
}

// writeArchiveFile resolves the host path and creates its parent before transferring file ownership to the writer.
func writeArchiveFile(outputName string) error {
	outputPath, err := darkpaths.ExpandHost(outputName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	destination, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	return writeAndCloseArchive(destination)
}

// writeAndCloseArchive always releases its destination while preserving archive-write errors over cleanup errors.
func writeAndCloseArchive(destination io.WriteCloser) error {
	// The write failure describes the invalid archive; a secondary cleanup failure must not replace it.
	if err := content.WriteD2LegacyArchive(destination); err != nil {
		_ = destination.Close()
		return err
	}

	return destination.Close()
}

// fail emits the command's stable error prefix and exits with the operational-failure status expected by callers.
func fail(err error) {
	fmt.Fprintf(os.Stderr, "d2legacy-pack: %v\n", err)
	os.Exit(failureExitCode)
}
