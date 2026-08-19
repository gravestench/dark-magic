package assetcatalog

import "fmt"

// CompareFixture returns every expected-asset mismatch in fixture order. Reporting all differences in one pass makes
// archive-version drift reviewable without allowing unexpected report-only assets to alter the fixture contract.
func CompareFixture(report Report, fixture Fixture) []string {
	if err := fixture.Validate(); err != nil {
		return []string{err.Error()}
	}

	actual := indexReportResults(report)

	var mismatches []string

	for _, expected := range fixture.Assets {
		if mismatch := compareFixtureAsset(actual, expected); mismatch != "" {
			mismatches = append(mismatches, mismatch)
		}
	}

	return mismatches
}

// indexReportResults retains last-entry-wins behavior for duplicate report IDs, matching the original comparison and
// leaving fixture validation responsible for enforcing uniqueness on the expected side.
func indexReportResults(report Report) map[string]Result {
	results := make(map[string]Result, len(report.Results))
	for _, result := range report.Results {
		results[result.ID] = result
	}

	return results
}

// compareFixtureAsset preserves missing-before-error precedence, then compares source identity and frame structure as
// one fingerprint. Returning an empty string keeps the caller's fixture-order aggregation straightforward.
func compareFixtureAsset(actual map[string]Result, expected AssetFixture) string {
	result, exists := actual[expected.ID]
	if !exists || !result.Found {
		return fmt.Sprintf("%s: missing", expected.ID)
	}

	if result.Error != "" {
		return fmt.Sprintf("%s: %s", expected.ID, result.Error)
	}

	observed := AssetFixture{Directions: result.Directions}
	observed.addFrames(result.Frames)

	if fixtureMatchesResult(expected, observed, result) {
		return ""
	}

	return fmt.Sprintf("%s: structural fingerprint differs", expected.ID)
}

// fixtureMatchesResult compares every field in the persisted structural contract. Separating source and frame facts
// into visual groups makes additions to AssetFixture harder to overlook during future contract changes.
func fixtureMatchesResult(expected, observed AssetFixture, result Result) bool {
	sourceMatches := result.Path == expected.Path &&
		result.SHA256 == expected.SHA256 &&
		result.Bytes == expected.Bytes &&
		result.Type == expected.Type

	framesMatch := observed.Directions == expected.Directions &&
		observed.FrameCount == expected.FrameCount &&
		observed.MinWidth == expected.MinWidth &&
		observed.MaxWidth == expected.MaxWidth &&
		observed.MinHeight == expected.MinHeight &&
		observed.MaxHeight == expected.MaxHeight &&
		observed.FramesHash == expected.FramesHash

	return sourceMatches && framesMatch
}
