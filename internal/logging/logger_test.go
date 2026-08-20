package logging

import (
	"context"
	"log/slog"
	"testing"
)

// TestHandlerObserverReceivesResolvedAttributes verifies that observers receive attributes from both derived handlers
// and the current record after slog has applied its normal value conversion.
func TestHandlerObserverReceivesResolvedAttributes(t *testing.T) {
	var observed Record

	handler := NewHandlerWithObserver(nil, func(record Record) { observed = record })
	logger := slog.New(handler).With("component", "shell")
	logger.InfoContext(context.Background(), "hello", "answer", 42)

	if observed.Message != "hello" || observed.Attributes["component"] != "shell" ||
		observed.Attributes["answer"] != float64(42) {
		t.Fatalf("observed = %#v", observed)
	}
}

// TestParseLevelSupportsTraceBelowDebug protects the ordering contract that lets trace logs be filtered independently
// from ordinary debug output.
func TestParseLevelSupportsTraceBelowDebug(t *testing.T) {
	level, err := ParseLevel("trace")
	if err != nil {
		t.Fatal(err)
	}

	if level != LevelTrace || level >= slog.LevelDebug {
		t.Fatalf("trace level = %v", level)
	}
}
