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
	Schema = "d2legacy.skill-evidence.report/v1"
	Target = "diablo-ii-lod-1.14d-expansion"
)

var (
	replacementToken = regexp.MustCompile(`%[-+0-9.]*[A-Za-z]`)
	skillReference   = regexp.MustCompile(`skill\('([^']+)'\.(blvl|lvl)\)`)
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
	Table         string `json:"table"`
	Column        string `json:"column"`
	Formula       string `json:"formula"`
	ReferencedID  int    `json:"referenced_skill_id"`
	Referenced    string `json:"referenced_skill"`
	LevelSelector string `json:"level_selector"`
}

func Build(skillIDs []int, skills, descriptions []map[string]string, locale Localizer) (Report, error) {
	if len(skillIDs) == 0 || locale == nil {
		return Report{}, fmt.Errorf("skill evidence: skill IDs and locale are required")
	}
	byID, byName := map[int]map[string]string{}, map[string]map[string]string{}
	for rowNumber, row := range skills {
		id, err := strconv.Atoi(strings.TrimSpace(row["Id"]))
		if err != nil {
			return Report{}, fmt.Errorf("skill evidence: Skills.txt row %d has invalid Id %q", rowNumber+2, row["Id"])
		}
		byID[id] = row
		byName[strings.ToLower(strings.TrimSpace(row["skill"]))] = row
	}
	descriptionsByName := map[string]map[string]string{}
	for _, row := range descriptions {
		descriptionsByName[strings.ToLower(strings.TrimSpace(row["skilldesc"]))] = row
	}

	report := Report{Schema: Schema, Target: Target, Skills: make([]Skill, 0, len(skillIDs))}
	for _, id := range skillIDs {
		row := byID[id]
		if row == nil {
			return Report{}, fmt.Errorf("skill evidence: skill %d is absent from Skills.txt", id)
		}
		descriptionName := strings.TrimSpace(row["skilldesc"])
		description := descriptionsByName[strings.ToLower(descriptionName)]
		if description == nil {
			return Report{}, fmt.Errorf("skill evidence: skill %d references missing SkillDesc %q", id, descriptionName)
		}
		evidence := Skill{
			ID: id, Name: row["skill"], SkillDesc: descriptionName,
			CrossSkillModifiers: make([]CrossSkillModifier, 0),
		}
		evidence.Localization = localizationReferences(description, locale)
		for _, candidate := range []struct {
			table string
			row   map[string]string
		}{{"Skills.txt", row}, {"SkillDesc.txt", description}} {
			columns := sortedColumns(candidate.row)
			for _, column := range columns {
				formula := candidate.row[column]
				for _, match := range skillReference.FindAllStringSubmatch(formula, -1) {
					referenced := byName[strings.ToLower(match[1])]
					if referenced == nil {
						return Report{}, fmt.Errorf("skill evidence: skill %d formula references unknown skill %q", id, match[1])
					}
					referencedID, _ := strconv.Atoi(referenced["Id"])
					evidence.CrossSkillModifiers = append(evidence.CrossSkillModifiers, CrossSkillModifier{
						Table: candidate.table, Column: column, Formula: formula,
						ReferencedID: referencedID, Referenced: referenced["skill"], LevelSelector: match[2],
					})
				}
			}
		}
		report.Skills = append(report.Skills, evidence)
	}
	return report, nil
}

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

func isLocalizationColumn(column string) bool {
	normalized := strings.ToLower(column)
	return strings.HasPrefix(normalized, "str ") || strings.Contains(normalized, "texta") ||
		strings.Contains(normalized, "textb")
}

func sortedColumns(row map[string]string) []string {
	columns := make([]string, 0, len(row))
	for column := range row {
		columns = append(columns, column)
	}
	sort.Strings(columns)
	return columns
}
