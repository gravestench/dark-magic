package assetcatalog

import (
	"errors"
	"testing"
	"testing/fstest"
)

// TestVerifyReportsFoundAndMissingAssetsWithoutStopping ensures one failed read remains local to its manifest entry.
// It also protects result ordering and the advisory provenance recorded for a successfully resolved asset.
func TestVerifyReportsFoundAndMissingAssetsWithoutStopping(t *testing.T) {
	manifest := Manifest{Version: 1, Assets: []Hypothesis{
		{ID: "found", Screen: "test", Path: "data/found.txt", Meaning: "fixture"},
		{ID: "missing", Screen: "test", Path: "data/missing.txt", Meaning: "fixture"},
	}}
	source := fstest.MapFS{"data/found.txt": {Data: []byte("screen knowledge")}}
	options := Options{Resolve: resolveFixtureSource}

	report := Verify(source, manifest, options)

	if len(report.Results) != 2 {
		t.Fatalf("got %d results", len(report.Results))
	}

	found := report.Results[0]
	if !found.Found || found.Source == nil || found.Source.Layer != "fixture.mpq" {
		t.Fatalf("unexpected found result: %+v", found)
	}

	missing := report.Results[1]
	if missing.Found || missing.Error == "" {
		t.Fatalf("unexpected missing result: %+v", missing)
	}
}

// resolveFixtureSource supplies provenance for the one path present in the test filesystem. Failing other lookups
// ensures advisory resolution cannot abort the scan before the filesystem records the missing entry.
func resolveFixtureSource(name string) (Source, error) {
	if name == "data/found.txt" {
		return Source{Layer: "fixture.mpq", Path: name}, nil
	}

	return Source{}, errors.New("missing")
}

// TestVerifyReportsByteIdenticalAliases protects manifest-order alias grouping for distinct paths with equal bytes.
func TestVerifyReportsByteIdenticalAliases(t *testing.T) {
	manifest := Manifest{Version: 1, Assets: []Hypothesis{
		{ID: "one", Screen: "test", Path: "one.bin", Meaning: "fixture"},
		{ID: "two", Screen: "test", Path: "two.bin", Meaning: "fixture alias"},
	}}
	source := fstest.MapFS{
		"one.bin": {Data: []byte("same")},
		"two.bin": {Data: []byte("same")},
	}

	report := Verify(source, manifest, Options{})

	if len(report.Aliases) != 1 || len(report.Aliases[0].Paths) != 2 {
		t.Fatalf("unexpected aliases: %+v", report.Aliases)
	}
}
