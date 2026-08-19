package clientapp

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestConnectedPersistentStatesBindToMirrorsWithoutGameplayFacts proves aura presentation binds to
// retained targets while effect magnitude and other authority-only state remain absent.
func TestConnectedPersistentStatesBindToMirrorsWithoutGameplayFacts(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	location := playeradapter.HUDLocation{Act: 1, LevelID: 2}

	peer := playeradapter.WorldEntity{
		ID: "player:peer", Kind: "player", Label: "Peer", Owner: "peer",
		Position: playeradapter.HUDPosition{X: 10, Y: 20}, Class: "Paladin", Token: "PA", Mode: "NU",
	}
	if err := app.syncRemoteMirrors([]playeradapter.WorldEntity{peer}, location); err != nil {
		t.Fatal(err)
	}

	client := newClientWorld()

	projected := []playeradapter.WorldState{{TargetID: peer.ID, StateID: "might", PeriodTicks: 50}}
	if err := client.reconcilePersistentStates(app, projected); err != nil {
		t.Fatal(err)
	}

	states, _ := akara.GetDynamicStore(engine.World(), "d2legacy.presentation.state")
	if states.Len() != 1 || len(client.stateEntities) != 1 {
		t.Fatalf("persistent states = %d/%d", states.Len(), len(client.stateEntities))
	}

	var entity akara.Entity
	for _, value := range client.stateEntities {
		entity = value
	}

	state, _ := states.Get(entity)
	target, _ := state.Get("target")
	stateID, _ := state.Get("state_id")

	period, _ := state.Get("period_ticks")
	if target != app.remoteMirrors[peer.ID] || stateID != "might" || period != int64(50) {
		t.Fatalf("persistent state target=%v state=%v period=%v", target, stateID, period)
	}

	projected[0].PeriodTicks = 25
	if err := client.reconcilePersistentStates(app, projected); err != nil {
		t.Fatal(err)
	}

	updated, _ := states.Get(entity)

	updatedPeriod, _ := updated.Get("period_ticks")
	if updatedPeriod != int64(25) {
		t.Fatal("persistent state period did not update in place")
	}

	if err := client.reconcilePersistentStates(app, nil); err != nil {
		t.Fatal(err)
	}

	if states.Len() != 0 || engine.World().EntityExists(entity) {
		t.Fatal("removed persistent state retained a disposable client entity")
	}
}

// TestConnectedMissilesUsePresentationOnlyECSLifecycle requires complete reliable views to create,
// update, and retire projectile mirrors without installing damage or collision authority.
func TestConnectedMissilesUsePresentationOnlyECSLifecycle(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	client := newClientWorld()

	fireball := playeradapter.WorldMissile{
		ID: "missile:42", Kind: "projectile", MissileID: "fireball",
		Position: playeradapter.HUDPosition{X: 10, Y: 20}, Velocity: playeradapter.HUDPosition{X: 1, Y: 0},
		Act: 1, LevelID: 2, DCC: "data/global/missiles/Fireball.dcc",
		LogicalDirection: -1, Directions: 16, FramesPerSecond: 16, Loop: true, TransparencyMode: 1,
	}
	if err := client.reconcileMissiles(app, []playeradapter.WorldMissile{fireball}); err != nil {
		t.Fatal(err)
	}

	entity := client.missileEntities[fireball.ID]

	visuals, ok := akara.GetDynamicStore(engine.World(), "d2legacy.presentation.missile")
	if !ok {
		t.Fatal("presentation missile store is unavailable")
	}

	visual, found := visuals.Get(entity)
	if !found {
		t.Fatal("connected missile has no presentation component")
	}

	dcc, _ := visual.Get("dcc")

	blend, _ := visual.Get("transparency_mode")
	if dcc != fireball.DCC || blend != int64(1) || currentPosition(engine.World(), entity) != fireball.Position {
		t.Fatalf("connected fireball dcc=%v blend=%v position=%+v", dcc, blend, currentPosition(engine.World(), entity))
	}

	if _, authorityPresent := akara.GetDynamicStore(engine.World(), "d2legacy.missile.projectile"); authorityPresent {
		t.Fatal("test schema unexpectedly gives connected presentation projectile authority")
	}

	updated := fireball

	updated.Position = playeradapter.HUDPosition{X: 11, Y: 20}
	if err := client.reconcileMissiles(app, []playeradapter.WorldMissile{updated}); err != nil {
		t.Fatal(err)
	}

	if client.missileEntities[fireball.ID] != entity || currentPosition(engine.World(), entity) != updated.Position {
		t.Fatalf("missile identity/position was not updated in place")
	}

	if err := client.reconcileMissiles(app, nil); err != nil {
		t.Fatal(err)
	}

	if len(client.missileEntities) != 0 || engine.World().EntityExists(entity) {
		t.Fatal("retired connected missile remains in the disposable ECS")
	}
}

