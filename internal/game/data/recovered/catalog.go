// Package recovered parses declarative game relationships recovered by
// clean-room engine projects and carried in the first-party d2legacy mod.
package recovered

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	// QuestsPath is the provenance-preserving recovered quest hierarchy.
	QuestsPath = "data/recovered/riiablo/quests.txt"
	// SpeechPath joins logical speech sounds to localization keys.
	SpeechPath = "data/recovered/riiablo/speech.txt"
	// DS1TypesPath maps executable-era DS1 definitions to level types.
	DS1TypesPath = "data/recovered/riiablo/ds1types.txt"
	// ObjectsPath maps act-local DS1 object IDs to Objects.txt identities.
	ObjectsPath = "data/recovered/riiablo/obj.txt"
)

// QuestStage is one ordered localized stage plus recovered alternate keys.
type QuestStage struct {
	Index      int
	StringKey  string
	Alternates []string
}

// Quest preserves recovered hierarchy and localization identities without
// embedding executable behavior in the catalog.
type Quest struct {
	ID             int
	Name           string
	Act            int
	Order          int
	Visible        bool
	Icon           string
	PrerequisiteID *int
	TitleStringKey string
	Stages         []QuestStage
}

// Speech is the recovered logical-sound to localized-string join.
type Speech struct {
	Sound     string
	StringKey string
}

// DS1Type resolves one DS1 definition number within the recovered table.
type DS1Type struct {
	Name       string
	Definition int
	LevelType  int
}

// MapObject resolves an act-local static map object to Objects.txt.
type MapObject struct {
	Act         int
	ID          int
	Description string
	ObjectID    int
}

// Snapshot is an immutable generation with ordered rows and lookup indexes.
type Snapshot struct {
	Quests           []Quest
	QuestsByID       map[int]Quest
	Speech           []Speech
	SpeechByName     map[string]Speech
	DS1Types         []DS1Type
	DS1TypeByDef     map[int]DS1Type
	MapObjects       []MapObject
	MapObjectByActID map[string]MapObject
}

// ReferenceIssue is a non-fatal cross-catalog diagnostic. Different editions
// and language packs legitimately omit some recovered references.
type ReferenceIssue struct {
	Kind       string
	Identifier string
	Detail     string
}

// ValidateReferences checks joins from recovered rows into mounted Sounds.txt
// and localization tables. Missing references are diagnostics because game
// editions and language packs legitimately contain different subsets.
func ValidateReferences(snapshot Snapshot, soundNames map[string]struct{}, text func(string) (string, error)) []ReferenceIssue {
	var issues []ReferenceIssue
	for _, speech := range snapshot.Speech {
		if _, found := soundNames[strings.ToLower(speech.Sound)]; !found {
			issues = append(issues, ReferenceIssue{Kind: "sound", Identifier: speech.Sound, Detail: "not present in Sounds.txt"})
		}
		if text != nil {
			if _, err := text(speech.StringKey); err != nil {
				issues = append(issues, ReferenceIssue{Kind: "string", Identifier: speech.StringKey, Detail: err.Error()})
			}
		}
	}
	for _, quest := range snapshot.Quests {
		keys := []string{quest.TitleStringKey}
		for _, stage := range quest.Stages {
			keys = append(keys, stage.StringKey)
			keys = append(keys, stage.Alternates...)
		}
		for _, key := range keys {
			if key == "" || text == nil {
				continue
			}
			if _, err := text(key); err != nil {
				issues = append(issues, ReferenceIssue{Kind: "string", Identifier: key, Detail: err.Error()})
			}
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Kind == issues[j].Kind {
			return issues[i].Identifier < issues[j].Identifier
		}
		return issues[i].Kind < issues[j].Kind
	})
	return issues
}

// Catalog loads the provenance-backed recovered family once and returns deep
// copies. Unlike user-authored TSV data, these relationships ship with d2legacy.
type Catalog struct {
	source fs.FS
	once   sync.Once
	data   Snapshot
	err    error
}

// New creates a lazy recovered-data catalog over layered content.
func New(source fs.FS) *Catalog { return &Catalog{source: source} }

