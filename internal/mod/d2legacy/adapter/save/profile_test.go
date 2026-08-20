package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestProfileFileRoundTripPreservesRosterSelectionAndPrivacy covers durability, permissions, and defensive ownership.
func TestProfileFileRoundTripPreservesRosterSelectionAndPrivacy(t *testing.T) {
	store := New(Character{
		ID:         "hero",
		Name:       "Hero",
		Class:      "Amazon",
		Level:      12,
		Stats:      &Stats{Health: 70},
		Appearance: &Appearance{Components: map[string]string{"HD": "head.dcc"}},
	})
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}

	// A nested destination proves profile writing creates private parent directories as part of the durable operation.
	path := filepath.Join(t.TempDir(), "profiles", "profile.json")
	if err := WriteProfileFile(path, store.Profile()); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("profile mode = %o", info.Mode().Perm())
	}

	profile, err := ReadProfileFile(path)
	if err != nil {
		t.Fatal(err)
	}

	restored, err := NewFromProfile(profile)
	if err != nil {
		t.Fatal(err)
	}

	selected, ok := restored.Selected()
	if !ok || selected.Name != "Hero" || selected.Stats.Health != 70 ||
		selected.Appearance.Components["HD"] != "head.dcc" {
		t.Fatalf("restored selection = %#v", selected)
	}

	// Mutating the decoded transport value must not reach the Store created from its defensive copy.
	profile.Characters[0].Stats.Health = 0
	if current, _ := restored.Selected(); current.Stats.Health != 70 {
		t.Fatal("restored store aliases decoded profile")
	}
}

// TestProfileDecodeRejectsTamperingUnknownTrailingAndInvalidRoster protects each trust-boundary validation phase.
func TestProfileDecodeRejectsTamperingUnknownTrailingAndInvalidRoster(t *testing.T) {
	var encoded bytes.Buffer

	profile := Profile{
		Version:    ProfileVersion,
		Characters: []Character{{ID: "hero", Name: "Hero"}},
		Selected:   "hero",
	}
	if err := EncodeProfile(&encoded, profile); err != nil {
		t.Fatal(err)
	}

	altered := bytes.Replace(encoded.Bytes(), []byte(`"name":"Hero"`), []byte(`"name":"Other"`), 1)
	assertProfileDecodeError(t, altered, "integrity mismatch")

	// Re-encoding altered selection creates valid JSON but still violates roster membership before integrity checking.
	var value map[string]any
	if err := json.Unmarshal(encoded.Bytes(), &value); err != nil {
		t.Fatal(err)
	}

	value["selected"] = "changed"

	tampered, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	assertProfileDecodeError(t, tampered, "selected character")

	unknown := bytes.Replace(encoded.Bytes(), []byte(`"version":1`), []byte(`"version":1,"unknown":true`), 1)
	if _, err := DecodeProfile(bytes.NewReader(unknown)); !errors.Is(err, ErrProfile) {
		t.Fatalf("unknown error = %v", err)
	}

	trailing := append(encoded.Bytes(), []byte(`{}`)...)
	assertProfileDecodeError(t, trailing, "trailing")

	duplicate := Profile{Version: ProfileVersion, Characters: []Character{{ID: "same"}, {ID: "same"}}}
	if _, err := NewFromProfile(duplicate); !errors.Is(err, ErrProfile) {
		t.Fatalf("duplicate error = %v", err)
	}
}

// assertProfileDecodeError verifies both the stable error identity and the phase-specific diagnostic.
func assertProfileDecodeError(t *testing.T, data []byte, diagnostic string) {
	t.Helper()

	_, err := DecodeProfile(bytes.NewReader(data))
	if !errors.Is(err, ErrProfile) || !strings.Contains(err.Error(), diagnostic) {
		t.Fatalf("profile decode error = %v, want diagnostic %q", err, diagnostic)
	}
}

// TestProfileDecodeBoundsInput verifies the reader limit rejects oversized data before JSON parsing.
func TestProfileDecodeBoundsInput(t *testing.T) {
	if _, err := DecodeProfile(strings.NewReader(strings.Repeat("x", MaxProfileBytes+1))); !errors.Is(err, ErrProfile) {
		t.Fatalf("oversize error = %v", err)
	}
}
