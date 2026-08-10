package clientapp

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/persistence"
)

func TestRealCharacterClassesReceiveAssignableStartingSkills(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	contentFS, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := gamedata.New(recordstore.New(contentFS)).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, class := range []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Assassin", "Druid"} {
		skills := learnedSkills(snapshot, persistence.Character{Class: class})
		if len(skills) == 0 {
			t.Errorf("%s received no assignable starting skills", class)
		}
	}
}
