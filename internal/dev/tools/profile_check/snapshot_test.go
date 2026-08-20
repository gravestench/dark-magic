package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const exceededBudgetSnapshot = `{
  "composition": {
    "Decoded": {"Weight": 12},
    "Retained": {"ActiveResources": 3, "RetainedTextureBytes": 40},
    "DecodeTime": 2000000
  },
  "frame_timing": {
    "title": {"samples": 2, "frame_p95": 30000000, "update_p95": 12000000}
  }
}`

// TestCheckSnapshotReportsEveryViolationInStableOrder protects the complete diagnostic contract for one capture.
func TestCheckSnapshotReportsEveryViolationInStableOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagnostics.json")
	writeTestFile(t, path, exceededBudgetSnapshot)

	limits := budget{
		MaxRetainedTextureBytes: 30,
		MaxActiveResources:      2,
		MaxDecodedWeight:        10,
		MaxDecodeTimeMS:         1,
		MinFrameSamples:         3,
		MaxFrameP95MS:           20,
		MaxUpdateP95MS:          10,
	}

	err := checkSnapshot("title", path, limits)
	if err == nil {
		t.Fatal("checkSnapshot passed a snapshot that exceeded every configured budget")
	}

	want := strings.Join([]string{
		"profile check: title retained texture bytes 40 exceed 30",
		"profile check: title active resources 3 exceed 2",
		"profile check: title decoded weight 12 exceeds 10",
		"profile check: title cumulative decode time 2ms exceeds 1ms",
		"profile check: title frame samples 2 below 3",
		"profile check: title p95 frame interval 30ms exceeds 20ms",
		"profile check: title p95 update time 12ms exceeds 10ms",
	}, "\n")
	if err.Error() != want {
		t.Fatalf("unexpected violations:\n got: %s\nwant: %s", err, want)
	}
}

// TestCheckSnapshotAcceptsBudgetBoundaries verifies limits remain inclusive and do not create false regressions.
func TestCheckSnapshotAcceptsBudgetBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagnostics.json")
	writeTestFile(t, path, exceededBudgetSnapshot)

	limits := budget{
		MaxRetainedTextureBytes: 40,
		MaxActiveResources:      3,
		MaxDecodedWeight:        12,
		MaxDecodeTimeMS:         2,
		MinFrameSamples:         2,
		MaxFrameP95MS:           30,
		MaxUpdateP95MS:          12,
	}
	if err := checkSnapshot("title", path, limits); err != nil {
		t.Fatalf("checkSnapshot rejected values at their inclusive limits: %v", err)
	}
}

// TestCheckSnapshotLeavesZeroTimingLimitsDisabled preserves compatibility with budgets that omit timing fields.
func TestCheckSnapshotLeavesZeroTimingLimitsDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagnostics.json")
	writeTestFile(t, path, exceededBudgetSnapshot)

	limits := budget{
		MaxRetainedTextureBytes: 40,
		MaxActiveResources:      3,
		MaxDecodedWeight:        12,
		MaxDecodeTimeMS:         2,
	}
	if err := checkSnapshot("title", path, limits); err != nil {
		t.Fatalf("checkSnapshot enforced omitted timing limits: %v", err)
	}
}

// TestCheckSnapshotDistinguishesReadAndParseFailures ensures corrupt artifacts are not reported as budget regressions.
func TestCheckSnapshotDistinguishesReadAndParseFailures(t *testing.T) {
	root := t.TempDir()

	t.Run("read", func(t *testing.T) {
		missingPath := filepath.Join(root, "missing.json")

		err := checkSnapshot("title", missingPath, budget{})
		if err == nil || !strings.HasPrefix(err.Error(), "profile check: read ") {
			t.Fatalf("expected snapshot read failure, got %v", err)
		}
	})

	t.Run("parse", func(t *testing.T) {
		invalidPath := filepath.Join(root, "invalid.json")
		writeTestFile(t, invalidPath, "{")

		err := checkSnapshot("title", invalidPath, budget{})
		if err == nil || !strings.HasPrefix(err.Error(), "profile check: parse ") {
			t.Fatalf("expected snapshot parse failure, got %v", err)
		}
	})
}
