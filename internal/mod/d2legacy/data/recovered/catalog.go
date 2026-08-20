// Package recovered parses declarative game relationships recovered by
// clean-room engine projects and carried in the first-party d2legacy mod.
package recovered

import (
	"fmt"
	"io"
	"io/fs"
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

// Catalog loads the provenance-backed recovered family once and returns deep
// copies. Unlike user-authored TSV data, these relationships ship with d2legacy.
type Catalog struct {
	source fs.FS
	mu     sync.Mutex
	once   sync.Once
	data   Snapshot
	err    error
}

// New creates a lazy recovered-data catalog over layered content. No files are
// opened until Snapshot, so callers can finish mounting package layers first.
func New(source fs.FS) *Catalog {
	return &Catalog{source: source}
}

// Snapshot loads and validates the recovered family once, then returns a copy
// so runtime adapters cannot mutate shared provenance data.
func (catalog *Catalog) Snapshot() (Snapshot, error) {
	if catalog == nil || catalog.source == nil {
		return Snapshot{}, fmt.Errorf("recovered data: no content source")
	}

	catalog.mu.Lock()
	defer catalog.mu.Unlock()
	// Hold the mutex across once.Do and cloning so Invalidate cannot replace a
	// generation while another caller is copying it.
	catalog.once.Do(func() {
		catalog.data, catalog.err = load(catalog.source)
	})

	if catalog.err != nil {
		return Snapshot{}, catalog.err
	}

	return clone(catalog.data), nil
}

// Invalidate discards the immutable generation after package-layer changes.
func (catalog *Catalog) Invalidate() {
	if catalog == nil {
		return
	}

	catalog.mu.Lock()
	catalog.once = sync.Once{}
	catalog.data = Snapshot{}
	catalog.err = nil
	catalog.mu.Unlock()
}

// load opens every provenance table in contract order and builds a generation.
// Each file is parsed and closed before the next opens, preserving error order.
func load(source fs.FS) (Snapshot, error) {
	quests, err := loadTable(source, QuestsPath, "quests", ParseQuests)
	if err != nil {
		return Snapshot{}, err
	}

	speech, err := loadTable(source, SpeechPath, "speech", ParseSpeech)
	if err != nil {
		return Snapshot{}, err
	}

	ds1Types, err := loadTable(source, DS1TypesPath, "DS1 types", ParseDS1Types)
	if err != nil {
		return Snapshot{}, err
	}

	objects, err := loadTable(source, ObjectsPath, "map objects", ParseMapObjects)
	if err != nil {
		return Snapshot{}, err
	}

	result := Snapshot{
		Quests:           quests,
		QuestsByID:       make(map[int]Quest, len(quests)),
		Speech:           speech,
		SpeechByName:     make(map[string]Speech, len(speech)),
		DS1Types:         ds1Types,
		DS1TypeByDef:     make(map[int]DS1Type, len(ds1Types)),
		MapObjects:       objects,
		MapObjectByActID: make(map[string]MapObject, len(objects)),
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

// loadTable owns one file from open through close. Parse errors retain priority
// over close errors, matching the catalog's established error behavior.
func loadTable[T any](
	source fs.FS,
	path, label string,
	parse func(io.Reader) ([]T, error),
) ([]T, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recovered data: open %s: %w", label, err)
	}

	values, parseErr := parse(file)
	closeErr := file.Close()

	if parseErr != nil {
		return nil, parseErr
	}

	if closeErr != nil {
		return nil, fmt.Errorf("recovered data: close %s: %w", label, closeErr)
	}

	return values, nil
}
