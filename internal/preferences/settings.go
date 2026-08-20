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

// Values is the versioned, serializable client-preference snapshot. These are
// presentation/backend choices, not authoritative character or game state.
type Values struct {
	Version               int     `json:"version"`
	SoundVolume           float64 `json:"sound_volume"`
	MusicVolume           float64 `json:"music_volume"`
	DebugTextureResidency bool    `json:"debug_texture_residency"`
	TextureUploadBudgetMB float64 `json:"texture_upload_budget_mb"`
	TextureCacheBudgetMB  float64 `json:"texture_cache_budget_mb"`
	CameraFollowStrategy  string  `json:"camera_follow_strategy"`
	CameraFollowDuration  float64 `json:"camera_follow_duration"`
	CameraFollowParam1    float64 `json:"camera_follow_param_1"`
	CameraFollowParam2    float64 `json:"camera_follow_param_2"`
	CameraFollowParam3    float64 `json:"camera_follow_param_3"`
	RealmGateway          string  `json:"realm_gateway"`
}

// Settings owns validated live values and their optional host-file lifetime.
// Callers receive copies, then use Update so validation and dirty tracking stay
// centralized instead of mutating the serialized structure directly.
type Settings struct {
	mu     sync.RWMutex
	path   string
	values Values
	dirty  bool
}

// Defaults returns a fresh copy of the built-in preference baseline.
func Defaults() Values {
	return Values{
		Version:               1,
		SoundVolume:           .5,
		MusicVolume:           .5,
		TextureUploadBudgetMB: 16,
		TextureCacheBudgetMB:  512,
		CameraFollowStrategy:  "instant",
		RealmGateway:          "127.0.0.1",
	}
}

// NewTransient creates in-memory preferences for tests and headless tools, where
// Save must remain unavailable rather than selecting a surprising host path.
func NewTransient() *Settings { return &Settings{values: Defaults()} }

// New loads preferences from path, or the platform configuration directory when
// path is empty. A missing file is normal first-run state; malformed data is not.
func New(path string) (*Settings, error) {
	path, err := resolvePath(path)
	if err != nil {
		return nil, err
	}

	settings := &Settings{path: path, values: Defaults()}

	data, err := os.ReadFile(settings.path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}

	if err != nil {
		return nil, fmt.Errorf("preferences: read %q: %w", settings.path, err)
	}

	if err := decode(settings.path, data, &settings.values); err != nil {
		return nil, err
	}

	applyCompatibilityDefaults(&settings.values)

	if err := validate(settings.values); err != nil {
		return nil, fmt.Errorf("preferences: %q: %w", settings.path, err)
	}

	return settings, nil
}

// resolvePath keeps the platform-specific default in one place so every load
// error can name the concrete path that was actually selected.
func resolvePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	directory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("preferences: user config directory: %w", err)
	}

	return filepath.Join(directory, "dark-magic", "preferences.json"), nil
}

