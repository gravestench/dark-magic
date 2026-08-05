package loot

import (
	"strings"
	"testing"
)

func TestParseSpecialItemAvailability(t *testing.T) {
	uniques, err := ParseUniqueItemsTSV(strings.NewReader("index\tversion\tenabled\tlvl\tcode\nNormal Unique\t0\t1\t10\tssd\nExpansion Unique\t100\t1\t20\tssd\nDisabled\t0\t0\t1\taxe\n"))
	if err != nil {
		t.Fatal(err)
	}
	sets, err := ParseSetItemsTSV(strings.NewReader("index\titem\tlvl\nSet Sword\tssd\t12\n"))
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
	unique := SpecialItems{"ssd": {{BaseCode: "ssd", Level: 20, Enabled: true}}}
	sets := SpecialItems{"ssd": {{BaseCode: "ssd", Level: 10, Enabled: true}}}

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
	uniques := SpecialItems{"ssd": {{BaseCode: "ssd", Level: 1, Version: 100, Enabled: true}}}
	quality, err := ApplyQualityEligibility(QualityUnique, BaseItem{Code: "ssd", Type: "rare"}, types, uniques, nil, EligibilityContext{Version: 0, DropLevel: 99})
	if err != nil || quality != QualityRare {
		t.Fatalf("quality = %q, error = %v", quality, err)
	}
}
