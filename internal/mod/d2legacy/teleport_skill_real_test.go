package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestOwnedTargetTeleportRecords pins teleport's movement and resource fields
// so archive updates cannot quietly alter traversal semantics.
func TestOwnedTargetTeleportRecords(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = assets.Close() }()

	store := recordstore.New(assets)
	store.SetLogger(nil)

	skills, err := store.Load("data/global/excel/skills.txt")
	if err != nil {
		t.Fatal(err)
	}

	teleport := rowBy(skills, "Id", "54")
	if teleport == nil {
		t.Fatal("owned expansion 1.14d Teleport row is missing")
	}

	for field, want := range map[string]string{
		"skill": "Teleport", "srvstfunc": "", "srvdofunc": "27", "cltstfunc": "", "cltdofunc": "",
		"anim": "SC", "monanim": "xx", "range": "none", "warp": "1", "mana": "24", "lvlmana": "-1",
		"manashift": "8", "minmana": "1", "interrupt": "1", "InTown": "", "TargetableOnly": "",
		"LineOfSight": "", "leftskill": "", "general": "", "InGame": "1",
	} {
		if teleport[field] != want {
			t.Fatalf("owned expansion 1.14d Teleport %s = %q, want %q", field, teleport[field], want)
		}
	}

	levels, err := store.Load("data/global/excel/Levels.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range levels {
		id := row["Id"]
		if id == "" {
			continue
		}

		want := "1"

		switch id {
		case "0":
			want = "0"
		case "73":
			want = "2"
		}

		if row["Teleport"] != want {
			t.Fatalf("owned expansion 1.14d Levels[%s] %q Teleport = %q, want %q", id, row["Name"], row["Teleport"], want)
		}
	}
}
