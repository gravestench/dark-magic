package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestHandlerObserverReceivesResolvedAttributes(t *testing.T) {
	var observed Record
	handler := NewHandlerWithObserver(nil, func(record Record) { observed = record })
	logger := slog.New(handler).With("component", "shell")
	logger.InfoContext(context.Background(), "hello", "answer", 42)
	if observed.Message != "hello" || observed.Attributes["component"] != "shell" || observed.Attributes["answer"] != float64(42) {
		t.Fatalf("observed = %#v", observed)
	}
}

func TestParseLevelSupportsTraceBelowDebug(t *testing.T) {
	level, err := ParseLevel("trace")
	if err != nil {
		t.Fatal(err)
	}
	if level != LevelTrace || level >= slog.LevelDebug {
		t.Fatalf("trace level = %v", level)
	}
}
