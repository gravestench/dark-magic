package main

import (
	"os"
	"path/filepath"
	"testing"

	assetcatalog "github.com/gravestench/dark-magic/internal/assets/catalog"
)

// TestWriteJSONCreatesParentsAndPreservesFormat protects fixture directory creation, indentation, and the trailing
// newline as observable artifact-format behavior while also exercising the matching reader.
func TestWriteJSONCreatesParentsAndPreservesFormat(t *testing.T) {
	type document struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	wantValue := document{Name: "catalog", Count: 2}

	path := filepath.Join(t.TempDir(), "nested", "report.json")
	if err := writeJSON(path, wantValue); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	wantJSON := "{\n  \"name\": \"catalog\",\n  \"count\": 2\n}\n"
	if string(data) != wantJSON {
		t.Fatalf("written JSON = %q, want %q", data, wantJSON)
	}

	var decoded document
	if err := readJSON(path, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded != wantValue {
		t.Fatalf("decoded JSON = %+v, want %+v", decoded, wantValue)
	}
}

// TestWriteVerificationReportUsesEstablishedFilename ensures the primary report stays directly beneath the selected
// output directory so scripts and the printed summary retain their existing path contract.
func TestWriteVerificationReportUsesEstablishedFilename(t *testing.T) {
	outputDirectory := t.TempDir()

	reportPath, err := writeVerificationReport(outputDirectory, assetcatalog.Report{})
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(outputDirectory, "report.json")
	if reportPath != want {
		t.Fatalf("report path = %q, want %q", reportPath, want)
	}
}
