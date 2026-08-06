package loot

import (
	"strings"
	"testing"
)

func TestParseAndResolveBaseItems(t *testing.T) {
	weapons, err := ParseBaseItemsTSV(strings.NewReader("code\tnamestr\ttype\tlevel\tlevelreq\tinvfile\tflippyfile\nssd\tShortSword\tswor\t1\t0\tinvssd\tflpssd\n"), ItemWeapon)
	if err != nil {
		t.Fatal(err)
	}
	misc, err := ParseBaseItemsTSV(strings.NewReader("code\tnamestr\ttype\tlevel\nrin\tRing\tring\t1\n"), ItemMisc)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := MergeItemCatalogs(weapons, misc)
	if err != nil {
		t.Fatal(err)
	}
	drops := []Drop{{Code: "ssd", Path: []string{"root"}}, {Code: "armo3", Path: []string{"root"}}}
	resolved := ResolveBaseItems(drops, catalog)
	if !resolved[0].Resolved || resolved[0].Item.Kind != ItemWeapon || resolved[0].Item.NameKey != "ShortSword" {
		t.Fatalf("resolved item = %#v", resolved[0])
	}
	if resolved[1].Resolved || resolved[1].Item != nil {
		t.Fatalf("dynamic code should remain unresolved: %#v", resolved[1])
	}
}

func TestMergeItemCatalogsRejectsAmbiguousCode(t *testing.T) {
	_, err := MergeItemCatalogs(
		ItemCatalog{"dup": {Code: "dup", Kind: ItemWeapon}},
		ItemCatalog{"dup": {Code: "dup", Kind: ItemArmor}},
	)
	if err == nil || !strings.Contains(err.Error(), "both weapon and armor") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseBaseItemsReportsBadLevel(t *testing.T) {
	_, err := ParseBaseItemsTSV(strings.NewReader("code\tlevel\nx\thigh\n"), ItemArmor)
	if err == nil || !strings.Contains(err.Error(), "armor row 2 column \"level\"") {
		t.Fatalf("error = %v", err)
	}
}
