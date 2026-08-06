package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
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
		Transforms map[string]string          `json:"palette_transforms"`
		Fonts      map[string]json.RawMessage `json:"fonts"`
		Styles     map[string]struct {
			Font      string `json:"font"`
			Transform string `json:"transform"`
		} `json:"text_styles"`
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
	for _, name := range []string{"panel_heading", "panel_label", "panel_value", "frontend_version", "frontend_legal", "character_select_title", "character_select_metadata", "character_create_heading", "character_create_class", "character_create_description", "character_create_option", "credits", "button_normal", "button_hover", "label_button_normal", "label_button_hover", "dialog_text", "tooltip", "disabled", "font_lab_heading", "font_lab_caption", "font_lab_exocet10", "font_lab_font6", "font_lab_font16", "font_lab_font30", "font_lab_font42", "font_lab_formal10", "font_lab_formal11", "font_lab_formal12", "font_lab_color", "font_lab_gold_sky", "font_lab_gold_fechar", "font_lab_gold_act1"} {
		style, ok := manifest.Styles[name]
		if !ok {
			t.Errorf("presentation manifest is missing text style %q", name)
			continue
		}
		if _, ok := manifest.Fonts[style.Font]; !ok {
			t.Errorf("text style %q references unknown font %q", name, style.Font)
		}
		if style.Transform != "" {
			if _, ok := manifest.Transforms[style.Transform]; !ok {
				t.Errorf("text style %q references unknown palette transform %q", name, style.Transform)
			}
		}
	}
	if len(manifest.Cursor) == 0 || len(manifest.Startup) == 0 {
		t.Fatal("presentation manifest must own cursor and startup facts")
	}
	for _, screen := range []string{
		"font_lab",
		"title",
		"main_menu",
		"tcpip",
		"character_create",
		"character_select",
		"cinematics",
		"game_loading",
		"game_world",
		"inventory",
		"character",
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
				Palette      string `json:"palette"`
				ClassPalette string `json:"class_palette"`
				Campfire     struct {
					Sheet string `json:"sheet"`
				} `json:"campfire"`
				Stage map[string]struct {
					Anchor struct{ X, Y int }                `json:"anchor"`
					Hit    struct{ X, Y, Width, Height int } `json:"hit"`
				} `json:"stage"`
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
	creation := manifest.Screens.CharacterCreate
	if creation.Palette != "fechar" || creation.ClassPalette != "fechar" || creation.Campfire.Sheet == "" {
		t.Fatalf("character creation palette/campfire facts = %#v", creation)
	}
	if len(classes) != 7 {
		t.Fatalf("character creation classes = %d, want 7", len(classes))
	}
	for _, class := range classes {
		if class.Class == "" || class.Forward == "" || class.Back == "" || class.ForwardFrames <= 0 || class.BackFrames <= 0 {
			t.Errorf("incomplete walk transition for %#v", class)
		}
		placement, ok := creation.Stage[class.Class]
		if !ok || placement.Anchor.X <= 0 || placement.Anchor.Y <= 0 || placement.Hit.Width <= 0 || placement.Hit.Height <= 0 {
			t.Errorf("missing calibrated stage placement for %q: %#v", class.Class, placement)
		}
	}
}

func TestGameHUDCompositionFacts(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Screens struct {
			GameWorld struct {
				HUD struct {
					PanelSheet string `json:"panel_sheet"`
					PanelParts []struct {
						Frame  int `json:"frame"`
						X      int `json:"x"`
						Bottom int `json:"bottom"`
					} `json:"panel_parts"`
					Globes struct {
						Sheet        string `json:"sheet"`
						OverlapSheet string `json:"overlap_sheet"`
					} `json:"globes"`
					Skills struct {
						Sheet string `json:"sheet"`
					} `json:"skills"`
					Run struct {
						Sheet     string `json:"sheet"`
						WalkFrame int    `json:"walk_frame"`
						RunFrame  int    `json:"run_frame"`
					} `json:"run"`
					Menu struct {
						Sheet       string `json:"sheet"`
						ClosedFrame int    `json:"closed_frame"`
						OpenFrame   int    `json:"open_frame"`
					} `json:"menu"`
					Minipanel struct {
						Sheet       string `json:"sheet"`
						ButtonSheet string `json:"button_sheet"`
						Buttons     []struct {
							ID      string `json:"id"`
							Frame   int    `json:"frame"`
							Enabled bool   `json:"enabled"`
						} `json:"buttons"`
					} `json:"minipanel"`
				} `json:"hud"`
			} `json:"game_world"`
		} `json:"screens"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	hud := manifest.Screens.GameWorld.HUD
	if hud.PanelSheet == "" || hud.Globes.Sheet == "" || hud.Globes.OverlapSheet == "" || hud.Skills.Sheet == "" {
		t.Fatalf("incomplete HUD asset facts: %#v", hud)
	}
	if hud.Run.Sheet == "" || hud.Menu.Sheet == "" || hud.Minipanel.Sheet == "" || hud.Minipanel.ButtonSheet == "" {
		t.Fatalf("incomplete HUD control assets: %#v", hud)
	}
	if hud.Run.WalkFrame != 0 || hud.Run.RunFrame != 2 || hud.Menu.ClosedFrame != 0 || hud.Menu.OpenFrame != 2 {
		t.Fatalf("unexpected HUD toggle frames: run=%#v menu=%#v", hud.Run, hud.Menu)
	}
	if len(hud.Minipanel.Buttons) != 7 {
		t.Fatalf("minipanel buttons = %d, want 7", len(hud.Minipanel.Buttons))
	}
	wantFrames := []int{0, 2, 4, 8, 10, 12, 14}
	for index, button := range hud.Minipanel.Buttons {
		if button.ID == "" || button.Frame != wantFrames[index] {
			t.Errorf("minipanel button %d = %#v", index, button)
		}
	}
	wantX := []int{0, 165, 293, 421, 549, 683}
	if len(hud.PanelParts) != len(wantX) {
		t.Fatalf("panel parts = %d, want %d", len(hud.PanelParts), len(wantX))
	}
	for index, part := range hud.PanelParts {
		if part.Frame != index || part.X != wantX[index] || part.Bottom != 600 {
			t.Errorf("panel part %d = %#v", index, part)
		}
	}
}

func TestInventoryPresentationUsesRecordGeometry(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(Shim(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Screens struct {
			Inventory struct {
				Records      string `json:"records"`
				RecordSuffix string `json:"record_suffix"`
				Panel        struct {
					Sheet  string `json:"sheet"`
					Frames []int  `json:"frames"`
				} `json:"panel"`
				Close struct {
					Sheet string `json:"sheet"`
				} `json:"close"`
				Slots []struct {
					ID     string `json:"id"`
					Prefix string `json:"prefix"`
				} `json:"slots"`
			} `json:"inventory"`
		} `json:"screens"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	inventory := manifest.Screens.Inventory
	if inventory.Records != "data/global/excel/Inventory.txt" || inventory.RecordSuffix != "2" {
		t.Fatalf("inventory record contract = %#v", inventory)
	}
	if inventory.Panel.Sheet == "" || inventory.Close.Sheet == "" || !reflect.DeepEqual(inventory.Panel.Frames, []int{4, 5, 7, 6}) {
		t.Fatalf("inventory presentation assets = %#v", inventory)
	}
	if len(inventory.Slots) != 10 {
		t.Fatalf("inventory equipment slots = %d, want 10", len(inventory.Slots))
	}
	for _, slot := range inventory.Slots {
		if slot.ID == "" || slot.Prefix == "" {
			t.Errorf("incomplete inventory slot = %#v", slot)
		}
	}
}