// decode reads both the values and the schema marker from the same bytes. The
// marker distinguishes an explicit legacy value from an old implicit default.
func decode(path string, data []byte, values *Values) error {
	if err := json.Unmarshal(data, values); err != nil {
		return fmt.Errorf("preferences: decode %q: %w", path, err)
	}

	var schema struct {
		Version *int `json:"version"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("preferences: decode schema %q: %w", path, err)
	}

	if schema.Version == nil && values.TextureUploadBudgetMB == 4 {
		values.TextureUploadBudgetMB = Defaults().TextureUploadBudgetMB
	}

	return nil
}

// applyCompatibilityDefaults advances the schema version and fills legacy zero-value options without replacing
// explicit nonzero option values.
func applyCompatibilityDefaults(values *Values) {
	defaults := Defaults()
	values.Version = defaults.Version

	if values.TextureUploadBudgetMB == 0 {
		values.TextureUploadBudgetMB = defaults.TextureUploadBudgetMB
	}

	if values.TextureCacheBudgetMB == 0 {
		values.TextureCacheBudgetMB = defaults.TextureCacheBudgetMB
	}

	if values.CameraFollowStrategy == "" {
		values.CameraFollowStrategy = defaults.CameraFollowStrategy
	}

	if values.RealmGateway == "" {
		values.RealmGateway = defaults.RealmGateway
	}
}

// validate rejects values that downstream presentation and networking code
// cannot interpret safely, keeping invalid snapshots out of live settings.
func validate(values Values) error {
	if values.SoundVolume < 0 || values.SoundVolume > 1 {
		return fmt.Errorf("sound_volume must be between 0 and 1 (got %g)", values.SoundVolume)
	}

	if values.MusicVolume < 0 || values.MusicVolume > 1 {
		return fmt.Errorf("music_volume must be between 0 and 1 (got %g)", values.MusicVolume)
	}

	if values.TextureUploadBudgetMB < .25 || values.TextureUploadBudgetMB > 64 {
		return fmt.Errorf("texture_upload_budget_mb must be between 0.25 and 64 (got %g)", values.TextureUploadBudgetMB)
	}

	if values.TextureCacheBudgetMB < 64 || values.TextureCacheBudgetMB > 4096 {
		return fmt.Errorf("texture_cache_budget_mb must be between 64 and 4096 (got %g)", values.TextureCacheBudgetMB)
	}

	strategies := map[string]bool{
		"instant": true, "linear": true, "quad_in": true, "quad_out": true,
		"quad_in_out": true, "cubic_in": true, "cubic_out": true, "cubic_in_out": true,
		"exponential_out": true, "back_out": true,
	}
	if !strategies[values.CameraFollowStrategy] {
		return fmt.Errorf("unknown camera_follow_strategy %q", values.CameraFollowStrategy)
	}

	if values.CameraFollowDuration < 0 || values.CameraFollowDuration > 5 {
		return fmt.Errorf("camera_follow_duration must be between 0 and 5 seconds (got %g)", values.CameraFollowDuration)
	}

	if values.RealmGateway == "" || len(values.RealmGateway) > 255 {
		return errors.New("realm_gateway must contain between 1 and 255 bytes")
	}

	parameters := map[string]float64{
		"camera_follow_param_1": values.CameraFollowParam1,
		"camera_follow_param_2": values.CameraFollowParam2,
		"camera_follow_param_3": values.CameraFollowParam3,
	}
	for name, value := range parameters {
		if value < -100 || value > 100 {
			return fmt.Errorf("%s must be between -100 and 100 (got %g)", name, value)
		}
	}

	return nil
}

// Values returns a copy so callers cannot bypass validation or dirty tracking
// by mutating the live snapshot through an alias.
func (settings *Settings) Values() Values {
	settings.mu.RLock()
	defer settings.mu.RUnlock()

	return settings.values
}

// Path reports the configured persistence target; an empty path identifies
// transient settings whose Save operation intentionally fails.
func (settings *Settings) Path() string { return settings.path }

// Dirty reports whether Update has installed values not yet written by Save.
func (settings *Settings) Dirty() bool {
	settings.mu.RLock()
	defer settings.mu.RUnlock()

	return settings.dirty
}

// Update validates the entire replacement snapshot before publishing it, so a
// rejected field cannot partially change the settings observed by readers.
func (settings *Settings) Update(values Values) error {
	if err := validate(values); err != nil {
		return err
	}

	settings.mu.Lock()
	settings.values, settings.dirty = values, true
	settings.mu.Unlock()

	return nil
}

// Save stages a complete preferences file and clears the dirty marker only after the final rename succeeds.
func (settings *Settings) Save() error {
	// Copy the serializable state under the read lock, then release it before host
	// I/O so readers are not blocked by a slow filesystem.
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
	defer removeTemporaryFile(temporaryPath)

	// Closing before rename is required on platforms that do not allow an open
	// temporary file to replace its destination.
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

// removeTemporaryFile preserves Save's established error contract by treating staging-file cleanup as best effort.
func removeTemporaryFile(path string) {
	_ = os.Remove(path)
}
