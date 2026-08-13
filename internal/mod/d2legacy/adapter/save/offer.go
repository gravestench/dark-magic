package save

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

const (
	CharacterOfferVersion  uint32 = 1
	MaxCharacterOfferBytes        = 128 << 10
)

type CharacterOffer struct {
	Version   uint32    `json:"version"`
	Character Character `json:"character"`
	Integrity string    `json:"integrity"`
}

func EncodeCharacterOffer(character Character) ([]byte, error) {
	offer := CharacterOffer{Version: CharacterOfferVersion, Character: cloneCharacter(character)}
	if err := validateCharacterOffer(offer); err != nil {
		return nil, err
	}
	integrity, err := characterOfferIntegrity(offer)
	if err != nil {
		return nil, err
	}
	offer.Integrity = integrity
	encoded, err := json.Marshal(offer)
	if err != nil {
		return nil, fmt.Errorf("%w: encode character offer: %v", ErrProfile, err)
	}
	if len(encoded) > MaxCharacterOfferBytes {
		return nil, fmt.Errorf("%w: character offer exceeds %d bytes", ErrProfile, MaxCharacterOfferBytes)
	}
	return encoded, nil
}

func DecodeCharacterOffer(data []byte) (Character, error) {
	if len(data) == 0 || len(data) > MaxCharacterOfferBytes {
		return Character{}, fmt.Errorf("%w: invalid character offer size", ErrProfile)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var offer CharacterOffer
	if err := decoder.Decode(&offer); err != nil {
		return Character{}, fmt.Errorf("%w: decode character offer: %v", ErrProfile, err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return Character{}, fmt.Errorf("%w: trailing character offer data", ErrProfile)
	}
	if err := validateCharacterOffer(offer); err != nil {
		return Character{}, err
	}
	want, err := characterOfferIntegrity(offer)
	if err != nil {
		return Character{}, err
	}
	if offer.Integrity == "" || offer.Integrity != want {
		return Character{}, fmt.Errorf("%w: character offer integrity mismatch", ErrProfile)
	}
	return cloneCharacter(offer.Character), nil
}

func validateCharacterOffer(offer CharacterOffer) error {
	if offer.Version != CharacterOfferVersion || offer.Character.ID == "" {
		return fmt.Errorf("%w: character offer version and ID are required", ErrProfile)
	}
	return nil
}

func characterOfferIntegrity(offer CharacterOffer) (string, error) {
	offer.Integrity = ""
	encoded, err := json.Marshal(offer)
	if err != nil {
		return "", fmt.Errorf("%w: character offer integrity: %v", ErrProfile, err)
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(encoded)), nil
}
