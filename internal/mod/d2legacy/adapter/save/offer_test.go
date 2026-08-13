package save

import (
	"bytes"
	"errors"
	"testing"
)

func TestCharacterOfferRoundTripAndIntegrity(t *testing.T) {
	encoded, err := EncodeCharacterOffer(Character{ID: "hero", Name: "Hero", Stats: &Stats{Health: 20}})
	if err != nil {
		t.Fatal(err)
	}
	character, err := DecodeCharacterOffer(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if character.ID != "hero" || character.Stats.Health != 20 {
		t.Fatalf("character = %#v", character)
	}
	tampered := bytes.Replace(encoded, []byte(`"health":20`), []byte(`"health":99`), 1)
	if _, err := DecodeCharacterOffer(tampered); !errors.Is(err, ErrProfile) {
		t.Fatalf("tamper error = %v", err)
	}
	if _, err := DecodeCharacterOffer(make([]byte, MaxCharacterOfferBytes+1)); !errors.Is(err, ErrProfile) {
		t.Fatalf("oversize error = %v", err)
	}
}
