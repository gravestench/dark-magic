// Package loot provides deterministic, renderer-independent treasure-class rolls.
package loot

import (
	"errors"
	"fmt"
)

const defaultMaxDepth = 64

// Entry names either an item code or another class and assigns its weight.
// For classes with negative Picks, Weight is instead the number of copies.
type Entry struct {
	Code   string `json:"code"`
	Weight int    `json:"weight"`
}

// Class is the subset of a Diablo II TreasureClass record needed to choose drops.
type Class struct {
	Name    string  `json:"name"`
	Picks   int     `json:"picks"`
	NoDrop  int     `json:"noDrop,omitempty"`
	Entries []Entry `json:"entries"`
}

// Catalog resolves treasure-class names. Codes absent from the catalog are items.
type Catalog map[string]Class

// Drop describes a terminal item and the treasure classes traversed to reach it.
type Drop struct {
	Code string   `json:"code"`
	Path []string `json:"path"`
}

// Roller owns deterministic random state. A Roller should not be shared between
// goroutines; construct one per gameplay event or serialized game simulation.
type Roller struct {
	catalog  Catalog
	rng      splitMix64
	maxDepth int
}

// New returns a roller whose results are reproducible for the same catalog and seed.
func New(catalog Catalog, seed uint64) *Roller {
	return &Roller{catalog: catalog, rng: splitMix64(seed), maxDepth: defaultMaxDepth}
}

// Roll expands a named treasure class into terminal item drops.
func (r *Roller) Roll(name string) ([]Drop, error) {
	if r == nil {
		return nil, errors.New("loot: nil roller")
	}
	if _, ok := r.catalog[name]; !ok {
		return nil, fmt.Errorf("loot: unknown treasure class %q", name)
	}

	return r.rollClass(name, nil, make(map[string]bool))
}

func (r *Roller) rollClass(name string, path []string, active map[string]bool) ([]Drop, error) {
	if len(path) >= r.maxDepth {
		return nil, fmt.Errorf("loot: maximum treasure-class depth %d exceeded at %q", r.maxDepth, name)
	}
	if active[name] {
		return nil, fmt.Errorf("loot: treasure-class cycle at %q along %v", name, append(path, name))
	}

	class := r.catalog[name]
	if err := validate(class); err != nil {
		return nil, fmt.Errorf("loot: class %q: %w", name, err)
	}

	active[name] = true
	defer delete(active, name)
	path = appendPath(path, name)

	entries := make([]Entry, 0, max(class.Picks, -class.Picks))
	if class.Picks < 0 {
		left := -class.Picks
		for _, entry := range class.Entries {
			for count := 0; count < entry.Weight && left > 0; count++ {
				entries = append(entries, entry)
				left--
			}
		}
	} else {
		total := class.NoDrop
		for _, entry := range class.Entries {
			total += entry.Weight
		}
		for pick := 0; pick < class.Picks; pick++ {
			roll := int(r.rng.next() % uint64(total))
			if roll < class.NoDrop {
				continue
			}
			roll -= class.NoDrop
			for _, entry := range class.Entries {
				if roll < entry.Weight {
					entries = append(entries, entry)
					break
				}
				roll -= entry.Weight
			}
		}
	}

	var drops []Drop
	for _, entry := range entries {
		if _, nested := r.catalog[entry.Code]; nested {
			resolved, err := r.rollClass(entry.Code, path, active)
			if err != nil {
				return nil, err
			}
			drops = append(drops, resolved...)
			continue
		}
		drops = append(drops, Drop{Code: entry.Code, Path: appendPath(path, "")[:len(path)]})
	}

	return drops, nil
}

func validate(class Class) error {
	if class.Picks == 0 {
		return errors.New("Picks must not be zero")
	}
	if class.NoDrop < 0 {
		return errors.New("NoDrop must not be negative")
	}
	total := class.NoDrop
	for index, entry := range class.Entries {
		if entry.Code == "" {
			return fmt.Errorf("entry %d has an empty code", index+1)
		}
		if entry.Weight < 0 {
			return fmt.Errorf("entry %q has a negative weight", entry.Code)
		}
		total += entry.Weight
	}
	if class.Picks > 0 && total == 0 {
		return errors.New("positive Picks requires at least one weighted outcome")
	}

	return nil
}

func appendPath(path []string, value string) []string {
	result := make([]string, len(path), len(path)+1)
	copy(result, path)
	if value != "" {
		result = append(result, value)
	}
	return result
}

// splitMix64 is small, stable across Go releases, and adequate for simulation.
type splitMix64 uint64

func (s *splitMix64) next() uint64 {
	*s += 0x9e3779b97f4a7c15
	z := uint64(*s)
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}
