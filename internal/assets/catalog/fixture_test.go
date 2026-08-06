package assetcatalog

import "testing"

func TestFixtureRoundTripAndMismatch(t *testing.T) {
	report := Report{ManifestVersion: 1, Results: []Result{{
		Hypothesis: Hypothesis{ID: "button", Path: "button.dc6"}, Found: true,
		Bytes: 42, SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Type: "dc6", Directions: 1,
		Frames: []Frame{{Direction: 0, Frame: 0, Width: 10, Height: 20}},
	}}}
	fixture, err := FixtureFromReport(report)
	if err != nil {
		t.Fatal(err)
	}
	if mismatches := CompareFixture(report, fixture); len(mismatches) != 0 {
		t.Fatalf("unexpected mismatches: %v", mismatches)
	}
	report.Results[0].Frames[0].Width++
	if mismatches := CompareFixture(report, fixture); len(mismatches) != 1 {
		t.Fatalf("mismatches = %v", mismatches)
	}
}

func TestFixtureRejectsPartialReport(t *testing.T) {
	_, err := FixtureFromReport(Report{ManifestVersion: 1, Results: []Result{{
		Hypothesis: Hypothesis{ID: "missing", Path: "missing.dc6"},
		Error:      "not found",
	}}})
	if err == nil {
		t.Fatal("expected partial report to be rejected")
	}
}
