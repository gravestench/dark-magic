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
	if !reflect.DeepEqual(publicColdPopulation(coldPopulationWorldView(t, "local-player", checkpoint)), publicColdPopulation(coldPopulationWorldView(t, "local-player", hostCheckpoint))) {
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
			if view.HUD.Tick != checkpoint.Tick || view.World.Tick != checkpoint.Tick || view.Events.Tick != checkpoint.Tick {
				t.Fatal("reliable connected projection does not acknowledge the completed movement tick")
			}
			if !reflect.DeepEqual(publicColdPopulation(view.World), publicColdPopulation(coldPopulationWorldView(t, "local-player", checkpoint))) {
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

// Search deterministically near a real resident, requiring a reachable start and a walkable player footprint.
func findColdMovementStep(t *testing.T, f *realD2LegacyFixture, names []string, ids map[string]akara.Entity, allowed bool) coldMovementStep {
	t.Helper()
	positions := requireRealStore(t, f, "d2legacy.world.position", "positions absent")
	colliders := requireRealStore(t, f, "d2legacy.world.collider", "colliders absent")
	brains := requireRealStore(t, f, "d2legacy.monster.ai", "AI absent")
	player := controlledColdPlainsPlayer(t, f)
	playerCollider, _ := colliders.Get(player)
	playerRadius, _ := playerCollider.Get("radius")
	for _, name := range names {
		entity := ids[name]
		brain, ok := brains.Get(entity)
		if !ok {
			continue
		}
		speed, _ := brain.Get("speed")
		collider, _ := colliders.Get(entity)
		radius, _ := collider.Get("radius")
		x, y := dynamicPosition(positions, entity)
		for distance := 0; distance <= 40; distance++ {
			for dy := -distance; dy <= distance; dy++ {
				for dx := -distance; dx <= distance; dx++ {
					if max(absCold(dx), absCold(dy)) != distance {
						continue
					}
					for _, direction := range []gameworld.Point{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
						start := gameworld.Point{X: math.Round(x) + float64(dx) + direction.X*.49, Y: math.Round(y) + float64(dy) + direction.Y*.49}
						goal := gameworld.Point{X: start.X + direction.X*speed.(float64)*.04, Y: start.Y + direction.Y*speed.(float64)*.04}
						target := gameworld.Point{X: start.X + direction.X*5, Y: start.Y + direction.Y*5}
						cold, blood := f.app.gameWorlds[3], f.app.gameWorlds[2]
						r := radius.(float64)
						if !cold.WalkableStep(start, start, r) || cold.WalkableStep(start, goal, r) != allowed || blood.WalkableStep(start, goal, r) == allowed || !cold.WalkableStep(target, target, playerRadius.(float64)) {
							continue
						}
						if _, err := cold.FindPath(gameworld.PathRequest{Start: gameworld.Point{X: x, Y: y}, Goal: start, Radius: r}); err != nil {
							continue
						}
						t.Logf("resident %s radius %.2f: start %+v target %+v, level 3 allowed=%v, level 2=%v", name, r, start, target, allowed, !allowed)
						return coldMovementStep{entity, player, start, target, r, allowed}
					}
				}
			}
		}
	}
	t.Fatalf("no reachable discriminating Cold Plains chase step (level 3 allowed=%v)", allowed)
	return coldMovementStep{}
}
func absCold(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

type coldMovementObservation struct {
	state, target string
	velocity      gameworld.Point
	tick          uint64
}

// Observe immediately after real AI and before native movement; never manufacture velocity.
func observeColdMovement(t *testing.T, engine *gameecs.Engine, step coldMovementStep) *coldMovementObservation {
	t.Helper()
	store := func(name string) *akara.DynamicStore {
		s, ok := akara.GetDynamicStore(engine.World(), name)
		if !ok {
			t.Fatalf("missing %s", name)
		}
		return s
	}
	positions, brains := store("d2legacy.world.position"), store("d2legacy.monster.ai")
	for entity, point := range map[akara.Entity]gameworld.Point{step.monster: step.start, step.player: step.playerPoint} {
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
	brain, _ := brains.Get(step.monster)
	if err := brain.Set("next_think_tick", int64(0)); err != nil {
		t.Fatal(err)
	}
	observation := &coldMovementObservation{}
	velocities := store("d2legacy.world.velocity")
	if err := engine.Register(gameecs.Definition{
		ID:    map[bool]string{true: "test.cold_movement.open", false: "test.cold_movement.blocked"}[step.allowed],
		Phase: gameecs.PhaseIntent,
		After: []string{"d2legacy.monster.basic_ai"},
		Read:  []akara.ComponentType{brains, velocities},
		Update: func(context gameecs.Context, _ []akara.Entity, _ *gameecs.StructuralCommands) error {
			b, _ := brains.Get(step.monster)
			s, _ := b.Get("state")
			target, _ := b.Get("target_id")
			vx, vy := dynamicPosition(velocities, step.monster)
			observation.state, observation.target = s.(string), target.(string)
			observation.velocity = gameworld.Point{X: vx, Y: vy}
			observation.tick = context.Tick
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	return observation
}

func (o *coldMovementObservation) check(t *testing.T, engine *gameecs.Engine, worlds map[int]*gameworld.Map, step coldMovementStep) {
	t.Helper()
	locations, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.location")
	location, present := locations.Get(step.monster)
	if !present {
		t.Fatal("spawned resident lost authoritative location")
	}
	level, _ := location.Get("level_id")
	if level != int64(3) {
		t.Fatalf("spawned resident level = %v, want 3", level)
	}
	selected, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	player, _ := selected.Get(step.player)
	target, _ := player.Get("id")
	if o.state != "chase" || o.target != target || math.Hypot(o.velocity.X, o.velocity.Y) == 0 {
		t.Fatalf("production AI did not chase player: %+v, want %v", o, target)
	}
	goal := gameworld.Point{X: step.start.X + o.velocity.X*.04, Y: step.start.Y + o.velocity.Y*.04}
	if gameworld.CollisionCell(goal.X) == gameworld.CollisionCell(step.start.X) && gameworld.CollisionCell(goal.Y) == gameworld.CollisionCell(step.start.Y) {
		t.Fatal("chase did not cross a collision boundary")
	}
	if worlds[3].WalkableStep(step.start, goal, step.radius) != step.allowed || worlds[2].WalkableStep(step.start, goal, step.radius) == step.allowed {
		t.Fatalf("actual AI velocity did not attempt opposing map step: %+v -> %+v radius %v", step.start, goal, step.radius)
	}
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	x, y := dynamicPosition(positions, step.monster)
	want := step.start
	if step.allowed {
		want = goal
	}
	if math.Hypot(x-want.X, y-want.Y) > 1e-9 {
		t.Fatalf("resident moved to (%v,%v), want %+v (own level allowed=%v)", x, y, want, step.allowed)
	}
	if !step.allowed {
		velocities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.velocity")
		vx, vy := dynamicPosition(velocities, step.monster)
		if vx != 0 || vy != 0 {
			t.Fatalf("blocked velocity not consumed: %v,%v", vx, vy)
		}
	}
}

// Serve only transport: the test owns all authority ticks and inspects reliable completed corrections.
func connectColdMovementClient(t *testing.T, f *realD2LegacyFixture, runtime *listenHostRuntime) *clientsession.Session {
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
	ticket, err := runtime.tickets.Issue(gameserver.Principal{ID: listenPrincipalID, CharacterID: character.ID, PlayerID: "local-player"}, time.Minute)
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
	if err := f.app.offlineSession.RestoreRecoveryCheckpoint(legacy); !errors.Is(err, simulation.ErrAuthoritativeState) {
		t.Fatalf("old movement recovery returned %v", err)
	}
	after, err := f.app.offlineSession.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("rejected native-version checkpoint changed owned authority")
	}
}
