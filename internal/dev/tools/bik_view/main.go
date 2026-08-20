// Command bik_view validates and plays a BIK from a file, directory, ZIP, MPQ,
// or standard Diablo II MPQ directory without permanently extracting it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// commandOptions records the two supported input modes and the player selected by the caller.
// Keeping parsing separate from validation makes the compatibility rules visible without changing flag behavior.
type commandOptions struct {
	fileName   string
	sourceName string
	assetName  string
	playerName string
}

// main coordinates the viewer phases and translates every failure into the command's stable one-line error format.
// Temporary extraction cleanup remains registered here so fatal exits retain their established defer behavior.
func main() {
	options := parseCommandOptions()
	if err := validateCommandOptions(options); err != nil {
		fatal(err)
	}

	playable, cleanup, err := resolvePlayableBIK(options)
	if err != nil {
		fatal(err)
	}

	// fatal exits bypass defers; normal return remains the only path that removes a successful extraction.
	defer cleanup()

	expanded, err := validateAndDescribeBIK(playable, os.Stdout)
	if err != nil {
		fatal(err)
	}

	if err := playBIK(options.playerName, expanded); err != nil {
		fatal(err)
	}
}

// parseCommandOptions registers and parses the command's historical flags on flag.CommandLine.
// Using the default flag set preserves standard help, parse-error, and process-exit behavior.
func parseCommandOptions() commandOptions {
	fileName := flag.String("file", "", "standalone BIK file")
	sourceName := flag.String("source", "", "directory, ZIP, MPQ, or MPQ directory")
	assetName := flag.String("asset", "", "BIK path inside the source")
	playerName := flag.String("player", "ffplay", "ffplay-compatible executable")

	flag.Parse()

	return commandOptions{
		fileName:   *fileName,
		sourceName: *sourceName,
		assetName:  *assetName,
		playerName: *playerName,
	}
}

// validateCommandOptions requires exactly one complete input mode so downstream code can choose a source directly.
// A file selection still ignores one incomplete source flag, preserving the command's existing compatibility behavior.
func validateCommandOptions(options commandOptions) error {
	usesFile := options.fileName != ""

	usesSource := options.sourceName != "" && options.assetName != ""
	if usesFile == usesSource {
		return errors.New("use either -file <movie.bik> or -source <source> -asset <path>")
	}

	return nil
}

// fatal emits the command prefix once and exits immediately, which intentionally prevents deferred work after errors.
// Trimming the supplied error keeps wrapped multi-line or padded errors within the established single-message shape.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "bik_view:", strings.TrimSpace(err.Error()))
	os.Exit(1)
}
