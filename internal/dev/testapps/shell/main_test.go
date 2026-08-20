package main

import (
	"errors"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestIsUnexpectedTerminalExit locks in which TUI outcomes may end the development harness without a fatal exit.
func TestIsUnexpectedTerminalExit(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "clean completion", err: nil, want: false},
		{name: "direct interruption", err: tea.ErrInterrupted, want: false},
		{name: "wrapped interruption", err: fmt.Errorf("terminal stopped: %w", tea.ErrInterrupted), want: false},
		{name: "unexpected failure", err: errors.New("terminal failed"), want: true},
	}

	for _, testCase := range testCases {
		got := isUnexpectedTerminalExit(testCase.err)
		if got != testCase.want {
			t.Errorf("isUnexpectedTerminalExit(%s) = %t, want %t", testCase.name, got, testCase.want)
		}
	}
}
