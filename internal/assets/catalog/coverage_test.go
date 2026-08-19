package assetcatalog

import (
	"slices"
	"testing"
	"testing/fstest"
)

// TestBuildCoverageClassifiesManifestCodeAndFixturePaths exercises every coverage category in one coherent fixture.
// Exact slice assertions protect both membership and deterministic ordering because fingerprints depend on both.
func TestBuildCoverageClassifiesManifestCodeAndFixturePaths(t *testing.T) {
	coverage, err := BuildCoverage(
		coverageTestSource(),
		coverageTestManifest(),
		coverageTestFixture(),
	)
	if err != nil {
		t.Fatal(err)
	}

	requireCoveragePaths(t, "verified", coverage.VerifiedPresentation, "data/ui/known.dc6")
	requireCoveragePaths(t, "unverified", coverage.UnverifiedPresentation, "data/ui/new.dc6")
	requireCoveragePaths(t, "code owned", coverage.CodeOwnedPaths, "data/ui/code-only.dc6")
	requireCoveragePaths(t, "dynamic", coverage.DynamicPrefixes, "data/ui/spells/")
	requireCoveragePaths(t, "catalog only", coverage.CatalogOnlyPaths, "data/ui/old.dc6")
	requireCoveragePaths(t, "fixture gaps", coverage.CatalogFixtureGaps)
}

// coverageTestSource models presentation, static Lua, and dynamic Lua references without proprietary game data.
func coverageTestSource() fstest.MapFS {
	return fstest.MapFS{
		"manifests/presentation.v1.json": {
			Data: []byte(`{"screen":{"sheet":"data/ui/known.dc6","other":"data/ui/new.dc6"}}`),
		},
		"boot.lua": {
			Data: []byte(`local a = "data/ui/known.dc6"`),
		},
		"lua/screen.lua": {
			Data: []byte(
				`local a = "data/ui/code-only.dc6"; local b = "data/ui/spells/" .. id .. ".dc6"`,
			),
		},
	}
}

// coverageTestManifest includes one presentation asset and one catalog-only asset to exercise both classifications.
func coverageTestManifest() Manifest {
	return Manifest{Version: 1, Assets: []Hypothesis{
		{ID: "known", Screen: "test", Path: "DATA/UI/KNOWN.DC6", Meaning: "known"},
		{ID: "old", Screen: "test", Path: "data/ui/old.dc6", Meaning: "old"},
	}}
}

// coverageTestFixture mirrors every catalog ID and normalized path so any reported gap indicates classification drift.
func coverageTestFixture() Fixture {
	return Fixture{Version: 1, ManifestVersion: 1, Assets: []AssetFixture{
		{ID: "known", Path: "data/ui/known.dc6"},
		{ID: "old", Path: "data/ui/old.dc6"},
	}}
}

// requireCoveragePaths compares complete ordered categories so unexpected additions cannot pass a first-item assertion.
func requireCoveragePaths(t *testing.T, name string, actual []string, expected ...string) {
	t.Helper()

	if !slices.Equal(actual, expected) {
		t.Fatalf("%s = %#v, want %#v", name, actual, expected)
	}
}
