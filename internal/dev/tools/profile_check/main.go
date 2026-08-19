// Command profile_check verifies scene diagnostic snapshots against tracked
// performance budgets. It intentionally consumes profiler artifacts rather
// than starting the graphical client itself.
package main

import (
	"flag"
	"fmt"
	"os"
)

// main keeps process concerns at the command boundary so validation remains reusable and independently testable.
// Reporting every discovered violation before exiting gives CI one complete, deterministic failure report.
func main() {
	profileDirectory := flag.String("profile-dir", "./profiles/acceptance", "profiling artifact directory")
	budgetPath := flag.String("budgets", "./docs/profile-budgets.json", "scene budget JSON")
	flag.Parse()

	if err := check(*profileDirectory, *budgetPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
