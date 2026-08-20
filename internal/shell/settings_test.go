package shell

import (
	"path/filepath"
	"testing"
)

// TestSettingsEditResetSaveAndReload exercises live validation, explicit
// persistence, defaults, and disk reload through one user workflow.
func TestSettingsEditResetSaveAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "shell.json")

	settings, err := NewSettings(path)
	if err != nil || settings.Values().FontSize != DefaultFontSize {
		t.Fatalf("defaults = %#v, %v", settings, err)
	}

	if err := settings.SetFontSize(22); err != nil || !settings.Dirty() {
		t.Fatalf("set = %v dirty=%v", err, settings.Dirty())
	}

	if err := settings.Save(); err != nil || settings.Dirty() {
		t.Fatalf("save = %v dirty=%v", err, settings.Dirty())
	}

	reloaded, err := NewSettings(path)
	if err != nil || reloaded.Values().FontSize != 22 || reloaded.Values().ConsoleHeight != 0.6 {
		t.Fatalf("reload = %#v, %v", reloaded, err)
	}

	reloaded.Reset()

	if reloaded.Values().FontSize != DefaultFontSize || !reloaded.Dirty() {
		t.Fatalf("reset = %#v", reloaded.Values())
	}

	if err := reloaded.Reload(); err != nil || reloaded.Values().FontSize != 22 || reloaded.Dirty() {
		t.Fatalf("reload = %#v, %v", reloaded.Values(), err)
	}

	if err := reloaded.SetFontSize(7); err == nil {
		t.Fatal("accepted invalid font size")
	}
}
