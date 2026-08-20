// Command missile_audio_probe validates and normalizes sanitized audio/video
// observations from an owned Expansion 1.14d runtime. It resolves when and how
// often record-referenced missile sounds occur without importing older engine,
// server, save, memory-tool, or community behavior.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// main owns the input file and output stream for one probe invocation. Errors remain fatal so callers never mistake
// a partial or invalid report for a successful observation.
func main() {
	input := flag.String("input", "", "sanitized owned-runtime missile-audio probe JSON")

	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: missile_audio_probe -input <capture.json>")
		os.Exit(2)
	}

	file, err := os.Open(*input)
	if err != nil {
		fatal(err)
	}
	// The input is read-only, so a close failure cannot invalidate the report or alter established CLI output.
	defer func() {
		_ = file.Close()
	}()

	result, err := analyze(file)
	if err != nil {
		fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		fatal(err)
	}
}

// fatal reports a command failure and exits immediately, preserving the probe's nonzero failure contract.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
