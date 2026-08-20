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

const (
	// These paths retain their source capitalization because archive-backed lookups may be case-sensitive.
	skillsPath            = "data/global/excel/skills.txt"
	skillDescriptionsPath = "data/global/excel/SkillDesc.txt"
)

// config groups the archive, skill-selection, and locale inputs that define one evidence report.
type config struct {
	mpqDirectory string
	skillIDs     string
	language     string
}

// evidenceTables keeps the two related record sets together so callers cannot accidentally swap their roles.
type evidenceTables struct {
	skills       []map[string]string
	descriptions []map[string]string
}

// main loads evidence from one mounted archive set and emits the established indented JSON document.
func main() {
	configuration := parseConfig()

	mounted, err := mountContent(configuration.mpqDirectory)
	if err != nil {
		fatal(err)
	}
	defer closeMountedContent(mounted)

	tables, err := loadEvidenceTables(mounted)
	if err != nil {
		fatal(err)
	}

	skillIDs, err := parseSkillIDs(configuration.skillIDs)
	if err != nil {
		fatal(err)
	}

	report, err := skillevidence.Build(
		skillIDs,
		tables.skills,
		tables.descriptions,
		localization.New(mounted, configuration.language),
	)
	if err != nil {
		fatal(err)
	}

	if err := encodeReport(report); err != nil {
		fatal(err)
	}
}

// parseConfig validates archive ownership input while retaining the command's established defaults.
func parseConfig() config {
	mpqDirectory := flag.String(
		"mpq-dir",
		os.Getenv("MPQ_DIRECTORY"),
		"owned Diablo II expansion 1.14d MPQ directory",
	)
	skillIDs := flag.String("skill-ids", "0,36,40", "comma-separated exact skill IDs")
	language := flag.String("language", "English", "locale display name or Diablo directory token")

	flag.Parse()

	if *mpqDirectory == "" {
		fatalMessage("-mpq-dir or MPQ_DIRECTORY is required")
	}

	return config{
		mpqDirectory: *mpqDirectory,
		skillIDs:     *skillIDs,
		language:     *language,
	}
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

// loadEvidenceTables disables record logging and loads tables in the historical order used by the command.
func loadEvidenceTables(mounted *content.FS) (evidenceTables, error) {
	store := recordstore.New(mounted)
	store.SetLogger(nil)

	skills, err := store.Load(skillsPath)
	if err != nil {
		return evidenceTables{}, err
	}

	descriptions, err := store.Load(skillDescriptionsPath)
	if err != nil {
		return evidenceTables{}, err
	}

	return evidenceTables{
		skills:       skills,
		descriptions: descriptions,
	}, nil
}

// parseSkillIDs preserves input order and duplicates because both can be meaningful in requested evidence output.
func parseSkillIDs(value string) ([]int, error) {
	rawIDs := strings.Split(value, ",")

	result := make([]int, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || id < 0 {
			return nil, fmt.Errorf("invalid skill ID %q", raw)
		}

		result = append(result, id)
	}

	return result, nil
}

// encodeReport writes one indented JSON document; preserving Encoder.Encode retains the trailing newline contract.
func encodeReport(report skillevidence.Report) error {
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
