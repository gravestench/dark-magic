package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	assetcatalog "github.com/gravestench/dark-magic/internal/assets/catalog"
	"github.com/gravestench/dark-magic/internal/content"
)

// writeVerificationReport persists the primary deterministic report at its established filename and returns the host
// path used by the CLI summary.
func writeVerificationReport(outputDirectory string, report assetcatalog.Report) (string, error) {
	reportPath := filepath.Join(outputDirectory, "report.json")
	if err := writeIndentedJSON(reportPath, report); err != nil {
		return "", err
	}

	return reportPath, nil
}

// writeStructuralFixture refuses partial reports through FixtureFromReport before creating the requested JSON file,
// preventing an incomplete installation from becoming a trusted structural baseline.
func writeStructuralFixture(name string, report assetcatalog.Report) error {
	fixture, err := assetcatalog.FixtureFromReport(report)
	if err != nil {
		return err
	}

	return writeJSON(name, fixture)
}

// verifyStructuralFixture loads the expected fingerprint and reports every mismatch together so users can distinguish
// broad archive-version drift from an isolated asset difference.
func verifyStructuralFixture(name string, report assetcatalog.Report) error {
	var fixture assetcatalog.Fixture
	if err := readJSON(name, &fixture); err != nil {
		return err
	}

	mismatches := assetcatalog.CompareFixture(report, fixture)
	if len(mismatches) == 0 {
		return nil
	}

	return errors.New("fixture mismatch:\n  " + strings.Join(mismatches, "\n  "))
}

// auditCommunityListfile parses, resolves, and persists the optional community inventory as one ordered phase. The
// report is written only after the input closes successfully, preserving input-error precedence.
func auditCommunityListfile(
	contentFS *content.FS,
	listfilePath string,
	outputDirectory string,
) (assetcatalog.ListfileReport, string, error) {
	entries, err := readListfileEntries(listfilePath)
	if err != nil {
		return assetcatalog.ListfileReport{}, "", err
	}

	audit := assetcatalog.AuditListfile(contentFS, entries)

	auditPath := filepath.Join(outputDirectory, "listfile-report.json")
	if err := writeIndentedJSON(auditPath, audit); err != nil {
		return assetcatalog.ListfileReport{}, "", err
	}

	return audit, auditPath, nil
}

// readListfileEntries closes the input before returning and gives parse failures precedence over close failures,
// matching the command's historical diagnostic order.
func readListfileEntries(name string) ([]assetcatalog.ListedPath, error) {
	listfile, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	entries, parseErr := assetcatalog.ParseListfile(listfile)
	closeErr := listfile.Close()

	if parseErr != nil {
		return nil, parseErr
	}

	if closeErr != nil {
		return nil, closeErr
	}

	return entries, nil
}

// readJSON decodes a complete host file in one operation, leaving schema validation to the destination contract.
func readJSON(name string, destination any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, destination)
}

// writeJSON creates parent directories before encoding optional user-requested artifacts such as fixture files.
func writeJSON(name string, value any) error {
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}

	return writeIndentedJSON(name, value)
}

// writeIndentedJSON always closes the destination and returns encoding failures before close failures, preserving the
// established error precedence while keeping every report in the same review-friendly format.
func writeIndentedJSON(name string, value any) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(value)
	closeErr := file.Close()

	if encodeErr != nil {
		return encodeErr
	}

	return closeErr
}
