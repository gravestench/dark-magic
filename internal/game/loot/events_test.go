package loot

import (
	"reflect"
	"testing"
)

func TestMonsterAndChestEventsProduceReplayableLootStreams(t *testing.T) {
	catalog := Catalog{"drop": {Name: "drop", Picks: 4, Entries: []Entry{{Code: "a", Weight: 1}, {Code: "b", Weight: 1}}}}
	event := Event{Kind: EventMonster, EntityID: 17, Sequence: 3}
	want, err := RollEvent(catalog, "drop", 99, event)
	if err != nil {
		t.Fatal(err)
	}
	got, err := RollEvent(catalog, "drop", 99, event)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("replay differs: %#v != %#v", got, want)
	}
	monsterSeed, _ := EventSeed(99, event)
	chestSeed, _ := EventSeed(99, Event{Kind: EventChest, EntityID: 17, Sequence: 3})
	if monsterSeed == chestSeed {
		t.Fatal("monster and chest streams collided")
	}
}
