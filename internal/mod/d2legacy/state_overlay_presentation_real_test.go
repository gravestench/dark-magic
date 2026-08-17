package d2legacy

import (
	"io/fs"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedStateOverlayPresentationRecordsAndAssets(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	defer assets.Close()
	store := recordstore.New(assets)
	store.SetLogger(nil)
	overlays, err := store.Load("data/global/excel/Overlay.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]string{
		"frozenarmor":        {"Filename": "FrozenArmor", "Frames": "24", "AnimRate": "16", "Trans": "3"},
		"curse_hit":          {"Filename": "CurseHit", "Frames": "10", "AnimRate": "16", "Trans": "3"},
		"curseamplifydamage": {"Filename": "CurseAmplifyDamageEffect", "Frames": "24", "AnimRate": "16", "Trans": "3"},
		"curseweaken":        {"Filename": "CurseWeakenEffect", "Frames": "24", "AnimRate": "16", "Trans": "3"},
		"enchant":            {"Filename": "FireEnchant", "Frames": "17", "AnimRate": "16", "Trans": "3"},
	}
	for id, fields := range want {
		row := rowBy(overlays, "overlay", id)
		if row == nil {
			t.Fatalf("owned expansion 1.14d overlay %q is missing", id)
		}
		for field, value := range fields {
			if row[field] != value {
				t.Fatalf("owned expansion 1.14d overlay %s %s = %q, want %q", id, field, row[field], value)
			}
		}
		path := "data/global/overlays/" + row["Filename"] + ".dcc"
		if _, err := fs.Stat(assets, path); err != nil {
			t.Fatalf("owned expansion 1.14d overlay %s asset %q: %v", id, path, err)
		}
	}
}
