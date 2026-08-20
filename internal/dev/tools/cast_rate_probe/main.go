// Command cast_rate_probe validates and normalizes sanitized cast-timing
// observations from an owned Expansion 1.14d runtime. The report separates
// target evidence from candidate formulas, so older breakpoint tables cannot
// silently become Dark Magic behavior.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

// main owns process-level flag parsing and exit behavior so the analysis path remains usable in tests without exits.
func main() {
	inputPath := flag.String("input", "", "sanitized owned-runtime cast-rate probe JSON")

	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: cast_rate_probe -input <capture.json>")
		os.Exit(2)
	}

	input, err := os.Open(*inputPath)
	if err != nil {
		fatal(err)
	}
	defer input.Close() //nolint:errcheck // A read-only close cannot change a report that has already been emitted.

	result, err := analyze(input)
	if err != nil {
		fatal(err)
	}

	if err := writeReport(os.Stdout, result); err != nil {
		fatal(err)
	}
}

// writeReport centralizes the stable indented JSON format, including the encoder's trailing newline.
func writeReport(output io.Writer, result report) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")

	return encoder.Encode(result)
}

// fatal writes operational failures to stderr before terminating with the command's established exit status.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
