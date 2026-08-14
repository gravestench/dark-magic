package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestSessionPinsAuthoritativeRuntimeStateInReplay(t *testing.T) {
	engine := gameecs.New()
	session, err := New(engine, Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	stores := simulation.NewStateStore()
	if err := stores.Register("d2legacy.test", "test/v1", []byte(`{"value":1}`)); err != nil {
		t.Fatal(err)
	}
	identity := simulation.RuntimeIdentity{ModID: "d2legacy", ContractVersion: "v1", PackageHash: "package", AuthoritativeHash: "rules", ConfigurationHash: "config", CapabilityVersions: map[string]string{"d2legacy.ecs": "v1"}}
	if err := session.RegisterAuthoritativeRuntime(identity, stores); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.InitialParticipants) != 2 || len(replay.Checkpoints) != 1 || len(replay.Checkpoints[0].Participants) != 2 {
		t.Fatalf("runtime participants missing from replay: %#v", replay)
	}

	replayStores := simulation.NewStateStore()
	if err := replayStores.Register("d2legacy.test", "test/v1", nil); err != nil {
		t.Fatal(err)
	}
	replayIdentity, err := simulation.NewIdentityParticipant(identity)
	if err != nil {
		t.Fatal(err)
	}
	if err := simulation.VerifyReplay(replay, nil, func(*gameecs.Engine, simulation.Command) error { return nil }, replayIdentity, replayStores); err != nil {
		t.Fatal(err)
	}
	got, found := replayStores.Read("d2legacy.test")
	if !found || !bytes.Equal(got.Data, []byte(`{"value":1}`)) {
		t.Fatalf("restored runtime state = %#v, %v", got, found)
	}
}

