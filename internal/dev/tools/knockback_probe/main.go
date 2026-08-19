// Command knockback_probe validates and normalizes sanitized observations from
// an owned Expansion 1.14d runtime. It compares observations with explicitly
// labeled hypotheses; it does not promote older recovered behavior to policy.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// main keeps process concerns at the command boundary so the analysis path can
// return errors and remain directly testable without intercepting os.Exit.
func main() {
	inputPath := flag.String("input", "", "sanitized owned-runtime knockback probe JSON")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: knockback_probe -input <capture.json>")
		os.Exit(2)
	}

	file, err := os.Open(*inputPath)
	if err != nil {
		fatal(err)
	}
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

// fatal reports operational failures on stderr and exits with status one;
// missing command arguments use the distinct usage status in main.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