// TestConnectedMonsterBecomesANonSelectableCorpseAndDeathCue proves death changes public structure,
// removes living collision, and emits a cue without leaking reward or loot facts.
func TestConnectedMonsterBecomesANonSelectableCorpseAndDeathCue(t *testing.T) {
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })

	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	client := newClientWorld()
	location := playeradapter.HUDLocation{Act: 1, LevelID: 2}
	alive := connectedLivingMonster()

	if err := app.syncRemoteMirrors([]playeradapter.WorldEntity{alive}, location); err != nil {
		t.Fatal(err)
	}

	entity := app.remoteMirrors[alive.ID]
	assertLivingRemoteMonster(t, engine, entity)

	dead := alive
	health := int64(0)
	dead.Kind, dead.Mode = "corpse", "DT"
	dead.Health, dead.Radius = &health, 0

	if err := app.syncRemoteMirrors([]playeradapter.WorldEntity{dead}, location); err != nil {
		t.Fatal(err)
	}

	assertRemoteMonsterCorpse(t, engine, app, alive.ID, entity)
	assertRemoteMonsterDeathCue(t, engine, client, app, dead)
}

// connectedLivingMonster supplies both collision facts and public presentation metadata for the corpse transition.
func connectedLivingMonster() playeradapter.WorldEntity {
	health, maximum := int64(8), int64(10)

	return playeradapter.WorldEntity{
		ID: "monster:fallen-a", Kind: "monster", Label: "Fallen", SpawnID: "fallen-a", DefinitionID: "fallen1",
		Position: playeradapter.HUDPosition{X: 10, Y: 20}, Radius: .75, Health: &health, MaxHealth: &maximum,
		Token: "FA", Mode: "A1", WeaponClass: "HTH", Components: "HD=LIT", DeathSound: "fallen_death",
		OverlayHeight: 3, Act: 1, LevelID: 2,
	}
}

// assertLivingRemoteMonster proves the public mirror is selectable and collidable before the death transition.
func assertLivingRemoteMonster(t *testing.T, engine *gameecs.Engine, entity akara.Entity) {
	t.Helper()

	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	if _, found := selectables.Get(entity); !found {
		t.Fatal("living monster mirror is not selectable")
	}

	colliders, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.collider")
	if _, found := colliders.Get(entity); !found {
		t.Fatal("living monster mirror has no collider")
	}

	appearances, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.appearance")
	appearance, _ := appearances.Get(entity)
	mode, _ := appearance.Get("mode")
	if mode != "A1" {
		t.Fatalf("living monster presentation mode=%v, want A1", mode)
	}
}

// assertRemoteMonsterCorpse checks the structural implications of death without consulting private authority state.
func assertRemoteMonsterCorpse(
	t *testing.T,
	engine *gameecs.Engine,
	app *application,
	publicID string,
	entity akara.Entity,
) {
	t.Helper()

	if app.remoteMirrors[publicID] != entity {
		t.Fatal("living monster and corpse did not retain one presentation identity")
	}

	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	deadSelectable, found := selectables.Get(entity)
	if !found {
		t.Fatal("corpse lost the selectable needed by corpse-target skills")
	}
	deadKind, _ := deadSelectable.Get("kind")
	if deadKind != "corpse" {
		t.Fatalf("corpse selectable kind=%v, want corpse", deadKind)
	}

	colliders, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.collider")
	if _, found := colliders.Get(entity); found {
		t.Fatal("corpse remained a locomotion obstacle in the disposable client ECS")
	}

	appearances, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.appearance")
	appearance, _ := appearances.Get(entity)
	mode, _ := appearance.Get("mode")
	sound, _ := appearance.Get("death_sound")
	if mode != "DT" || sound != "fallen_death" {
		t.Fatalf("corpse appearance mode=%v death_sound=%v", mode, sound)
	}
}

