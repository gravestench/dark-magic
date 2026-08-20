package recovered

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// ParseDS1Types parses unique executable-era definitions while ignoring the
// recovered Expansion section marker, which is metadata rather than a row.
func ParseDS1Types(input io.Reader) ([]DS1Type, error) {
	rows, err := readTSV(input)
	if err != nil {
		return nil, fmt.Errorf("recovered data: parse DS1 types: %w", err)
	}

	result := make([]DS1Type, 0, len(rows))

	seen := make(map[int]struct{}, len(rows))
	for index, row := range rows {
		if strings.TrimSpace(row["Def"]) == "" {
			// Riiablo preserves an "Expansion" section marker between the
			// classic and Act V definitions. It is not itself a definition.
			continue
		}

		definition, parseErr := requiredInt(row, "Def")
		if parseErr != nil {
			return nil, fmt.Errorf("recovered data: DS1 types line %d: %w", index+2, parseErr)
		}

		if _, exists := seen[definition]; exists {
			return nil, fmt.Errorf(
				"recovered data: DS1 types line %d: duplicate definition %d",
				index+2,
				definition,
			)
		}

		seen[definition] = struct{}{}

		levelType, parseErr := requiredInt(row, "LevelType")
		if parseErr != nil {
			return nil, fmt.Errorf("recovered data: DS1 types line %d: %w", index+2, parseErr)
		}

		name := strings.TrimSpace(row["Name"])
		if name == "" {
			return nil, fmt.Errorf("recovered data: DS1 types line %d: Name is required", index+2)
		}

		result = append(result, DS1Type{Name: name, Definition: definition, LevelType: levelType})
	}

	return result, nil
}

// ParseMapObjects parses act-local static object mappings and rejects duplicate
// composite keys so lookup construction cannot silently overwrite evidence.
func ParseMapObjects(input io.Reader) ([]MapObject, error) {
	rows, err := readTSV(input)
	if err != nil {
		return nil, fmt.Errorf("recovered data: parse map objects: %w", err)
	}

	result := make([]MapObject, 0, len(rows))

	seen := make(map[string]struct{}, len(rows))
	for index, row := range rows {
		act, actErr := requiredInt(row, "Act")
		id, idErr := requiredInt(row, "Id")

		objectID, objectErr := requiredInt(row, "ObjectId")
		if actErr != nil || idErr != nil || objectErr != nil || act < 1 || act > 5 {
			return nil, fmt.Errorf(
				"recovered data: map objects line %d: invalid Act, Id, or ObjectId",
				index+2,
			)
		}

		key := mapObjectKey(act, id)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf(
				"recovered data: map objects line %d: duplicate act/id %s",
				index+2,
				key,
			)
		}

		seen[key] = struct{}{}

		result = append(result, MapObject{
			Act: act, ID: id, Description: strings.TrimSpace(row["Description"]), ObjectID: objectID,
		})
	}

	return result, nil
}

// mapObjectKey creates the stable act-local lookup identity used by snapshots.
func mapObjectKey(act, id int) string {
	return fmt.Sprintf("%d:%d", act, id)
}

