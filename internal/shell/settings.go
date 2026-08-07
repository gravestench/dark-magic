package shell

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const DefaultFontSize = 18.0

// SettingsValues contains the user-editable presentation settings shared by
// shell adapters. Values change immediately; Save explicitly persists them.
type SettingsValues struct {
	FontSize float64 `json:"font_size"`
}

// Settings owns validated runtime values, defaults, and optional persistence.
type Settings struct {
	mu       sync.RWMutex
	path     string
	values   SettingsValues
	defaults SettingsValues
	dirty    bool
}

// NewTransientSettings returns defaults without a persistence destination.
// It is useful for adapters embedded without an application configuration.
func NewTransientSettings() *Settings {
	defaults := SettingsValues{FontSize: DefaultFontSize}
	return &Settings{values: defaults, defaults: defaults}
}

// NewSettings loads a settings file. An empty path selects the platform's user
// configuration directory. A missing file is equivalent to defaults.
func NewSettings(path string) (*Settings, error) {
	if path == "" {
		directory, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("shell settings: user config directory: %w", err)
		}
		path = filepath.Join(directory, "dark-magic", "shell.json")
	}
	defaults := SettingsValues{FontSize: DefaultFontSize}
	settings := &Settings{path: path, values: defaults, defaults: defaults}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return nil, fmt.Errorf("shell settings: read %q: %w", path, err)
	}
	var values SettingsValues
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("shell settings: decode %q: %w", path, err)
	}
	if err := validateSettings(values); err != nil {
		return nil, fmt.Errorf("shell settings: %q: %w", path, err)
	}
	settings.values = values
	return settings, nil
}

func validateSettings(values SettingsValues) error {
	if values.FontSize < 8 || values.FontSize > 48 {
		return fmt.Errorf("font_size must be between 8 and 48 (got %g)", values.FontSize)
	}
	return nil
}

func (s *Settings) Values() SettingsValues {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values
}

func (s *Settings) Defaults() SettingsValues { return s.defaults }
func (s *Settings) Path() string             { return s.path }

func (s *Settings) Dirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

func (s *Settings) SetFontSize(value float64) error {
	values := s.Values()
	values.FontSize = value
	if err := validateSettings(values); err != nil {
		return err
	}
	s.mu.Lock()
	s.values = values
	s.dirty = true
	s.mu.Unlock()
	return nil
}

func (s *Settings) Reset() {
	s.mu.Lock()
	s.values = s.defaults
	s.dirty = true
	s.mu.Unlock()
}

func (s *Settings) Save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.values, "", "  ")
	path := s.path
	s.mu.RUnlock()
	if path == "" {
		return errors.New("shell settings: no persistence path configured")
	}
	if err != nil {
		return fmt.Errorf("shell settings: encode: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("shell settings: create directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".shell-*.json")
	if err != nil {
		return fmt.Errorf("shell settings: create temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = temporary.Write(append(data, '\n')); err == nil {
		err = temporary.Close()
	} else {
		_ = temporary.Close()
	}
	if err != nil {
		return fmt.Errorf("shell settings: write: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("shell settings: replace %q: %w", path, err)
	}
	s.mu.Lock()
	s.dirty = false
	s.mu.Unlock()
	return nil
}
