package loot

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolveDynamicItemCodes(t *testing.T) {
	types, err := ParseItemTypesTSV(strings.NewReader("ItemType\tCode\tEquiv1\tEquiv2\nWeapon\tweap\t\t\nSword\tswor\tweap\t\nKnife\tknif\tweap\t\n"))
	if err != nil {
		t.Fatal(err)
	}
	items := ItemCatalog{
		"clb": {Code: "clb", Kind: ItemWeapon, Type: "weap", Level: 1},
		"ssd": {Code: "ssd", Kind: ItemWeapon, Type: "swor", Level: 3},
		"lsd": {Code: "lsd", Kind: ItemWeapon, Type: "swor", Level: 5},
		"kri": {Code: "kri", Kind: ItemWeapon, Type: "knif", Level: 4},
		"hax": {Code: "hax", Kind: ItemWeapon, Type: "weap", Level: 8},
	}
	drops := []Drop{{Code: "weap3"}, {Code: "swor"}, {Code: "ssd"}, {Code: "weap20"}, {Code: "weap0"}}
	want, err := ResolveItems(drops, items, types, 44)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ResolveItems(drops, items, types, 44)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("same seed differs: %#v != %#v", got, want)
	}
	if !got[0].Resolved || got[0].Item.Level < 3 || got[0].Item.Level >= 6 {
		t.Fatalf("dynamic level result = %#v", got[0])
	}
	if !got[1].Resolved || got[1].Item.Type != "swor" {
		t.Fatalf("generic type result = %#v", got[1])
	}
	if got[2].Item.Code != "ssd" || got[3].Resolved {
		t.Fatalf("direct/unmatched results = %#v, %#v", got[2], got[3])
	}
	if !got[4].Resolved || got[4].Item.Level < 0 || got[4].Item.Level >= 3 {
		t.Fatalf("zero-level dynamic result = %#v", got[4])
	}
}

func TestResolveItemsDetectsTypeCycle(t *testing.T) {
	types := ItemTypes{"a": {Code: "a", Equiv1: "b"}, "b": {Code: "b", Equiv1: "a"}, "root": {Code: "root"}}
	items := ItemCatalog{"x": {Code: "x", Type: "a"}}
	_, err := ResolveItems([]Drop{{Code: "root"}}, items, types, 1)
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("error = %v", err)
	}
}
