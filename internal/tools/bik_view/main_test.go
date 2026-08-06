package main

import "testing"

func TestPlayerArguments(t *testing.T) {
	arguments := playerArguments("movie with spaces.bik")
	if got := arguments[len(arguments)-1]; got != "movie with spaces.bik" {
		t.Fatalf("input argument = %q", got)
	}
	if arguments[0] != "-autoexit" {
		t.Fatalf("arguments = %v", arguments)
	}
}
