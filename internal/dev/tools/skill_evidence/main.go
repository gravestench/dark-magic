// Command skill_evidence joins exact target skill rows to layered localized
// descriptions and cross-skill formula references.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/skillevidence"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

func main() {
	mpqDirectory := flag.String("mpq-dir", os.Getenv("MPQ_DIRECTORY"), "owned Diablo II expansion 1.14d MPQ directory")
	skillList := flag.String("skill-ids", "0,36,40", "comma-separated exact skill IDs")
	language := flag.String("language", "English", "locale display name or Diablo directory token")
	flag.Parse()
	if *mpqDirectory == "" {
		fatal("-mpq-dir or MPQ_DIRECTORY is required")
	}
	expanded, err := darkpaths.ExpandHost(*mpqDirectory)
	if err != nil {
		fatal(err.Error())
	}
	if err := os.Setenv("MPQ_DIRECTORY", expanded); err != nil {
		fatal(err.Error())
	}
	mounted, err := content.FromEnvironment()
	if err != nil {
		fatal(err.Error())
	}
	defer mounted.Close()
	store := recordstore.New(mounted)
	store.SetLogger(nil)
	skills, err := store.Load("data/global/excel/skills.txt")
	if err != nil {
		fatal(err.Error())
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		fatal(err.Error())
	}
	report, err := skillevidence.Build(parseIDs(*skillList), skills, descriptions, localization.New(mounted, *language))
	if err != nil {
		fatal(err.Error())
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fatal(err.Error())
	}
}

func parseIDs(value string) []int {
	result := make([]int, 0)
	for _, raw := range strings.Split(value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || id < 0 {
			fatal(fmt.Sprintf("invalid skill ID %q", raw))
		}
		result = append(result, id)
	}
	return result
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
