package content

import (
	"encoding/json"
	"io/fs"
	"testing"

	"github.com/gravestench/dark-magic/pkg/assetcatalog"
)

// TestShimPresentationManifestContract protects the architectural boundary
// between native engine code and mod-owned presentation knowledge. Go should
// provide capabilities; the shim manifest should name and describe assets.
func TestShimPresentationManifestContract(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Schema     string                     `json:"schema"`
		Version    int                        `json:"version"`
		Palettes   map[string]string          `json:"palettes"`
		Fonts      map[string]json.RawMessage `json:"fonts"`
		Sounds     map[string]string          `json:"sounds"`
		Cursor     json.RawMessage            `json:"cursor"`
		Startup    json.RawMessage            `json:"startup"`
		Screens    map[string]json.RawMessage `json:"screens"`
		Resolution struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"resolution"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode presentation manifest: %v", err)
	}
	if manifest.Schema != "darkmagic.presentation/v1" || manifest.Version != 1 {
		t.Fatalf("unexpected presentation contract %q version %d", manifest.Schema, manifest.Version)
	}
	if manifest.Resolution.Width <= 0 || manifest.Resolution.Height <= 0 {
		t.Fatalf("invalid presentation resolution: %#v", manifest.Resolution)
	}
	if len(manifest.Palettes) == 0 || len(manifest.Fonts) == 0 || len(manifest.Sounds) == 0 {
		t.Fatal("presentation manifest must own palette, font, and sound facts")
	}
	if len(manifest.Cursor) == 0 || len(manifest.Startup) == 0 {
		t.Fatal("presentation manifest must own cursor and startup facts")
	}
	for _, screen := range []string{
		"title",
		"main_menu",
		"tcpip",
		"character_create",
		"cinematics",
		"game_loading",
		"credits",
	} {
		if len(manifest.Screens[screen]) == 0 {
			t.Errorf("presentation manifest is missing %q", screen)
		}
	}
}

func TestShimAssetFixtureContract(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "manifests/asset-fixture.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture assetcatalog.Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode asset fixture: %v", err)
	}
	if err := fixture.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Assets) != 90 {
		t.Fatalf("asset fixture contains %d entries, want 90", len(fixture.Assets))
	}
}

func TestCharacterCreationTransitionFacts(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Screens struct {
			CharacterCreate struct {
				Classes []struct {
					Class         string `json:"class"`
					Forward       string `json:"forward"`
					ForwardFrames int    `json:"forward_frames"`
					Back          string `json:"back"`
					BackFrames    int    `json:"back_frames"`
				} `json:"classes"`
			} `json:"character_create"`
		} `json:"screens"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	classes := manifest.Screens.CharacterCreate.Classes
	if len(classes) != 7 {
		t.Fatalf("character creation classes = %d, want 7", len(classes))
	}
	for _, class := range classes {
		if class.Class == "" || class.Forward == "" || class.Back == "" || class.ForwardFrames <= 0 || class.BackFrames <= 0 {
			t.Errorf("incomplete walk transition for %#v", class)
		}
	}
}
