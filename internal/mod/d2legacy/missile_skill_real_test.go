package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

// TestTargetArchivesBootSkillBehaviorFamilies verifies every explicitly
// admitted definition against the user's owned expansion 1.14d tables. CI uses
// synthetic rows because copyrighted archives never enter the repository.
func TestTargetArchivesBootSkillBehaviorFamilies(t *testing.T) {
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

	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	pinned, _, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}
	missiles, err := pinned.Load("data/global/excel/Missiles.txt")
	if err != nil {
		t.Fatal(err)
	}
	for missile, want := range map[string]string{
		"firebolt":        "",
		"amphibiangoo1":   "33",
		"leapknockback":   "1",
		"moltenboulder":   "1",
		"baal cold maker": "75",
	} {
		row := rowBy(missiles, "Missile", missile)
		if row == nil || row["KnockBack"] != want {
			t.Fatalf("owned expansion 1.14d missile %q KnockBack = %#v, want %q", missile, row, want)
		}
	}
	authority, err := Start(t.Context(), content.D2Legacy(), pinned, engine, session, 314)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(t.Context())
}
