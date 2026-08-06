package content

import (
	"encoding/json"
	"fmt"
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
		Schema   string                     `json:"schema"`
		Version  int                        `json:"version"`
		Palettes map[string]string          `json:"palettes"`
		Fonts    map[string]json.RawMessage `json:"fonts"`
		Sounds   map[string]string          `json:"sounds"`
		Cursor   json.RawMessage            `json:"cursor"`
		Startup  json.RawMessage            `json:"startup"`
		Screens  map[string]json.RawMessage `json:"screens"`
		Profiles []struct {
			ID          string `json:"id"`
			GameVersion string `json:"game_version"`
			Language    string `json:"language"`
			Resolution  struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"resolution"`
		} `json:"supported_profiles"`
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
	if len(manifest.Profiles) != 1 {
		t.Fatalf("supported presentation profiles = %d, want 1", len(manifest.Profiles))
	}
	profile := manifest.Profiles[0]
	if profile.ID != "lod-english-800x600" || profile.GameVersion != "diablo-ii-lod" || profile.Language != "English" ||
		profile.Resolution.Width != manifest.Resolution.Width || profile.Resolution.Height != manifest.Resolution.Height {
		t.Fatalf("unsupported or inconsistent presentation profile: %#v", profile)
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
		"character_select",
		"cinematics",
		"game_loading",
		"credits",
	} {
		if len(manifest.Screens[screen]) == 0 {
			t.Errorf("presentation manifest is missing %q", screen)
		}
	}
}

func TestSupportedPresentationCompositionMatrix(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	profiles, ok := document["supported_profiles"].([]any)
	if !ok || len(profiles) == 0 {
		t.Fatal("presentation manifest must declare supported profiles")
	}
	for _, rawProfile := range profiles {
		profile := rawProfile.(map[string]any)
		name := profile["id"].(string)
		t.Run(name, func(t *testing.T) {
			resolution := profile["resolution"].(map[string]any)
			width, height := resolution["width"].(float64), resolution["height"].(float64)
			if width <= 0 || height <= 0 {
				t.Fatalf("invalid resolution %.0fx%.0f", width, height)
			}
			validatePresentationGeometry(t, document["screens"], width, height, "screens")
			layouts := document["layouts"].(map[string]any)
			tiles := layouts["frontend_tiles"].(map[string]any)
			if sumJSONNumbers(tiles["columns"].([]any)) != width || sumJSONNumbers(tiles["rows"].([]any)) != height {
				t.Fatalf("frontend tile grid does not cover %.0fx%.0f", width, height)
			}
			language := profile["language"].(string)
			localeData, err := fs.ReadFile(Shim(), "locales/"+language+".json")
			if err != nil {
				t.Fatalf("read %s locale: %v", language, err)
			}
			var localized map[string]string
			if err := json.Unmarshal(localeData, &localized); err != nil {
				t.Fatalf("decode %s locale: %v", language, err)
			}
			for key := range presentationLocaleKeys(document["screens"]) {
				if localized[key] == "" {
					t.Errorf("%s locale is missing presentation key %q", language, key)
				}
			}
		})
	}
}

func presentationLocaleKeys(value any) map[string]bool {
	result := make(map[string]bool)
	var visit func(any)
	visit = func(value any) {
		switch current := value.(type) {
		case []any:
			for _, item := range current {
				visit(item)
			}
		case map[string]any:
			for key, item := range current {
				if key == "label" || key == "key" {
					if text, ok := item.(string); ok {
						result[text] = true
					}
				}
				visit(item)
			}
		}
	}
	visit(value)
	return result
}

func validatePresentationGeometry(t *testing.T, value any, width, height float64, path string) {
	t.Helper()
	switch current := value.(type) {
	case []any:
		for index, item := range current {
			validatePresentationGeometry(t, item, width, height, fmt.Sprintf("%s[%d]", path, index))
		}
	case map[string]any:
		if x, exists := current["x"].(float64); exists && (x < 0 || x > width) {
			t.Errorf("%s.x = %v outside width %v", path, x, width)
		}
		if y, exists := current["y"].(float64); exists && (y < 0 || y > height) {
			t.Errorf("%s.y = %v outside height %v", path, y, height)
		}
		if itemWidth, exists := current["width"].(float64); exists && (itemWidth <= 0 || itemWidth > width) {
			t.Errorf("%s.width = %v invalid for width %v", path, itemWidth, width)
		}
		if itemHeight, exists := current["height"].(float64); exists && (itemHeight <= 0 || itemHeight > height) {
			t.Errorf("%s.height = %v invalid for height %v", path, itemHeight, height)
		}
		for key, item := range current {
			validatePresentationGeometry(t, item, width, height, path+"."+key)
		}
	}
}

func sumJSONNumbers(values []any) float64 {
	var result float64
	for _, value := range values {
		result += value.(float64)
	}
	return result
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
