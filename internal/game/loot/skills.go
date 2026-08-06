package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Skill is the Skills.txt subset required by property function 36.
type Skill struct {
	ID    int    `json:"id"`
	Code  string `json:"code"`
	Class int    `json:"class"`
}

type Skills []Skill

var skillClass = map[string]int{"ama": 0, "sor": 1, "nec": 2, "pal": 3, "bar": 4, "dru": 5, "ass": 6}

func ParseSkillsTSV(input io.Reader) (Skills, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read skills header: %w", err)
	}
	columns := columnsByName(header)
	for _, name := range []string{"Id", "skill", "charclass"} {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("loot: skills table missing required column %q", name)
		}
	}
	var result Skills
	seen := make(map[int]bool)
	for rowNumber := 2; ; rowNumber++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("loot: read skills row %d: %w", rowNumber, err)
		}
		code := field(row, columns, "skill")
		classCode := strings.ToLower(field(row, columns, "charclass"))
		if code == "" || classCode == "" {
			continue
		}
		id, err := integerField(row, columns, "Id", rowNumber)
		if err != nil {
			return nil, err
		}
		class, ok := skillClass[classCode]
		if !ok {
			return nil, fmt.Errorf("loot: skills row %d has unknown character class %q", rowNumber, classCode)
		}
		if seen[id] {
			return nil, fmt.Errorf("loot: duplicate skill ID %d", id)
		}
		seen[id] = true
		result = append(result, Skill{ID: id, Code: code, Class: class})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
