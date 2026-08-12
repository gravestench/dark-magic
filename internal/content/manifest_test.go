package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/assets/catalog"
)

// TestShimPresentationManifestContract protects the architectural boundary
// between native engine code and mod-owned presentation knowledge. Go should
// provide capabilities; the shim manifest should name and describe assets.
func TestShimPresentationManifestContract(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(D2Legacy(), "manifests/presentation.v1.json")
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
			Color     struct {
				Red   int `json:"red"`
				Green int `json:"green"`
				Blue  int `json:"blue"`
				Alpha int `json:"alpha"`
			} `json:"color"`
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
			Screens []string `json:"screens"`
		} `json:"supported_profiles"`
		Resolution struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"resolution"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode presentation manifest: %v", err)
	}
	if manifest.Schema != "d2.presentation/v1" || manifest.Version != 1 {
		t.Fatalf("unexpected presentation contract %q version %d", manifest.Schema, manifest.Version)
	}
	if manifest.Resolution.Width <= 0 || manifest.Resolution.Height <= 0 {
		t.Fatalf("invalid presentation resolution: %#v", manifest.Resolution)
	}
	if len(manifest.Profiles) != 2 {
		t.Fatalf("supported presentation profiles = %d, want 2", len(manifest.Profiles))
	}
	desktop, gameplay := manifest.Profiles[0], manifest.Profiles[1]
	if desktop.ID != "lod-english-800x600" || desktop.GameVersion != "diablo-ii-lod" || desktop.Language != "English" ||
		desktop.Resolution.Width != manifest.Resolution.Width || desktop.Resolution.Height != manifest.Resolution.Height {
		t.Fatalf("unsupported or inconsistent desktop presentation profile: %#v", desktop)
	}
	if gameplay.ID != "lod-english-640x480-gameplay" || gameplay.Resolution.Width != 640 || gameplay.Resolution.Height != 480 || !reflect.DeepEqual(gameplay.Screens, []string{"game_world", "inventory", "character", "skills", "quests", "party", "help", "stash", "cube", "hireling", "vendor", "waypoint", "pause", "options", "automap", "death", "loading", "game_loading"}) {
		t.Fatalf("unsupported or inconsistent gameplay presentation profile: %#v", gameplay)
	}
	if len(manifest.Palettes) == 0 || len(manifest.Fonts) == 0 || len(manifest.Sounds) == 0 {
		t.Fatal("presentation manifest must own palette, font, and sound facts")
	}
	for _, name := range []string{"panel_heading", "panel_label", "panel_value", "frontend_version", "frontend_legal", "character_select_title", "character_select_metadata", "character_create_heading", "character_create_class", "character_create_description", "character_create_option", "credits", "button_normal", "button_hover", "label_button_normal", "label_button_hover", "dialog_text", "tooltip", "disabled", "font_lab_heading", "font_lab_caption", "font_lab_font6", "font_lab_font16", "font_lab_font30", "font_lab_font42", "font_lab_formal10", "font_lab_formal11", "font_lab_formal12", "font_lab_color", "font_lab_gold_sky", "font_lab_gold_fechar", "font_lab_gold_act1"} {
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
	for _, name := range []string{"button_normal", "button_hover", "label_button_normal", "label_button_hover", "dialog_text"} {
		if manifest.Styles[name].Transform != "" {
			t.Errorf("Exocet UI style %q must use its Units palette directly, not PL2 transform %q", name, manifest.Styles[name].Transform)
		}
	}
	for _, name := range []string{"button_normal", "label_button_normal", "dialog_text"} {
		got := manifest.Styles[name].Color
		if got.Red != 100 || got.Green != 100 || got.Blue != 100 || got.Alpha != 255 {
			t.Errorf("Exocet UI style %q modulation = %#v, want neutral 0x646464ff", name, got)
		}
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

	data, err := fs.ReadFile(D2Legacy(), "manifests/presentation.v1.json")
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
			selected, _, err := ApplyPresentationProfile(document, name)
			if err != nil {
				t.Fatal(err)
			}
			resolution := selected["resolution"].(map[string]any)
			width, height := resolution["width"].(float64), resolution["height"].(float64)
			if width <= 0 || height <= 0 {
				t.Fatalf("invalid resolution %.0fx%.0f", width, height)
			}
			screens := selected["screens"].(map[string]any)
			validatedScreens := any(screens)
			if scope, ok := profile["screens"].([]any); ok {
				scoped := make(map[string]any, len(scope))
				for _, value := range scope {
					id := value.(string)
					scoped[id] = screens[id]
				}
				validatedScreens = scoped
			} else {
				layouts := selected["layouts"].(map[string]any)
				tiles := layouts["frontend_tiles"].(map[string]any)
				if sumJSONNumbers(tiles["columns"].([]any)) != width || sumJSONNumbers(tiles["rows"].([]any)) != height {
					t.Fatalf("frontend tile grid does not cover %.0fx%.0f", width, height)
				}
			}
			validatePresentationGeometry(t, validatedScreens, width, height, "screens")
			language := profile["language"].(string)
			localeData, err := fs.ReadFile(D2Legacy(), "locales/"+language+".json")
			if err != nil {
				t.Fatalf("read %s locale: %v", language, err)
			}
			var localized map[string]string
			if err := json.Unmarshal(localeData, &localized); err != nil {
				t.Fatalf("decode %s locale: %v", language, err)
			}
			for key := range presentationLocaleKeys(validatedScreens) {
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

	data, err := fs.ReadFile(D2Legacy(), "manifests/asset-fixture.v1.json")
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
	if len(fixture.Assets) != 100 {
		t.Fatalf("asset fixture contains %d entries, want 100", len(fixture.Assets))
	}
}

func TestShimPresentationAssetCoverageBaseline(t *testing.T) {
	t.Parallel()

	manifestData, err := fs.ReadFile(D2Legacy(), "manifests/asset-catalog.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	fixtureData, err := fs.ReadFile(D2Legacy(), "manifests/asset-fixture.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest assetcatalog.Manifest
	var fixture assetcatalog.Fixture
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(fixtureData, &fixture); err != nil {
		t.Fatal(err)
	}
	coverage, err := assetcatalog.BuildCoverage(D2Legacy(), manifest, fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(coverage.CatalogFixtureGaps) != 0 {
		t.Fatalf("catalog/fixture join gaps: %v", coverage.CatalogFixtureGaps)
	}
	// Map Labs intentionally add the Act II-V DAT/PL2 palettes as code-owned
	// developer choices. Warp Lab likewise names the verified TP/PP ON-mode COF
	// and HD/TR components recovered from OpenDiablo2 and visually checked against
	// mounted production assets. Random map/tile paths remain dynamic VFS
	// discoveries and therefore do not pretend every mounted asset is manifest-owned.
	// The d2legacy Fire Bolt slice now declares its immutable Skills.txt and
	// Missiles.txt inputs directly. They are code-owned data rather than
	// presentation assets; missile art remains covered by the audited dynamic
	// data/global/missiles prefix.
	const auditedFingerprint = "3716f2b4c45fd566467250872f0213d06a6ebcbaee962c11efbce4953d32b2ac"
	if coverage.Fingerprint != auditedFingerprint {
		t.Fatalf("presentation asset coverage changed: got %s, want audited %s; run `make presentation-coverage` and classify every changed path", coverage.Fingerprint, auditedFingerprint)
	}
}

func TestCharacterCreationTransitionFacts(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(D2Legacy(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Screens struct {
			CharacterCreate struct {
				Palette       string `json:"palette"`
				ClassPalette  string `json:"class_palette"`
				IdleFPS       int    `json:"idle_frames_per_second"`
				TransitionFPS int    `json:"transition_frames_per_second"`
				Campfire      struct {
					Sheet string `json:"sheet"`
					Z     int    `json:"z"`
				} `json:"campfire"`
				Stage map[string]struct {
					Anchor struct{ X, Y int }                `json:"anchor"`
					Z      int                               `json:"z"`
					Hit    struct{ X, Y, Width, Height int } `json:"hit"`
				} `json:"stage"`
				Classes []struct {
					Class          string `json:"class"`
					Forward        string `json:"forward"`
					ForwardOverlay string `json:"forward_overlay"`
					ForwardFrames  int    `json:"forward_frames"`
					Back           string `json:"back"`
					BackOverlay    string `json:"back_overlay"`
					BackFrames     int    `json:"back_frames"`
				} `json:"classes"`
			} `json:"character_create"`
		} `json:"screens"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	classes := manifest.Screens.CharacterCreate.Classes
	creation := manifest.Screens.CharacterCreate
	if creation.Palette != "fechar" || creation.ClassPalette != "fechar" || creation.IdleFPS != 15 || creation.TransitionFPS != 25 || creation.Campfire.Sheet == "" || creation.Campfire.Z <= 0 {
		t.Fatalf("character creation palette/campfire facts = %#v", creation)
	}
	if len(classes) != 7 {
		t.Fatalf("character creation classes = %d, want 7", len(classes))
	}
	for _, class := range classes {
		if class.Class == "" || class.Forward == "" || class.Back == "" || class.ForwardFrames <= 0 || class.BackFrames <= 0 {
			t.Errorf("incomplete walk transition for %#v", class)
		}
		wantForwardOverlay := map[string]bool{"Amazon": true, "Sorceress": true, "Necromancer": true, "Paladin": true, "Barbarian": true}[class.Class]
		wantBackOverlay := map[string]bool{"Sorceress": true, "Necromancer": true}[class.Class]
		if (class.ForwardOverlay != "") != wantForwardOverlay || (class.BackOverlay != "") != wantBackOverlay {
			t.Errorf("unexpected shipped overlay pairing for %q: forward=%q back=%q", class.Class, class.ForwardOverlay, class.BackOverlay)
		}
		placement, ok := creation.Stage[class.Class]
		if !ok || placement.Anchor.X <= 0 || placement.Anchor.Y <= 0 || placement.Z <= 0 || placement.Hit.Width <= 0 || placement.Hit.Height <= 0 {
			t.Errorf("missing calibrated stage placement for %q: %#v", class.Class, placement)
		}
	}
	if creation.Stage["Paladin"].Z <= creation.Stage["Barbarian"].Z {
		t.Errorf("Paladin depth %d must render in front of Barbarian depth %d", creation.Stage["Paladin"].Z, creation.Stage["Barbarian"].Z)
	}
	for class, placement := range creation.Stage {
		if placement.Z >= creation.Campfire.Z {
			t.Errorf("class %q depth %d must remain behind foreground campfire depth %d", class, placement.Z, creation.Campfire.Z)
		}
	}
}

func TestGameHUDCompositionFacts(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(D2Legacy(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Screens struct {
			GameWorld struct {
				Map struct {
					DS1     string   `json:"ds1"`
					DT1     []string `json:"dt1"`
					Palette string   `json:"palette"`
				} `json:"map"`
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
					Belt struct {
						Sheet      string `json:"sheet"`
						X          int    `json:"x"`
						Y          int    `json:"y"`
						Columns    int    `json:"columns"`
						Rows       int    `json:"rows"`
						CellWidth  int    `json:"cell_width"`
						CellHeight int    `json:"cell_height"`
					} `json:"belt"`
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
							Scene   string `json:"scene"`
							Slot    string `json:"slot"`
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
	worldMap := manifest.Screens.GameWorld.Map
	if worldMap.DS1 == "" || len(worldMap.DT1) == 0 || worldMap.Palette == "" {
		t.Fatalf("incomplete world-map asset facts: %#v", worldMap)
	}
	if hud.PanelSheet == "" || hud.Globes.Sheet == "" || hud.Globes.OverlapSheet == "" || hud.Skills.Sheet == "" {
		t.Fatalf("incomplete HUD asset facts: %#v", hud)
	}
	// The expansion panel begins with a 117px globe and 48px skill well.
	// Riiablo places the belt at x=177 inside the following control widget:
	// 117 + 48 + 177 = 342.
	if hud.Belt.Sheet == "" || hud.Belt.X != 342 || hud.Belt.Y != 561 || hud.Belt.Columns != 4 || hud.Belt.Rows != 4 || hud.Belt.CellWidth != 31 || hud.Belt.CellHeight != 31 {
		t.Fatalf("unexpected desktop belt facts: %#v", hud.Belt)
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
		if button.ID == "messages" && (!button.Enabled || button.Scene != "messages") {
			t.Errorf("messages minipanel route = %#v, want enabled messages scene", button)
		}
		if button.Slot != "left" && button.Slot != "right" && button.Slot != "full" {
			t.Errorf("minipanel button %q has invalid overlay slot %q", button.ID, button.Slot)
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

	data, err := fs.ReadFile(D2Legacy(), "manifests/presentation.v1.json")
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

func Test640GameplayProfileUsesClassicOverlayGeometry(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(D2Legacy(), "manifests/presentation.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	selected, _, err := ApplyPresentationProfile(document, "lod-english-640x480-gameplay")
	if err != nil {
		t.Fatal(err)
	}
	screens := selected["screens"].(map[string]any)
	inventory := screens["inventory"].(map[string]any)
	panel := inventory["panel"].(map[string]any)
	if inventory["record_suffix"] != "" || panel["origin_y_correction"] != float64(4) || inventory["offset_x"] != float64(-80) || inventory["offset_y"] != float64(-60) {
		t.Fatalf("classic inventory geometry = %#v", inventory)
	}
	for _, id := range []string{"character", "quests", "party"} {
		screen := screens[id].(map[string]any)
		if screen["offset_x"] != float64(-80) || screen["offset_y"] != float64(-60) {
			t.Errorf("classic %s offset = %v,%v", id, screen["offset_x"], screen["offset_y"])
		}
	}
	skills := screens["skills"].(map[string]any)
	if skills["x"] != float64(320) || skills["y"] != float64(4) {
		t.Fatalf("classic skill-tree origin = %v,%v", skills["x"], skills["y"])
	}
	help := screens["help"].(map[string]any)
	helpBorder := help["border"].(map[string]any)
	placements := helpBorder["placements"].([]any)
	if helpBorder["sheet"] != "data/global/ui/MENU/helpborder.DC6" || len(placements) != 8 {
		t.Fatalf("classic help border = %#v", helpBorder)
	}
	last := placements[7].(map[string]any)
	if last["frame"] != float64(7) || last["x"] != float64(576) || last["y"] != float64(256) {
		t.Fatalf("classic help lower-right placement = %#v", last)
	}
	for _, id := range []string{"stash", "cube", "hireling", "vendor", "waypoint"} {
		screen := screens[id].(map[string]any)
		if screen["x"] != float64(0) || screen["y"] != float64(4) {
			t.Errorf("classic fixed panel %s origin = %v,%v", id, screen["x"], screen["y"])
		}
		if screen["sheet"] == "" || screen["close"] == nil {
			t.Errorf("incomplete fixed panel %s = %#v", id, screen)
		}
	}
	automap := screens["automap"].(map[string]any)
	if automap["x"] != float64(320) || automap["y"] != float64(240) || automap["width"] != float64(540) || automap["height"] != float64(380) {
		t.Fatalf("classic automap viewport = %#v", automap)
	}
	death := screens["death"].(map[string]any)
	if death["x"] != float64(320) || death["width"] != float64(600) || death["died_y"] != float64(145) {
		t.Fatalf("classic death presentation = %#v", death)
	}
	for _, id := range []string{"loading", "game_loading"} {
		screen := screens[id].(map[string]any)
		if screen["x"] != float64(320) || screen["y"] != float64(240) || screen["width"] != float64(640) || screen["height"] != float64(480) {
			t.Errorf("classic %s viewport = %#v", id, screen)
		}
	}
}
