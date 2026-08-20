// Command defense_outcome_probe validates and normalizes sanitized visual
// observations from an owned Expansion 1.14d runtime. It records evidence for
// attack/block/avoid ordering; it does not implement or infer that policy.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

const (
	probeSchema = "d2legacy.defense_outcome_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

// main keeps command ownership visible from input selection through report emission, including the process exit codes.
func main() {
	inputPath := flag.String("input", "", "sanitized owned-runtime defense-outcome probe JSON")

	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: defense_outcome_probe -input <capture.json>")
		os.Exit(2)
	}

	input, err := os.Open(*inputPath)
	if err != nil {
		exitWithError(err)
	}
	defer func() {
		// The read-only close cannot invalidate an emitted report, so preserve the original non-fatal cleanup behavior.
		_ = input.Close()
	}()

	result, err := analyze(input)
	if err != nil {
		exitWithError(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		exitWithError(err)
	}
}

// exitWithError preserves the command's one-line stderr contract and failure status for every fatal error path.
func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