// Snapshot loads and validates the recovered family once, then returns a copy
// so runtime adapters cannot mutate shared provenance data.
func (catalog *Catalog) Snapshot() (Snapshot, error) {
	if catalog == nil || catalog.source == nil {
		return Snapshot{}, fmt.Errorf("recovered data: no content source")
	}
	catalog.once.Do(func() { catalog.data, catalog.err = load(catalog.source) })
	if catalog.err != nil {
		return Snapshot{}, catalog.err
	}
	return clone(catalog.data), nil
}

func load(source fs.FS) (Snapshot, error) {
	questsFile, err := source.Open(QuestsPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovered data: open quests: %w", err)
	}
	quests, err := ParseQuests(questsFile)
	closeErr := questsFile.Close()
	if err != nil {
		return Snapshot{}, err
	}
	if closeErr != nil {
		return Snapshot{}, fmt.Errorf("recovered data: close quests: %w", closeErr)
	}
	speechFile, err := source.Open(SpeechPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovered data: open speech: %w", err)
	}
	speech, err := ParseSpeech(speechFile)
	closeErr = speechFile.Close()
	if err != nil {
		return Snapshot{}, err
	}
	if closeErr != nil {
		return Snapshot{}, fmt.Errorf("recovered data: close speech: %w", closeErr)
	}
	ds1TypesFile, err := source.Open(DS1TypesPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovered data: open DS1 types: %w", err)
	}
	ds1Types, err := ParseDS1Types(ds1TypesFile)
	closeErr = ds1TypesFile.Close()
	if err != nil {
		return Snapshot{}, err
	}
	if closeErr != nil {
		return Snapshot{}, fmt.Errorf("recovered data: close DS1 types: %w", closeErr)
	}
	objectsFile, err := source.Open(ObjectsPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("recovered data: open map objects: %w", err)
	}
	objects, err := ParseMapObjects(objectsFile)
	closeErr = objectsFile.Close()
	if err != nil {
		return Snapshot{}, err
	}
	if closeErr != nil {
		return Snapshot{}, fmt.Errorf("recovered data: close map objects: %w", closeErr)
	}
	result := Snapshot{
		Quests: quests, QuestsByID: make(map[int]Quest, len(quests)),
		Speech: speech, SpeechByName: make(map[string]Speech, len(speech)),
		DS1Types: ds1Types, DS1TypeByDef: make(map[int]DS1Type, len(ds1Types)),
		MapObjects: objects, MapObjectByActID: make(map[string]MapObject, len(objects)),
	}
	for _, quest := range quests {
		result.QuestsByID[quest.ID] = quest
	}
	for _, entry := range speech {
		result.SpeechByName[strings.ToLower(entry.Sound)] = entry
	}
	for _, entry := range ds1Types {
		result.DS1TypeByDef[entry.Definition] = entry
	}
	for _, entry := range objects {
		result.MapObjectByActID[mapObjectKey(entry.Act, entry.ID)] = entry
	}
	return result, nil
}

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
			return nil, fmt.Errorf("recovered data: DS1 types line %d: duplicate definition %d", index+2, definition)
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
			return nil, fmt.Errorf("recovered data: map objects line %d: invalid Act, Id, or ObjectId", index+2)
		}
		key := mapObjectKey(act, id)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("recovered data: map objects line %d: duplicate act/id %s", index+2, key)
		}
		seen[key] = struct{}{}
		result = append(result, MapObject{Act: act, ID: id, Description: strings.TrimSpace(row["Description"]), ObjectID: objectID})
	}
	return result, nil
}

func mapObjectKey(act, id int) string { return fmt.Sprintf("%d:%d", act, id) }

