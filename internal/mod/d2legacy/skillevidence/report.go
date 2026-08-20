// Package skillevidence joins target Skills/SkillDesc records to layered TBL
// text and makes cross-skill formula references explicit for lawful research.
package skillevidence

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	Schema = "d2legacy.skill-evidence.report/v2"
	Target = "diablo-ii-lod-1.14d-expansion"
)

var (
	replacementToken = regexp.MustCompile(`%[-+0-9.]*[A-Za-z]`)
	skillReference   = regexp.MustCompile(`skill\('([^']+)'\.([A-Za-z][A-Za-z0-9_]*)\)`)
)

type Localizer interface {
	Resolve(key string) (value, source string, err error)
}

type Report struct {
	Schema string  `json:"schema"`
	Target string  `json:"target"`
	Skills []Skill `json:"skills"`
}

type Skill struct {
	ID                  int                     `json:"skill_id"`
	Name                string                  `json:"skill"`
	SkillDesc           string                  `json:"skilldesc"`
	Localization        []LocalizationReference `json:"localization"`
	CrossSkillModifiers []CrossSkillModifier    `json:"cross_skill_modifiers"`
}

type LocalizationReference struct {
	Column            string   `json:"column"`
	Key               string   `json:"key"`
	Text              string   `json:"text,omitempty"`
	Source            string   `json:"source,omitempty"`
	ReplacementTokens []string `json:"replacement_tokens"`
	Missing           bool     `json:"missing"`
}

type CrossSkillModifier struct {
	Table        string `json:"table"`
	Column       string `json:"column"`
	Formula      string `json:"formula"`
	ReferencedID int    `json:"referenced_skill_id"`
	Referenced   string `json:"referenced_skill"`
	Selector     string `json:"selector"`
}

// Build joins requested skills to descriptions, localization, and formula
// references. Requested order is preserved because callers use it for reports.
func Build(skillIDs []int, skills, descriptions []map[string]string, locale Localizer) (Report, error) {
	if len(skillIDs) == 0 || locale == nil {
		return Report{}, fmt.Errorf("skill evidence: skill IDs and locale are required")
	}

	byID, byName, err := indexSkills(skills)
	if err != nil {
		return Report{}, err
	}

	descriptionsByName := indexDescriptions(descriptions)

	report := Report{Schema: Schema, Target: Target, Skills: make([]Skill, 0, len(skillIDs))}
	for _, id := range skillIDs {
		evidence, buildErr := buildSkillEvidence(id, byID, byName, descriptionsByName, locale)
		if buildErr != nil {
			return Report{}, buildErr
		}

		report.Skills = append(report.Skills, evidence)
	}

	return report, nil
}

// indexSkills creates numeric and case-insensitive name indexes from Skills.txt.
// Invalid IDs fail with their original source row for actionable diagnostics.
func indexSkills(
	skills []map[string]string,
) (map[int]map[string]string, map[string]map[string]string, error) {
	byID := make(map[int]map[string]string, len(skills))

	byName := make(map[string]map[string]string, len(skills))
	for rowNumber, row := range skills {
		id, err := strconv.Atoi(strings.TrimSpace(row["Id"]))
		if err != nil {
			return nil, nil, fmt.Errorf(
				"skill evidence: Skills.txt row %d has invalid Id %q",
				rowNumber+2,
				row["Id"],
			)
		}

		byID[id] = row
		byName[strings.ToLower(strings.TrimSpace(row["skill"]))] = row
	}

	return byID, byName, nil
}

// indexDescriptions mirrors the target table's case-insensitive skilldesc join.
// Later duplicate rows retain the existing last-row-wins behavior.
func indexDescriptions(descriptions []map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(descriptions))
	for _, row := range descriptions {
		result[strings.ToLower(strings.TrimSpace(row["skilldesc"]))] = row
	}

	return result
}

