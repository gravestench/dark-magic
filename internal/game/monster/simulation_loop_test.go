package monster

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameloot "github.com/gravestench/dark-magic/internal/game/loot"
	gamemissile "github.com/gravestench/dark-magic/internal/game/missile"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
	gamestate "github.com/gravestench/dark-magic/internal/game/state"
)

const acceptanceSkillID int64 = 42
const acceptanceActionCommand = "test.acceptance.action"

// This is the first complete authoritative gameplay spine. The policies are
// intentionally synthetic; the test proves owner boundaries and replay, not
// that temporary M21 numbers reproduce retail Diablo II combat formulas.
func TestGeneratedBloodMoorSimulationLoopRestoresExactly(t *testing.T) {
	zone := populationZone(t, 42)
	records := populationSnapshot()
	stats := records.MonstersByID["fallen"]
	stats.TreasureClass1 = "fallen-drop"
	records.MonstersByID["fallen"] = stats
	plan, err := BuildBloodMoorPopulation(zone, openPlacement{}, records)
	if err != nil || len(plan.Spawns) == 0 {
		t.Fatalf("population plan: spawns=%d err=%v", len(plan.Spawns), err)
	}
	plan.Spawns = plan.Spawns[:1]
	policy := DeathPolicy{WorldSeed: zone.Request().Seed, Loot: gameloot.Catalog{
		"fallen-drop": {Name: "fallen-drop", Picks: 1, Entries: []gameloot.Entry{{Code: "gld", Weight: 1}}},
	}}
	skills, missiles := acceptanceRegistries(t)
	engine := gameecs.New()
	defer engine.Close()
	registerAcceptanceSystems(t, engine, policy, skills, missiles)
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	if err := gameplayer.Register(session); err != nil {
		t.Fatal(err)
	}
	actionApply := func(target *gameecs.Engine, command simulation.Command) error {
		return applyAcceptanceAction(target)
	}
	if err := session.Register(acceptanceActionCommand, gamesession.CommandHandler{Validate: func(simulation.Command) error { return nil }, Apply: actionApply, Allowed: []simulation.Authority{simulation.AuthoritySystem}}); err != nil {
		t.Fatal(err)
	}
	if err := SubmitPopulation(session, plan, "population", 1); err != nil {
		t.Fatal(err)
	}
	entry := gameplayer.Entry{
		CharacterID: "acceptance-hero", Player: "hero", Name: "Hero", Class: "Amazon",
		Level: 1, Health: 100, MaxHealth: 100, Mana: 10, MaxMana: 10,
		Token: "AM", Palette: "data/global/palette/units/pal.dat", Direction: 0, Mode: "NU", WeaponClass: "HTH",
		MeleeRange: 2, PhysicalMinRaw: gamecombat.MustWhole(1).Raw(), PhysicalMaxRaw: gamecombat.MustWhole(2).Raw(),
		X: plan.Spawns[0].X - 0.5, Y: plan.Spawns[0].Y, WorldWidth: 200, WorldHeight: 200, Act: 1, LevelID: 2,
		Skills: []gameplayer.Skill{{ID: acceptanceSkillID, Level: 1, ListRow: 0, RightAllowed: true}},
	}
	entryCommand, err := gameplayer.Command(entry, "entry", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(entryCommand); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	monster := firstEntity(t, engine, "dm.monster.identity")
	player := firstEntity(t, engine, "dm.player.identity")
	if err := session.Submit(simulation.Command{Tick: 2, Player: "acceptance", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: acceptanceActionCommand, Payload: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	midpoint, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	beforeObservation, _ := engine.Snapshot()
	beforeObservationChecksum, _ := beforeObservation.Checksum()
	assertAcceptanceConsequences(t, engine, monster, player)
	afterObservation, _ := engine.Snapshot()
	afterObservationChecksum, _ := afterObservation.Checksum()
	if afterObservationChecksum != beforeObservationChecksum {
		t.Fatal("semantic event observation mutated authoritative state")
	}
	want, _ := engine.Snapshot()
	wantChecksum, _ := want.Checksum()

	restored, err := gameecs.RestoreSnapshot(midpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	registerAcceptanceSystems(t, restored, policy, skills, missiles)
	if err := restored.Update(gameecs.DefaultStep); err != nil {
		t.Fatal(err)
	}
	got, _ := restored.Snapshot()
	gotChecksum, _ := got.Checksum()
	if gotChecksum != wantChecksum {
		t.Fatalf("restored loop checksum = %s, want %s", gotChecksum, wantChecksum)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	prepare := func(target *gameecs.Engine) error {
		for _, register := range acceptanceSystemRegistrations(target, policy, skills, missiles) {
			if err := register(); err != nil {
				return err
			}
		}
		return nil
	}
	apply := func(target *gameecs.Engine, command simulation.Command) error {
		switch command.Kind {
		case SpawnCommand:
			return ApplySpawnCommand(target, command)
		case gameplayer.EnterCommand:
			return gameplayer.ApplyEntryCommand(target, command)
		case acceptanceActionCommand:
			return applyAcceptanceAction(target)
		default:
			return nil
		}
	}
	if err := simulation.VerifyReplay(replay, prepare, apply); err != nil {
		t.Fatal(err)
	}
}

func acceptanceRegistries(t *testing.T) (gameskill.Registry, gamemissile.Registry) {
	t.Helper()
	skills, err := gameskill.NewRegistry(gameskill.Definition{SkillID: acceptanceSkillID, Behavior: gameskill.BehaviorStraightMissile, TargetPolicy: gameskill.TargetPoint, EffectDelay: 1, CompleteDelay: 2})
	if err != nil {
		t.Fatal(err)
	}
	missiles, err := gamemissile.NewRegistry(gamemissile.Definition{SkillID: acceptanceSkillID, SpeedPerTick: 1, MaxRange: 4, LifetimeTicks: 8, CollisionRadius: 0.25, PhysicalDamage: gamecombat.MustWhole(1000)})
	if err != nil {
		t.Fatal(err)
	}
	return skills, missiles
}

func registerAcceptanceSystems(t *testing.T, engine *gameecs.Engine, policy DeathPolicy, skills gameskill.Registry, missiles gamemissile.Registry) {
	t.Helper()
	for _, register := range acceptanceSystemRegistrations(engine, policy, skills, missiles) {
		if err := register(); err != nil {
			t.Fatal(err)
		}
	}
}

func acceptanceSystemRegistrations(engine *gameecs.Engine, policy DeathPolicy, skills gameskill.Registry, missiles gamemissile.Registry) []func() error {
	return []func() error{
		func() error { return RegisterAI(engine, nil) },
		func() error { return RegisterMovement(engine, nil) },
		func() error { return gameskill.RegisterIntentConsumer(engine) },
		func() error { return gameskill.RegisterCastLifecycle(engine, skills) },
		func() error { return gamemissile.Register(engine, missiles) },
		func() error {
			return gamecombat.RegisterBasicMelee(engine, gamecombat.BasicMeleePolicy{HitChance: 100})
		},
		func() error { return gamestate.Register(engine) },
		func() error { return RegisterDeath(engine, policy) },
	}
}

func applyAcceptanceAction(engine *gameecs.Engine) error {
	world := engine.World()
	monster, found := entityFromStore(world, "dm.monster.identity")
	if !found {
		return nil
	}
	player, found := entityFromStore(world, "dm.player.identity")
	if !found {
		return nil
	}
	positions, _ := akara.GetDynamicStore(world, "dm.world.position")
	intents, _ := akara.GetDynamicStore(world, "dm.player.skill_intent")
	monsterPosition, _ := positions.Get(monster)
	mx, _ := monsterPosition.Get("x")
	my, _ := monsterPosition.Get("y")
	brains, _ := akara.GetDynamicStore(world, AIComponent)
	brain, _ := brains.Get(monster)
	if err := brain.Set("next_think_tick", int64(2)); err != nil {
		return err
	}
	if _, err := intents.Set(player, map[string]any{"side": "right", "skill_id": acceptanceSkillID, "target_x": mx.(float64), "target_y": my.(float64), "target_id": "monster:" + spawnIDValue(world, monster)}); err != nil {
		return err
	}
	_, err := gamestate.Apply(engine, player, "acceptance-focus", "skill:42", 1)
	return err
}

func assertAcceptanceConsequences(t *testing.T, engine *gameecs.Engine, monster, player akara.Entity) {
	t.Helper()
	deaths, _ := akara.GetDynamicStore(engine.World(), DeathTransaction)
	if !deaths.Has(monster) {
		t.Fatal("lethal missile produced no monster death transaction")
	}
	progress, _ := akara.GetDynamicStore(engine.World(), "dm.player.progress")
	component, _ := progress.Get(player)
	experience, _ := component.Get("experience")
	if experience.(int64) <= 0 {
		t.Fatalf("player experience = %v", experience)
	}
	deathEvents, _ := akara.GetDynamicStore(engine.World(), DeathEvent)
	if deathEvents.Len() != 4 {
		t.Fatalf("death semantic events = %d", deathEvents.Len())
	}
	stateEvents, _ := akara.GetDynamicStore(engine.World(), gamestate.EventComponent)
	if stateEvents.Len() < 2 {
		t.Fatalf("timed-state events = %d", stateEvents.Len())
	}
	missileEvents, _ := akara.GetDynamicStore(engine.World(), gamemissile.EventComponent)
	if missileEvents.Len() < 2 {
		t.Fatalf("missile events = %d", missileEvents.Len())
	}
	combatEvents, _ := akara.GetDynamicStore(engine.World(), gamecombat.CombatEvent)
	if combatEvents.Len() < 3 {
		t.Fatalf("combat events = %d", combatEvents.Len())
	}
}

func firstEntity(t *testing.T, engine *gameecs.Engine, name string) akara.Entity {
	t.Helper()
	store, found := akara.GetDynamicStore(engine.World(), name)
	if !found || store.Len() == 0 {
		t.Fatalf("store %q has no entities", name)
	}
	return store.Entities()[0]
}

func spawnIDValue(world *akara.World, entity akara.Entity) string {
	store, _ := akara.GetDynamicStore(world, "dm.monster.identity")
	component, _ := store.Get(entity)
	value, _ := component.Get("spawn_id")
	return value.(string)
}

func entityFromStore(world *akara.World, name string) (akara.Entity, bool) {
	store, found := akara.GetDynamicStore(world, name)
	if !found || store.Len() == 0 {
		return 0, false
	}
	return store.Entities()[0], true
}

func setComponent(t *testing.T, store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	t.Helper()
	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}
