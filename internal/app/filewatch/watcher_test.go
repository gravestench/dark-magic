package filewatch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestScanReportsCreateChangeAndRemoval proves baseline files stay silent while later edits and removals are ordered.
func TestScanReportsCreateChangeAndRemoval(t *testing.T) {
	root := t.TempDir()

	name := filepath.Join(root, "boot.lua")
	if err := os.WriteFile(name, []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}

	var changed []string

	watcher := New(root, time.Hour, func(_ context.Context, name string) error {
		changed = append(changed, name)
		return nil
	})
	if err := watcher.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = watcher.Stop(context.Background()) })

	if err := os.WriteFile(name, []byte("two-longer"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := watcher.Scan(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(name); err != nil {
		t.Fatal(err)
	}

	if err := watcher.Scan(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := watcher.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(changed, []string{"boot.lua", "boot.lua"}) {
		t.Fatalf("changes = %v", changed)
	}
}

// TestScanOrdersChangesAndDoesNotRetryHandlerFailures documents delivery ordering and generation ownership.
func TestScanOrdersChangesAndDoesNotRetryHandlerFailures(t *testing.T) {
	root := t.TempDir()

	var handled []string

	wanted := errors.New("handler rejected file")

	watcher := New(root, time.Hour, func(_ context.Context, name string) error {
		handled = append(handled, name)
		if name == "a.lua" {
			return wanted
		}

		return nil
	})
	if err := watcher.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = watcher.Stop(context.Background()) })

	for _, name := range []string{"z.lua", "a.lua"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	err := watcher.Scan(context.Background())
	if err == nil || !strings.Contains(err.Error(), wanted.Error()) {
		t.Fatalf("scan error = %v, want handler failure", err)
	}

	if !reflect.DeepEqual(handled, []string{"a.lua", "z.lua"}) {
		t.Fatalf("handled paths = %v", handled)
	}

	handled = nil

	if err := watcher.Scan(context.Background()); err != nil {
		t.Fatalf("unchanged rescan: %v", err)
	}

	if len(handled) != 0 {
		t.Fatalf("failed handler was retried without a new change: %v", handled)
	}
}
