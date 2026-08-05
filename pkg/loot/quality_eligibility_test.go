package loot

import (
	"strings"
	"testing"
)

func TestParseSpecialItemAvailability(t *testing.T) {
	uniques, err := ParseUniqueItemsTSV(strings.NewReader("index\tversion\tenabled\trarity\tlvl\tcode\nNormal Unique\t0\t1\t2\t10\tssd\nExpansion Unique\t100\t1\t1\t20\tssd\nDisabled\t0\t0\t5\t1\taxe\n"))
	if err != nil {
		t.Fatal(err)
	}
	sets, err := ParseSetItemsTSV(strings.NewReader("index\titem\trarity\tlvl\nSet Sword\tssd\t1\t12\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(uniques["ssd"]) != 2 || !sets["ssd"][0].Enabled {
		t.Fatalf("uniques=%#v sets=%#v", uniques, sets)
	}
}

func TestApplyQualityEligibilityFallbacks(t *testing.T) {
	types := ItemTypes{
		"rare":   {Code: "rare", Rare: true},
		"common": {Code: "common"},
		"magic":  {Code: "magic", Magic: true, Rare: true},
		"normal": {Code: "normal", Normal: true, Magic: true, Rare: true},
	}
	unique := SpecialItems{"ssd": {{Name: "Unique Sword", BaseCode: "ssd", Level: 20, Rarity: 1, Enabled: true}}}
	sets := SpecialItems{"ssd": {{Name: "Set Sword", BaseCode: "ssd", Level: 10, Rarity: 1, Enabled: true}}}

	tests := []struct {
		name     string
		rolled   Quality
		typeCode string
		level    int
		want     Quality
	}{
		{name: "unique available", rolled: QualityUnique, typeCode: "rare", level: 20, want: QualityUnique},
		{name: "unique unavailable falls rare", rolled: QualityUnique, typeCode: "rare", level: 19, want: QualityRare},
		{name: "set unavailable then rare forbidden", rolled: QualitySet, typeCode: "common", level: 9, want: QualityMagic},
		{name: "rare forbidden", rolled: QualityRare, typeCode: "common", level: 20, want: QualityMagic},
		{name: "forced magic", rolled: QualityUnique, typeCode: "magic", level: 20, want: QualityMagic},
		{name: "forced normal wins", rolled: QualityUnique, typeCode: "normal", level: 20, want: QualityNormal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := BaseItem{Code: "ssd", Type: test.typeCode}
			got, err := ApplyQualityEligibility(test.rolled, item, types, unique, sets, EligibilityContext{Version: 100, DropLevel: test.level})
			if err != nil || got != test.want {
				t.Fatalf("quality = %q, error = %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestClassicCannotUseExpansionUnique(t *testing.T) {
	types := ItemTypes{"rare": {Code: "rare", Rare: true}}
	uniques := SpecialItems{"ssd": {{Name: "Expansion Sword", BaseCode: "ssd", Level: 1, Version: 100, Rarity: 1, Enabled: true}}}
	quality, err := ApplyQualityEligibility(QualityUnique, BaseItem{Code: "ssd", Type: "rare"}, types, uniques, nil, EligibilityContext{Version: 0, DropLevel: 99})
	if err != nil || quality != QualityRare {
		t.Fatalf("quality = %q, error = %v", quality, err)
	}
}

func TestSelectSpecialItemIsWeightedAndDeterministic(t *testing.T) {
	items := []SpecialItem{
		{Name: "Zulu", BaseCode: "ssd", Level: 1, Rarity: 1, Enabled: true},
		{Name: "Alpha", BaseCode: "ssd", Level: 1, Rarity: 3, Enabled: true},
		{Name: "Disabled", BaseCode: "ssd", Level: 1, Rarity: 100, Enabled: false},
		{Name: "Too High", BaseCode: "ssd", Level: 99, Rarity: 100, Enabled: true},
	}
	context := EligibilityContext{Version: 100, DropLevel: 20}
	want, err := SelectSpecialItem(items, context, 77)
	if err != nil {
		t.Fatal(err)
	}
	got, err := SelectSpecialItem([]SpecialItem{items[3], items[1], items[0], items[2]}, context, 77)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("selection depends on input order: %#v != %#v", got, want)
	}
}

func TestSelectSpecialItemRequiresPositiveEligibleWeight(t *testing.T) {
	_, err := SelectSpecialItem([]SpecialItem{{Name: "Zero", Enabled: true}}, EligibilityContext{}, 1)
	if err == nil || !strings.Contains(err.Error(), "no eligible") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseSpecialItemRejectsNegativeRarity(t *testing.T) {
	_, err := ParseSetItemsTSV(strings.NewReader("index\titem\trarity\tlvl\nBad\tssd\t-1\t1\n"))
	if err == nil || !strings.Contains(err.Error(), "must not be negative") {
		t.Fatalf("error = %v", err)
	}
}
