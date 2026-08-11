package models

import "testing"

func TestLevelLinksKeepVisibilityAndWarpSlotsPaired(t *testing.T) {
	level := LevelData{Vis0: 2, Warp0: 10, Vis3: 7, Warp3: 13, Vis7: 9, Warp7: -1}
	links := level.Links()
	want := []LevelLink{{Slot: 0, DestinationLevel: 2, WarpID: 10}, {Slot: 3, DestinationLevel: 7, WarpID: 13}, {Slot: 7, DestinationLevel: 9, WarpID: -1}}
	if len(links) != len(want) {
		t.Fatalf("links = %#v", links)
	}
	for index := range want {
		if links[index] != want[index] {
			t.Fatalf("link %d = %#v, want %#v", index, links[index], want[index])
		}
	}
}
