package modcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
		if err := writeJSONAtomic(fileName, profile); err != nil {
			return Profile{}, false, fmt.Errorf("modcache: create profile: %w", err)
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
