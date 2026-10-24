package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy"
)

// The owned maps exercise Dark Magic's registration, not retail collision parity.
// Moving a real resident near its spawn is test setup; production Lua chooses its chase velocity.
func TestColdPlainsMovementUsesResidentLevelOfflineAndListen(t *testing.T) {
	f := newColdPopulationFixture(t)
	if err := f.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	f.advanceGame(t, 10)
	seam := f.app.transitionSeams[1]
	moveAcrossColdPlainsSeam(t, f, 2, 3, seam.First, seam.Second)
	f.advanceGame(t, 2)

	runtime := &listenHostRuntime{}
	t.Cleanup(runtime.close)

	if err := f.app.network.startListenAuthority(t.Context(), runtime); err != nil {
		t.Fatal(err)
	}

	if err := f.app.network.installListenWorlds(t.Context(), runtime.host); err != nil {
		t.Fatal(err)
	}
	// Replay the admitted offline inputs at exactly the same completed ticks. No host goroutines run.
	replay, err := f.app.offlineSession.Replay()
	if err != nil {
		t.Fatal(err)
	}

	bootstrap, err := f.app.populationBootstrapCommand()
	if err != nil {
		t.Fatal(err)
	}

	checkpoint, err := f.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	for tick := uint64(1); tick <= checkpoint.Tick; tick++ {
		for _, command := range replay.Commands {
			if command.Tick != tick || command.Kind == bootstrap.Kind {
				continue
			}

			if err := runtime.host.Session.Submit(command); err != nil {
				t.Fatal(err)
			}
		}

		if _, err := runtime.host.Session.Advance(time.Second / 25); err != nil {
			t.Fatal(err)
		}
	}

	hostCheckpoint, err := runtime.host.Session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	offlinePopulation := publicColdPopulation(coldPopulationWorldView(t, "local-player", checkpoint))

	hostedPopulation := publicColdPopulation(coldPopulationWorldView(t, "local-player", hostCheckpoint))
	if !reflect.DeepEqual(offlinePopulation, hostedPopulation) {
		t.Fatal("matched admitted inputs produced different offline/listen population")
	}

	client := connectColdMovementClient(t, f, runtime)

	ids := populationResidentIDs(t, f, 3, false)

	names := make([]string, 0, len(ids))
	for name := range ids {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, allowed := range []bool{true, false} {
		t.Run(map[bool]string{true: "open", false: "blocked"}[allowed], func(t *testing.T) {
			candidate := findColdMovementStep(t, f, names, ids, allowed)
			offline := observeColdMovement(t, f.app.entitySimulation, candidate)

			hosted := observeColdMovement(t, runtime.host.Engine, candidate)
			if _, err := f.app.offlineSession.Advance(time.Second / 25); err != nil {
				t.Fatal(err)
			}

			if _, err := runtime.host.Session.Advance(time.Second / 25); err != nil {
				t.Fatal(err)
			}

			offline.check(t, f.app.entitySimulation, f.app.gameWorlds, candidate)
			hosted.check(t, runtime.host.Engine, f.app.gameWorlds, candidate)

			if _, err := client.Refresh(t.Context()); err != nil {
				t.Fatal(err)
			}

			checkpoint, err := runtime.host.Session.CanonicalCheckpoint()
			if err != nil {
				t.Fatal(err)
			}

			view := client.PresentationSnapshot()
			if view.HUD.Tick != checkpoint.Tick || view.World.Tick != checkpoint.Tick ||
				view.Events.Tick != checkpoint.Tick {
				t.Fatal("reliable connected projection does not acknowledge the completed movement tick")
			}

			hostedPopulation := publicColdPopulation(coldPopulationWorldView(t, "local-player", checkpoint))
			if !reflect.DeepEqual(publicColdPopulation(view.World), hostedPopulation) {
				t.Fatal("connected monster positions differ from acknowledged authority")
			}

			if !reflect.DeepEqual(*offline, *hosted) {
				t.Fatalf("offline/listen intents differ: %+v vs %+v", offline, hosted)
			}
		})
	}
}

type coldMovementStep struct {
	monster, player    akara.Entity
	start, playerPoint gameworld.Point
	radius             float64
	allowed            bool
}

// Search residents in stable name order so map discrimination remains reproducible across runs.
func findColdMovementStep(
	t *testing.T, f *realD2LegacyFixture, names []string, ids map[string]akara.Entity, allowed bool,
) coldMovementStep {
	t.Helper()
	positions := requireRealStore(t, f, "d2legacy.world.position", "positions absent")
	colliders := requireRealStore(t, f, "d2legacy.world.collider", "colliders absent")
	brains := requireRealStore(t, f, "d2legacy.monster.ai", "AI absent")
	player := controlledColdPlainsPlayer(t, f)
	playerRadius := coldMovementField[float64](t, colliders, player, "radius")

	for _, name := range names {
		entity := ids[name]
		if _, ok := brains.Get(entity); !ok {
			continue
		}

		speed := coldMovementField[float64](t, brains, entity, "speed")
		radius := coldMovementField[float64](t, colliders, entity, "radius")
		origin := coldMovementPoint(t, positions, entity)
		search := coldMovementSearch{
			f.app.gameWorlds[3], f.app.gameWorlds[2], origin, speed, radius, playerRadius, allowed,
		}

		start, target, found := search.find()
		if !found {
			continue
		}

		t.Logf("resident %s radius %.2f: start %+v target %+v, level 3 allowed=%v, level 2=%v",
			name, radius, start, target, allowed, !allowed)

		return coldMovementStep{entity, player, start, target, radius, allowed}
	}

	t.Fatalf("no reachable discriminating Cold Plains chase step (level 3 allowed=%v)", allowed)

	return coldMovementStep{}
}

type coldMovementSearch struct {
	cold, blood                 *gameworld.Map
	origin                      gameworld.Point
	speed, radius, playerRadius float64
	allowed                     bool
}

// Search square rings in distance/y/x order, preserving the original first matching candidate.
func (s coldMovementSearch) find() (gameworld.Point, gameworld.Point, bool) {
	for distance := 0; distance <= 40; distance++ {
		for dy := -distance; dy <= distance; dy++ {
			for dx := -distance; dx <= distance; dx++ {
				if max(absCold(dx), absCold(dy)) != distance {
					continue
				}

				center := gameworld.Point{
					X: math.Round(s.origin.X) + float64(dx), Y: math.Round(s.origin.Y) + float64(dy),
				}
				if start, target, found := s.atCell(center); found {
					return start, target, true
				}
			}
		}
	}

	return gameworld.Point{}, gameworld.Point{}, false
}

// Try cardinal directions in their established order and retain only reachable opposing-map steps.
func (s coldMovementSearch) atCell(center gameworld.Point) (gameworld.Point, gameworld.Point, bool) {
	for _, direction := range []gameworld.Point{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		start := gameworld.Point{X: center.X + direction.X*.49, Y: center.Y + direction.Y*.49}
		goal := gameworld.Point{X: start.X + direction.X*s.speed*.04, Y: start.Y + direction.Y*s.speed*.04}

		target := gameworld.Point{X: start.X + direction.X*5, Y: start.Y + direction.Y*5}
		if !s.cold.WalkableStep(start, start, s.radius) || !s.cold.WalkableStep(target, target, s.playerRadius) {
			continue
		}

		if s.cold.WalkableStep(start, goal, s.radius) != s.allowed ||
			s.blood.WalkableStep(start, goal, s.radius) == s.allowed {
			continue
		}

		request := gameworld.PathRequest{Start: s.origin, Goal: start, Radius: s.radius}
		if _, err := s.cold.FindPath(request); err != nil {
			continue
		}

		return start, target, true
	}

	return gameworld.Point{}, gameworld.Point{}, false
}

// Absolute offsets identify ring edges without changing candidate visitation order.
func absCold(v int) int {
	if v < 0 {
		return -v
	}

	return v
}

// Missing stores fail at the fixture boundary instead of panicking during a movement assertion.
func coldMovementStore(t *testing.T, engine *gameecs.Engine, name string) *akara.DynamicStore {
	t.Helper()

	store, ok := akara.GetDynamicStore(engine.World(), name)
	if !ok {
		t.Fatalf("missing movement store %s", name)
	}

	return store
}

// Check entity membership before reading fields so broken fixtures identify the missing resident.
func coldMovementComponent(t *testing.T, store *akara.DynamicStore, entity akara.Entity) *akara.DynamicComponent {
	t.Helper()

	component, ok := store.Get(entity)
	if !ok {
		t.Fatalf("missing movement component for entity %d", entity)
	}

	return component
}

// Type checking keeps schema failures diagnostic instead of hiding them behind assertion panics.
func coldMovementField[T any](t *testing.T, store *akara.DynamicStore, entity akara.Entity, field string) T {
	t.Helper()
	component := coldMovementComponent(t, store, entity)

	value, err := component.Get(field)
	if err != nil {
		t.Fatalf("entity %d field %s: %v", entity, field, err)
	}

	typed, ok := value.(T)
	if !ok {
		t.Fatalf("entity %d field %s has unexpected type %T", entity, field, value)
	}

	return typed
}

// Validate both vector fields so missing position or velocity data cannot masquerade as a zero vector.
func coldMovementPoint(t *testing.T, store *akara.DynamicStore, entity akara.Entity) gameworld.Point {
	t.Helper()

	return gameworld.Point{
		X: coldMovementField[float64](t, store, entity, "x"),
		Y: coldMovementField[float64](t, store, entity, "y"),
	}
}

type coldMovementObservation struct {
	state, target string
	velocity      gameworld.Point
	tick          uint64
}

// Observe immediately after real AI and before native movement; never manufacture velocity.
func observeColdMovement(t *testing.T, engine *gameecs.Engine, step coldMovementStep) *coldMovementObservation {
	t.Helper()

	positions := coldMovementStore(t, engine, "d2legacy.world.position")
	brains := coldMovementStore(t, engine, "d2legacy.monster.ai")

	points := map[akara.Entity]gameworld.Point{step.monster: step.start, step.player: step.playerPoint}
	for entity, point := range points {
		p, ok := positions.Get(entity)
		if !ok {
			t.Fatal("matched entity absent")
		}

		if err := p.Set("x", point.X); err != nil {
			t.Fatal(err)
		}

		if err := p.Set("y", point.Y); err != nil {
			t.Fatal(err)
		}
	}
	// Force a think on this tick while preserving the resident's production speed, range, and target policy.
	brain := coldMovementComponent(t, brains, step.monster)
	if err := brain.Set("next_think_tick", int64(0)); err != nil {
		t.Fatal(err)
	}

	observation := &coldMovementObservation{}

	velocities := coldMovementStore(t, engine, "d2legacy.world.velocity")
	if err := engine.Register(gameecs.Definition{
		ID:    map[bool]string{true: "test.cold_movement.open", false: "test.cold_movement.blocked"}[step.allowed],
		Phase: gameecs.PhaseIntent,
		After: []string{"d2legacy.monster.basic_ai"},
		Read:  []akara.ComponentType{brains, velocities},
		Update: func(context gameecs.Context, _ []akara.Entity, _ *gameecs.StructuralCommands) error {
			observation.state = coldMovementField[string](t, brains, step.monster, "state")
			observation.target = coldMovementField[string](t, brains, step.monster, "target_id")
			observation.velocity = coldMovementPoint(t, velocities, step.monster)
			observation.tick = context.Tick

			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	return observation
}

// Check the real intent against both maps before asserting movement, so a vacuous chase cannot pass.
func (o *coldMovementObservation) check(
	t *testing.T, engine *gameecs.Engine, worlds map[int]*gameworld.Map, step coldMovementStep,
) {
	t.Helper()
	locations := coldMovementStore(t, engine, "d2legacy.world.location")

	level := coldMovementField[int64](t, locations, step.monster, "level_id")
	if level != int64(3) {
		t.Fatalf("spawned resident level = %v, want 3", level)
	}

	selected := coldMovementStore(t, engine, "d2legacy.world.selectable")

	target := coldMovementField[string](t, selected, step.player, "id")
	if o.state != "chase" || o.target != target || math.Hypot(o.velocity.X, o.velocity.Y) == 0 {
		t.Fatalf("production AI did not chase player: %+v, want %v", o, target)
	}

	goal := gameworld.Point{X: step.start.X + o.velocity.X*.04, Y: step.start.Y + o.velocity.Y*.04}
	if gameworld.CollisionCell(goal.X) == gameworld.CollisionCell(step.start.X) &&
		gameworld.CollisionCell(goal.Y) == gameworld.CollisionCell(step.start.Y) {
		t.Fatal("chase did not cross a collision boundary")
	}

	if worlds[3].WalkableStep(step.start, goal, step.radius) != step.allowed ||
		worlds[2].WalkableStep(step.start, goal, step.radius) == step.allowed {
		t.Fatalf("actual AI velocity did not attempt opposing map step: %+v -> %+v radius %v",
			step.start, goal, step.radius)
	}

	positions := coldMovementStore(t, engine, "d2legacy.world.position")
	position := coldMovementPoint(t, positions, step.monster)

	want := step.start
	if step.allowed {
		want = goal
	}

	if math.Hypot(position.X-want.X, position.Y-want.Y) > 1e-9 {
		t.Fatalf("resident moved to (%v,%v), want %+v (own level allowed=%v)",
			position.X, position.Y, want, step.allowed)
	}

	if !step.allowed {
		velocities := coldMovementStore(t, engine, "d2legacy.world.velocity")

		velocity := coldMovementPoint(t, velocities, step.monster)
		if velocity.X != 0 || velocity.Y != 0 {
			t.Fatalf("blocked velocity not consumed: %v,%v", velocity.X, velocity.Y)
		}
	}
}

// Serve only transport: the test owns all authority ticks and inspects reliable completed corrections.
func connectColdMovementClient(
	t *testing.T, f *realD2LegacyFixture, runtime *listenHostRuntime,
) *clientsession.Session {
	t.Helper()

	var err error

	f.app.networkTrust, err = networktrust.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if err := f.app.network.bindListenTransport(runtime); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	go func() { _ = runtime.server.Serve(ctx) }()

	character, ok := f.app.saves.Selected()
	if !ok {
		t.Fatal("admitted character absent")
	}

	principal := gameserver.Principal{ID: listenPrincipalID, CharacterID: character.ID, PlayerID: "local-player"}

	ticket, err := runtime.tickets.Issue(principal, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	client, err := clientsession.Connect(ctx, realm.JoinAssignment{
		GameID: listenGameID, Ticket: ticket, Runtime: runtime.host.Authority.Identity,
		Endpoint: realm.GameEndpoint{Address: listenDialAddress, TLSFingerprint: runtime.fingerprint},
	}, runtime.clientTLS)
	if err != nil {
		t.Fatal(err)
	}

	runtime.client = client

	return client
}

// Keep real ECS/participant data intact while changing only the omitted native-version identity.
func assertColdRecoveryRejectsOldMovement(t *testing.T, f *realD2LegacyFixture) {
	t.Helper()

	before, err := f.app.offlineSession.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	legacy, err := f.app.offlineSession.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	data := f.app.sessionInitialData()
	delete(data, "engine.world.velocity_movement_version")

	source, err := f.app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}

	old, err := d2legacy.Identity(source, data)
	if err != nil {
		t.Fatal(err)
	}

	identity, err := simulation.RuntimeIdentityFromParticipants(legacy.State.Participants)
	if err != nil {
		t.Fatal(err)
	}

	if identity.Recipe.ConfigurationHash == old.Recipe.ConfigurationHash {
		t.Fatal("missing native version kept current identity")
	}

	identity.Recipe.ConfigurationHash = old.Recipe.ConfigurationHash

	encoded, err := json.Marshal(identity)
	if err != nil {
		t.Fatal(err)
	}

	for i := range legacy.State.Participants {
		if legacy.State.Participants[i].ID == (&simulation.IdentityParticipant{}).StateID() {
			legacy.State.Participants[i].Data = encoded
		}
	}

	legacy.State.Checksum, err = simulation.CompositeChecksum(*legacy.State.Snapshot, legacy.State.Participants)
	if err != nil {
		t.Fatal(err)
	}

	legacy, err = gamesession.NewRecoveryCheckpoint(legacy.State, legacy.Pending, legacy.Accepted, legacy.Sequences)
	if err != nil {
		t.Fatal(err)
	}

	restoreErr := f.app.offlineSession.RestoreRecoveryCheckpoint(legacy)
	if !errors.Is(restoreErr, simulation.ErrAuthoritativeState) {
		t.Fatalf("old movement recovery returned %v", restoreErr)
	}

	after, err := f.app.offlineSession.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(before, after) {
		t.Fatal("rejected native-version checkpoint changed owned authority")
	}
}
