package preferences

import (
	"os"
	"path/filepath"
	"testing"
)

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

func TestLegacyPreferencesGainTextureBudgetDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	if err := os.WriteFile(path, []byte(`{"sound_volume":0.5,"music_volume":0.5}`), 0o600); err != nil {
		t.Fatal(err)
	}
	settings, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	if values := settings.Values(); values.TextureUploadBudgetMB != 4 || values.TextureCacheBudgetMB != 512 {
		t.Fatalf("migrated values = %#v", values)
	}
}
