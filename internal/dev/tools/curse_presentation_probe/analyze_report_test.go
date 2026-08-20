package main

import (
	"strings"
	"testing"
)

// TestAnalyzeNormalizesCompleteOwned114dMatrix verifies that trusted evidence yields complete coverage and stable
// anchor-relative timing and motion facts.
func TestAnalyzeNormalizesCompleteOwned114dMatrix(t *testing.T) {
	rawCapture := marshalCapture(t, ownedCapture())

	got, err := analyze(strings.NewReader(rawCapture))
	if err != nil {
		t.Fatal(err)
	}

	if !got.Coverage.Complete {
		t.Fatalf("report coverage = %#v, want complete", got.Coverage)
	}

	if len(got.Coverage.Missing) != 0 {
		t.Fatalf("report missing coverage = %#v, want none", got.Coverage.Missing)
	}

	if len(got.Cases) != 6 {
		t.Fatalf("report cases = %d, want 6", len(got.Cases))
	}

	if len(got.CaptureFingerprint) != 64 {
		t.Fatalf("capture fingerprint length = %d, want 64", len(got.CaptureFingerprint))
	}

	areaInstance := got.Cases[0].Layers[0].Instances[0]
	if areaInstance.Frames != 20 ||
		areaInstance.StartOffset != (point{X: 0, Y: -10}) ||
		areaInstance.Translated {
		t.Fatalf("area instance = %#v", areaInstance)
	}

	castInstance := got.Cases[0].Layers[1].Instances[0]
	if castInstance.Frames != 25 ||
		castInstance.StartOffset != (point{X: 0, Y: -20}) ||
		castInstance.EndOffset != (point{X: 200, Y: 10}) ||
		!castInstance.Translated {
		t.Fatalf("cast instance = %#v", castInstance)
	}
}

// TestCoverageReportsMissingTargetBandsWithoutPromotingRoles verifies that partial evidence reports every absent
// matrix cell and never marks inferred role coverage as complete.
func TestCoverageReportsMissingTargetBandsWithoutPromotingRoles(t *testing.T) {
	captured := ownedCapture()
	captured.Cases = captured.Cases[:1]

	got := coverageFor(captured.Cases)

	if got.Complete || len(got.Missing) != 5 {
		t.Fatalf("coverage = %#v", got)
	}
}
