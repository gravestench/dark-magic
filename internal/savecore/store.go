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
	ID    string
	Name  string
	Class string
	Level int
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
	character := Character{ID: id, Name: name, Class: class, Level: 1}
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
	copyEntries := append([]Character(nil), entries...)
	return &Store{entries: copyEntries}
}

func (s *Store) Characters() []Character {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Character(nil), s.entries...)
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

func (s *Store) Selected() (Character, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, character := range s.entries {
		if character.ID == s.selected {
			return character, true
		}
	}
	return Character{}, false
}
