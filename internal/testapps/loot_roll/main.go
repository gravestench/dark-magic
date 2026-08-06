// Command loot_roll exercises deterministic treasure-class rolls without a renderer.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/pkg/loot"
)

func main() {
	file := flag.String("file", "", "TreasureClass TSV or JSON class file")
	class := flag.String("class", "", "treasure class to roll")
	seed := flag.Uint64("seed", 1, "deterministic random seed")
	weapons := flag.String("weapons", "", "optional weapons.txt path for item resolution")
	armor := flag.String("armor", "", "optional armor.txt path for item resolution")
	misc := flag.String("misc", "", "optional misc.txt path for item resolution")
	itemTypes := flag.String("item-types", "", "optional ItemTypes.txt path for dynamic-code resolution")
	flag.Parse()
	if *file == "" || *class == "" {
		fmt.Fprintln(os.Stderr, "usage: loot_roll -file classes.json -class CLASS [-seed N]")
		os.Exit(2)
	}

	expandedFile, err := darkpaths.ExpandHost(*file)
	if err != nil {
		fatal(err)
	}
	contents, err := os.ReadFile(expandedFile)
	if err != nil {
		fatal(err)
	}
	var catalog loot.Catalog
	if filepath.Ext(*file) == ".json" {
		var classes []loot.Class
		if err := json.Unmarshal(contents, &classes); err != nil {
			fatal(fmt.Errorf("decode %s: %w", *file, err))
		}
		catalog = make(loot.Catalog, len(classes))
		for _, treasureClass := range classes {
			if treasureClass.Name == "" {
				fatal(fmt.Errorf("class has an empty name"))
			}
			if _, duplicate := catalog[treasureClass.Name]; duplicate {
				fatal(fmt.Errorf("duplicate class %q", treasureClass.Name))
			}
			catalog[treasureClass.Name] = treasureClass
		}
	} else {
		catalog, err = loot.ParseTreasureClassTSV(bytes.NewReader(contents))
		if err != nil {
			fatal(err)
		}
	}

	drops, err := loot.New(catalog, *seed).Roll(*class)
	if err != nil {
		fatal(err)
	}
	var output any = drops
	itemFiles := []struct {
		path string
		kind loot.ItemKind
	}{{*weapons, loot.ItemWeapon}, {*armor, loot.ItemArmor}, {*misc, loot.ItemMisc}}
	var itemCatalogs []loot.ItemCatalog
	for _, itemFile := range itemFiles {
		if itemFile.path == "" {
			continue
		}
		expandedPath, err := darkpaths.ExpandHost(itemFile.path)
		if err != nil {
			fatal(err)
		}
		contents, err := os.ReadFile(expandedPath)
		if err != nil {
			fatal(err)
		}
		items, err := loot.ParseBaseItemsTSV(bytes.NewReader(contents), itemFile.kind)
		if err != nil {
			fatal(err)
		}
		itemCatalogs = append(itemCatalogs, items)
	}
	if len(itemCatalogs) > 0 {
		items, err := loot.MergeItemCatalogs(itemCatalogs...)
		if err != nil {
			fatal(err)
		}
		if *itemTypes == "" {
			output = loot.ResolveBaseItems(drops, items)
		} else {
			expandedPath, err := darkpaths.ExpandHost(*itemTypes)
			if err != nil {
				fatal(err)
			}
			contents, err := os.ReadFile(expandedPath)
			if err != nil {
				fatal(err)
			}
			types, err := loot.ParseItemTypesTSV(bytes.NewReader(contents))
			if err != nil {
				fatal(err)
			}
			output, err = loot.ResolveItems(drops, items, types, *seed)
			if err != nil {
				fatal(err)
			}
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "loot_roll:", err)
	os.Exit(1)
}
