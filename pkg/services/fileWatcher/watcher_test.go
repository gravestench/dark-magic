package fileWatcher

import (
	"io"
	"log/slog"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestHandleEventFiltersAndDispatchesSynchronously(t *testing.T) {
	called := 0
	service := &Service{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		activeWatchers: map[string]FileHandlerFunc{
			"mod.lua": func(path string) error {
				called++
				if path != "mod.lua" {
					t.Fatalf("path = %q", path)
				}
				return nil
			},
		},
	}
	service.handleEvent(fsnotify.Event{Name: "mod.lua", Op: fsnotify.Chmod})
	service.handleEvent(fsnotify.Event{Name: "mod.lua", Op: fsnotify.Write})
	if called != 1 {
		t.Fatalf("callback count = %d, want 1", called)
	}
}

func TestCloseWatcherIsIdempotentWithoutInitialization(t *testing.T) {
	service := &Service{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	service.CloseWatcher()
	service.CloseWatcher()
}