// assertRemoteMonsterDeathCue proves reliable presentation emits identity but no private reward or loot facts.
func assertRemoteMonsterDeathCue(
	t *testing.T,
	engine *gameecs.Engine,
	client *clientWorld,
	app *application,
	dead playeradapter.WorldEntity,
) {
	t.Helper()

	baseline := playeradapter.EventView{Version: playeradapter.EventViewVersion, Tick: 10, FromTick: 0}
	if err := client.reconcileSemanticEvents(app, baseline, 1); err != nil {
		t.Fatal(err)
	}

	death := playeradapter.SemanticEvent{
		ID: 40, Type: "monster_death", Tick: 11, Position: dead.Position, Act: 1, LevelID: 2,
		MonsterDeath: &playeradapter.SemanticMonsterDeathCue{Kind: "monster_death_presented", MonsterID: "fallen-a"},
	}
	if err := client.reconcileSemanticEvents(app, playeradapter.EventView{
		Version: playeradapter.EventViewVersion, Tick: 11, FromTick: 0, Events: []playeradapter.SemanticEvent{death},
	}, 1); err != nil {
		t.Fatal(err)
	}

	deathEvents, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death_event")
	if deathEvents.Len() != 1 {
		t.Fatalf("connected monster death cues = %d, want 1", deathEvents.Len())
	}

	cue, _ := deathEvents.Get(deathEvents.Entities()[0])
	monsterID, _ := cue.Get("monster_id")

	drops, _ := cue.Get("drops")
	if monsterID != "fallen-a" || drops != "" {
		t.Fatalf("connected death cue monster=%v drops=%v", monsterID, drops)
	}
}

// TestConnectedSemanticEventsBaselineHistoryAndMirrorOnlyNewCues ensures join baselines durable history
// and presents only later reliable cues once, even across overlapping corrections.
func TestConnectedSemanticEventsBaselineHistoryAndMirrorOnlyNewCues(t *testing.T) {
	engine := gameecs.New()

	t.Cleanup(func() { _ = engine.Close() })

	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	client := newClientWorld()
	historical := connectedHistoricalCast()

	baseline := playeradapter.EventView{
		Version:  playeradapter.EventViewVersion,
		Tick:     100,
		FromTick: 37,
		Events:   []playeradapter.SemanticEvent{historical},
	}
	if err := client.reconcileSemanticEvents(app, baseline, 1); err != nil {
		t.Fatal(err)
	}

	casts, _ := akara.GetDynamicStore(engine.World(), "d2legacy.skill.cast_cue")
	if casts.Len() != 0 {
		t.Fatalf("baseline replayed %d historical cues", casts.Len())
	}

	current := connectedCurrentCast()

	view := playeradapter.EventView{
		Version:  playeradapter.EventViewVersion,
		Tick:     101,
		FromTick: 38,
		Events:   []playeradapter.SemanticEvent{historical, current},
	}
	if err := client.reconcileSemanticEvents(app, view, 1); err != nil {
		t.Fatal(err)
	}

	if casts.Len() != 1 {
		t.Fatalf("new connected cues = %d, want 1", casts.Len())
	}

	assertConnectedCastMirror(t, engine, casts, current)

	if err := client.reconcileSemanticEvents(app, view, 1); err != nil {
		t.Fatal(err)
	}

	if casts.Len() != 1 {
		t.Fatalf("same correction duplicated cue count=%d", casts.Len())
	}
}

// connectedHistoricalCast represents durable pre-join history that must establish a cursor without replaying a cue.
func connectedHistoricalCast() playeradapter.SemanticEvent {
	return playeradapter.SemanticEvent{
		ID:       20,
		Type:     "cast",
		Tick:     98,
		Position: playeradapter.HUDPosition{X: 4, Y: 5},
		Act:      1,
		LevelID:  2,
		Cast: &playeradapter.SemanticCastCue{
			Kind:       "cast_started",
			EffectTick: 100,
			Player:     "alice",
			SkillID:    47,
		},
	}
}

// connectedCurrentCast supplies the first post-baseline cue and every allowlisted presentation field it should mirror.
func connectedCurrentCast() playeradapter.SemanticEvent {
	return playeradapter.SemanticEvent{
		ID:            21,
		Type:          "cast",
		Tick:          101,
		Position:      playeradapter.HUDPosition{X: 12, Y: 13},
		Act:           1,
		LevelID:       2,
		Direction:     4,
		OverlayHeight: 3,
		Cast: &playeradapter.SemanticCastCue{
			Kind:       "cast_effect",
			EffectTick: 101,
			Player:     "alice",
			SkillID:    47,
			Target:     playeradapter.HUDPosition{X: 20, Y: 21},
		},
	}
}

