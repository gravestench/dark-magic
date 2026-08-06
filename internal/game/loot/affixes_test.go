package loot

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseAffixesTSV(t *testing.T) {
	header := "Name\tversion\tspawnable\trare\tlevel\tmaxlevel\tfrequency\tgroup\titype1\tetype1\tmod1code\tmod1param\tmod1min\tmod1max\n"
	affixes, err := ParseAffixesTSV(strings.NewReader(header+"Strong\t0\t1\t1\t5\t20\t4\t7\tweap\tbow\tdmg%\t0\t10\t20\n"), AffixPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(affixes) != 1 || affixes[0].Name != "Strong" || affixes[0].Kind != AffixPrefix || affixes[0].Frequency != 4 || affixes[0].Includes[0] != "weap" || affixes[0].Modifiers[0].Maximum != 20 {
		t.Fatalf("affixes = %#v", affixes)
	}
}

func TestSelectAffixesIsDeterministicAndHonorsEligibility(t *testing.T) {
	types := ItemTypes{
		"weap": {Code: "weap"},
		"swor": {Code: "swor", Equiv1: "weap"},
	}
	item := BaseItem{Code: "ssd", Type: "swor"}
	prefixes := []Affix{
		{Name: "Alpha", Kind: AffixPrefix, Spawnable: true, Rare: true, Level: 1, Frequency: 10, Group: 1, Includes: []string{"weap"}},
		{Name: "Beta", Kind: AffixPrefix, Spawnable: true, Rare: true, Level: 1, Frequency: 1, Group: 1, Includes: []string{"weap"}},
		{Name: "Excluded", Kind: AffixPrefix, Spawnable: true, Rare: true, Level: 1, Frequency: 100, Group: 2, Includes: []string{"weap"}, Excludes: []string{"swor"}},
	}
	suffixes := []Affix{
		{Name: "Power", Kind: AffixSuffix, Spawnable: true, Rare: true, Level: 1, Frequency: 1, Group: 1, Includes: []string{"weap"}},
		{Name: "Speed", Kind: AffixSuffix, Spawnable: true, Rare: true, Level: 1, Frequency: 1, Group: 3, Includes: []string{"weap"}},
	}
	options := AffixOptions{Version: 100, AffixLevel: 10, MaxPrefixes: 3, MaxSuffixes: 3, MaxTotal: 4, Quality: QualityRare}
	wantPrefixes, wantSuffixes, err := SelectAffixes(item, types, prefixes, suffixes, options, 12)
	if err != nil {
		t.Fatal(err)
	}
	gotPrefixes, gotSuffixes, err := SelectAffixes(item, types, []Affix{prefixes[2], prefixes[1], prefixes[0]}, []Affix{suffixes[1], suffixes[0]}, options, 12)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotPrefixes, wantPrefixes) || !reflect.DeepEqual(gotSuffixes, wantSuffixes) {
		t.Fatalf("selection depends on input order: %#v/%#v != %#v/%#v", gotPrefixes, gotSuffixes, wantPrefixes, wantSuffixes)
	}
	seenGroups := make(map[int]bool)
	seenNames := make(map[string]bool)
	for _, affix := range append(append([]Affix{}, gotPrefixes...), gotSuffixes...) {
		if affix.Name == "Excluded" {
			t.Fatal("excluded affix was selected")
		}
		if affix.Group != 0 && seenGroups[affix.Group] {
			t.Fatalf("duplicate affix group %d", affix.Group)
		}
		seenGroups[affix.Group] = true
		key := string(affix.Kind) + ":" + affix.Name
		if seenNames[key] {
			t.Fatalf("duplicate affix %s", key)
		}
		seenNames[key] = true
	}
}

func TestRareCannotUseMagicOnlyAffix(t *testing.T) {
	types := ItemTypes{"weap": {Code: "weap"}}
	affix := Affix{Name: "Magic Only", Spawnable: true, Frequency: 1, Includes: []string{"weap"}}
	eligible, err := affixEligible(affix, BaseItem{Type: "weap"}, types, AffixOptions{AffixLevel: 1, Quality: QualityRare}, nil)
	if err != nil || eligible {
		t.Fatalf("eligible = %t, error = %v", eligible, err)
	}
}