// ParseQuests reconstructs quest hierarchy, stage order, and prerequisite
// relationships. Cross-row prerequisites are validated after every ID is known.
func ParseQuests(input io.Reader) ([]Quest, error) {
	rows, err := readTSV(input)
	if err != nil {
		return nil, fmt.Errorf("recovered data: parse quests: %w", err)
	}

	result := make([]Quest, 0, len(rows))

	seen := make(map[int]struct{}, len(rows))
	for index, row := range rows {
		line := index + 2

		id, parseErr := requiredInt(row, "id")
		if parseErr != nil {
			return nil, fmt.Errorf("recovered data: quests line %d: %w", line, parseErr)
		}

		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("recovered data: quests line %d: duplicate id %d", line, id)
		}

		// Reserve the ID before parsing later fields to preserve the parser's
		// duplicate-first error precedence for repeated rows.
		seen[id] = struct{}{}

		quest, parseErr := parseQuest(row, line, id)
		if parseErr != nil {
			return nil, parseErr
		}

		result = append(result, quest)
	}

	for _, quest := range result {
		if quest.PrerequisiteID == nil {
			continue
		}

		if _, exists := seen[*quest.PrerequisiteID]; !exists {
			return nil, fmt.Errorf(
				"recovered data: quest %d references unknown prerequisite %d",
				quest.ID,
				*quest.PrerequisiteID,
			)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// parseQuest validates nonidentity fields and preserves six authored stage slots.
// The caller handles IDs first to retain duplicate-error precedence.
func parseQuest(row map[string]string, line, id int) (Quest, error) {
	act, err := requiredInt(row, "act")
	if err != nil || act < 0 || act > 4 {
		return Quest{}, fmt.Errorf("recovered data: quests line %d: invalid act %q", line, row["act"])
	}

	quest := Quest{
		ID:             id,
		Name:           strings.TrimSpace(row["name"]),
		Act:            act,
		Icon:           strings.TrimSpace(row["icon"]),
		TitleStringKey: strings.TrimSpace(row["qstr"]),
	}
	if quest.Name == "" || quest.TitleStringKey == "" {
		return Quest{}, fmt.Errorf("recovered data: quests line %d: name and qstr are required", line)
	}

	quest.Order, err = optionalInt(row, "order")
	if err != nil {
		return Quest{}, fmt.Errorf("recovered data: quests line %d: %w", line, err)
	}

	quest.Visible, err = optionalBool(row, "visible")
	if err != nil {
		return Quest{}, fmt.Errorf("recovered data: quests line %d: %w", line, err)
	}

	if value := strings.TrimSpace(row["questdone"]); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return Quest{}, fmt.Errorf(
				"recovered data: quests line %d: invalid questdone %q",
				line,
				value,
			)
		}

		quest.PrerequisiteID = &parsed
	}

	for stage := 1; stage <= 6; stage++ {
		primary := strings.TrimSpace(row[fmt.Sprintf("qsts%d", stage)])

		alternates := compact(
			row[fmt.Sprintf("qsts%da", stage)],
			row[fmt.Sprintf("qsts%db", stage)],
		)
		if primary != "" || len(alternates) > 0 {
			quest.Stages = append(quest.Stages, QuestStage{
				Index: stage, StringKey: primary, Alternates: alternates,
			})
		}
	}

	return quest, nil
}

// ParseSpeech parses unique case-insensitive logical sound names while keeping
// source row order for diagnostics and snapshot iteration.
func ParseSpeech(input io.Reader) ([]Speech, error) {
	rows, err := readTSV(input)
	if err != nil {
		return nil, fmt.Errorf("recovered data: parse speech: %w", err)
	}

	result := make([]Speech, 0, len(rows))

	seen := make(map[string]struct{}, len(rows))
	for index, row := range rows {
		entry := Speech{
			Sound: strings.TrimSpace(row["sound"]), StringKey: strings.TrimSpace(row["soundstr"]),
		}
		if entry.Sound == "" || entry.StringKey == "" {
			return nil, fmt.Errorf(
				"recovered data: speech line %d: sound and soundstr are required",
				index+2,
			)
		}

		key := strings.ToLower(entry.Sound)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf(
				"recovered data: speech line %d: duplicate sound %q",
				index+2,
				entry.Sound,
			)
		}

		seen[key] = struct{}{}

		result = append(result, entry)
	}

	return result, nil
}

// readTSV decodes irregular recovered tables by allowing short and long rows.
// Missing trailing columns remain absent map entries and therefore zero values.
func readTSV(input io.Reader) ([]map[string]string, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = false

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	for index := range header {
		header[index] = strings.TrimSpace(header[index])
	}

	var rows []map[string]string

	for {
		values, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			return nil, readErr
		}

		row := make(map[string]string, len(header))
		for index, name := range header {
			if index < len(values) {
				row[name] = values[index]
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// requiredInt parses a mandatory integer while keeping field names in errors.
func requiredInt(row map[string]string, name string) (int, error) {
	value := strings.TrimSpace(row[name])
	if value == "" {
		return 0, fmt.Errorf("%s is required", name)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", name, value)
	}

	return parsed, nil
}

// optionalInt treats blank recovered cells as zero but still rejects malformed
// nonblank values, preserving the table's optional-number convention.
func optionalInt(row map[string]string, name string) (int, error) {
	value := strings.TrimSpace(row[name])
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", name, value)
	}

	return parsed, nil
}

// optionalBool accepts only the recovered table's numeric zero/one vocabulary;
// blanks inherit optionalInt's zero value.
func optionalBool(row map[string]string, name string) (bool, error) {
	value, err := optionalInt(row, name)
	if err != nil || (value != 0 && value != 1) {
		return false, fmt.Errorf("invalid %s %q", name, row[name])
	}

	return value == 1, nil
}

// compact trims and retains nonempty alternate keys in authored order.
func compact(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}

	return result
}
