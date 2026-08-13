package realm

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy"
)

type fixtureRecords struct{}

func (fixtureRecords) Load(string) ([]map[string]string, error) { return nil, nil }
func (fixtureRecords) Invalidate(string)                        {}
func (fixtureRecords) Loaded(string) bool                       { return true }

func TestManagerAllocatesSharedRealmModeAndRejectsDuplicate(t *testing.T) {
	manager, err := NewManager(fstest.MapFS{"boot.lua": {}}, fixtureRecords{})
	if err != nil {
		t.Fatal(err)
	}
	started := 0
	manager.start = func(_ context.Context, _ fs.FS, _ d2legacy.Records,
		config gameserver.Config,
	) (*gameserver.Host, error) {
		started++
		return &gameserver.Host{Mode: config.Mode}, nil
	}
	host, err := manager.Allocate(t.Context(), gameserver.Config{SessionID: "game-1"})
	if err != nil {
		t.Fatal(err)
	}
	if host.Mode != gameserver.ModeRealm || started != 1 {
		t.Fatalf("host mode/start count = %q/%d", host.Mode, started)
	}
	if _, err := manager.Allocate(t.Context(), gameserver.Config{SessionID: "game-1"}); !errors.Is(err, ErrGameExists) || started != 1 {
		t.Fatalf("duplicate allocation error/count = %v/%d", err, started)
	}
	if found, ok := manager.Game("game-1"); !ok || found != host {
		t.Fatal("allocated game is not discoverable")
	}
}

func TestManagerReleaseRemovesGame(t *testing.T) {
	manager, err := NewManager(fstest.MapFS{"boot.lua": {}}, fixtureRecords{})
	if err != nil {
		t.Fatal(err)
	}
	manager.start = func(context.Context, fs.FS, d2legacy.Records,
		gameserver.Config,
	) (*gameserver.Host, error) {
		return &gameserver.Host{}, nil
	}
	if _, err := manager.Allocate(t.Context(), gameserver.Config{SessionID: "game-1"}); err != nil {
		t.Fatal(err)
	}
	if err := manager.Release(t.Context(), "game-1"); err != nil {
		t.Fatal(err)
	}
	if _, found := manager.Game("game-1"); found {
		t.Fatal("released game remains discoverable")
	}
	if err := manager.Release(t.Context(), "game-1"); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("second release error = %v", err)
	}
}
