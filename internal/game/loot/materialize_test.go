package loot

import (
	"reflect"
	"strings"
	"testing"
)

func TestMaterializeMagicItemRollsPropertiesDeterministically(t *testing.T) {
	base := BaseItem{Code: "ssd", Kind: ItemWeapon, LevelReq: 3}
	prefix := Affix{
		Name: "Strong", Kind: AffixPrefix, LevelReq: 8,
		Modifiers: []AffixModifier{{Code: "dmg%", Minimum: 10, Maximum: 20}, {Code: "str", Parameter: 1, Minimum: 2, Maximum: 2}},
	}
	suffix := Affix{Name: "Readiness", Kind: AffixSuffix, LevelReq: 6, Modifiers: []AffixModifier{{Code: "swing2", Minimum: 5, Maximum: 10}}}
	want, err := MaterializeItem(base, QualityMagic, nil, []Affix{prefix}, []Affix{suffix}, 42)
	if err != nil {
		t.Fatal(err)
	}
	got, err := MaterializeItem(base, QualityMagic, nil, []Affix{prefix}, []Affix{suffix}, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("same seed differs: %#v != %#v", got, want)
	}
	if got.LevelRequirement != 8 || got.Prefixes[0].Modifiers[0].Value < 10 || got.Prefixes[0].Modifiers[0].Value > 20 || got.Prefixes[0].Modifiers[1].Value != 2 {
		t.Fatalf("generated item = %#v", got)
	}
}

func TestMaterializeSpecialItemCopiesRecord(t *testing.T) {
	base := BaseItem{Code: "ssd"}
	special := SpecialItem{Name: "The Sword", BaseCode: "ssd", LevelReq: 12, Enabled: true, Rarity: 1}
	item, err := MaterializeItem(base, QualityUnique, &special, nil, nil, 1)
	if err != nil || item.Special == nil || item.Special.Name != "The Sword" || item.LevelRequirement != 12 {
		t.Fatalf("item = %#v, error = %v", item, err)
	}
	special.Name = "mutated"
	if item.Special.Name != "The Sword" {
		t.Fatal("generated item retained caller's special-item pointer")
	}
}

func TestMaterializeItemRejectsInconsistentInputs(t *testing.T) {
	tests := []struct {
		name    string
		quality Quality
		special *SpecialItem
		prefix  []Affix
		want    string
	}{
		{name: "missing unique", quality: QualityUnique, want: "requires a concrete"},
		{name: "special on normal", quality: QualityNormal, special: &SpecialItem{Name: "Wrong", BaseCode: "ssd"}, want: "must not have special"},
		{name: "affix on normal", quality: QualityNormal, prefix: []Affix{{Name: "Wrong"}}, want: "must not have magic affixes"},
		{name: "empty modifier", quality: QualityMagic, prefix: []Affix{{Name: "Bad", Modifiers: []AffixModifier{{Minimum: 2, Maximum: 1}}}}, want: "empty modifier code"},
		{name: "unknown quality", quality: Quality("mythic"), want: "unsupported item quality"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := MaterializeItem(BaseItem{Code: "ssd"}, test.quality, test.special, test.prefix, nil, 1)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}
