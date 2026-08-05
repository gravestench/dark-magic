package loot

import (
	"strings"
	"testing"
)

func TestParseAndInterpretItemProperties(t *testing.T) {
	propertyData := "code\tfunc1\tstat1\tset1\tval1\tfunc2\tstat2\tset2\tval2\n" +
		"all-stats\t1\tstrength\t0\t0\t3\tdexterity\t1\t0\n" +
		"on-hit\t11\titem_skillonattack\t0\t0\t0\t\t0\t0\n"
	properties, err := ParsePropertiesTSV(strings.NewReader(propertyData))
	if err != nil {
		t.Fatal(err)
	}
	statData := "Stat\tSigned\tValShift\tMinAccr\nstrength\t1\t0\t-100\ndexterity\t1\t0\t-100\nitem_skillonattack\t0\t0\t0\n"
	stats, err := ParseItemStatCostTSV(strings.NewReader(statData))
	if err != nil {
		t.Fatal(err)
	}
	item := GeneratedItem{Prefixes: []RolledAffix{{Name: "Balanced", Modifiers: []RolledModifier{{Code: "all-stats", Value: 5}, {Code: "on-hit", Parameter: 7, Value: 3}}}}}
	item, err = InterpretItemProperties(item, properties, stats)
	if err != nil {
		t.Fatal(err)
	}
	if len(item.Stats) != 2 || item.Stats[0].Code != "dexterity" || !item.Stats[0].Set || item.Stats[0].Function != 1 || item.Stats[1].Code != "strength" || item.Stats[1].Value != 5 {
		t.Fatalf("stats = %#v", item.Stats)
	}
	if len(item.Unsupported) != 1 || item.Unsupported[0].Function != 11 || item.Unsupported[0].Parameter != 7 {
		t.Fatalf("unsupported = %#v", item.Unsupported)
	}
}

func TestInterpretItemPropertiesRejectsUnknownReferences(t *testing.T) {
	item := GeneratedItem{Prefixes: []RolledAffix{{Name: "Bad", Modifiers: []RolledModifier{{Code: "missing", Value: 1}}}}}
	if _, err := InterpretItemProperties(item, nil, nil); err == nil || !strings.Contains(err.Error(), "unknown property") {
		t.Fatalf("unknown property error = %v", err)
	}
	properties := PropertyCatalog{"known": {Code: "known", Steps: []PropertyStep{{Function: 1, Stat: "missing"}}}}
	item.Prefixes[0].Modifiers[0].Code = "known"
	if _, err := InterpretItemProperties(item, properties, nil); err == nil || !strings.Contains(err.Error(), "unknown item stat") {
		t.Fatalf("unknown stat error = %v", err)
	}
}

func TestPropertyParsersReportMalformedNumbers(t *testing.T) {
	_, err := ParsePropertiesTSV(strings.NewReader("code\tfunc1\tstat1\np\tnope\tstrength\n"))
	if err == nil || !strings.Contains(err.Error(), "column \"func1\"") {
		t.Fatalf("property error = %v", err)
	}
	_, err = ParseItemStatCostTSV(strings.NewReader("Stat\tValShift\ns\tnope\n"))
	if err == nil || !strings.Contains(err.Error(), "column \"ValShift\"") {
		t.Fatalf("stat error = %v", err)
	}
}
