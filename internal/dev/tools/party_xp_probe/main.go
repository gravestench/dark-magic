// Command party_xp_probe validates and normalizes sanitized observations from
// an owned expansion 1.14d runtime. It does not emulate party XP: its output is
// evidence used to choose exact distance and integer-ordering vectors later.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

// main keeps process concerns at the boundary so capture validation and analysis remain independently testable.
func main() {
	inputPath := flag.String("input", "", "sanitized owned-runtime party-XP probe JSON")

	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: party_xp_probe -input <capture.json>")
		os.Exit(2)
	}

	input, err := os.Open(*inputPath)
	if err != nil {
		fatal(err)
	}
	// The process cannot recover from a read-only close failure after analysis, so cleanup is explicitly best-effort.
	defer func() {
		_ = input.Close()
	}()

	result, err := analyze(input)
	if err != nil {
		fatal(err)
	}

	if err := writeReport(os.Stdout, result); err != nil {
		fatal(err)
	}
}

// writeReport emits stable, indented JSON so captures remain both machine-readable and practical to review as evidence.
func writeReport(output io.Writer, result report) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")

	return encoder.Encode(result)
}

// fatal reserves stdout for valid report JSON and gives every operational failure the same exit status.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
