// Command loot_roll exercises deterministic treasure-class rolls without a renderer.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gravestench/dark-magic/pkg/loot"
)

func main() {
	file := flag.String("file", "", "JSON file containing an array of treasure classes")
	class := flag.String("class", "", "treasure class to roll")
	seed := flag.Uint64("seed", 1, "deterministic random seed")
	flag.Parse()
	if *file == "" || *class == "" {
		fmt.Fprintln(os.Stderr, "usage: loot_roll -file classes.json -class CLASS [-seed N]")
		os.Exit(2)
	}

	contents, err := os.ReadFile(*file)
	if err != nil {
		fatal(err)
	}
	var classes []loot.Class
	if err := json.Unmarshal(contents, &classes); err != nil {
		fatal(fmt.Errorf("decode %s: %w", *file, err))
	}
	catalog := make(loot.Catalog, len(classes))
	for _, treasureClass := range classes {
		if treasureClass.Name == "" {
			fatal(fmt.Errorf("class has an empty name"))
		}
		if _, duplicate := catalog[treasureClass.Name]; duplicate {
			fatal(fmt.Errorf("duplicate class %q", treasureClass.Name))
		}
		catalog[treasureClass.Name] = treasureClass
	}

	drops, err := loot.New(catalog, *seed).Roll(*class)
	if err != nil {
		fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(drops); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "loot_roll:", err)
	os.Exit(1)
}
