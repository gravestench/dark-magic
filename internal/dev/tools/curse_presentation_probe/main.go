// Command curse_presentation_probe validates and normalizes sanitized visual
// observations from an owned Expansion 1.14d runtime. It identifies the
// client-function-30 presentation roles without reading memory/save data and
// without treating an older or community implementation as behavior evidence.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// main keeps command-line validation, capture analysis, and report emission in execution order. The distinct exit
// statuses remain part of the probe's scripting contract.
func main() {
	inputPath := flag.String("input", "", "sanitized owned-runtime curse-presentation probe JSON")

	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: curse_presentation_probe -input <capture.json>")
		os.Exit(2)
	}

	captureFile, err := os.Open(*inputPath)
	if err != nil {
		exitWithProbeError(err)
	}
	// The input is read-only and already consumed before exit, so a close failure must not replace report behavior.
	defer captureFile.Close() //nolint:errcheck

	result, err := analyze(captureFile)
	if err != nil {
		exitWithProbeError(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		exitWithProbeError(err)
	}
}

// exitWithProbeError writes operational failures to stderr and preserves the command's generic failure status.
func exitWithProbeError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
