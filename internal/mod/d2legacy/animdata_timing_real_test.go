package d2legacy

import (
	"os"
	"testing"

	cof "github.com/gravestench/cof/pkg"
	"github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedTargetArchivesPinAttackAnimationTiming(t *testing.T) {
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
	if source, found := pinned.Source(recordstore.AnimationDataPath); !found || source.Layer == "" {
		t.Fatalf("AnimData provenance = %#v/%v", source, found)
	}
	catalog, err := assetdecode.AnimationData(pinned, recordstore.AnimationDataPath)
	if err != nil {
		t.Fatal(err)
	}
	record := catalog.GetRecord("AMA1HTH")
	if record == nil {
		t.Fatal("owned target is missing AMA1HTH")
	}
	events := record.Events()
	if generation.ID == "" || record.FramesPerDirection() != 13 || record.Speed() != 256 ||
		len(events) != 1 || events[8] != cof.EventAttack {
		t.Fatalf("generation=%s AMA1HTH frames=%d speed=%d events=%v", generation.ID,
			record.FramesPerDirection(), record.Speed(), events)
	}
}
