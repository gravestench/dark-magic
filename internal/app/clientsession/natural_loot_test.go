package clientsession

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	worldobjects "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	recovered "github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
)

// Owned records, ordinary commands and the real transport share one authority.
func TestConnectedNaturalLootPickup(t *testing.T) {
	for _, seed := range []uint64{10, 8} {
		t.Run(fmt.Sprint(seed), func(t *testing.T) { connectedNaturalLoot(t, seed) })
	}
}

func connectedNaturalLoot(t *testing.T, seed uint64) {
	t.Helper()

	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to owned Diablo II assets")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	source, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = source.Close() }()

	records := recordstore.New(source)

	data, err := recovered.New(source).Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	objects, err := worldobjects.New(data, records)
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := entryworld.Build(ctx, source, content.D2Legacy(), records, objects, 1, 0)
	if err != nil {
		t.Fatal(err)
	}

	assetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	host,
		err := gameserver.Start(ctx,
		content.D2Legacy(),
		records,
		gameserver.Config{Mode: gameserver.ModeStandalone,
			SessionID:  "live-gameworld",
			Seed:       seed,
			Content:    source,
			AssetSetID: assetID,
			InitialData: prepared.InitialData("",
				false),
			Prediction: gamesession.PredictionLimited,
			Session: gamesession.Config{Step: 40 * time.Millisecond,
				CheckpointInterval: 1,
				RollbackWindow:     64}})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = host.Close(context.Background()) })

	if err := prepared.InstallCollision(ctx, host.Authority.Runtime); err != nil {
		t.Fatal(err)
	}

	population, err := prepared.PopulationCommand(0)
	if err != nil {
		t.Fatal(err)
	}

	if err := host.Session.Submit(population); err != nil {
		t.Fatal(err)
	}

	destination, err := prepared.Destination(2)
	if err != nil {
		t.Fatal(err)
	}

	town, err := prepared.Destination(1)
	if err != nil {
		t.Fatal(err)
	}

	fixture := serveLiveGameworld(t, ctx, host, destination)
	defer func() {
		fixture.stopServe()
		fixture.stopRun()
		assertCanceledLoop(t, ctx, fixture.runErrors, "session")
		assertCanceledLoop(t, ctx, fixture.serveErrors, "QUIC")
	}()

	character := d2save.Character{ID: "hero",
		Name:      "Hero",
		Class:     "Amazon",
		Level:     1,
		Expansion: true,
		Stats: &d2save.Stats{Dexterity: 20,
			Vitality:   20,
			Health:     50,
			MaxHealth:  50,
			Mana:       20,
			MaxMana:    20,
			Stamina:    84,
			MaxStamina: 84}}
	alice := connectLiveCharacter(t, ctx, fixture, character)
	actor := newLootClient(t, ctx, alice, prepared.Worlds[2])

	var target playeradapter.WorldEntity

	best := math.Inf(1)

	for _, entity := range alice.PresentationSnapshot().World.Entities {
		if entity.Kind != "hostile" || entity.Health == nil || *entity.Health <= 0 {
			continue
		}

		position := alice.PresentationSnapshot().HUD.Position

		path,
			err := prepared.Worlds[2].FindPath(gameworld.PathRequest{Start: gameworld.Point{X: position.X,
			Y: position.Y},
			Goal: gameworld.Point{X: entity.Position.X,
				Y: entity.Position.Y},
			Radius:     1,
			StopRadius: 3.5})
		if err == nil && float64(len(path)) < best {
			target, best = entity, float64(len(path))
		}
	}

	if target.ID == "" {
		t.Fatal("no naturally activated reachable public monster")
	}

	monster := naturalMonsterID(t, host, target.ID)
	actor.walk(gameworld.Point{X: target.Position.X, Y: target.Position.Y}, 3.5)

	var itemID, code, targetID string

	for attempt := 0; attempt < 100; attempt++ {
		actor.refresh()

		itemID, code, targetID = naturalDeathItem(t, host, monster)
		if itemID != "" {
			break
		}

		for _, entity := range alice.PresentationSnapshot().World.Entities {
			if entity.ID == target.ID {
				target = entity
			}
		}

		actor.submit("player.use_skill",
			map[string]any{"side": "left",
				"target_x":  target.Position.X,
				"target_y":  target.Position.Y,
				"target_id": target.ID})
		actor.wait(200 * time.Millisecond)
	}

	if itemID == "" {
		t.Fatal("ordinary connected attack did not commit a supported drop")
	}

	checkpoint, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	projected, err := playeradapter.ProjectWorldView(alice.PresentationSnapshot().HUD.Player.PlayerID, checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	var direct playeradapter.WorldView
	if err := json.Unmarshal(projected, &direct); err != nil {
		t.Fatal(err)
	}

	directVisible := false

	for _, entity := range direct.Entities {
		if entity.ID == targetID {
			directVisible = true
		}
	}

	t.Logf("death observation canonical tick=%d client tick=%d exact item public=%v",
		direct.Tick,
		alice.PresentationSnapshot().World.Tick,
		directVisible)

	if !directVisible {
		t.Fatal("exact generated item missing from same-owner canonical public projection")
	}

	drop := actor.item(targetID)
	if drop.WorldDC6 == "" {
		t.Fatal("connected public item has no ground art")
	}

	t.Logf("seed=%d monster=%d exact committed item=%s code=%s public art=%s", seed, monster, itemID, code, drop.WorldDC6)

	character.ID, character.Name = "peer", "Peer"
	bob := connectLiveCharacter(t, ctx, fixture, character)
	peer := newLootClient(t, ctx, bob, prepared.Worlds[2])
	character.ID, character.Name = "town", "Town"
	// The third principal uses the existing trusted ticket/admission seam, with a separate system source.
	if err := host.Session.SubmitNext(func(tick uint64) (simulation.Command, error) {
		return playeradapter.AdmissionCommand(character, "carol", town, "town-fixture", 1, tick, simulation.AuthoritySystem)
	}); err != nil {
		t.Fatal(err)
	}

	ticket,
		err := fixture.tickets.Issue(gameserver.Principal{ID: "town-account",
		PlayerID:            "carol",
		CharacterID:         character.ID,
		RuntimeIdentityHash: host.Allocation.IdentityHash},
		time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	carol,
		err := Connect(ctx,
		realm.JoinAssignment{GameID: "live-gameworld",
			Endpoint: fixture.endpoint,
			Runtime:  host.Authority.Identity,
			Ticket:   ticket},
		fixture.clientTLS)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = carol.Close(context.Background()) })
	other := newLootClient(t, ctx, carol, prepared.Worlds[1])

	actor.walk(gameworld.Point{X: drop.Position.X, Y: drop.Position.Y}, 2.5)
	actor.refresh()
	peer.refresh()
	other.refresh()
	actor.item(targetID)
	peer.item(targetID)

	for _, entity := range carol.PresentationSnapshot().World.Entities {
		if entity.ID == targetID {
			t.Fatal("different-level connected client saw item")
		}
	}

	actor.submit("item.pickup", map[string]any{"target_id": targetID})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		actor.refresh()

		for _, item := range alice.PresentationSnapshot().Private.Items.Items {
			if item.ID == itemID && item.Code == code && item.Container == "inventory" {
				goto acquired
			}
		}

		actor.wait(20 * time.Millisecond)
	}

	t.Fatal("connected admitted pickup did not acquire exact generated item")

