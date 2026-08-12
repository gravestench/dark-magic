package worldgen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// Zone is immutable after construction. Accessors return copies so a consumer
// cannot accidentally change authoritative generation results.
type Zone struct{ definition Definition }

// Definition returns a defensive copy for generic adapters which serialize or
// transport an already-admitted recipe.
func (zone *Zone) Definition() Definition { return cloneDefinition(zone.definition) }

func NewZone(definition Definition) (*Zone, error) {
	definition = cloneDefinition(definition)
	canonicalize(&definition)
	if err := validateDefinition(definition); err != nil {
		return nil, err
	}
	return &Zone{definition: definition}, nil
}

func (zone *Zone) Request() Request  { return zone.definition.Request }
func (zone *Zone) Kind() Kind        { return zone.definition.Kind }
func (zone *Zone) Bounds() Bounds    { return zone.definition.Bounds }
func (zone *Zone) Stamps() []Stamp   { return cloneDefinition(zone.definition).Stamps }
func (zone *Zone) Rooms() []Room     { return append([]Room(nil), zone.definition.Rooms...) }
func (zone *Zone) Links() []Link     { return append([]Link(nil), zone.definition.Links...) }
func (zone *Zone) Warps() []Warp     { return append([]Warp(nil), zone.definition.Warps...) }
func (zone *Zone) Spawns() []Spawn   { return append([]Spawn(nil), zone.definition.Spawns...) }
func (zone *Zone) Paths() []PathTile { return append([]PathTile(nil), zone.definition.Paths...) }
func (zone *Zone) Structures() []StructureTile {
	return append([]StructureTile(nil), zone.definition.Structures...)
}
func (zone *Zone) Trace() []string { return append([]string(nil), zone.definition.Trace...) }

// MarshalJSON is the canonical, versioned representation used by replays,
// diagnostics, and server/client verification.
func (zone *Zone) MarshalJSON() ([]byte, error) { return json.Marshal(zone.definition) }

func (zone *Zone) Checksum() (string, error) {
	encoded, err := zone.MarshalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func cloneDefinition(source Definition) Definition {
	result := source
	result.Stamps = append([]Stamp(nil), source.Stamps...)
	for index := range result.Stamps {
		result.Stamps[index].TilePaths = append([]string(nil), source.Stamps[index].TilePaths...)
	}
	result.Rooms = append([]Room(nil), source.Rooms...)
	result.Links = append([]Link(nil), source.Links...)
	result.Warps = append([]Warp(nil), source.Warps...)
	result.Spawns = append([]Spawn(nil), source.Spawns...)
	result.Paths = append([]PathTile(nil), source.Paths...)
	result.Structures = append([]StructureTile(nil), source.Structures...)
	result.Trace = append([]string(nil), source.Trace...)
	return result
}

func canonicalize(definition *Definition) {
	for index := range definition.Stamps {
		sort.Strings(definition.Stamps[index].TilePaths)
	}
	sort.Slice(definition.Stamps, func(i, j int) bool { return definition.Stamps[i].ID < definition.Stamps[j].ID })
	sort.Slice(definition.Rooms, func(i, j int) bool { return definition.Rooms[i].ID < definition.Rooms[j].ID })
	for index := range definition.Links {
		if definition.Links[index].From > definition.Links[index].To {
			definition.Links[index].From, definition.Links[index].To = definition.Links[index].To, definition.Links[index].From
		}
	}
	sort.Slice(definition.Links, func(i, j int) bool {
		if definition.Links[i].From == definition.Links[j].From {
			return definition.Links[i].To < definition.Links[j].To
		}
		return definition.Links[i].From < definition.Links[j].From
	})
	sort.Slice(definition.Warps, func(i, j int) bool { return definition.Warps[i].ID < definition.Warps[j].ID })
	sort.Slice(definition.Spawns, func(i, j int) bool { return definition.Spawns[i].ID < definition.Spawns[j].ID })
	sort.Slice(definition.Paths, func(i, j int) bool {
		if definition.Paths[i].Y == definition.Paths[j].Y {
			return definition.Paths[i].X < definition.Paths[j].X
		}
		return definition.Paths[i].Y < definition.Paths[j].Y
	})
	sort.Slice(definition.Structures, func(i, j int) bool {
		if definition.Structures[i].Y == definition.Structures[j].Y {
			return definition.Structures[i].X < definition.Structures[j].X
		}
		return definition.Structures[i].Y < definition.Structures[j].Y
	})
}
