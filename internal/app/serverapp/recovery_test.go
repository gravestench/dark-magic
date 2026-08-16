package serverapp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/realm"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestReadGameRecoveryRequiresOwnerOnlyValidatedFile(t *testing.T) {
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 4}
	checksum, err := simulation.CompositeChecksum(snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	recovery, err := gamesession.NewRecoveryCheckpoint(simulation.Checkpoint{
		Tick: 4, Checksum: checksum, Snapshot: &snapshot,
	}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := realm.NewGameRecovery(recovery, []string{"player"})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(handoff)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "recovery.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadGameRecovery(path)
	if err != nil || loaded.Checkpoint.State.Tick != 4 || loaded.Checkpoint.Checksum != recovery.Checksum ||
		len(loaded.PlayerIDs) != 1 || loaded.PlayerIDs[0] != "player" {
		t.Fatalf("loaded checkpoint=%#v error=%v", loaded, err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGameRecovery(path); err == nil {
		t.Fatal("group/world-readable recovery checkpoint was accepted")
	}
}