acquired:
	actor.waitAbsent(targetID)

	peer.waitAbsent(targetID)
	other.waitAbsent(targetID)

	for _, client := range []*Session{alice, bob, carol} {
		for _, entity := range client.PresentationSnapshot().World.Entities {
			if entity.ID == targetID {
				t.Fatal("picked item remained public")
			}
		}
	}

	for _, client := range []*Session{bob, carol} {
		for _, item := range client.PresentationSnapshot().Private.Items.Items {
			if item.ID == itemID {
				t.Fatal("item leaked into another private inventory")
			}
		}
	}

	actor.submit("item.pickup", map[string]any{"target_id": targetID})
	actor.wait(120 * time.Millisecond)
	actor.refresh()

	count := 0

	for _, item := range alice.PresentationSnapshot().Private.Items.Items {
		if item.ID == itemID {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("repeat connected pickup produced %d copies", count)
	}
}

type lootClient struct {
	t           *testing.T
	ctx         context.Context
	session     *Session
	world       *gameworld.Map
	sequence    uint64
	watchErrors <-chan error
}

func newLootClient(t *testing.T, ctx context.Context, session *Session, world *gameworld.Map) *lootClient {
	_, errors, err := session.Watch(ctx)
	if err != nil {
		t.Fatal(err)
	}

	client := &lootClient{t: t, ctx: ctx, session: session, world: world, watchErrors: errors}
	client.wait(80 * time.Millisecond)

	return client
}
func (c *lootClient) refresh() {
	c.t.Helper()

	select {
	case err, open := <-c.watchErrors:
		if open && err != nil {
			c.t.Fatal(err)
		}
	default:
	}
}

func (c *lootClient) wait(duration time.Duration) {
	c.t.Helper()

	select {
	case <-time.After(duration):
	case <-c.ctx.Done():
		c.t.Fatal(c.ctx.Err())
	}
}
func (c *lootClient) submit(kind string, payload any) {
	c.t.Helper()
	c.refresh()
	c.sequence++

	encoded, err := json.Marshal(payload)
	if err != nil {
		c.t.Fatal(err)
	}

	if err := c.session.Submit(c.ctx,
		gameserver.CommandIntent{TargetTick: c.session.NextInputTick(time.Now()),
			Sequence: c.sequence,
			Kind:     kind,
			Payload:  encoded}); err != nil {
		c.t.Fatal(err)
	}
}
func (c *lootClient) item(id string) playeradapter.WorldEntity {
	c.t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c.refresh()

		for _, entity := range c.session.PresentationSnapshot().World.Entities {
			if entity.ID == id && entity.Kind == "item" {
				return entity
			}
		}

		c.wait(20 * time.Millisecond)
	}

	c.t.Fatalf("connected item %s missing at observed tick %d", id, c.session.PresentationSnapshot().World.Tick)

	return playeradapter.WorldEntity{}
}

