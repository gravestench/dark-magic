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

// NewZone defensively copies, canonicalizes, and validates a mutable recipe before admitting it as authoritative.
// Canonicalization before validation makes equivalent input orderings produce the same immutable representation.
func NewZone(definition Definition) (*Zone, error) {
	definition = cloneDefinition(definition)
	canonicalize(&definition)

	if err := validateDefinition(definition); err != nil {
		return nil, err
	}

	return &Zone{definition: definition}, nil
}

// Request returns the value-only replay identity that produced the zone.
func (zone *Zone) Request() Request { return zone.definition.Request }

// Kind returns the opaque mod policy identifier without interpreting it in engine code.
func (zone *Zone) Kind() Kind { return zone.definition.Kind }

// Bounds returns the value-only authoritative world extent.
func (zone *Zone) Bounds() Bounds { return zone.definition.Bounds }

// Stamps returns a deep copy because each stamp owns a mutable TilePaths slice.
func (zone *Zone) Stamps() []Stamp { return cloneDefinition(zone.definition).Stamps }

// Rooms returns a copy so callers cannot mutate the admitted topology through shared slice storage.
func (zone *Zone) Rooms() []Room { return append([]Room(nil), zone.definition.Rooms...) }

// Links returns a copy that preserves the canonical undirected edge order.
func (zone *Zone) Links() []Link { return append([]Link(nil), zone.definition.Links...) }

// Warps returns a copy so transport and presentation adapters cannot alter authoritative transitions.
func (zone *Zone) Warps() []Warp { return append([]Warp(nil), zone.definition.Warps...) }

// Spawns returns a copy so consumers may annotate their own view without changing the recipe.
func (zone *Zone) Spawns() []Spawn { return append([]Spawn(nil), zone.definition.Spawns...) }

// Paths returns a copy of the semantic route cells admitted during generation.
func (zone *Zone) Paths() []PathTile { return append([]PathTile(nil), zone.definition.Paths...) }

// Structures returns a copy of the authoritative structural and passability decisions.
func (zone *Zone) Structures() []StructureTile {
	return append([]StructureTile(nil), zone.definition.Structures...)
}

// Trace returns a copy so diagnostics may filter messages without modifying replay evidence.
func (zone *Zone) Trace() []string { return append([]string(nil), zone.definition.Trace...) }

// MarshalJSON is the canonical, versioned representation used by replays,
// diagnostics, and server/client verification.
func (zone *Zone) MarshalJSON() ([]byte, error) { return json.Marshal(zone.definition) }

// Checksum hashes the canonical JSON representation used for server/client and replay agreement.
// Any admitted recipe difference therefore changes the identity while equivalent input ordering does not.
func (zone *Zone) Checksum() (string, error) {
	encoded, err := zone.MarshalJSON()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(encoded)

	return hex.EncodeToString(sum[:]), nil
}

// cloneDefinition severs every slice alias, including the nested tile-path slices owned by stamps.
// This copy boundary is what makes Zone immutable despite accepting and returning ordinary Go slices.
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

// canonicalize establishes stable ordering for every collection that contributes to JSON and checksums.
// It runs only on a private copy, so sorting never mutates the generator's input.
func canonicalize(definition *Definition) {
	canonicalizeStamps(definition.Stamps)
	canonicalizeLinks(definition.Links)

	sort.Slice(definition.Rooms, func(i, j int) bool {
		return definition.Rooms[i].ID < definition.Rooms[j].ID
	})
	sort.Slice(definition.Warps, func(i, j int) bool {
		return definition.Warps[i].ID < definition.Warps[j].ID
	})
	sort.Slice(definition.Spawns, func(i, j int) bool {
		return definition.Spawns[i].ID < definition.Spawns[j].ID
	})
	canonicalizePositions(definition.Paths, func(tile PathTile) (int, int) {
		return tile.X, tile.Y
	})
	canonicalizePositions(definition.Structures, func(tile StructureTile) (int, int) {
		return tile.X, tile.Y
	})
}

// canonicalizeStamps sorts nested asset identities before ordering stamps by their stable numeric identity.
func canonicalizeStamps(stamps []Stamp) {
	for index := range stamps {
		sort.Strings(stamps[index].TilePaths)
	}

	sort.Slice(stamps, func(i, j int) bool {
		return stamps[i].ID < stamps[j].ID
	})
}

// canonicalizeLinks normalizes each undirected edge before sorting the complete topology.
func canonicalizeLinks(links []Link) {
	for index := range links {
		if links[index].From > links[index].To {
			links[index].From, links[index].To = links[index].To, links[index].From
		}
	}

	sort.Slice(links, func(i, j int) bool {
		if links[i].From == links[j].From {
			return links[i].To < links[j].To
		}

		return links[i].From < links[j].From
	})
}

// canonicalizePositions orders coordinate-bearing values by row and then column for deterministic traversal.
func canonicalizePositions[T any](values []T, position func(T) (int, int)) {
	sort.Slice(values, func(i, j int) bool {
		xi, yi := position(values[i])

		xj, yj := position(values[j])
		if yi == yj {
			return xi < xj
		}

		return yi < yj
	})
}
