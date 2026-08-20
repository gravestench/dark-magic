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
	// ProfileVersion identifies the only durable local-profile schema currently supported.
	ProfileVersion uint32 = 1
	// MaxProfileBytes bounds profile reads before strict JSON decoding allocates nested roster data.
	MaxProfileBytes = 4 << 20
	// MaxProfileCharacters prevents malformed profiles from constructing an unbounded local roster.
	MaxProfileCharacters = 64
)

// ErrProfile is the stable sentinel wrapped by profile encoding, validation, integrity, and file failures.
var ErrProfile = errors.New("d2legacy player profile: invalid data")

// Profile is the versioned durable envelope for a player-owned roster and active selection.
type Profile struct {
	Version    uint32      `json:"version"`
	Characters []Character `json:"characters"`
	Selected   string      `json:"selected,omitempty"`
	Integrity  string      `json:"integrity"`
}

// Profile returns a deep roster snapshot without an integrity digest; encoding computes the digest canonically.
func (s *Store) Profile() Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	characters := make([]Character, len(s.entries))
	for index, character := range s.entries {
		characters[index] = cloneCharacter(character)
	}

	return Profile{Version: ProfileVersion, Characters: characters, Selected: s.selected}
}

// NewFromProfile validates roster relationships before constructing a defensive in-memory store.
func NewFromProfile(profile Profile) (*Store, error) {
	if err := validateProfile(profile); err != nil {
		return nil, err
	}

	store := New(profile.Characters...)
	store.selected = profile.Selected

	return store, nil
}

// EncodeProfile writes one newline-terminated, integrity-protected profile without closing the caller's writer.
func EncodeProfile(destination io.Writer, profile Profile) error {
	if destination == nil {
		return fmt.Errorf("%w: destination is required", ErrProfile)
	}

	encoded, err := encodeProfileDocument(profile)
	if err != nil {
		return err
	}

	if _, err := destination.Write(append(encoded, '\n')); err != nil {
		return fmt.Errorf("%w: write: %v", ErrProfile, err)
	}

	return nil
}

// encodeProfileDocument validates and signs the canonical JSON document before enforcing the durable size limit.
func encodeProfileDocument(profile Profile) ([]byte, error) {
	profile.Integrity = ""
	if err := validateProfile(profile); err != nil {
		return nil, err
	}

	integrity, err := profileIntegrity(profile)
	if err != nil {
		return nil, err
	}

	profile.Integrity = integrity

	encoded, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", ErrProfile, err)
	}

	if len(encoded)+1 > MaxProfileBytes {
		return nil, fmt.Errorf("%w: encoded profile exceeds %d bytes", ErrProfile, MaxProfileBytes)
	}

	return encoded, nil
}

// DecodeProfile bounds and strictly decodes one profile document, then verifies its canonical digest.
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

	return decodeProfileDocument(data)
}

// decodeProfileDocument rejects schema extensions and trailing values before checking roster and integrity invariants.
func decodeProfileDocument(data []byte) (Profile, error) {
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

// validateProfile enforces version, roster size, unique opaque identities, and selection membership.
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

// profileIntegrity hashes canonical JSON with the integrity field cleared so encoding and decoding agree exactly.
func profileIntegrity(profile Profile) (string, error) {
	profile.Integrity = ""

	encoded, err := json.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("%w: integrity encoding: %v", ErrProfile, err)
	}

	return fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)), nil
}
