// Command quality_roll diagnoses deterministic ItemRatio quality calculations.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gravestench/dark-magic/internal/game/loot"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

func main() {
	file := flag.String("file", "", "path to itemratio.txt")
	version := flag.Int("version", 100, "item version (0 classic, 100 expansion)")
	uber := flag.Bool("uber", false, "use the exceptional/elite ratio row")
	classSpecific := flag.Bool("class-specific", false, "use the class-specific ratio row")
	mlvl := flag.Int("monster-level", 1, "monster or chest level")
	ilvl := flag.Int("item-level", 1, "base item level")
	mf := flag.Int("magic-find", 0, "player magic-find bonus percentage")
	unique := flag.Int("tc-unique", 0, "TreasureClass Unique modifier (0..1024)")
	set := flag.Int("tc-set", 0, "TreasureClass Set modifier (0..1024)")
	rare := flag.Int("tc-rare", 0, "TreasureClass Rare modifier (0..1024)")
	magic := flag.Int("tc-magic", 0, "TreasureClass Magic modifier (0..1024)")
	seed := flag.Uint64("seed", 1, "deterministic random seed")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "usage: quality_roll -file itemratio.txt [quality options]")
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
	ratios, err := loot.ParseItemRatiosTSV(bytes.NewReader(contents))
	if err != nil {
		fatal(err)
	}
	ratio, err := loot.SelectQualityRatio(ratios, *version, *uber, *classSpecific)
	if err != nil {
		fatal(err)
	}
	context := loot.QualityContext{
		MonsterLevel: *mlvl, ItemLevel: *ilvl, MagicFind: *mf,
		Modifiers: loot.QualityModifiers{Unique: *unique, Set: *set, Rare: *rare, Magic: *magic},
	}
	quality, chances, err := loot.RollQuality(ratio, context, *seed)
	if err != nil {
		fatal(err)
	}
	result := struct {
		Quality loot.Quality        `json:"quality"`
		Chances loot.QualityChances `json:"denominators"`
	}{quality, chances}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "quality_roll:", err)
	os.Exit(1)
}
