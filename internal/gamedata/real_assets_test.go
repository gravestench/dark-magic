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
}
