package filewatch

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

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
