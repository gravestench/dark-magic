package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandUserPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := expandUserPath("~/d2_english_mpq")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "d2_english_mpq")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandUserPathLeavesOrdinaryPath(t *testing.T) {
	const path = "/tmp/mpqs"
	got, err := expandUserPath(path)
	if err != nil || got != path {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestExpandUserPathRejectsNamedHome(t *testing.T) {
	if _, err := expandUserPath("~someone/mpqs"); err == nil {
		t.Fatal("expected named home path to fail")
	}
}
