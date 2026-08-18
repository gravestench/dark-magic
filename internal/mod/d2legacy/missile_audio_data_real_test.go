package d2legacy

import (
	"os"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedTargetArchivesPinMissileAudioEvidence(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	defer assets.Close()
	pinned, generation, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}
	missiles, err := pinned.Load("data/global/excel/Missiles.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(generation.ID, "sha256:") {
		t.Fatalf("record generation ID = %q", generation.ID)
	}
	for id, expected := range map[string][4]string{
		"firebolt":     {"sorceress_firebolt_1", "sorceress_firebolt_impact_1", "fireexplode", ""},
		"fireball":     {"sorceress_fireball_1", "sorceress_fireball_impact_1", "explodingarrowexp", "fireexplosion2"},
		"nova":         {"sorceress_nova", "", "", ""},
		"iceblast":     {"sorceress_icebolt_1", "sorceress_iceblast_impact_1", "freezingarrowexp1", ""},
		"glacialspike": {"sorceress_glacialspike_1", "sorceress_iceblast_impact_1", "", "freezingarrowexp1"},
	} {
		row := rowBy(missiles, "Missile", id)
		if row == nil || row["TravelSound"] != expected[0] || row["HitSound"] != expected[1] ||
			row["ExplosionMissile"] != expected[2] || row["CltHitSubMissile1"] != expected[3] {
			t.Fatalf("missile %s audio/effect references = %#v", id, row)
		}
	}
	sounds, err := pinned.Load("data/global/excel/Sounds.txt")
	if err != nil {
		t.Fatal(err)
	}
	for id, expected := range map[string][3]string{
		"sorceress_firebolt_1":        {"skill\\sorceress\\firebolt1.wav", "3", "1"},
		"sorceress_firebolt_impact_1": {"skill\\sorceress\\largefireimpact1.wav", "3", "0"},
		"sorceress_fireball_1":        {"skill\\sorceress\\fireball1.wav", "3", "1"},
		"sorceress_fireball_impact_1": {"skill\\sorceress\\boom1.wav", "3", "0"},
		"sorceress_nova":              {"skill\\sorceress\\novaelec.wav", "0", "0"},
		"sorceress_icebolt_1":         {"skill\\sorceress\\icebolt1.wav", "3", "1"},
		"sorceress_iceblast_impact_1": {"skill\\sorceress\\blastimpact1.wav", "3", "0"},
		"sorceress_glacialspike_1":    {"skill\\sorceress\\icespike1.wav", "3", "1"},
	} {
		row := rowBy(sounds, "Sound", id)
		if row == nil || row["FileName"] != expected[0] || row["Group Size"] != expected[1] ||
			row["Loop"] != expected[2] || row["Stream"] != "0" {
			t.Fatalf("sound %s playback record = %#v", id, row)
		}
	}
}
