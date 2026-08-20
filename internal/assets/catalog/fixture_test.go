package assetcatalog

import "testing"

// TestFixtureRoundTripAndMismatch confirms a complete report fixtures cleanly and that frame metadata participates in
// comparison. The mutation isolates the frame hash from unchanged source identity fields.
func TestFixtureRoundTripAndMismatch(t *testing.T) {
	report := completeFixtureReport()

	fixture, err := FixtureFromReport(report)
	if err != nil {
		t.Fatal(err)
	}

	if mismatches := CompareFixture(report, fixture); len(mismatches) != 0 {
		t.Fatalf("unexpected mismatches: %v", mismatches)
	}

	// Alter one structural fact while retaining source identity to prove the frame fingerprint detects the drift.
	report.Results[0].Frames[0].Width++
	if mismatches := CompareFixture(report, fixture); len(mismatches) != 1 {
		t.Fatalf("mismatches = %v", mismatches)
	}
}

// completeFixtureReport owns all mutable slices used by the round-trip test so its deliberate mutation cannot leak.
func completeFixtureReport() Report {
	return Report{ManifestVersion: 1, Results: []Result{{
		Hypothesis: Hypothesis{ID: "button", Path: "button.dc6"},
		Found:      true,
		Bytes:      42,
		SHA256:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Type:       "dc6",
		Directions: 1,
		Frames:     []Frame{{Direction: 0, Frame: 0, Width: 10, Height: 20}},
	}}}
}

// TestFixtureRejectsPartialReport ensures missing archive observations cannot be promoted into a distributable fixture.
func TestFixtureRejectsPartialReport(t *testing.T) {
	_, err := FixtureFromReport(Report{ManifestVersion: 1, Results: []Result{{
		Hypothesis: Hypothesis{ID: "missing", Path: "missing.dc6"},
		Error:      "not found",
	}}})
	if err == nil {
		t.Fatal("expected partial report to be rejected")
	}
}
