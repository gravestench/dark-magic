package luaModLoader

import "testing"

func TestManifestValidationAndAPIKey(t *testing.T) {
	manifest := Manifest{Name: "Dark Magic Terminal", Version: "1.0", Requires: []string{"api.ui"}}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	if got, want := manifest.ApiKey(), "darkmagicterminal10"; got != want {
		t.Fatalf("API key = %q, want %q", got, want)
	}
}

func TestManifestRejectsInitScriptOutsideMod(t *testing.T) {
	manifest := Manifest{Name: "Example", Version: "1.0", InitScript: "../init.lua"}
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected invalid init script error")
	}
}

func TestManifestValidationRejectsMissingFields(t *testing.T) {
	for _, manifest := range []Manifest{{Version: "1.0"}, {Name: "Example"}} {
		if err := manifest.Validate(); err == nil {
			t.Fatalf("expected invalid manifest: %+v", manifest)
		}
	}
}
