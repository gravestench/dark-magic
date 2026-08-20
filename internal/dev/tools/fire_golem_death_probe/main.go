// Command fire_golem_death_probe validates and normalizes an instrumented
// observation from an owned Expansion 1.14d runtime. It records the facts
// needed to implement Fire Golem's death splash without importing an older
// patch's general monster-death routine into current gameplay policy.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// main owns the command lifecycle so every input, analysis, and output failure retains its established exit behavior.
func main() {
	input := flag.String("input", "", "sanitized owned-runtime Fire Golem death probe JSON")

	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: fire_golem_death_probe -input <capture.json>")
		os.Exit(2)
	}

	file, err := os.Open(*input)
	if err != nil {
		fatal(err)
	}
	defer closeInput(file)

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

// closeInput releases the capture file at command exit while preserving the historical choice to ignore close errors.
func closeInput(file *os.File) {
	_ = file.Close()
}

// fatal reports an operational error without decoration and exits with the command's established failure status.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
