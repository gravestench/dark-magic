package loot

import "fmt"

// EventKind separates RNG domains that may otherwise share entity identities.
type EventKind string

const (
	// EventMonster is a monster-owned drop opportunity.
	EventMonster EventKind = "monster"
	// EventChest is an interactable-container drop opportunity.
	EventChest EventKind = "chest"
)

// Event identifies one deterministic gameplay drop opportunity.
type Event struct {
	Kind     EventKind `json:"kind"`
	EntityID uint64    `json:"entityId"`
	Sequence uint64    `json:"sequence"`
}

// EventSeed derives independent, replayable loot streams from world identity
// and gameplay event identity.
func EventSeed(worldSeed uint64, event Event) (uint64, error) {
	if event.Kind != EventMonster && event.Kind != EventChest {
		return 0, fmt.Errorf("loot: unsupported event kind %q", event.Kind)
	}
	if event.EntityID == 0 {
		return 0, fmt.Errorf("loot: event entity ID is required")
	}
	hash := uint64(1469598103934665603)
	for _, value := range []byte(event.Kind) {
		hash ^= uint64(value)
		hash *= 1099511628211
	}
	for _, value := range []uint64{worldSeed, event.EntityID, event.Sequence} {
		hash ^= value + 0x9e3779b97f4a7c15 + (hash << 6) + (hash >> 2)
	}
	rng := splitMix64(hash)
	return rng.next(), nil
}

// RollEvent derives a purpose-specific seed, then runs ordinary treasure-class
// selection. Presentation timing and global RNG order cannot affect the result.
func RollEvent(catalog Catalog, class string, worldSeed uint64, event Event) ([]Drop, error) {
	seed, err := EventSeed(worldSeed, event)
	if err != nil {
		return nil, err
	}
	return New(catalog, seed).Roll(class)
}
