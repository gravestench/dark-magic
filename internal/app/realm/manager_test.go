package realm

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

type fixtureRecords struct{}

func (fixtureRecords) Load(string) ([]map[string]string, error) { return nil, nil }
func (fixtureRecords) Invalidate(string)                        {}
func (fixtureRecords) Loaded(string) bool                       { return true }

type managerWorker struct{ closed bool }

func (*managerWorker) Describe(context.Context) (WorkerDescription, error) {
	return WorkerDescription{}, nil
}
func (*managerWorker) Status(context.Context) (WorkerStatus, error) {
	return WorkerStatus{Ready: true}, nil
}
func (*managerWorker) Checkpoint(context.Context) (gamesession.RecoveryCheckpoint, error) {
	return testCheckpoint(orchestrationIdentity(), 0)
}
func (*managerWorker) AdmitCharacter(context.Context, WorkerAdmission) error { return nil }
func (*managerWorker) RemoveCharacter(context.Context, string) error         { return nil }
func (*managerWorker) ProjectCharacter(_ context.Context, _ string, character d2save.Character) (d2save.Character, error) {
	return character, nil
}
func (worker *managerWorker) Close(context.Context) error {
	worker.closed = true
	return nil
}

func TestManagerAllocatesSharedRealmModeAndRejectsDuplicate(t *testing.T) {
	manager, err := NewManager(fstest.MapFS{"boot.lua": {}}, fixtureRecords{})
	if err != nil {
		t.Fatal(err)
	}
	started := 0
	var mode gameserver.Mode
	allocatedWorker := &managerWorker{}
	manager.start = func(_ context.Context, config gameserver.Config) (WorkerClient, error) {
		started++
		mode = config.Mode
		return allocatedWorker, nil
	}
	worker, err := manager.Allocate(t.Context(), gameserver.Config{SessionID: "game-1"})
	if err != nil {
		t.Fatal(err)
	}
	if mode != gameserver.ModeRealm || started != 1 {
		t.Fatalf("host mode/start count = %q/%d", mode, started)
	}
	if _, err := manager.Allocate(t.Context(), gameserver.Config{SessionID: "game-1"}); !errors.Is(err, ErrGameExists) || started != 1 {
		t.Fatalf("duplicate allocation error/count = %v/%d", err, started)
	}
	if found, ok := manager.Game("game-1"); !ok || found != worker {
		t.Fatal("allocated game is not discoverable")
	}
}

func TestManagerReleaseRemovesGame(t *testing.T) {
	manager, err := NewManager(fstest.MapFS{"boot.lua": {}}, fixtureRecords{})
	if err != nil {
		t.Fatal(err)
	}
	worker := &managerWorker{}
	manager.start = func(context.Context, gameserver.Config) (WorkerClient, error) {
		return worker, nil
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
	if !worker.closed {
		t.Fatal("released worker was not closed")
	}
	if err := manager.Release(t.Context(), "game-1"); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("second release error = %v", err)
	}
}
