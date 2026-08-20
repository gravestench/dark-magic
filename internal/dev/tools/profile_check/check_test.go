package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckReportsMissingDiagnosticsInSceneOrder verifies map iteration cannot make CI output nondeterministic.
func TestCheckReportsMissingDiagnosticsInSceneOrder(t *testing.T) {
	root := t.TempDir()
	budgetPath := filepath.Join(root, "budgets.json")
	writeTestFile(t, budgetPath, `{"zeta":{},"alpha":{}}`)

	err := check(filepath.Join(root, "profiles"), budgetPath)
	if err == nil {
		t.Fatal("check passed without required scene diagnostics")
	}

	want := strings.Join([]string{
		`profile check: scene "alpha" has no diagnostics`,
		`profile check: scene "zeta" has no diagnostics`,
	}, "\n")
	if err.Error() != want {
		t.Fatalf("unexpected error order:\n got: %s\nwant: %s", err, want)
	}
}

// TestCheckReportsSnapshotViolationsInPathOrder verifies multiple captures remain ordered by artifact name.
func TestCheckReportsSnapshotViolationsInPathOrder(t *testing.T) {
	root := t.TempDir()
	budgetPath := filepath.Join(root, "budgets.json")
	writeTestFile(t, budgetPath, `{"title":{"max_decoded_weight":1}}`)

	sceneDirectory := filepath.Join(root, "profiles", "scenes", "title")
	writeTestFile(t, filepath.Join(sceneDirectory, "diagnostics-b.json"), snapshotWithDecodedWeight(3))
	writeTestFile(t, filepath.Join(sceneDirectory, "diagnostics-a.json"), snapshotWithDecodedWeight(2))

	err := check(filepath.Join(root, "profiles"), budgetPath)
	if err == nil {
		t.Fatal("check passed snapshots that exceeded the decoded-weight budget")
	}

	want := strings.Join([]string{
		"profile check: title decoded weight 2 exceeds 1",
		"profile check: title decoded weight 3 exceeds 1",
	}, "\n")
	if err.Error() != want {
		t.Fatalf("unexpected snapshot error order:\n got: %s\nwant: %s", err, want)
	}
}

// TestCheckDistinguishesBudgetReadAndParseFailures keeps setup diagnostics actionable for missing and corrupt inputs.
func TestCheckDistinguishesBudgetReadAndParseFailures(t *testing.T) {
	root := t.TempDir()

	t.Run("read", func(t *testing.T) {
		missingPath := filepath.Join(root, "missing.json")

		err := check(filepath.Join(root, "profiles"), missingPath)
		if err == nil || !strings.HasPrefix(err.Error(), "profile check: read budgets:") {
			t.Fatalf("expected budget read failure, got %v", err)
		}
	})

	t.Run("parse", func(t *testing.T) {
		invalidPath := filepath.Join(root, "invalid.json")
		writeTestFile(t, invalidPath, "{")

		err := check(filepath.Join(root, "profiles"), invalidPath)
		if err == nil || !strings.HasPrefix(err.Error(), "profile check: parse budgets:") {
			t.Fatalf("expected budget parse failure, got %v", err)
		}
	})
}

// writeTestFile creates parent directories with fixture-local ownership, then fails immediately on unusable input.
// Keeping fixture construction here makes every test's cleanup the responsibility of its t.TempDir root.
func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}

// snapshotWithDecodedWeight returns the smallest valid profiler document needed to exercise composition ordering.
func snapshotWithDecodedWeight(weight int) string {
	return `{"composition":{"Decoded":{"Weight":` + fmt.Sprint(weight) + `}}}`
}
