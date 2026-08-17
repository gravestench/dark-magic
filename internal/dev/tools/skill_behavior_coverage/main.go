// Command skill_behavior_coverage reports target Skills.txt behavior-family
// coverage from owned, mounted expansion 1.14d archives.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/behaviorcoverage"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

const (
	manifestPath = "manifests/skill-behavior-coverage.v1.json"
	skillsPath   = "data/global/excel/skills.txt"
	missilesPath = "data/global/excel/Missiles.txt"
)

func main() {
	mpqDirectory := flag.String("mpq-dir", os.Getenv("MPQ_DIRECTORY"), "directory containing owned Diablo II expansion 1.14d MPQ files")
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

	manifestData, err := fs.ReadFile(content.D2Legacy(), manifestPath)
	if err != nil {
		fatal(err.Error())
	}
	manifest, err := behaviorcoverage.DecodeManifest(bytes.NewReader(manifestData))
	if err != nil {
		fatal(err.Error())
	}
	store := recordstore.New(mounted)
	store.SetLogger(nil)
	skills, err := store.Load(skillsPath)
	if err != nil {
		fatal(err.Error())
	}
	missiles, err := store.Load(missilesPath)
	if err != nil {
		fatal(err.Error())
	}
	skillsSource, err := mounted.Resolve(skillsPath)
	if err != nil {
		fatal(err.Error())
	}
	missilesSource, err := mounted.Resolve(missilesPath)
	if err != nil {
		fatal(err.Error())
	}
	report, err := behaviorcoverage.Build(manifest, skills, missiles, behaviorcoverage.Sources{
		SkillsTable: skillsPath, SkillsLayer: skillsSource.Layer,
		MissilesTable: missilesPath, MissilesLayer: missilesSource.Layer,
	})
	if err != nil {
		fatal(err.Error())
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
