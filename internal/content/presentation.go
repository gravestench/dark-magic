package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

const presentationManifest = "manifests/presentation.v1.json"

// PresentationBootstrap contains the small set of authored presentation assets
// needed before Lua scenes take ownership. Asset identity remains shim data;
// the native composition root only schedules the returned dependencies.
type PresentationBootstrap struct {
	TitleBackground string
	LoadingAssets   []string
	GameWorldMap    PresentationMapRecipe
}

// PresentationMapRecipe identifies the authored stamp used by the current
// playable-world fixture. Geometry and collision still come from DS1/DT1.
type PresentationMapRecipe struct {
	DS1 string
	DT1 []string
}

// LoadPresentationBootstrap reads startup facts from the versioned shim
// manifest without duplicating Blizzard paths or palette choices in Go.
func LoadPresentationBootstrap(source fs.FS) (PresentationBootstrap, error) {
	data, err := fs.ReadFile(source, presentationManifest)
	if err != nil {
		return PresentationBootstrap{}, fmt.Errorf("content: read presentation manifest: %w", err)
	}
	var document struct {
		Schema   string            `json:"schema"`
		Palettes map[string]string `json:"palettes"`
		Screens  struct {
			Title struct {
				Background string `json:"background"`
			} `json:"title"`
			GameLoading struct {
				Sheet   string `json:"sheet"`
				Palette string `json:"palette"`
			} `json:"game_loading"`
			GameWorld struct {
				Map struct {
					DS1 string   `json:"ds1"`
					DT1 []string `json:"dt1"`
				} `json:"map"`
			} `json:"game_world"`
		} `json:"screens"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		return PresentationBootstrap{}, fmt.Errorf("content: decode presentation manifest: %w", err)
	}
	if document.Schema != "d2.presentation/v1" {
		return PresentationBootstrap{}, fmt.Errorf("content: presentation schema is %q, want %q", document.Schema, "d2.presentation/v1")
	}
	loadingPalette := document.Palettes[document.Screens.GameLoading.Palette]
	if document.Screens.Title.Background == "" || document.Screens.GameLoading.Sheet == "" || loadingPalette == "" {
		return PresentationBootstrap{}, fmt.Errorf("content: presentation manifest has incomplete title or loading assets")
	}
	return PresentationBootstrap{
		TitleBackground: document.Screens.Title.Background,
		LoadingAssets:   []string{document.Screens.GameLoading.Sheet, loadingPalette},
		GameWorldMap: PresentationMapRecipe{
			DS1: document.Screens.GameWorld.Map.DS1,
			DT1: append([]string(nil), document.Screens.GameWorld.Map.DT1...),
		},
	}, nil
}

// ValidateClientAssets confirms that the manifest-selected bootstrap asset is
// supplied by the mounted game data and provides actionable setup guidance.
func ValidateClientAssets(source fs.FS) error {
	bootstrap, err := LoadPresentationBootstrap(source)
	if err != nil {
		return err
	}
	if _, err := fs.Stat(source, bootstrap.TitleBackground); err != nil {
		return fmt.Errorf("required Diablo II asset %q is unavailable; set MPQ_DIRECTORY to the directory containing the game MPQs: %w", bootstrap.TitleBackground, err)
	}
	return nil
}
