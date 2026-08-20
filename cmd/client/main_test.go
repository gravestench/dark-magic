package main

import (
	"log/slog"
	"testing"

	"github.com/gravestench/dark-magic/internal/logging"
)

// TestParseWindowSize covers explicit, omitted, and malformed native workspace dimensions.
func TestParseWindowSize(t *testing.T) {
	t.Parallel()
	width, height, err := parseWindowSize("1600x1000")
	if err != nil || width != 1600 || height != 1000 {
		t.Fatalf("parseWindowSize = %d×%d, %v", width, height, err)
	}
	if width, height, err := parseWindowSize(""); err != nil || width != 0 || height != 0 {
		t.Fatalf("empty parseWindowSize = %d×%d, %v", width, height, err)
	}
	if _, _, err := parseWindowSize("wide"); err == nil {
		t.Fatal("parseWindowSize accepted malformed dimensions")
	}
}

// TestParseLogLevel protects the CLI vocabulary so configuration examples do
// not silently diverge from the typed logging levels accepted at startup.
func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	for input, want := range map[string]slog.Level{
		"trace":   logging.LevelTrace,
		"debug":   slog.LevelDebug,
		"INFO":    slog.LevelInfo,
		" warn ":  slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
	} {
		got, err := parseLogLevel(input)
		if err != nil {
			t.Errorf("parseLogLevel(%q): %v", input, err)
		} else if got != want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", input, got, want)
		}
	}

	if _, err := parseLogLevel("verbose"); err == nil {
		t.Fatal("parseLogLevel accepted an unsupported level")
	}
}

// TestDevelopmentCharacters protects fixture determinism because capture and
// acceptance comparisons rely on stable character identity and ordering.
func TestDevelopmentCharacters(t *testing.T) {
	t.Parallel()

	if characters := developmentCharacters(0); characters != nil {
		t.Fatalf("developmentCharacters(0) = %#v, want nil", characters)
	}

	characters := developmentCharacters(10)
	if len(characters) != 10 {
		t.Fatalf("developmentCharacters(10) length = %d", len(characters))
	}

	wantClasses := []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Assassin", "Druid", "Amazon"}
	for index, wantClass := range wantClasses {
		character := characters[index]
		if character.Class != wantClass || character.Level != index+1 || !character.Expansion {
			t.Errorf("character %d = %#v", index, character)
		}
	}

	if characters[0].ID != "fixture-01" || characters[0].Name != "Hero01" {
		t.Fatalf("first character = %#v", characters[0])
	}

	if !characters[2].Hardcore || characters[1].Hardcore || characters[3].Hardcore {
		t.Fatalf("unexpected hardcore sequence: %#v", characters[:4])
	}
}
