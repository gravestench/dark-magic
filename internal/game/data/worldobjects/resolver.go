// Package worldobjects joins DS1 act-local identities to admitted game data.
package worldobjects

import (
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/recovered"
)

type staticDefinition struct {
	objectID    int
	description string
}

// Resolver is an immutable O(1) lookup built once per game-data generation.
type Resolver struct {
	static  map[string]staticDefinition
	dynamic map[string]string
}

func New(recoveredData recovered.Snapshot, gameData gamedata.Snapshot) *Resolver {
	resolver := &Resolver{
		static:  make(map[string]staticDefinition, len(recoveredData.MapObjects)),
		dynamic: make(map[string]string, len(gameData.MonsterPresets)),
	}
	for _, entry := range recoveredData.MapObjects {
		resolver.static[key(entry.Act, entry.ID)] = staticDefinition{objectID: entry.ObjectID, description: entry.Description}
	}
	actOffsets := make(map[int]int, 5)
	for _, entry := range gameData.MonsterPresets {
		index := actOffsets[entry.Act]
		resolver.dynamic[key(entry.Act, index)] = entry.Place
		actOffsets[entry.Act] = index + 1
	}
	return resolver
}

func (resolver *Resolver) ResolveStaticObject(act, id int) (int, string, bool) {
	if resolver == nil {
		return 0, "", false
	}
	entry, found := resolver.static[key(act, id)]
	return entry.objectID, entry.description, found
}

func (resolver *Resolver) ResolveDynamicObject(act, id int) (string, bool) {
	if resolver == nil {
		return "", false
	}
	entry, found := resolver.dynamic[key(act, id)]
	return entry, found
}

func key(act, id int) string { return fmt.Sprintf("%d:%d", act, id) }
