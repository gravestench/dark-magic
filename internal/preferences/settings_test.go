package preferences

import (
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
