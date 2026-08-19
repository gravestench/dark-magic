package typed

import (
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/data/model"
)

// TestIndexRejectsDuplicatePrimaryKeys ensures the strict index never hides an
// ambiguous authored key by silently overwriting an earlier record.
func TestIndexRejectsDuplicatePrimaryKeys(t *testing.T) {
	t.Parallel()

	records := []models.CharStats{{Class: "ama"}, {Class: "ama"}}

	_, err := Index(records, charStatsClass)
	if err == nil || !strings.Contains(err.Error(), "duplicate key ama") {
		t.Fatalf("duplicate index error = %v", err)
	}
}

// TestObservedIndexPreservesRowsAndReportsDuplicate verifies that tolerant
// indexing retains the first record while exposing the later source row.
func TestObservedIndexPreservesRowsAndReportsDuplicate(t *testing.T) {
	t.Parallel()

	records := []models.CharStats{
		{Class: "unused", Strength: 1},
		{Class: "unused", Strength: 2},
	}

	index, issues, err := ObservedIndex("charstats", records, charStatsClass)
	if err != nil {
		t.Fatal(err)
	}

	if index["unused"].Strength != 1 {
		t.Fatalf("observed index = %#v, want first record", index)
	}

	if len(issues) != 1 || issues[0].Row != 3 || issues[0].Kind != "duplicate-key" {
		t.Fatalf("observed issues = %#v", issues)
	}
}

// charStatsClass names the primary-key contract shared by the index scenarios,
// keeping anonymous callbacks from obscuring the behavior under test.
func charStatsClass(record models.CharStats) string {
	return record.Class
}
