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
	// These paths retain their source capitalization because archive-backed lookups may be case-sensitive.
	manifestPath = "manifests/skill-behavior-coverage.v1.json"
	skillsPath   = "data/global/excel/skills.txt"
	missilesPath = "data/global/excel/Missiles.txt"
)

// config identifies the archive directory that supplies every layer used by the coverage report.
type config struct {
	mpqDirectory string
}

// coverageInputs keeps decoded records beside their source metadata so the report remains auditable.
type coverageInputs struct {
	manifest behaviorcoverage.Manifest
	skills   []map[string]string
	missiles []map[string]string
	sources  behaviorcoverage.Sources
}

// main mounts the owned archives, loads one consistent input set, and emits the established indented JSON schema.
func main() {
	configuration := parseConfig()

	mounted, err := mountContent(configuration.mpqDirectory)
	if err != nil {
		fatal(err)
	}
	defer closeMountedContent(mounted)

	inputs, err := loadCoverageInputs(mounted)
	if err != nil {
		fatal(err)
	}

	report, err := behaviorcoverage.Build(
		inputs.manifest,
		inputs.skills,
		inputs.missiles,
		inputs.sources,
	)
	if err != nil {
		fatal(err)
	}

	if err := encodeReport(report); err != nil {
		fatal(err)
	}
}

// parseConfig validates the archive directory before mounting so missing ownership input fails immediately.
func parseConfig() config {
	mpqDirectory := flag.String(
		"mpq-dir",
		os.Getenv("MPQ_DIRECTORY"),
		"directory containing owned Diablo II expansion 1.14d MPQ files",
	)

	flag.Parse()

	if *mpqDirectory == "" {
		fatalMessage("-mpq-dir or MPQ_DIRECTORY is required")
	}

	return config{mpqDirectory: *mpqDirectory}
}

// mountContent normalizes host syntax before updating the environment consumed by the layered content loader.
func mountContent(mpqDirectory string) (*content.FS, error) {
	expanded, err := darkpaths.ExpandHost(mpqDirectory)
	if err != nil {
		return nil, err
	}

	// FromEnvironment is the shared construction path, so update its input only after expansion succeeds.
	if err := os.Setenv("MPQ_DIRECTORY", expanded); err != nil {
		return nil, err
	}

	return content.FromEnvironment()
}

// closeMountedContent performs best-effort cleanup because command output cannot recover from a close failure.
func closeMountedContent(mounted *content.FS) {
	_ = mounted.Close()
}

// loadCoverageInputs preserves load order and records resolved layers so the report identifies its exact evidence.
func loadCoverageInputs(mounted *content.FS) (coverageInputs, error) {
	manifestData, err := fs.ReadFile(content.D2Legacy(), manifestPath)
	if err != nil {
		return coverageInputs{}, err
	}

	manifest, err := behaviorcoverage.DecodeManifest(bytes.NewReader(manifestData))
	if err != nil {
		return coverageInputs{}, err
	}

	store := recordstore.New(mounted)
	store.SetLogger(nil)

	skills, err := store.Load(skillsPath)
	if err != nil {
		return coverageInputs{}, err
	}

	missiles, err := store.Load(missilesPath)
	if err != nil {
		return coverageInputs{}, err
	}

	sources, err := resolveCoverageSources(mounted)
	if err != nil {
		return coverageInputs{}, err
	}

	return coverageInputs{
		manifest: manifest,
		skills:   skills,
		missiles: missiles,
		sources:  sources,
	}, nil
}

// resolveCoverageSources identifies the winning content layers after loading both tables from the same mount.
func resolveCoverageSources(mounted *content.FS) (behaviorcoverage.Sources, error) {
	skillsSource, err := mounted.Resolve(skillsPath)
	if err != nil {
		return behaviorcoverage.Sources{}, err
	}

	missilesSource, err := mounted.Resolve(missilesPath)
	if err != nil {
		return behaviorcoverage.Sources{}, err
	}

	return behaviorcoverage.Sources{
		SkillsTable:   skillsPath,
		SkillsLayer:   skillsSource.Layer,
		MissilesTable: missilesPath,
		MissilesLayer: missilesSource.Layer,
	}, nil
}

// encodeReport writes one indented JSON document; preserving Encoder.Encode retains the trailing newline contract.
func encodeReport(report behaviorcoverage.Report) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(report)
}

// fatal reports the original error text without wrapping it, preserving this diagnostic command's CLI contract.
func fatal(err error) {
	fatalMessage(err.Error())
}

// fatalMessage prints validation failures in the same form as runtime failures and exits with status one.
func fatalMessage(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
