package content

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestPresentationBootstrapComesFromManifest proves startup facts remain manifest-owned and missing assets produce
// actionable setup guidance.
func TestPresentationBootstrapComesFromManifest(t *testing.T) {
	const manifest = `{
		"schema":"d2legacy.presentation/v1",
		"palettes":{"loading":"palette.dat"},
		"screens":{
			"title":{"background":"title.dc6"},
			"game_loading":{"sheet":"loading.dc6","palette":"loading"},
			"game_world":{"map":{"ds1":"town.ds1","dt1":["floor.dt1"]}}
		}
	}`

	source := fstest.MapFS{
		presentationManifest: {Data: []byte(manifest)},
		"title.dc6":          {Data: []byte("fixture")},
	}

	bootstrap, err := LoadPresentationBootstrap(source)
	if err != nil {
		t.Fatal(err)
	}

	if bootstrap.TitleBackground != "title.dc6" || len(bootstrap.LoadingAssets) != 2 ||
		bootstrap.LoadingAssets[0] != "loading.dc6" || bootstrap.LoadingAssets[1] != "palette.dat" {
		t.Fatalf("bootstrap = %#v", bootstrap)
	}

	if bootstrap.GameWorldMap.DS1 != "town.ds1" || len(bootstrap.GameWorldMap.DT1) != 1 ||
		bootstrap.GameWorldMap.DT1[0] != "floor.dt1" {
		t.Fatalf("game-world recipe = %#v", bootstrap.GameWorldMap)
	}

	if err := ValidateClientAssets(source); err != nil {
		t.Fatal(err)
	}

	delete(source, "title.dc6")

	if err := ValidateClientAssets(source); err == nil || !strings.Contains(err.Error(), "MPQ_DIRECTORY") {
		t.Fatalf("missing asset error = %v", err)
	}
}
