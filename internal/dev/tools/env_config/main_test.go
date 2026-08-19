package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestAssignmentsString verifies usage output documents the input shape without leaking any accumulated settings.
func TestAssignmentsString(t *testing.T) {
	values := assignments{"SECRET": "private"}

	if got, want := values.String(), "NAME=VALUE"; got != want {
		t.Fatalf("assignments.String() = %q, want %q", got, want)
	}
}

// TestAssignmentsSetPreservesCompleteValues locks in first-separator parsing and last-write-wins repeat behavior.
func TestAssignmentsSetPreservesCompleteValues(t *testing.T) {
	values := assignments{}

	if err := values.Set("NAME=left=right"); err != nil {
		t.Fatalf("assignments.Set() first value error = %v", err)
	}

	if err := values.Set("NAME="); err != nil {
		t.Fatalf("assignments.Set() replacement error = %v", err)
	}

	if got, want := values["NAME"], ""; got != want {
		t.Fatalf("stored NAME = %q, want %q", got, want)
	}
}

// TestAssignmentsSetRejectsMissingNames prevents malformed flags from reaching environment file mutation.
func TestAssignmentsSetRejectsMissingNames(t *testing.T) {
	testCases := []string{"NAME", "=value"}

	for _, value := range testCases {
		t.Run(value, func(t *testing.T) {
			got := assignments{}.Set(value)
			if got == nil || got.Error() != "set value must be NAME=VALUE" {
				t.Fatalf("assignments.Set(%q) error = %v, want %q", value, got, "set value must be NAME=VALUE")
			}
		})
	}
}

// TestSelectUpdateRoles verifies expansion order and keeps command-only validation distinct from role validation.
func TestSelectUpdateRoles(t *testing.T) {
	testCases := []struct {
		name      string
		role      string
		updates   assignments
		wantRoles []string
		wantError string
	}{
		{
			name:      "missing role",
			role:      "",
			updates:   assignments{},
			wantError: "--role is required",
		},
		{
			name:      "all roles",
			role:      "all",
			updates:   assignments{},
			wantRoles: []string{"client", "server", "realm"},
		},
		{
			name:      "all roles with updates",
			role:      "all",
			updates:   assignments{"NAME": "value"},
			wantError: "--set requires one explicit role",
		},
		{
			name:      "explicit known role",
			role:      "realm",
			updates:   assignments{"NAME": "value"},
			wantRoles: []string{"realm"},
		},
		{
			name:      "explicit unknown role",
			role:      "unknown",
			updates:   assignments{},
			wantRoles: []string{"unknown"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			roles, err := selectUpdateRoles(testCase.role, testCase.updates)
			if !slices.Equal(roles, testCase.wantRoles) {
				t.Fatalf("selectUpdateRoles() roles = %v, want %v", roles, testCase.wantRoles)
			}

			if testCase.wantError == "" && err != nil {
				t.Fatalf("selectUpdateRoles() unexpected error = %v", err)
			}

			if testCase.wantError != "" && (err == nil || err.Error() != testCase.wantError) {
				t.Fatalf("selectUpdateRoles() error = %v, want %q", err, testCase.wantError)
			}
		})
	}
}

// TestUpdateEnvironmentFilesReportsUsageErrors ensures invalid cross-role requests cannot create configuration files.
func TestUpdateEnvironmentFilesReportsUsageErrors(t *testing.T) {
	testCases := []struct {
		name       string
		role       string
		updates    assignments
		wantStderr string
	}{
		{
			name:       "missing role",
			role:       "",
			updates:    assignments{},
			wantStderr: "--role is required\n",
		},
		{
			name:       "all roles with updates",
			role:       "all",
			updates:    assignments{"NAME": "value"},
			wantStderr: "--set requires one explicit role\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			exitCode := updateEnvironmentFiles(testCase.role, testCase.updates, &stdout, &stderr)

			if exitCode != usageErrorExitCode {
				t.Fatalf("updateEnvironmentFiles() exit code = %d, want %d", exitCode, usageErrorExitCode)
			}

			if stdout.Len() != 0 {
				t.Fatalf("updateEnvironmentFiles() stdout = %q, want empty", stdout.String())
			}

			if got := stderr.String(); got != testCase.wantStderr {
				t.Fatalf("updateEnvironmentFiles() stderr = %q, want %q", got, testCase.wantStderr)
			}
		})
	}
}

// TestUpdateEnvironmentFilesReportsUpdateErrors preserves envconfig failures and prevents false success output.
func TestUpdateEnvironmentFilesReportsUpdateErrors(t *testing.T) {
	// Isolate the command from the user's configuration even though an unknown role fails before filesystem access.
	t.Setenv("DARK_MAGIC_CONFIG_DIR", t.TempDir())

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	exitCode := updateEnvironmentFiles("unknown", assignments{}, &stdout, &stderr)

	if exitCode != updateFailureExitCode {
		t.Fatalf("updateEnvironmentFiles() exit code = %d, want %d", exitCode, updateFailureExitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("updateEnvironmentFiles() stdout = %q, want empty", stdout.String())
	}

	if got, want := stderr.String(), "unknown environment role \"unknown\"\n"; got != want {
		t.Fatalf("updateEnvironmentFiles() stderr = %q, want %q", got, want)
	}
}

// TestUpdateEnvironmentFilesUpdatesAllRoles verifies completed paths remain ordered for shell consumers.
func TestUpdateEnvironmentFilesUpdatesAllRoles(t *testing.T) {
	configDirectory := t.TempDir()
	// A private directory keeps the integration-style command check deterministic and free of operator state.
	t.Setenv("DARK_MAGIC_CONFIG_DIR", configDirectory)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	exitCode := updateEnvironmentFiles("all", assignments{}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("updateEnvironmentFiles() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("updateEnvironmentFiles() stderr = %q, want empty", stderr.String())
	}

	expectedPaths := []string{
		filepath.Join(configDirectory, "client.env"),
		filepath.Join(configDirectory, "server.env"),
		filepath.Join(configDirectory, "realm.env"),
	}

	wantOutput := strings.Join(expectedPaths, "\n") + "\n"
	if got := stdout.String(); got != wantOutput {
		t.Fatalf("updateEnvironmentFiles() stdout = %q, want %q", got, wantOutput)
	}

	for _, path := range expectedPaths {
		information, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat completed environment file %q: %v", path, err)
		}

		if got, want := information.Mode().Perm(), os.FileMode(0o600); got != want {
			t.Fatalf("environment file %q mode = %o, want %o", path, got, want)
		}
	}
}

// TestUpdateEnvironmentFilesAppliesExplicitRoleUpdates confirms command values reach the selected role without loss.
func TestUpdateEnvironmentFilesAppliesExplicitRoleUpdates(t *testing.T) {
	configDirectory := t.TempDir()
	t.Setenv("DARK_MAGIC_CONFIG_DIR", configDirectory)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	updates := assignments{"MPQ_DIRECTORY": "/fixtures/archive=legacy"}

	exitCode := updateEnvironmentFiles("realm", updates, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("updateEnvironmentFiles() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("updateEnvironmentFiles() stderr = %q, want empty", stderr.String())
	}

	path := filepath.Join(configDirectory, "realm.env")
	if got, want := stdout.String(), path+"\n"; got != want {
		t.Fatalf("updateEnvironmentFiles() stdout = %q, want %q", got, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated environment file: %v", err)
	}

	if want := "MPQ_DIRECTORY=\"/fixtures/archive=legacy\"\n"; !strings.Contains(string(data), want) {
		t.Fatalf("updated environment file does not contain %q:\n%s", want, data)
	}
}