func (c *lootClient) waitAbsent(id string) {
	c.t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c.refresh()

		found := false

		for _, entity := range c.session.PresentationSnapshot().World.Entities {
			if entity.ID == id {
				found = true
			}
		}

		if !found {
			return
		}

		c.wait(20 * time.Millisecond)
	}

	c.t.Fatalf("connected item %s remained public at tick %d", id, c.session.PresentationSnapshot().World.Tick)
}

func (c *lootClient) walk(goal gameworld.Point, stop float64) {
	c.t.Helper()
	c.refresh()
	start := c.session.PresentationSnapshot().HUD.Position

	path,
		err := c.world.FindPath(gameworld.PathRequest{Start: gameworld.Point{X: start.X,
		Y: start.Y},
		Goal:       goal,
		Radius:     1,
		StopRadius: stop})
	if err != nil {
		c.t.Fatal(err)
	}

	for _, point := range path {
		current := c.session.PresentationSnapshot().HUD.Position
		if math.Hypot(current.X-goal.X, current.Y-goal.Y) <= stop {
			break
		}

		c.submit("player.move",
			map[string]any{"x": 0,
				"y":       0,
				"running": false,
				"target": map[string]any{"x": point.X,
					"y": point.Y}})

		reached := false

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			c.wait(20 * time.Millisecond)
			c.refresh()

			current = c.session.PresentationSnapshot().HUD.Position
			if math.Hypot(current.X-point.X, current.Y-point.Y) < 0.6 || math.Hypot(current.X-goal.X, current.Y-goal.Y) <= stop {
				reached = true
				break
			}
		}

		if !reached {
			c.t.Fatalf("connected movement stalled for %s at %v toward %v final goal %v",
				c.session.PresentationSnapshot().HUD.Player.PlayerID,
				current,
				point,
				goal)
		}
	}

	current := c.session.PresentationSnapshot().HUD.Position
	c.submit("player.move",
		map[string]any{"x": 0,
			"y":       0,
			"running": false,
			"target": map[string]any{"x": current.X,
				"y": current.Y}})
	c.wait(80 * time.Millisecond)
}

func lootSnapshot(t *testing.T, host *gameserver.Host) *gameecs.Engine {
	t.Helper()

	checkpoint, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	engine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}

	return engine
}
func lootStore(t *testing.T, engine *gameecs.Engine, name string) *akara.DynamicStore {
	t.Helper()

	store, ok := akara.GetDynamicStore(engine.World(), name)
	if !ok {
		t.Fatalf("missing store %s", name)
	}

	return store
}
func naturalMonsterID(t *testing.T, host *gameserver.Host, target string) akara.Entity {
	t.Helper()

	engine := lootSnapshot(t, host)
	defer func() { _ = engine.Close() }()

	store := lootStore(t, engine, "d2legacy.world.selectable")
	for _, id := range store.Entities() {
		component, _ := store.Get(id)

		value, _ := component.Get("id")
		if value == target {
			return id
		}
	}

	t.Fatal("chosen public target has no canonical identity")

	return 0
}
func naturalDeathItem(t *testing.T, host *gameserver.Host, monster akara.Entity) (string, string, string) {
	t.Helper()

	engine := lootSnapshot(t, host)
	defer func() { _ = engine.Close() }()

	death, ok := lootStore(t, engine, "d2legacy.monster.death").Get(monster)
	if !ok {
		return "", "", ""
	}

	raw, _ := death.Get("drops")

	var drops []map[string]any
	if err := json.Unmarshal([]byte(raw.(string)), &drops); err != nil {
		t.Fatal(err)
	}

	if len(drops) != 1 {
		t.Fatalf("natural connected death needs one supported drop, got %v", drops)
	}

	tick, _ := death.Get("tick")
	id := fmt.Sprintf("loot:%v:%d:1", tick, monster)

	for _, entity := range lootStore(t, engine, "d2legacy.item.identity").Entities() {
		identity, _ := lootStore(t, engine, "d2legacy.item.identity").Get(entity)

		value, _ := identity.Get("id")
		if value == id {
			code, _ := identity.Get("code")
			if code != drops[0]["code"] {
				t.Fatal("materialized item differs from committed drop")
			}

			selected, ok := lootStore(t, engine, "d2legacy.world.selectable").Get(entity)
			if !ok {
				t.Fatal("generated item has no public selector")
			}

			selector, _ := selected.Get("id")

			return id, code.(string), selector.(string)
		}
	}

	t.Fatal("committed death lacks exact generated item")

	return "", "", ""
}
