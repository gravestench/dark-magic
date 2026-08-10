// Package preferences owns validated, persistent client preferences.
package preferences

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Values struct {
	SoundVolume float64 `json:"sound_volume"`
	MusicVolume float64 `json:"music_volume"`
}

type Settings struct {
	mu     sync.RWMutex
	path   string
	values Values
	dirty  bool
}

func Defaults() Values { return Values{SoundVolume: .5, MusicVolume: .5} }

func NewTransient() *Settings { return &Settings{values: Defaults()} }

func New(path string) (*Settings, error) {
	if path == "" {
		directory, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("preferences: user config directory: %w", err)
		}
		path = filepath.Join(directory, "dark-magic", "preferences.json")
	}
	settings := &Settings{path: path, values: Defaults()}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return nil, fmt.Errorf("preferences: read %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &settings.values); err != nil {
		return nil, fmt.Errorf("preferences: decode %q: %w", path, err)
	}
	if err := validate(settings.values); err != nil {
		return nil, fmt.Errorf("preferences: %q: %w", path, err)
	}
	return settings, nil
}

func validate(values Values) error {
	if values.SoundVolume < 0 || values.SoundVolume > 1 {
		return fmt.Errorf("sound_volume must be between 0 and 1 (got %g)", values.SoundVolume)
	}
	if values.MusicVolume < 0 || values.MusicVolume > 1 {
		return fmt.Errorf("music_volume must be between 0 and 1 (got %g)", values.MusicVolume)
	}
	return nil
}

func (settings *Settings) Values() Values {
	settings.mu.RLock()
	defer settings.mu.RUnlock()
	return settings.values
}

func (settings *Settings) Path() string { return settings.path }
func (settings *Settings) Dirty() bool {
	settings.mu.RLock()
	defer settings.mu.RUnlock()
	return settings.dirty
}

func (settings *Settings) Update(values Values) error {
	if err := validate(values); err != nil {
		return err
	}
	settings.mu.Lock()
	settings.values, settings.dirty = values, true
	settings.mu.Unlock()
	return nil
}

func (settings *Settings) Save() error {
	settings.mu.RLock()
	data, err := json.MarshalIndent(settings.values, "", "  ")
	path := settings.path
	settings.mu.RUnlock()
	if path == "" {
		return errors.New("preferences: no persistence path configured")
	}
	if err != nil {
		return fmt.Errorf("preferences: encode: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("preferences: create directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".preferences-*.json")
	if err != nil {
		return fmt.Errorf("preferences: create temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = temporary.Write(append(data, '\n')); err == nil {
		err = temporary.Close()
	} else {
		_ = temporary.Close()
	}
	if err != nil {
		return fmt.Errorf("preferences: write: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("preferences: replace %q: %w", path, err)
	}
	settings.mu.Lock()
	settings.dirty = false
	settings.mu.Unlock()
	return nil
}
