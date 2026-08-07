package simulation

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestReplayRestoresAppliesAndDiagnosesFirstDivergence(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "test.counter", Version: 1, Fields: []akara.Field{{Name: "value", Kind: akara.FieldInt64}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := store.Set(entity, map[string]any{"value": int64(0)}); err != nil {
		t.Fatal(err)
	}
	initial, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	commands := []Command{
		{Tick: 1, Player: "p1", Sequence: 1, Kind: "set", Payload: json.RawMessage(`{"value":4}`)},
		{Tick: 2, Player: "p1", Sequence: 2, Kind: "set", Payload: json.RawMessage(`{"value":9}`)},
	}
	apply := func(target *gameecs.Engine, command Command) error {
		var payload struct {
			Value int64 `json:"value"`
		}
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return err
		}
		componentStore, _ := akara.GetDynamicStore(target.World(), "test.counter")
		component, _ := componentStore.Get(entity)
		return component.Set("value", payload.Value)
	}
	checkpoints := make([]Checkpoint, 0, len(commands))
	for _, command := range commands {
		if err := apply(engine, command); err != nil {
			t.Fatal(err)
		}
		if err := engine.Update(40 * time.Millisecond); err != nil {
			t.Fatal(err)
		}
		snapshot, err := engine.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		checksum, err := snapshot.Checksum()
		if err != nil {
			t.Fatal(err)
		}
		copy := snapshot
		checkpoints = append(checkpoints, Checkpoint{Tick: command.Tick, Checksum: checksum, Snapshot: &copy})
	}
	replay := Replay{Version: ReplayVersion, StepNanos: int64(40 * time.Millisecond), Initial: initial, Commands: commands, Checkpoints: checkpoints}
	if err := VerifyReplay(replay, nil, apply); err != nil {
		t.Fatal(err)
	}
	replay.Commands[1].Payload = json.RawMessage(`{"value":10}`)
	var desync *DesyncError
	if err := VerifyReplay(replay, nil, apply); !errors.As(err, &desync) || !strings.Contains(desync.Detail, `component "test.counter" entity 1 field "value" differs`) {
		t.Fatalf("desync = %#v, %v", desync, err)
	}
}
