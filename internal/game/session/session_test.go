package session

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestSessionCanonicalizesArrivalOrderAndExportsVerifiableReplay(t *testing.T) {
	build := func() (*Session, func(*gameecs.Engine, simulation.Command) error) {
		engine := gameecs.New()
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "test.total", Version: 1, Fields: []akara.Field{{Name: "value", Kind: akara.FieldInt64}}})
		if err != nil {
			t.Fatal(err)
		}
		entity := engine.World().MustCreateEntity()
		if _, err := store.Set(entity, nil); err != nil {
			t.Fatal(err)
		}
		apply := func(target *gameecs.Engine, command simulation.Command) error {
			var payload struct {
				Add int64 `json:"add"`
			}
			if err := json.Unmarshal(command.Payload, &payload); err != nil {
				return err
			}
			total, _ := akara.GetDynamicStore(target.World(), "test.total")
			component, _ := total.Get(entity)
			value, _ := component.Get("value")
			return component.Set("value", value.(int64)*10+payload.Add)
		}
		session, err := New(engine, Config{Step: 40 * time.Millisecond, MaxCommandLead: 2, CheckpointInterval: 1})
		if err != nil {
			t.Fatal(err)
		}
		if err := session.Register("add", CommandHandler{Validate: func(simulation.Command) error { return nil }, Apply: apply}); err != nil {
			t.Fatal(err)
		}
		return session, apply
	}
	first, apply := build()
	defer first.Close()
	commands := []simulation.Command{
		{Tick: 1, Player: "bravo", Sequence: 1, Kind: "add", Payload: json.RawMessage(`{"add":2}`)},
		{Tick: 1, Player: "alpha", Sequence: 1, Kind: "add", Payload: json.RawMessage(`{"add":1}`)},
	}
	for _, command := range commands {
		if err := first.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := first.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if replay.Commands[0].Player != "alpha" || replay.Commands[1].Player != "bravo" {
		t.Fatalf("recorded order = %#v", replay.Commands)
	}
	if err := simulation.VerifyReplay(replay, nil, apply); err != nil {
		t.Fatal(err)
	}
	replay.Commands[0].Player = "mutated"
	replay.Commands[0].Payload[0] = '['
	replay.Initial.Entities[0] = 99
	replay.Checkpoints[0].Snapshot.Entities[0] = 99
	again, err := first.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if again.Commands[0].Player != "alpha" || string(again.Commands[0].Payload) != `{"add":1}` || again.Initial.Entities[0] != 1 || again.Checkpoints[0].Snapshot.Entities[0] != 1 {
		t.Fatalf("export mutated session recording: %#v", again)
	}
	second, _ := build()
	defer second.Close()
	for index := len(commands) - 1; index >= 0; index-- {
		if err := second.Submit(commands[index]); err != nil {
			t.Fatal(err)
		}
	}
	if err := second.Step(); err != nil {
		t.Fatal(err)
	}
	left, _ := first.engine.Snapshot()
	right, _ := second.engine.Snapshot()
	leftChecksum, _ := left.Checksum()
	rightChecksum, _ := right.Checksum()
	if leftChecksum != rightChecksum {
		t.Fatalf("arrival order changed state: %s != %s", leftChecksum, rightChecksum)
	}
}
