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
	if len(item.Unsupported) != 0 || len(item.Effects) != 1 || item.Effects[0].Kind != "proc" || item.Effects[0].SkillID != 7 {
		t.Fatalf("effects = %#v, unsupported = %#v", item.Effects, item.Unsupported)
	}
}

func TestInterpretSpecializedPropertyFunctions(t *testing.T) {
	properties := PropertyCatalog{
		"dmg-min":    {Code: "dmg-min", Steps: []PropertyStep{{Function: 5}}},
		"skilltab":   {Code: "skilltab", Steps: []PropertyStep{{Function: 10, Stat: "item_addskill_tab"}}},
		"charged":    {Code: "charged", Steps: []PropertyStep{{Function: 19, Stat: "item_charged_skill"}}},
		"indestruct": {Code: "indestruct", Steps: []PropertyStep{{Function: 20}}},
		"class":      {Code: "class", Steps: []PropertyStep{{Function: 21, Stat: "item_addclassskills", Value: 3}}},
		"ethereal":   {Code: "ethereal", Steps: []PropertyStep{{Function: 23}}},
		"future":     {Code: "future", Steps: []PropertyStep{{Function: 36, Stat: "future_stat"}}},
	}
	stats := StatCatalog{
		"item_addskill_tab":   {Code: "item_addskill_tab"},
		"item_charged_skill":  {Code: "item_charged_skill"},
		"item_addclassskills": {Code: "item_addclassskills"},
	}
	modifiers := []RolledModifier{
		{Code: "dmg-min", Minimum: 2, Maximum: 5, Value: 4},
		{Code: "skilltab", Parameter: 10, Minimum: 1, Maximum: 2, Value: 2},
		{Code: "charged", Parameter: 42, Minimum: 20, Maximum: 3, Value: 20},
		{Code: "indestruct", Minimum: 1, Maximum: 1, Value: 1},
		{Code: "class", Minimum: 1, Maximum: 2, Value: 2},
		{Code: "ethereal", Minimum: 1, Maximum: 1, Value: 1},
		{Code: "future", Value: 9},
	}
	item, err := InterpretItemProperties(GeneratedItem{Prefixes: []RolledAffix{{Name: "Special", Modifiers: modifiers}}}, properties, stats)
	if err != nil {
		t.Fatal(err)
	}
	if !item.Indestructible || !item.Ethereal {
		t.Fatalf("flags: indestructible=%t ethereal=%t", item.Indestructible, item.Ethereal)
	}
	if len(item.Effects) != 3 || item.Effects[0].Kind != "minimum_damage" || item.Effects[1].Class != 3 || item.Effects[1].SkillTab != 1 || item.Effects[2].Kind != "charged_skill" || item.Effects[2].SkillID != 42 || item.Effects[2].Charges != 20 || item.Effects[2].SkillLevel != 3 {
		t.Fatalf("effects = %#v", item.Effects)
	}
	if len(item.Stats) != 1 || item.Stats[0].Parameter != 3 || item.Stats[0].Value != 2 {
		t.Fatalf("stats = %#v", item.Stats)
	}
	if len(item.Unsupported) != 1 || item.Unsupported[0].Function != 36 {
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
