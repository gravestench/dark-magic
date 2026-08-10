package assetcatalog

import (
	"testing"
	"testing/fstest"
)

func TestBuildCoverageClassifiesManifestCodeAndFixturePaths(t *testing.T) {
	source := fstest.MapFS{
		"manifests/presentation.v1.json": {Data: []byte(`{"screen":{"sheet":"data/ui/known.dc6","other":"data/ui/new.dc6"}}`)},
		"boot.lua":                       {Data: []byte(`local a = "data/ui/known.dc6"`)},
		"lua/screen.lua":                 {Data: []byte(`local a = "data/ui/code-only.dc6"; local b = "data/ui/spells/" .. id .. ".dc6"`)},
	}
	manifest := Manifest{Version: 1, Assets: []Hypothesis{{ID: "known", Screen: "test", Path: "DATA/UI/KNOWN.DC6", Meaning: "known"}, {ID: "old", Screen: "test", Path: "data/ui/old.dc6", Meaning: "old"}}}
	fixture := Fixture{Version: 1, ManifestVersion: 1, Assets: []AssetFixture{{ID: "known", Path: "data/ui/known.dc6"}, {ID: "old", Path: "data/ui/old.dc6"}}}
	coverage, err := BuildCoverage(source, manifest, fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(coverage.VerifiedPresentation) != 1 || coverage.VerifiedPresentation[0] != "data/ui/known.dc6" {
		t.Fatalf("verified = %#v", coverage.VerifiedPresentation)
	}
	if len(coverage.UnverifiedPresentation) != 1 || coverage.UnverifiedPresentation[0] != "data/ui/new.dc6" {
		t.Fatalf("unverified = %#v", coverage.UnverifiedPresentation)
	}
	if len(coverage.CodeOwnedPaths) != 1 || coverage.CodeOwnedPaths[0] != "data/ui/code-only.dc6" {
		t.Fatalf("code owned = %#v", coverage.CodeOwnedPaths)
	}
	if len(coverage.DynamicPrefixes) != 1 || coverage.DynamicPrefixes[0] != "data/ui/spells/" {
		t.Fatalf("dynamic = %#v", coverage.DynamicPrefixes)
	}
	if len(coverage.CatalogOnlyPaths) != 1 || coverage.CatalogOnlyPaths[0] != "data/ui/old.dc6" || len(coverage.CatalogFixtureGaps) != 0 {
		t.Fatalf("catalog-only=%#v gaps=%#v", coverage.CatalogOnlyPaths, coverage.CatalogFixtureGaps)
	}
}
