package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type commandOptions struct {
	mpqDirectory    string
	manifestPath    string
	listfilePath    string
	outputDirectory string
	noSheets        bool
	writeFixture    string
	fixturePath     string
}

// main keeps process termination at the command boundary so the workflow can return failures without changing
// the CLI's fail-fast exit behavior.
func main() {
	if err := executeCommand(parseCommandOptions()); err != nil {
		fatal(err.Error())
	}
}

// parseCommandOptions defines the command's existing flag contract in one place, keeping defaults and help text
// visible without mixing them into path validation or catalog verification.
func parseCommandOptions() commandOptions {
	mpqDirectory := flag.String("mpq-dir", os.Getenv("MPQ_DIRECTORY"), "directory containing Diablo II MPQ files")
	manifestPath := flag.String("manifest", "", "optional JSON manifest; defaults to the curated screen catalog")
	listfilePath := flag.String("listfile", "", "optional community MPQ listfile to audit against this installation")
	outputDirectory := flag.String("out", "asset-catalog", "output directory for report.json and contact sheets")
	noSheets := flag.Bool("no-sheets", false, "skip DC6 contact sheet generation")
	writeFixture := flag.String("write-fixture", "", "write a structural fixture after every manifest asset verifies")
	fixturePath := flag.String("fixture", "", "validate the installation against a structural fixture")

	flag.Parse()

	return commandOptions{
		mpqDirectory:    *mpqDirectory,
		manifestPath:    *manifestPath,
		listfilePath:    *listfilePath,
		outputDirectory: *outputDirectory,
		noSheets:        *noSheets,
		writeFixture:    *writeFixture,
		fixturePath:     *fixturePath,
	}
}

// expandHostPaths normalizes every user-owned filesystem path before any I/O. Expansion remains ordered by the
// historical flag sequence so the first reported configuration error does not change.
func (options commandOptions) expandHostPaths() (commandOptions, error) {
	if options.mpqDirectory == "" {
		return commandOptions{}, errors.New("-mpq-dir or MPQ_DIRECTORY is required")
	}

	var err error

	options.mpqDirectory, err = expandHostPath(options.mpqDirectory)
	if err != nil {
		return commandOptions{}, err
	}

	options.manifestPath, err = expandHostPath(options.manifestPath)
	if err != nil {
		return commandOptions{}, err
	}

	options.outputDirectory, err = expandHostPath(options.outputDirectory)
	if err != nil {
		return commandOptions{}, err
	}

	options.listfilePath, err = expandHostPath(options.listfilePath)
	if err != nil {
		return commandOptions{}, err
	}

	options.writeFixture, err = expandHostPath(options.writeFixture)
	if err != nil {
		return commandOptions{}, err
	}

	options.fixturePath, err = expandHostPath(options.fixturePath)
	if err != nil {
		return commandOptions{}, err
	}

	if options.writeFixture != "" && options.fixturePath != "" {
		return commandOptions{}, errors.New("-write-fixture and -fixture are mutually exclusive")
	}

	return options, nil
}

// executeCommand runs verification and optional follow-up reports in their established order. Earlier output remains
// visible when a later fixture or listfile phase fails, which is part of the command's diagnostic behavior.
func executeCommand(options commandOptions) error {
	expandedOptions, err := options.expandHostPaths()
	if err != nil {
		return err
	}

	contentFS, report, err := verifyCatalog(expandedOptions)
	if err != nil {
		return err
	}

	reportPath, err := writeVerificationReport(expandedOptions.outputDirectory, report)
	if err != nil {
		return err
	}

	found := foundHypothesisCount(report)
	fmt.Printf("verified %d/%d hypotheses; report: %s\n", found, len(report.Results), reportPath)

	if expandedOptions.writeFixture != "" {
		if err := writeStructuralFixture(expandedOptions.writeFixture, report); err != nil {
			return err
		}

		fmt.Printf("wrote structural fixture: %s\n", expandedOptions.writeFixture)
	}

	if expandedOptions.fixturePath != "" {
		if err := verifyStructuralFixture(expandedOptions.fixturePath, report); err != nil {
			return err
		}

		fmt.Printf("fixture verified: %s\n", expandedOptions.fixturePath)
	}

	if expandedOptions.listfilePath != "" {
		audit, auditPath, err := auditCommunityListfile(
			contentFS,
			expandedOptions.listfilePath,
			expandedOptions.outputDirectory,
		)
		if err != nil {
			return err
		}

		fmt.Printf("resolved %d/%d listed paths; report: %s\n", audit.Found, audit.Listed, auditPath)
	}

	return nil
}

// fatal emits one normalized diagnostic and exits immediately, preserving the command's established handling for
// empty error strings as well as ordinary failures.
func fatal(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = fs.ErrInvalid.Error()
	}

	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
