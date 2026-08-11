package clientapp

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func TestEntryWorldSpawnsKeepGameplayAndSeamCaptureDistinct(t *testing.T) {
	seam := gameworld.Seam{
		Town:       gameworld.SeamEndpoint{ArrivalX: 11, ArrivalY: 12},
		Wilderness: gameworld.SeamEndpoint{ArrivalX: 21, ArrivalY: 22},
	}

	entry, err := entryWorldSpawns("entry", seam, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if entry[1] != [2]float64{1, 2} || entry[2] != [2]float64{21, 22} {
		t.Fatalf("ordinary entry spawns = %#v", entry)
	}

	capture, err := entryWorldSpawns("seam", seam, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if capture[1] != [2]float64{11, 12} || capture[2] != [2]float64{21, 22} {
		t.Fatalf("seam capture spawns = %#v", capture)
	}

	if _, err := entryWorldSpawns("somewhere", seam, 1, 2); err == nil {
		t.Fatal("accepted an unknown fixture spawn")
	}
}
