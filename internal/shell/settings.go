package shell

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const DefaultFontSize = 22.0

// SettingsValues contains the user-editable presentation settings shared by
// shell adapters. Values change immediately; Save explicitly persists them.
type SettingsValues struct {
	FontSize        float64 `json:"font_size"`
	ConsoleHeight   float64 `json:"console_height"`
	Opacity         float64 `json:"opacity"`
	TranscriptLimit int     `json:"transcript_limit"`
	AnimationSpeed  float64 `json:"animation_speed"`
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
	defaults := defaultSettingsValues()

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

	defaults := defaultSettingsValues()
	settings := &Settings{path: path, values: defaults, defaults: defaults}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}

	if err != nil {
		return nil, fmt.Errorf("shell settings: read %q: %w", path, err)
	}

	values := defaults
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("shell settings: decode %q: %w", path, err)
	}

	if err := validateSettings(values); err != nil {
		return nil, fmt.Errorf("shell settings: %q: %w", path, err)
	}

	settings.values = values

	return settings, nil
}

// validateSettings centralizes presentation limits so loading and live updates
// reject the same values before adapters observe them.
func validateSettings(values SettingsValues) error {
	if values.FontSize < 8 || values.FontSize > 48 {
		return fmt.Errorf("font_size must be between 8 and 48 (got %g)", values.FontSize)
	}

	if values.ConsoleHeight < 0.3 || values.ConsoleHeight > 1 {
		return fmt.Errorf("console_height must be between 0.3 and 1 (got %g)", values.ConsoleHeight)
	}

	if values.Opacity < 0.2 || values.Opacity > 1 {
		return fmt.Errorf("opacity must be between 0.2 and 1 (got %g)", values.Opacity)
	}

	if values.TranscriptLimit < 100 || values.TranscriptLimit > 100000 {
		return fmt.Errorf("transcript_limit must be between 100 and 100000 (got %d)", values.TranscriptLimit)
	}

	if values.AnimationSpeed < 0.25 || values.AnimationSpeed > 4 {
		return fmt.Errorf("animation_speed must be between 0.25 and 4 (got %g)", values.AnimationSpeed)
	}

	return nil
}

// defaultSettingsValues defines the single baseline shared by transient and persisted settings.
func defaultSettingsValues() SettingsValues {
	return SettingsValues{
		FontSize:        DefaultFontSize,
		ConsoleHeight:   0.6,
		Opacity:         0.93,
		TranscriptLimit: 2000,
		AnimationSpeed:  1,
	}
}

// Values returns a synchronized value snapshot suitable for adapter rendering.
func (s *Settings) Values() SettingsValues {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.values
}

// Defaults returns the immutable baseline captured when Settings was constructed.
func (s *Settings) Defaults() SettingsValues { return s.defaults }

// Path returns the persistence destination, or empty for transient settings.
func (s *Settings) Path() string { return s.path }

// Dirty reports whether a mutation marked the in-memory snapshot for persistence
// since the most recent load or successful save operation.
func (s *Settings) Dirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.dirty
}

// SetFontSize validates the complete prospective snapshot before publishing a
// live font change, keeping all Settings mutations on one invariant.
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

// Update validates and atomically applies a complete settings snapshot.
func (s *Settings) Update(values SettingsValues) error {
	if err := validateSettings(values); err != nil {
		return err
	}

	s.mu.Lock()
	s.values = values
	s.dirty = true
	s.mu.Unlock()

	return nil
}

// Reset restores construction-time defaults in memory and marks them for explicit persistence.
func (s *Settings) Reset() {
	s.mu.Lock()
	s.values = s.defaults
	s.dirty = true
	s.mu.Unlock()
}

// Save snapshots values and destination under the read lock, then releases it
// before file I/O. A successful replacement clears the established dirty marker.
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
	defer removeTemporarySettings(temporaryPath)

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

// removeTemporarySettings performs best-effort cleanup after failure or successful rename.
func removeTemporarySettings(path string) {
	_ = os.Remove(path)
}

// Reload discards unsaved values and reloads the configured file. A missing
// file restores defaults.
func (s *Settings) Reload() error {
	if s.path == "" {
		return errors.New("shell settings: no persistence path configured")
	}

	loaded, err := NewSettings(s.path)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.values = loaded.values
	s.dirty = false
	s.mu.Unlock()

	return nil
}
