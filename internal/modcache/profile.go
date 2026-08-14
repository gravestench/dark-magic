package modcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Profile struct {
	Schema  string   `json:"schema"`
	Enabled []string `json:"enabled"`
}

func LoadOrCreateProfile(fileName string, defaults []string) (Profile, bool, error) {
	if strings.TrimSpace(fileName) == "" {
		return Profile{}, false, errors.New("modcache: profile path is required")
	}
	data, err := os.ReadFile(fileName)
	if os.IsNotExist(err) {
		profile := Profile{Schema: ProfileSchema, Enabled: append([]string(nil), defaults...)}
		if err := ValidateProfile(profile); err != nil {
			return Profile{}, false, err
		}
		created, err := writeJSONExclusive(fileName, profile)
		if err != nil {
			return Profile{}, false, fmt.Errorf("modcache: create profile: %w", err)
		}
		if !created {
			return LoadOrCreateProfile(fileName, defaults)
		}
		return profile, true, nil
	}
	if err != nil {
		return Profile{}, false, fmt.Errorf("modcache: read profile: %w", err)
	}
	var profile Profile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, false, fmt.Errorf("modcache: decode profile: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Profile{}, false, errors.New("modcache: profile has trailing data")
	}
	if err := ValidateProfile(profile); err != nil {
		return Profile{}, false, err
	}
	return profile, false, nil
}

// SaveProfile validates and atomically replaces a user-owned extension
// selection. Product composition uses this for explicit schema-preserving
// migrations, never for temporary command-line overrides.
func SaveProfile(fileName string, profile Profile) error {
	if strings.TrimSpace(fileName) == "" {
		return errors.New("modcache: profile path is required")
	}
	if err := ValidateProfile(profile); err != nil {
		return err
	}
	if err := writeJSONAtomic(fileName, profile); err != nil {
		return fmt.Errorf("modcache: save profile: %w", err)
	}
	return nil
}

func writeJSONExclusive(fileName string, value any) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(fileName), 0o700); err != nil {
		return false, err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return false, err
	}
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, writeErr := file.Write(append(data, '\n'))
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		_ = os.Remove(fileName)
		return false, err
	}
	return true, nil
}

func ValidateProfile(profile Profile) error {
	if profile.Schema != ProfileSchema {
		return errors.New("modcache: invalid profile schema")
	}
	seen := make(map[string]struct{}, len(profile.Enabled))
	for _, id := range profile.Enabled {
		if !validID(id) {
			return fmt.Errorf("modcache: invalid enabled mod %q", id)
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("modcache: duplicate enabled mod %q", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}
