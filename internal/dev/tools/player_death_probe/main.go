// Command player_death_probe validates and normalizes sanitized visual
// observations from a probe-created character in an owned Expansion 1.14d
// runtime. It records death/corpse consequence evidence; it never reads or
// writes retail save data and does not implement inferred gameplay policy.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// main keeps operating-system concerns at the command boundary so capture analysis remains deterministic and testable.
func main() {
	inputPath := flag.String("input", "", "sanitized owned-runtime player-death probe JSON")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: player_death_probe -input <capture.json>")
		os.Exit(2)
	}

	file, err := os.Open(*inputPath)
	if err != nil {
		fatal(err)
	}
	// Analysis consumes the file synchronously; deferring closure keeps ownership visible without changing error output.
	defer file.Close()

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

// fatal writes command failures to stderr so stdout is either a complete JSON report or empty.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
