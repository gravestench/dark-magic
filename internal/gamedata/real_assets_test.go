package gamedata

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/recordstore"
	"github.com/gravestench/dark-magic/pkg/models"
)

func TestRealArchivesDecodeTypedCoreTables(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	store := recordstore.New(assets)
	characters, err := Load[models.CharStats](store, CharStatsTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(characters) < 7 {
		t.Fatalf("typed charstats rows = %d, want at least seven classes", len(characters))
	}
	byClass, err := Index(characters, func(record models.CharStats) string { return record.Class })
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := byClass["Amazon"]; !exists {
		t.Fatal("typed charstats index is missing Amazon")
	}
	sounds, err := Load[models.SoundEntry](store, SoundsTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(sounds) == 0 {
		t.Fatal("typed sounds table is empty")
	}
	catalog := New(store)
	snapshot, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.LevelsByID) == 0 || len(snapshot.ObjectsByClass) == 0 || len(snapshot.SkillsByID) == 0 || len(snapshot.TreasureByName) == 0 {
		t.Fatal("typed core catalog indexes are incomplete")
	}
	if len(snapshot.ArmorByCode) == 0 || len(snapshot.WeaponsByCode) == 0 || len(snapshot.MiscByCode) == 0 || len(snapshot.ItemTypesByCode) == 0 {
		t.Fatal("typed base-item catalog indexes are incomplete")
	}
	if len(snapshot.Issues) == 0 {
		t.Fatal("expected shipped-data diagnostics for known duplicate/sentinel records")
	}
	for name, load := range map[string]func() (int, error){
		"levels": func() (int, error) {
			records, err := Load[models.LevelData](store, LevelsTable)
			return len(records), err
		},
		"objects": func() (int, error) {
			records, err := Load[models.Object](store, ObjectsTable)
			return len(records), err
		},
		"skills": func() (int, error) {
			records, err := Load[models.SkillData](store, SkillsTable)
			return len(records), err
		},
		"treasure classes": func() (int, error) {
			records, err := Load[models.TreasureClassEx](store, TreasureClassExTable)
			return len(records), err
		},
		"armor": func() (int, error) {
			records, err := Load[models.ItemArmor](store, ArmorTable)
			return len(records), err
		},
		"weapons": func() (int, error) {
			records, err := Load[models.ItemWeapon](store, WeaponsTable)
			return len(records), err
		},
		"misc items": func() (int, error) {
			records, err := Load[models.MiscItem](store, MiscTable)
			return len(records), err
		},
		"item types": func() (int, error) {
			records, err := Load[models.ItemType](store, ItemTypesTable)
			return len(records), err
		},
	} {
		count, err := load()
		if err != nil {
			t.Errorf("decode typed %s: %v", name, err)
		} else if count == 0 {
			t.Errorf("typed %s table is empty", name)
		}
	}
}
