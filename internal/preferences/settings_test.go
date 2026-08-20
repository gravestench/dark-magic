package preferences

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSettingsValidatePersistAndReload verifies accepted values survive saving while rejected replacements leave the
// live snapshot untouched.
func TestSettingsValidatePersistAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")

	settings, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	values := settings.Values()
	values.SoundVolume, values.MusicVolume = .25, .75
	values.DebugTextureResidency, values.TextureUploadBudgetMB = true, 8
	values.TextureCacheBudgetMB = 768
	values.CameraFollowStrategy, values.CameraFollowDuration = "back_out", .25
	values.CameraFollowParam1 = 2.25
	values.RealmGateway = "realm.example"

	if err := settings.Update(values); err != nil {
		t.Fatal(err)
	}

	if !settings.Dirty() {
		t.Fatal("updated settings are not dirty")
	}

	if err := settings.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := reloaded.Values(); got != values {
		t.Fatalf("values = %#v, want %#v", got, values)
	}

	values.SoundVolume = 2
	if err := settings.Update(values); err == nil {
		t.Fatal("invalid sound volume accepted")
	}
}

// TestCameraFollowPreferencesValidateStrategyAndParameters protects the finite
// strategy vocabulary and duration bound consumed by camera interpolation.
func TestCameraFollowPreferencesValidateStrategyAndParameters(t *testing.T) {
	settings := NewTransient()

	values := settings.Values()
	if values.CameraFollowStrategy != "instant" || values.CameraFollowDuration != 0 {
		t.Fatalf("camera defaults = %#v", values)
	}

	values.CameraFollowStrategy, values.CameraFollowDuration = "cubic_out", .2
	if err := settings.Update(values); err != nil {
		t.Fatal(err)
	}

	values.CameraFollowStrategy = "teleportish"
	if err := settings.Update(values); err == nil {
		t.Fatal("unknown camera strategy accepted")
	}

	values.CameraFollowStrategy, values.CameraFollowDuration = "linear", 6
	if err := settings.Update(values); err == nil {
		t.Fatal("out-of-range camera duration accepted")
	}
}

// TestLegacyPreferencesGainTextureBudgetDefaults ensures fields absent from old
// files receive current safe defaults instead of failing validation at zero.
func TestLegacyPreferencesGainTextureBudgetDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte(`{"sound_volume":0.5,"music_volume":0.5}`), 0o600); err != nil {
		t.Fatal(err)
	}

	settings, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	if values := settings.Values(); values.TextureUploadBudgetMB != 16 || values.TextureCacheBudgetMB != 512 {
		t.Fatalf("migrated values = %#v", values)
	}
}

// TestLegacyDefaultUploadBudgetMigratesOnlyOnce distinguishes the old implicit
// 4 MB default from a versioned user's explicit choice of the same value.
func TestLegacyDefaultUploadBudgetMigratesOnlyOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")

	legacy := `{"sound_volume":0.5,"music_volume":0.5,"texture_upload_budget_mb":4,"texture_cache_budget_mb":512}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	settings, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := settings.Values().TextureUploadBudgetMB; got != 16 {
		t.Fatalf("migrated upload budget = %g", got)
	}

	values := settings.Values()

	values.TextureUploadBudgetMB = 4
	if err := settings.Update(values); err != nil {
		t.Fatal(err)
	}

	if err := settings.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := New(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := reloaded.Values().TextureUploadBudgetMB; got != 4 {
		t.Fatalf("explicit upload budget changed after reload: %g", got)
	}
}