// assertConnectedCastMirror checks spatial anchoring and owner mapping without depending on private skill state.
func assertConnectedCastMirror(
	t *testing.T,
	engine *gameecs.Engine,
	casts *akara.DynamicStore,
	current playeradapter.SemanticEvent,
) {
	t.Helper()

	entity := casts.Entities()[0]
	position := currentPosition(engine.World(), entity)
	if position != current.Position {
		t.Fatalf("semantic anchor = %+v, want %+v", position, current.Position)
	}

	anchors, _ := akara.GetDynamicStore(engine.World(), "d2legacy.presentation.overlay_anchor")
	anchor, _ := anchors.Get(entity)
	height, _ := anchor.Get("height")
	if height != int64(3) {
		t.Fatalf("semantic overlay height = %v, want 3", height)
	}

	cast, _ := casts.Get(entity)
	skillID, _ := cast.Get("skill_id")
	caster, _ := cast.Get("caster")
	if skillID != int64(47) || caster != entity {
		t.Fatalf("cast mirror skill=%v caster=%v entity=%v", skillID, caster, entity)
	}
}

// TestConnectedSemanticEffectPreservesRecordOverlayAndSoundOnly protects the semantic allowlist: record
// overlay and sound reach presentation, while gameplay cause and outcome do not.
func TestConnectedSemanticEffectPreservesRecordOverlayAndSoundOnly(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}

	client := newClientWorld()
	if err := client.reconcileSemanticEvents(app, playeradapter.EventView{
		Version: playeradapter.EventViewVersion, Tick: 10, FromTick: 0,
	}, 1); err != nil {
		t.Fatal(err)
	}

	effect := playeradapter.SemanticEvent{
		ID: 50, Type: "effect", Tick: 11, Position: playeradapter.HUDPosition{X: 8, Y: 9}, Act: 1, LevelID: 2,
		Direction: 3, OverlayHeight: 2,
		Effect: &playeradapter.SemanticEffectCue{
			Kind: "state_reaction", OverlayID: "frozenarmor_hit", Sound: "impact_cold_1",
		},
	}
	if err := client.reconcileSemanticEvents(app, playeradapter.EventView{
		Version: playeradapter.EventViewVersion, Tick: 11, FromTick: 0,
		Events: []playeradapter.SemanticEvent{effect},
	}, 1); err != nil {
		t.Fatal(err)
	}

	store, _ := akara.GetDynamicStore(engine.World(), "d2legacy.presentation.effect_cue")
	if store.Len() != 1 {
		t.Fatalf("connected effect cues = %d, want 1", store.Len())
	}

	entity := store.Entities()[0]
	instance, _ := store.Get(entity)
	overlay, _ := instance.Get("overlay_id")
	sound, _ := instance.Get("sound")

	target, _ := instance.Get("target")
	if overlay != "frozenarmor_hit" || sound != "impact_cold_1" || target != entity ||
		currentPosition(engine.World(), entity) != effect.Position {
		t.Fatalf(
			"connected effect overlay=%v sound=%v target=%v position=%+v",
			overlay,
			sound,
			target,
			currentPosition(engine.World(), entity),
		)
	}
}

// TestConnectedSemanticEventsRejectCorrectionGapAndTruncation requires reliable projection to fail
// closed when authority cannot provide a contiguous event window.
func TestConnectedSemanticEventsRejectCorrectionGapAndTruncation(t *testing.T) {
	engine := gameecs.New()
	registerRemoteViewSchemas(t, engine)
	app := &application{clientSimulation: engine}
	client := newClientWorld()

	baseline := playeradapter.EventView{
		Version:  playeradapter.EventViewVersion,
		Tick:     100,
		FromTick: 37,
		Events:   []playeradapter.SemanticEvent{},
	}
	if err := client.reconcileSemanticEvents(app, baseline, 1); err != nil {
		t.Fatal(err)
	}

	gap := playeradapter.EventView{
		Version:  playeradapter.EventViewVersion,
		Tick:     170,
		FromTick: 107,
	}
	if err := client.reconcileSemanticEvents(app, gap, 1); err == nil {
		t.Fatal("semantic correction gap was silently accepted")
	}

	truncated := playeradapter.EventView{
		Version:   playeradapter.EventViewVersion,
		Tick:      101,
		FromTick:  38,
		Truncated: true,
	}
	if err := client.reconcileSemanticEvents(app, truncated, 1); err == nil {
		t.Fatal("truncated semantic correction was silently accepted")
	}
}