func ParseQuests(input io.Reader) ([]Quest, error) {
	rows, err := readTSV(input)
	if err != nil {
		return nil, fmt.Errorf("recovered data: parse quests: %w", err)
	}
	result := make([]Quest, 0, len(rows))
	seen := make(map[int]struct{}, len(rows))
	for index, row := range rows {
		line := index + 2
		id, err := requiredInt(row, "id")
		if err != nil {
			return nil, fmt.Errorf("recovered data: quests line %d: %w", line, err)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("recovered data: quests line %d: duplicate id %d", line, id)
		}
		seen[id] = struct{}{}
		act, err := requiredInt(row, "act")
		if err != nil || act < 0 || act > 4 {
			return nil, fmt.Errorf("recovered data: quests line %d: invalid act %q", line, row["act"])
		}
		quest := Quest{ID: id, Name: strings.TrimSpace(row["name"]), Act: act, Icon: strings.TrimSpace(row["icon"]), TitleStringKey: strings.TrimSpace(row["qstr"])}
		if quest.Name == "" || quest.TitleStringKey == "" {
			return nil, fmt.Errorf("recovered data: quests line %d: name and qstr are required", line)
		}
		quest.Order, err = optionalInt(row, "order")
		if err != nil {
			return nil, fmt.Errorf("recovered data: quests line %d: %w", line, err)
		}
		quest.Visible, err = optionalBool(row, "visible")
		if err != nil {
			return nil, fmt.Errorf("recovered data: quests line %d: %w", line, err)
		}
		if value := strings.TrimSpace(row["questdone"]); value != "" {
			parsed, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return nil, fmt.Errorf("recovered data: quests line %d: invalid questdone %q", line, value)
			}
			quest.PrerequisiteID = &parsed
		}
		for stage := 1; stage <= 6; stage++ {
			primary := strings.TrimSpace(row[fmt.Sprintf("qsts%d", stage)])
			alternates := compact(row[fmt.Sprintf("qsts%da", stage)], row[fmt.Sprintf("qsts%db", stage)])
			if primary != "" || len(alternates) > 0 {
				quest.Stages = append(quest.Stages, QuestStage{Index: stage, StringKey: primary, Alternates: alternates})
			}
		}
		result = append(result, quest)
	}
	for _, quest := range result {
		if quest.PrerequisiteID != nil {
			if _, exists := seen[*quest.PrerequisiteID]; !exists {
				return nil, fmt.Errorf("recovered data: quest %d references unknown prerequisite %d", quest.ID, *quest.PrerequisiteID)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func ParseSpeech(input io.Reader) ([]Speech, error) {
	rows, err := readTSV(input)
	if err != nil {
		return nil, fmt.Errorf("recovered data: parse speech: %w", err)
	}
	result := make([]Speech, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for index, row := range rows {
		entry := Speech{Sound: strings.TrimSpace(row["sound"]), StringKey: strings.TrimSpace(row["soundstr"])}
		if entry.Sound == "" || entry.StringKey == "" {
			return nil, fmt.Errorf("recovered data: speech line %d: sound and soundstr are required", index+2)
		}
		key := strings.ToLower(entry.Sound)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("recovered data: speech line %d: duplicate sound %q", index+2, entry.Sound)
		}
		seen[key] = struct{}{}
		result = append(result, entry)
	}
	return result, nil
}

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

func optionalBool(row map[string]string, name string) (bool, error) {
	value, err := optionalInt(row, name)
	if err != nil || (value != 0 && value != 1) {
		return false, fmt.Errorf("invalid %s %q", name, row[name])
	}
	return value == 1, nil
}

func compact(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func clone(source Snapshot) Snapshot {
	result := Snapshot{
		Quests: make([]Quest, len(source.Quests)), QuestsByID: make(map[int]Quest, len(source.QuestsByID)),
		Speech: append([]Speech(nil), source.Speech...), SpeechByName: make(map[string]Speech, len(source.SpeechByName)),
		DS1Types: append([]DS1Type(nil), source.DS1Types...), DS1TypeByDef: make(map[int]DS1Type, len(source.DS1TypeByDef)),
		MapObjects: append([]MapObject(nil), source.MapObjects...), MapObjectByActID: make(map[string]MapObject, len(source.MapObjectByActID)),
	}
	for index, quest := range source.Quests {
		quest.Stages = append([]QuestStage(nil), quest.Stages...)
		for stage := range quest.Stages {
			quest.Stages[stage].Alternates = append([]string(nil), quest.Stages[stage].Alternates...)
		}
		result.Quests[index] = quest
		result.QuestsByID[quest.ID] = quest
	}
	for key, value := range source.SpeechByName {
		result.SpeechByName[key] = value
	}
	for key, value := range source.DS1TypeByDef {
		result.DS1TypeByDef[key] = value
	}
	for key, value := range source.MapObjectByActID {
		result.MapObjectByActID[key] = value
	}
	return result
}
