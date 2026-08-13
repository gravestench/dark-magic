package save

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	ProfileVersion       uint32 = 1
	MaxProfileBytes             = 4 << 20
	MaxProfileCharacters        = 64
)

var ErrProfile = errors.New("d2legacy player profile: invalid data")

type Profile struct {
	Version    uint32      `json:"version"`
	Characters []Character `json:"characters"`
	Selected   string      `json:"selected,omitempty"`
	Integrity  string      `json:"integrity"`
}

func (s *Store) Profile() Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	characters := make([]Character, len(s.entries))
	for index, character := range s.entries {
		characters[index] = cloneCharacter(character)
	}
	return Profile{Version: ProfileVersion, Characters: characters, Selected: s.selected}
}

func NewFromProfile(profile Profile) (*Store, error) {
	if err := validateProfile(profile); err != nil {
		return nil, err
	}
	store := New(profile.Characters...)
	store.selected = profile.Selected
	return store, nil
}

func EncodeProfile(destination io.Writer, profile Profile) error {
	if destination == nil {
		return fmt.Errorf("%w: destination is required", ErrProfile)
	}
	profile.Integrity = ""
	if err := validateProfile(profile); err != nil {
		return err
	}
	integrity, err := profileIntegrity(profile)
	if err != nil {
		return err
	}
	profile.Integrity = integrity
	encoded, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("%w: encode: %v", ErrProfile, err)
	}
	if len(encoded)+1 > MaxProfileBytes {
		return fmt.Errorf("%w: encoded profile exceeds %d bytes", ErrProfile, MaxProfileBytes)
	}
	_, err = destination.Write(append(encoded, '\n'))
	if err != nil {
		return fmt.Errorf("%w: write: %v", ErrProfile, err)
	}
	return nil
}

func DecodeProfile(source io.Reader) (Profile, error) {
	if source == nil {
		return Profile{}, fmt.Errorf("%w: source is required", ErrProfile)
	}
	data, err := io.ReadAll(io.LimitReader(source, MaxProfileBytes+1))
	if err != nil {
		return Profile{}, fmt.Errorf("%w: read: %v", ErrProfile, err)
	}
	if len(data) > MaxProfileBytes {
		return Profile{}, fmt.Errorf("%w: profile exceeds %d bytes", ErrProfile, MaxProfileBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var profile Profile
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("%w: decode: %v", ErrProfile, err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return Profile{}, fmt.Errorf("%w: trailing data", ErrProfile)
	}
	if err := validateProfile(profile); err != nil {
		return Profile{}, err
	}
	want, err := profileIntegrity(profile)
	if err != nil {
		return Profile{}, err
	}
	if profile.Integrity == "" || profile.Integrity != want {
		return Profile{}, fmt.Errorf("%w: integrity mismatch", ErrProfile)
	}
	return profile, nil
}

func validateProfile(profile Profile) error {
	if profile.Version != ProfileVersion || len(profile.Characters) > MaxProfileCharacters {
		return fmt.Errorf("%w: unsupported version or character count", ErrProfile)
	}
	identities := make(map[string]struct{}, len(profile.Characters))
	for _, character := range profile.Characters {
		if strings.TrimSpace(character.ID) == "" {
			return fmt.Errorf("%w: character ID is required", ErrProfile)
		}
		if _, exists := identities[character.ID]; exists {
			return fmt.Errorf("%w: duplicate character %q", ErrProfile, character.ID)
		}
		identities[character.ID] = struct{}{}
	}
	if profile.Selected != "" {
		if _, exists := identities[profile.Selected]; !exists {
			return fmt.Errorf("%w: selected character is absent", ErrProfile)
		}
	}
	return nil
}

func profileIntegrity(profile Profile) (string, error) {
	profile.Integrity = ""
	encoded, err := json.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("%w: integrity encoding: %v", ErrProfile, err)
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)), nil
}