func TestSessionExportsDefensiveReplayContainerEvidence(t *testing.T) {
	engine := gameecs.New()
	session, err := New(engine, Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	manifests := map[string]simulation.ReplayManifest{
		"session": {Schema: "dark-magic.session/v1", Data: json.RawMessage(`{"mod":"d2legacy"}`)},
	}
	events := []simulation.ReplayEvent{{Tick: 1, Kind: "session.started", Payload: json.RawMessage(`{}`)}}
	container, err := session.ReplayContainer(manifests, events)
	if err != nil {
		t.Fatal(err)
	}
	manifests["session"] = simulation.ReplayManifest{}
	events[0].Payload[0] = '['
	if container.Manifests["session"].Schema != "dark-magic.session/v1" ||
		string(container.Manifests["session"].Data) != `{"mod":"d2legacy"}` ||
		string(container.Events[0].Payload) != `{}` || len(container.Replay.Checkpoints) != 1 {
		t.Fatalf("container changed with caller inputs: %#v", container)
	}
}

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

func TestNetworkInputRollsBackAndReplaysToCanonicalOutcome(t *testing.T) {
	build := func() *Session {
		engine := gameecs.New()
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "test.rollback", Version: 1,
			Fields: []akara.Field{{Name: "value", Kind: akara.FieldInt64}}})
		if err != nil {
			t.Fatal(err)
		}
		entity := engine.World().MustCreateEntity()
		if _, err := store.Set(entity, nil); err != nil {
			t.Fatal(err)
		}
		result, err := New(engine, Config{CheckpointInterval: 1, RollbackWindow: 8})
		if err != nil {
			t.Fatal(err)
		}
		if err := result.Register("add", CommandHandler{Validate: func(simulation.Command) error { return nil }, Apply: func(target *gameecs.Engine, command simulation.Command) error {
			components, _ := akara.GetDynamicStore(target.World(), "test.rollback")
			component, _ := components.Get(entity)
			value, _ := component.Get("value")
			var payload struct {
				Add int64 `json:"add"`
			}
			if err := json.Unmarshal(command.Payload, &payload); err != nil {
				return err
			}
			return component.Set("value", value.(int64)*10+payload.Add)
		}}); err != nil {
			t.Fatal(err)
		}
		return result
	}
	command := func(tick, sequence uint64, add int) simulation.Command {
		return simulation.Command{Tick: tick, Player: "player", Sequence: sequence, Kind: "add", Payload: json.RawMessage(fmt.Sprintf(`{"add":%d}`, add))}
	}
	canonical := build()
	defer canonical.Close()
	if err := canonical.Submit(command(1, 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := canonical.Step(); err != nil {
		t.Fatal(err)
	}
	if err := canonical.Step(); err != nil {
		t.Fatal(err)
	}
	if err := canonical.Submit(command(3, 2, 2)); err != nil {
		t.Fatal(err)
	}
	if err := canonical.Step(); err != nil {
		t.Fatal(err)
	}

	late := build()
	defer late.Close()
	if err := late.Step(); err != nil {
		t.Fatal(err)
	}
	if err := late.Step(); err != nil {
		t.Fatal(err)
	}
	if _, err := late.SubmitNetwork(command(1, 1, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := late.SubmitNetwork(command(3, 2, 2)); err != nil {
		t.Fatal(err)
	}
	if err := late.Step(); err != nil {
		t.Fatal(err)
	}
	left, _ := canonical.engine.Snapshot()
	right, _ := late.engine.Snapshot()
	leftChecksum, _ := left.Checksum()
	rightChecksum, _ := right.Checksum()
	if leftChecksum != rightChecksum || late.ProcessedSequence("player") != 2 {
		t.Fatalf("rollback outcome checksum=%s want=%s ack=%d", rightChecksum, leftChecksum, late.ProcessedSequence("player"))
	}
	if _, err := late.SubmitNetwork(command(3, 2, 9)); !errors.Is(err, ErrCommandSequence) {
		t.Fatalf("duplicate sequence error = %v", err)
	}
}

func TestSessionRecordsOnlyExecutedPrivilegedCommandsInAudit(t *testing.T) {
	engine := gameecs.New()
	session, err := New(engine, Config{MaxCommandLead: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.Register("admin.spawn_item", CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
		Allowed:  []simulation.Authority{simulation.AuthorityAdmin, simulation.AuthoritySystem},
	}); err != nil {
		t.Fatal(err)
	}
	command := simulation.Command{Tick: 1, Player: "operator-1", Authority: simulation.AuthorityAdmin, Sequence: 1, Kind: "admin.spawn_item", Payload: json.RawMessage(`{"code":"rin"}`)}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if len(session.Audit()) != 0 {
		t.Fatal("queued command appeared in executed audit")
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	audit := session.Audit()
	if len(audit) != 1 || audit[0].Player != "operator-1" || audit[0].Authority != simulation.AuthorityAdmin {
		t.Fatalf("audit = %#v", audit)
	}
	audit[0].Payload[0] = '['
	if string(session.Audit()[0].Payload) != `{"code":"rin"}` {
		t.Fatal("audit payload was not defensively copied")
	}
}

func TestSubmitNextDerivesAndQueuesTickAtomically(t *testing.T) {
	engine := gameecs.New()
	session, err := New(engine, Config{MaxCommandLead: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.Register("system.enter", CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
		Allowed:  []simulation.Authority{simulation.AuthoritySystem},
	}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	var builtTick uint64
	if err := session.SubmitNext(func(tick uint64) (simulation.Command, error) {
		builtTick = tick
		return simulation.Command{Tick: tick, Player: "system", Authority: simulation.AuthoritySystem,
			Sequence: 1, Kind: "system.enter", Payload: json.RawMessage(`{}`)}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if builtTick != 2 {
		t.Fatalf("next tick = %d, want 2", builtTick)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Commands) != 1 || replay.Commands[0].Tick != 2 {
		t.Fatalf("executed commands = %#v", replay.Commands)
	}
	if err := session.SubmitNext(nil); !errors.Is(err, ErrHandler) {
		t.Fatalf("nil builder error = %v, want ErrHandler", err)
	}
}
