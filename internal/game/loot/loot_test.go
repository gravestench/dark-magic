package loot

import (
	"reflect"
	"strings"
	"testing"
)

func TestRollIsDeterministic(t *testing.T) {
	catalog := Catalog{"root": {Picks: 8, Entries: []Entry{{Code: "a", Weight: 1}, {Code: "b", Weight: 2}}}}
	want, err := New(catalog, 42).Roll("root")
	if err != nil {
		t.Fatal(err)
	}
	got, err := New(catalog, 42).Roll("root")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("same seed produced different drops: %#v != %#v", got, want)
	}
}

func TestRollNoDropAndNestedPath(t *testing.T) {
	catalog := Catalog{
		"root":   {Picks: 3, Entries: []Entry{{Code: "nested", Weight: 1}}},
		"nested": {Picks: -2, NoDrop: 99, Entries: []Entry{{Code: "rin", Weight: 1}, {Code: "amu", Weight: 1}}},
	}
	drops, err := New(catalog, 1).Roll("root")
	if err != nil {
		t.Fatal(err)
	}
	if len(drops) != 6 || drops[0].Code != "rin" || !reflect.DeepEqual(drops[0].Path, []string{"root", "nested"}) {
		t.Fatalf("unexpected nested drops: %#v", drops)
	}

	onlyNoDrop := Catalog{"empty": {Picks: 4, NoDrop: 1}}
	drops, err = New(onlyNoDrop, 7).Roll("empty")
	if err != nil || len(drops) != 0 {
		t.Fatalf("NoDrop roll = %#v, %v", drops, err)
	}
}

func TestNegativePicksUseWeightsAsCountsAndCapTotal(t *testing.T) {
	catalog := Catalog{"fixed": {Picks: -3, Entries: []Entry{{Code: "a", Weight: 2}, {Code: "b", Weight: 4}}}}
	drops, err := New(catalog, 99).Roll("fixed")
	if err != nil {
		t.Fatal(err)
	}
	want := []Drop{{Code: "a", Path: []string{"fixed"}}, {Code: "a", Path: []string{"fixed"}}, {Code: "b", Path: []string{"fixed"}}}
	if !reflect.DeepEqual(drops, want) {
		t.Fatalf("drops = %#v, want %#v", drops, want)
	}
}

func TestRollRejectsCyclesAndInvalidClasses(t *testing.T) {
	cycle := Catalog{"a": {Picks: -1, Entries: []Entry{{Code: "b", Weight: 1}}}, "b": {Picks: -1, Entries: []Entry{{Code: "a", Weight: 1}}}}
	if _, err := New(cycle, 0).Roll("a"); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle error = %v", err)
	}

	invalid := Catalog{"bad": {Picks: 1, Entries: []Entry{{Code: "x", Weight: -1}}}}
	if _, err := New(invalid, 0).Roll("bad"); err == nil || !strings.Contains(err.Error(), "negative") {
		t.Fatalf("validation error = %v", err)
	}
}
