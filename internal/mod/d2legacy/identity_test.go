package d2legacy

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

func TestIdentityIncludesCanonicalGameplayConfiguration(t *testing.T) {
	left, err := Identity(content.D2Legacy(), map[string]any{"difficulty": "normal", "seed": float64(7)})
	if err != nil {
		t.Fatal(err)
	}
	same, err := Identity(content.D2Legacy(), map[string]any{"seed": float64(7), "difficulty": "normal"})
	if err != nil {
		t.Fatal(err)
	}
	different, err := Identity(content.D2Legacy(), map[string]any{"difficulty": "nightmare", "seed": float64(7)})
	if err != nil {
		t.Fatal(err)
	}
	if left.ConfigurationHash != same.ConfigurationHash {
		t.Fatal("map insertion order changed configuration identity")
	}
	if left.ConfigurationHash == different.ConfigurationHash {
		t.Fatal("changed gameplay configuration retained identity")
	}
}
