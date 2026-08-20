package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestOwnedTargetSkillsPinSharedCastActionRecords ensures representative skills
// continue to share the cast action records expected by timing policy.
func TestOwnedTargetSkillsPinSharedCastActionRecords(t *testing.T) {
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

	wants := map[string]map[string]string{
		"0": {"anim": "A1", "seqtrans": "A1"},
		"36": {
			"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_cast_fire",
			"castoverlay": "fire_cast_1", "cltmissile": "firebolt",
		},
		"40": {
			"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_cast_cold",
			"dosound": "sorceress_frozenarmor", "castoverlay": "ice_cast_1",
		},
		"45": {
			"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_cast_cold",
			"castoverlay": "ice_cast_1", "cltmissile": "iceblast",
		},
		"47": {
			"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_cast_fire",
			"castoverlay": "fire_cast_2", "cltmissile": "fireball",
		},
		"48": {
			"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_cast_lightning",
			"castoverlay": "light_cast_1", "cltmissilea": "nova",
		},
		"52": {"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_enchant"},
		"54": {"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_teleport", "castoverlay": "teleport"},
		"55": {
			"anim": "SC", "seqtrans": "SC", "stsound": "sorceress_cast_cold",
			"castoverlay": "ice_cast_2", "cltmissile": "glacialspike",
		},
		"66": {
			"anim": "SC", "seqtrans": "SC", "stsound": "necromancer_curse_cast",
			"cltmissilea": "curseamplifydamage", "cltmissilec": "cursecast",
		},
		"72": {
			"anim": "SC", "seqtrans": "SC", "stsound": "necromancer_curse_cast",
			"cltmissilea": "curseweaken", "cltmissilec": "cursecast",
		},
	}
	for id, fields := range wants {
		row := rowBy(skills, "Id", id)
		if row == nil {
			t.Fatalf("owned expansion 1.14d skill %s is missing", id)
		}

		for field, want := range fields {
			if row[field] != want {
				t.Fatalf("owned expansion 1.14d skill %s %s = %q, want %q", id, field, row[field], want)
			}
		}
	}
}
