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

func TestProfileFileRoundTripPreservesRosterSelectionAndPrivacy(t *testing.T) {
	store := New(Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 12,
		Stats: &Stats{Health: 70}, Appearance: &Appearance{Components: map[string]string{"HD": "head.dcc"}}})
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}
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
	if !ok || selected.Name != "Hero" || selected.Stats.Health != 70 || selected.Appearance.Components["HD"] != "head.dcc" {
		t.Fatalf("restored selection = %#v", selected)
	}
	profile.Characters[0].Stats.Health = 0
	if current, _ := restored.Selected(); current.Stats.Health != 70 {
		t.Fatal("restored store aliases decoded profile")
	}
}

func TestProfileDecodeRejectsTamperingUnknownTrailingAndInvalidRoster(t *testing.T) {
	var encoded bytes.Buffer
	if err := EncodeProfile(&encoded, Profile{Version: ProfileVersion, Characters: []Character{{ID: "hero", Name: "Hero"}}, Selected: "hero"}); err != nil {
		t.Fatal(err)
	}
	altered := bytes.Replace(encoded.Bytes(), []byte(`"name":"Hero"`), []byte(`"name":"Other"`), 1)
	if _, err := DecodeProfile(bytes.NewReader(altered)); !errors.Is(err, ErrProfile) || !strings.Contains(err.Error(), "integrity mismatch") {
		t.Fatalf("integrity error = %v", err)
	}
	var value map[string]any
	if err := json.Unmarshal(encoded.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	value["selected"] = "changed"
	tampered, _ := json.Marshal(value)
	if _, err := DecodeProfile(bytes.NewReader(tampered)); !errors.Is(err, ErrProfile) || !strings.Contains(err.Error(), "selected character") {
		t.Fatalf("tamper error = %v", err)
	}
	unknown := bytes.Replace(encoded.Bytes(), []byte(`"version":1`), []byte(`"version":1,"unknown":true`), 1)
	if _, err := DecodeProfile(bytes.NewReader(unknown)); !errors.Is(err, ErrProfile) {
		t.Fatalf("unknown error = %v", err)
	}
	if _, err := DecodeProfile(bytes.NewReader(append(encoded.Bytes(), []byte(`{}`)...))); !errors.Is(err, ErrProfile) || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("trailing error = %v", err)
	}
	if _, err := NewFromProfile(Profile{Version: ProfileVersion, Characters: []Character{{ID: "same"}, {ID: "same"}}}); !errors.Is(err, ErrProfile) {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestProfileDecodeBoundsInput(t *testing.T) {
	if _, err := DecodeProfile(strings.NewReader(strings.Repeat("x", MaxProfileBytes+1))); !errors.Is(err, ErrProfile) {
		t.Fatalf("oversize error = %v", err)
	}
}
