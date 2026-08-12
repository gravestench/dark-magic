// Package worldobjects joins Diablo DS1 act-local identities to recovered and
// decoded d2legacy records.
package worldobjects

import (
	"fmt"
	"strconv"

	"github.com/gravestench/dark-magic/internal/game/data/recovered"
)

type Records interface {
	Load(string) ([]map[string]string, error)
}

type staticDefinition struct {
	objectID    int
	description string
}

// Resolver is an immutable O(1) lookup built once per game-data generation.
type Resolver struct {
	static  map[string]staticDefinition
	dynamic map[string]string
}

// New freezes both data generations into act-local lookup indexes. Dynamic
// monster preset IDs are positional, so their authored per-act order is kept.
func New(recoveredData recovered.Snapshot, records Records) (*Resolver, error) {
	presets, err := records.Load("data/global/excel/monpreset.txt")
	if err != nil {
		return nil, fmt.Errorf("d2legacy world objects: load MonPreset.txt: %w", err)
	}
	resolver := &Resolver{
		static:  make(map[string]staticDefinition, len(recoveredData.MapObjects)),
		dynamic: make(map[string]string, len(presets)),
	}
	for _, entry := range recoveredData.MapObjects {
		resolver.static[key(entry.Act, entry.ID)] = staticDefinition{objectID: entry.ObjectID, description: entry.Description}
	}
	actOffsets := make(map[int]int, 5)
	for row, entry := range presets {
		act, err := strconv.Atoi(entry["Act"])
		if err != nil {
			return nil, fmt.Errorf("d2legacy world objects: MonPreset.txt row %d has invalid Act %q", row+2, entry["Act"])
		}
		index := actOffsets[act]
		resolver.dynamic[key(act, index)] = entry["Place"]
		actOffsets[act] = index + 1
	}
	return resolver, nil
}

// ResolveStaticObject maps a DS1 static ID to Objects.txt and its recovered
// human-readable description.
func (resolver *Resolver) ResolveStaticObject(act, id int) (int, string, bool) {
	if resolver == nil {
		return 0, "", false
	}
	entry, found := resolver.static[key(act, id)]
	return entry.objectID, entry.description, found
}

// ResolveDynamicObject maps a DS1 dynamic ID to the act-local MonPreset place
// token. It returns identity only; spawning remains authoritative game logic.
func (resolver *Resolver) ResolveDynamicObject(act, id int) (string, bool) {
	if resolver == nil {
		return "", false
	}
	entry, found := resolver.dynamic[key(act, id)]
	return entry, found
}

func key(act, id int) string { return fmt.Sprintf("%d:%d", act, id) }
