package main

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	for input, want := range map[string]slog.Level{
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