// buildSkillEvidence resolves one requested skill and scans both source tables
// in stable column order so modifier output remains deterministic.
func buildSkillEvidence(
	id int,
	byID map[int]map[string]string,
	byName map[string]map[string]string,
	descriptionsByName map[string]map[string]string,
	locale Localizer,
) (Skill, error) {
	row := byID[id]
	if row == nil {
		return Skill{}, fmt.Errorf("skill evidence: skill %d is absent from Skills.txt", id)
	}

	descriptionName := strings.TrimSpace(row["skilldesc"])

	description := descriptionsByName[strings.ToLower(descriptionName)]
	if description == nil {
		return Skill{}, fmt.Errorf(
			"skill evidence: skill %d references missing SkillDesc %q",
			id,
			descriptionName,
		)
	}

	evidence := Skill{
		ID:                  id,
		Name:                row["skill"],
		SkillDesc:           descriptionName,
		Localization:        localizationReferences(description, locale),
		CrossSkillModifiers: make([]CrossSkillModifier, 0),
	}
	for _, candidate := range []formulaSource{
		{table: "Skills.txt", row: row},
		{table: "SkillDesc.txt", row: description},
	} {
		modifiers, err := crossSkillModifiers(id, candidate, byName)
		if err != nil {
			return Skill{}, err
		}

		evidence.CrossSkillModifiers = append(evidence.CrossSkillModifiers, modifiers...)
	}

	return evidence, nil
}

type formulaSource struct {
	table string
	row   map[string]string
}

// crossSkillModifiers extracts explicit skill('name'.selector) references from
// one table row. Unknown names fail rather than emitting incomplete evidence.
func crossSkillModifiers(
	skillID int,
	source formulaSource,
	byName map[string]map[string]string,
) ([]CrossSkillModifier, error) {
	result := make([]CrossSkillModifier, 0)

	for _, column := range sortedColumns(source.row) {
		formula := source.row[column]
		for _, match := range skillReference.FindAllStringSubmatch(formula, -1) {
			referenced := byName[strings.ToLower(match[1])]
			if referenced == nil {
				return nil, fmt.Errorf(
					"skill evidence: skill %d formula references unknown skill %q",
					skillID,
					match[1],
				)
			}

			referencedID, _ := strconv.Atoi(referenced["Id"])
			result = append(result, CrossSkillModifier{
				Table:        source.table,
				Column:       column,
				Formula:      formula,
				ReferencedID: referencedID,
				Referenced:   referenced["skill"],
				Selector:     match[2],
			})
		}
	}

	return result, nil
}

// localizationReferences resolves every localization-bearing description field
// while retaining missing keys as explicit evidence instead of dropping them.
func localizationReferences(description map[string]string, locale Localizer) []LocalizationReference {
	result := make([]LocalizationReference, 0)

	for _, column := range sortedColumns(description) {
		if !isLocalizationColumn(column) || strings.TrimSpace(description[column]) == "" {
			continue
		}

		key := strings.TrimSpace(description[column])
		value, source, err := locale.Resolve(key)

		reference := LocalizationReference{Column: column, Key: key, ReplacementTokens: []string{}}
		if err != nil {
			reference.Missing = true
		} else {
			reference.Text, reference.Source = value, source
			if tokens := replacementToken.FindAllString(value, -1); len(tokens) > 0 {
				reference.ReplacementTokens = tokens
			}
		}

		result = append(result, reference)
	}

	return result
}

// isLocalizationColumn applies the recovered SkillDesc naming convention; it
// intentionally stays permissive because the original table is irregular.
func isLocalizationColumn(column string) bool {
	normalized := strings.ToLower(column)

	return strings.HasPrefix(normalized, "str ") || strings.Contains(normalized, "texta") ||
		strings.Contains(normalized, "textb")
}

// sortedColumns removes Go map iteration from localization and formula evidence
// ordering, keeping reports reproducible.
func sortedColumns(row map[string]string) []string {
	columns := make([]string, 0, len(row))
	for column := range row {
		columns = append(columns, column)
	}

	sort.Strings(columns)

	return columns
}
