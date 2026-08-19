// Command mpq2file extracts one asset from a directory, ZIP, or MPQ source.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

const usage = "usage: mpq2file -source <directory-or-archive> -asset <asset> [-output file]"

type options struct {
	sourcePath string
	assetPath  string
	outputPath string
}

// main keeps source ownership visible around the read and write phases so normal completion closes archives last.
func main() {
	config := parseOptions()

	source, err := content.OpenSource(config.sourcePath)
	if err != nil {
		fatal(err)
	}
	// Register ownership as soon as the source opens. fatal exits directly, so only normal completion runs this cleanup.
	defer content.Close(source)

	data, err := fs.ReadFile(source, config.assetPath)
	if err != nil {
		fatal(err)
	}

	if err := writeAsset(config.outputPath, data, os.Stdout); err != nil {
		fatal(err)
	}
}

// parseOptions retains the process-wide flag package's help and error behavior while enforcing both required paths.
func parseOptions() options {
	sourcePath := flag.String("source", "", "directory, ZIP, or MPQ containing the asset")
	assetPath := flag.String("asset", "", "asset path inside the source")
	outputPath := flag.String("output", "", "output file; omit to write bytes to stdout")

	flag.Parse()

	config := options{
		sourcePath: *sourcePath,
		assetPath:  *assetPath,
		outputPath: *outputPath,
	}
	if !config.hasRequiredPaths() {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	return config
}

// hasRequiredPaths centralizes the CLI's only cross-flag invariant so invalid invocations keep exit status 2.
func (o options) hasRequiredPaths() bool {
	return o.sourcePath != "" && o.assetPath != ""
}

// writeAsset selects stdout for an empty destination or creates the expanded host path with the established mode.
func writeAsset(outputPath string, data []byte, stdout io.Writer) error {
	if outputPath == "" {
		_, err := stdout.Write(data)

		return err
	}

	expandedOutput, err := darkpaths.ExpandHost(outputPath)
	if err != nil {
		return err
	}

	return os.WriteFile(expandedOutput, data, 0o644)
}

// fatal preserves the command's error prefix and exits immediately; deferred cleanup does not run after this call.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "mpq2file:", err)
	os.Exit(1)
}
