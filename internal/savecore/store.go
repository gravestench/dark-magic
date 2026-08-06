// Package savecore owns engine-side character metadata exposed to the Lua shell.
package savecore

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var characterClasses = map[string]string{
	"amazon": "Amazon", "sorceress": "Sorceress", "necromancer": "Necromancer",
	"paladin": "Paladin", "barbarian": "Barbarian", "assassin": "Assassin", "druid": "Druid",
}

type Character struct {
	ID         string
	Name       string
	Class      string
	Level      int
	Expansion  bool
	Hardcore   bool
	Appearance *Appearance
	Stats      *Stats
}

// Stats is an immutable character-sheet snapshot. Authoritative simulation and
// save importers replace it as values change; Lua receives copies for display.
type Stats struct {
	Experience, NextLevelExperience       int
	Strength, Dexterity, Vitality, Energy int
	Defense                               int
	Health, MaxHealth, Mana, MaxMana      int
	Stamina, MaxStamina                   int
	FireResistance, ColdResistance        int
	LightningResistance, PoisonResistance int
}

// Appearance is an immutable rendering snapshot decoded from a character save.
// Asset resolution belongs to the save importer; the store carries only the
// authoritative COF and DCC paths needed by presentation code.
type Appearance struct {
	COF        string
	Palette    string
	Direction  int
	Components map[string]string
}

func (s *Store) Create(character Character) error {
	character.ID = strings.TrimSpace(character.ID)
	character.Name = strings.TrimSpace(character.Name)
	character.Class = characterClasses[strings.ToLower(strings.TrimSpace(character.Class))]
	if character.ID == "" || character.Name == "" || character.Class == "" {
		return errors.New("savecore: character ID, name, and a supported class are required")
	}
	if err := validateCharacterName(character.Name); err != nil {
		return err
	}
	if character.Level < 1 {
		character.Level = 1
	}
	character = cloneCharacter(character)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.entries {
		if existing.ID == character.ID {
			return fmt.Errorf("savecore: character %q already exists", character.ID)
		}
		if strings.EqualFold(existing.Name, character.Name) {
			return fmt.Errorf("savecore: character name %q already exists", character.Name)
		}
	}
	s.entries = append(s.entries, character)
	return nil
}

// CreateNamed assigns the storage identity. Scripted presentation code chooses
// player-facing metadata but never owns save-file naming or collision policy.
func (s *Store) CreateNamed(name, class string) (Character, error) {
	return s.CreateNamedWithOptions(name, class, true, false)
}

// CreateNamedWithOptions creates the character metadata selected by the
// front-end while retaining save identity and validation inside the engine.
func (s *Store) CreateNamedWithOptions(name, class string, expansion, hardcore bool) (Character, error) {
	id := strings.ToLower(strings.TrimSpace(class)) + "-" + strings.ToLower(strings.TrimSpace(name))
	id = strings.Map(func(current rune) rune {
		if current >= 'a' && current <= 'z' || current >= '0' && current <= '9' {
			return current
		}
		if current == ' ' || current == '-' || current == '\'' {
			return '-'
		}
		return -1
	}, id)
	id = strings.Trim(id, "-")
	character := Character{ID: id, Name: name, Class: class, Level: 1, Expansion: expansion, Hardcore: hardcore}
	if err := s.Create(character); err != nil {
		return Character{}, err
	}
	return s.characterByID(id)
}

func (s *Store) characterByID(id string) (Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, character := range s.entries {
		if character.ID == id {
			return character, nil
		}
	}
	return Character{}, errors.New("savecore: created character is unavailable")
}

func validateCharacterName(name string) error {
	runes := []rune(name)
	if len(runes) < 2 || len(runes) > 15 {
		return errors.New("savecore: character name must contain 2 to 15 characters")
	}
	punctuation := false
	for index, current := range runes {
		if current >= 'A' && current <= 'Z' || current >= 'a' && current <= 'z' {
			punctuation = false
			continue
		}
		if (current != '-' && current != '\'') || index == 0 || index == len(runes)-1 || punctuation {
			return errors.New("savecore: character name may contain ASCII letters and single internal hyphens or apostrophes")
		}
		punctuation = true
	}
	return nil
}

type Store struct {
	mu       sync.RWMutex
	entries  []Character
	selected string
}

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
	return errors.New("savecore: unknown character")
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
	return errors.New("savecore: unknown character")
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
