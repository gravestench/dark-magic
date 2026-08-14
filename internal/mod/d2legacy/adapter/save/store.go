package save

import (
	"errors"
	"fmt"
	"sync"
)

// Character is d2legacy's player-profile character-selection record. It may be
// used for single-player, listen-server, or self-hosted dedicated-server play,
// but is never realm authority. Realm persistence wraps equivalent game data
// in account ownership, revision, compatibility, and lease checks; copying this
// value never grants authority to update an account-owned realm character.
type Character struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Class      string      `json:"class"`
	Level      int         `json:"level"`
	Expansion  bool        `json:"expansion"`
	Hardcore   bool        `json:"hardcore"`
	Appearance *Appearance `json:"appearance,omitempty"`
	Stats      *Stats      `json:"stats,omitempty"`
}

// Stats is an immutable character-sheet snapshot. Authoritative simulation and
// save importers replace it as values change; Lua receives copies for display.
type Stats struct {
	Experience          int `json:"experience"`
	NextLevelExperience int `json:"next_level_experience"`
	Strength            int `json:"strength"`
	Dexterity           int `json:"dexterity"`
	Vitality            int `json:"vitality"`
	Energy              int `json:"energy"`
	Defense             int `json:"defense"`
	Health              int `json:"health"`
	MaxHealth           int `json:"max_health"`
	Mana                int `json:"mana"`
	MaxMana             int `json:"max_mana"`
	Stamina             int `json:"stamina"`
	MaxStamina          int `json:"max_stamina"`
	FireResistance      int `json:"fire_resistance"`
	ColdResistance      int `json:"cold_resistance"`
	LightningResistance int `json:"lightning_resistance"`
	PoisonResistance    int `json:"poison_resistance"`
}

// Appearance is an immutable rendering snapshot decoded from a character save.
// Asset resolution belongs to the save importer; the store carries only the
// authoritative COF and DCC paths needed by presentation code.
type Appearance struct {
	COF        string            `json:"cof"`
	Palette    string            `json:"palette"`
	Direction  int               `json:"direction"`
	Components map[string]string `json:"components,omitempty"`
}

func (s *Store) Create(character Character) error {
	// This store is deliberately policy-free. The owning mod validates and
	// normalizes its record before crossing this persistence boundary.
	if character.ID == "" {
		return errors.New("persistence: record ID is required")
	}
	character = cloneCharacter(character)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.entries {
		if existing.ID == character.ID {
			return fmt.Errorf("persistence: character %q already exists", character.ID)
		}
	}
	s.entries = append(s.entries, character)
	return nil
}

// CreateSelected atomically adds a character and makes it the active profile
// choice. Frontend creation uses this operation so an older selection cannot
// survive between persistence and launching a game session.
func (s *Store) CreateSelected(character Character) error {
	if character.ID == "" {
		return errors.New("persistence: record ID is required")
	}
	character = cloneCharacter(character)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.entries {
		if existing.ID == character.ID {
			return fmt.Errorf("persistence: character %q already exists", character.ID)
		}
	}
	s.entries = append(s.entries, character)
	s.selected = character.ID
	return nil
}

// Store owns the player-controlled selectable character roster and returns
// defensive copies. It intentionally does not implement the account-owned
// realm character repository contract.
// Selection is application state; authoritative in-session player state is
// materialized by the game session rather than mutated through this store.
type Store struct {
	mu       sync.RWMutex
	entries  []Character
	selected string
}

// New creates a roster from the supplied fixtures or imported characters.
func New(entries ...Character) *Store {
	copyEntries := make([]Character, len(entries))
	for index, entry := range entries {
		copyEntries[index] = cloneCharacter(entry)
	}
	return &Store{entries: copyEntries}
}

func (s *Store) Characters() []Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Character, len(s.entries))
	for index, entry := range s.entries {
		result[index] = cloneCharacter(entry)
	}
	return result
}

func (s *Store) Select(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, character := range s.entries {
		if character.ID == id {
			s.selected = id
			return nil
		}
	}
	return errors.New("persistence: unknown character")
}

// Delete removes one character identity and clears the active selection when
// it refers to that character. Save-file persistence remains a Store concern;
// presentation code only requests deletion by opaque ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, character := range s.entries {
		if character.ID != id {
			continue
		}
		copy(s.entries[index:], s.entries[index+1:])
		s.entries = s.entries[:len(s.entries)-1]
		if s.selected == id {
			s.selected = ""
		}
		return nil
	}
	return errors.New("persistence: unknown character")
}

func (s *Store) Selected() (Character, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, character := range s.entries {
		if character.ID == s.selected {
			return cloneCharacter(character), true
		}
	}
	return Character{}, false
}

// UpdateSelected replaces the active player-owned character without allowing
// a network projection to change roster identity or selection. The caller owns
// the policy that produced the updated durable fields.
func (s *Store) UpdateSelected(character Character) error {
	if character.ID == "" {
		return errors.New("persistence: record ID is required")
	}
	character = cloneCharacter(character)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.selected == "" || character.ID != s.selected {
		return errors.New("persistence: updated character is not selected")
	}
	for index := range s.entries {
		if s.entries[index].ID == s.selected {
			s.entries[index] = character
			return nil
		}
	}
	return errors.New("persistence: selected character is absent")
}

func cloneCharacter(character Character) Character {
	if character.Stats != nil {
		stats := *character.Stats
		character.Stats = &stats
	}
	if character.Appearance == nil {
		return character
	}
	appearance := *character.Appearance
	appearance.Components = make(map[string]string, len(character.Appearance.Components))
	for component, path := range character.Appearance.Components {
		appearance.Components[component] = path
	}
	character.Appearance = &appearance
	return character
}
