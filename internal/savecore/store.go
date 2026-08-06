// Package savecore owns engine-side character metadata exposed to the Lua shell.
package savecore

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Character struct {
	ID    string
	Name  string
	Class string
	Level int
}

func (s *Store) Create(character Character) error {
	character.ID = strings.TrimSpace(character.ID)
	character.Name = strings.TrimSpace(character.Name)
	character.Class = strings.TrimSpace(character.Class)
	if character.ID == "" || character.Name == "" || character.Class == "" {
		return errors.New("savecore: character ID, name, and class are required")
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
	}
	s.entries = append(s.entries, character)
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
