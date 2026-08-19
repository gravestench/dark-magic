package clientapp

import (
	"fmt"

	"github.com/gravestench/akara"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// reconcileSemanticEvents mirrors only reliable events newer than the retained cursor. Epoch changes
// establish a baseline instead of replaying history, so joining players do not see old transient
// cues as if they just happened.
func (client *clientWorld) reconcileSemanticEvents(
	app *application,
	view playeradapter.EventView,
	epoch uint64,
) error {
	if client == nil || app == nil || app.clientSimulation == nil || view.Version == 0 {
		return nil
	}

	world := app.clientSimulation.World()
	if epoch != client.lastEventEpoch {
		client.baselineSemanticEvents(world, view, epoch)

		return nil
	}

	if view.Tick <= client.lastEventViewTick {
		return nil
	}

	if err := client.validateSemanticEventWindow(view); err != nil {
		return err
	}

	client.destroySemanticEvents(world)

	if err := client.appendSemanticEvents(world, view.Events); err != nil {
		client.destroySemanticEvents(world)

		return err
	}

	client.lastEventViewTick = view.Tick

	return nil
}

// baselineSemanticEvents clears disposable cues and advances directly to the new epoch's latest
// event. Event history is durable for recovery, but presentation effects are not.
func (client *clientWorld) baselineSemanticEvents(
	world *akara.World,
	view playeradapter.EventView,
	epoch uint64,
) {
	client.destroySemanticEvents(world)
	client.lastEventEpoch = epoch
	client.lastEventViewTick = view.Tick
	client.eventCursorTick, client.eventCursorID = latestSemanticCursor(view.Events)
}

// validateSemanticEventWindow fails closed on gaps or truncation. Silently advancing the cursor
// would permanently lose reliable cues and leave client presentation inconsistent with authority.
func (client *clientWorld) validateSemanticEventWindow(view playeradapter.EventView) error {
	if view.FromTick > client.lastEventViewTick+1 {
		return fmt.Errorf(
			"remote presentation: semantic event window gap after tick %d (starts at %d)",
			client.lastEventViewTick,
			view.FromTick,
		)
	}

	if view.Truncated {
		return fmt.Errorf(
			"remote presentation: semantic event window truncated at tick %d",
			view.Tick,
		)
	}

	return nil
}

// destroySemanticEvents retires the previous correction's disposable ECS entities before installing
// new cues; semantic events describe occurrences, not persistent gameplay components.
func (client *clientWorld) destroySemanticEvents(world *akara.World) {
	for _, entity := range client.semanticEventEntities {
		world.DestroyEntity(entity)
	}

	client.semanticEventEntities = nil
}

// appendSemanticEvents advances the cursor only after each event is installed successfully. A
// failed projection therefore cannot acknowledge events the client never presented.
func (client *clientWorld) appendSemanticEvents(
	world *akara.World,
	events []playeradapter.SemanticEvent,
) error {
	nextTick := client.eventCursorTick
	nextID := client.eventCursorID

	for _, event := range events {
		if semanticEventSeen(event, nextTick, nextID) {
			continue
		}

		entity, err := installSemanticEvent(world, event)
		if err != nil {
			return err
		}

		client.semanticEventEntities = append(client.semanticEventEntities, entity)
		nextTick = event.Tick
		nextID = event.ID
	}

	client.eventCursorTick = nextTick
	client.eventCursorID = nextID

	return nil
}

// semanticEventSeen compares the ordered tick/ID pair, allowing several events on one simulation
// tick without duplicates during overlapping reliable windows.
func semanticEventSeen(
	event playeradapter.SemanticEvent,
	cursorTick uint64,
	cursorID uint64,
) bool {
	return event.Tick < cursorTick || event.Tick == cursorTick && event.ID <= cursorID
}

// latestSemanticCursor relies on the authority's ordered event window and returns zero for an empty
// baseline, preserving the first real event.
func latestSemanticCursor(events []playeradapter.SemanticEvent) (uint64, uint64) {
	if len(events) == 0 {
		return 0, 0
	}

	last := events[len(events)-1]

	return last.Tick, last.ID
}

// installSemanticEvent builds anchor and allowlisted payload as one disposable entity. Any partial
// failure destroys the entity so systems never observe an incomplete cue.
func installSemanticEvent(
	world *akara.World,
	event playeradapter.SemanticEvent,
) (akara.Entity, error) {
	entity, err := world.CreateEntity()
	if err != nil {
		return 0, fmt.Errorf("remote presentation: create semantic event: %w", err)
	}

	if err := setSemanticAnchor(world, entity, event); err != nil {
		world.DestroyEntity(entity)

		return 0, err
	}

	if err := setSemanticPayload(world, entity, event); err != nil {
		world.DestroyEntity(entity)

		return 0, err
	}

	return entity, nil
}

// setSemanticAnchor gives every cue the shared spatial facts needed by overlays, sounds, and
// animation systems. Missing schemas are fatal because a non-spatial cue could render misleadingly.
func setSemanticAnchor(
	world *akara.World,
	entity akara.Entity,
	event playeradapter.SemanticEvent,
) error {
	updates := map[string]map[string]any{
		"d2legacy.world.position": {
			"x": event.Position.X,
			"y": event.Position.Y,
		},
		"d2legacy.world.location": {
			"act":      event.Act,
			"level_id": event.LevelID,
		},
		"d2legacy.world.facing": {
			"direction":  event.Direction,
			"directions": int64(16),
		},
		"d2legacy.presentation.overlay_anchor": {
			"height": event.OverlayHeight,
		},
	}

	for name, values := range updates {
		store, found := akara.GetDynamicStore(world, name)
		if !found {
			return fmt.Errorf("remote presentation: semantic component %q is unavailable", name)
		}

		if _, err := store.Set(entity, values); err != nil {
			return fmt.Errorf("remote presentation: set semantic %s: %w", name, err)
		}
	}

	return nil
}

// setSemanticPayload accepts only explicitly supported presentation event types. Unknown server
// data fails rather than being copied generically into the client ECS.
func setSemanticPayload(
	world *akara.World,
	entity akara.Entity,
	event playeradapter.SemanticEvent,
) error {
	var err error

	switch event.Type {
	case "cast":
		err = setSemanticCast(world, entity, event)
	case "state":
		err = setSemanticState(world, entity, event)
	case "effect":
		err = setSemanticEffect(world, entity, event)
	case "monster_death":
		err = setSemanticMonsterDeath(world, entity, event)
	default:
		return fmt.Errorf("remote presentation: unsupported semantic event %q", event.Type)
	}

	if err != nil {
		return fmt.Errorf("remote presentation: set %s semantic event: %w", event.Type, err)
	}

	return nil
}

// setSemanticCast projects timing and visual targeting while referring to the disposable cue as the
// caster entity. It carries no authority to execute or resolve the skill locally.
func setSemanticCast(
	world *akara.World,
	entity akara.Entity,
	event playeradapter.SemanticEvent,
) error {
	store, found := akara.GetDynamicStore(world, "d2legacy.skill.cast_cue")
	if !found || event.Cast == nil {
		return fmt.Errorf("remote presentation: cast cue component is unavailable")
	}

	cue := event.Cast
	_, err := store.Set(entity, map[string]any{
		"kind":        cue.Kind,
		"tick":        int64(event.Tick),
		"effect_tick": int64(cue.EffectTick),
		"caster":      entity,
		"player":      cue.Player,
		"skill_id":    cue.SkillID,
		"target_x":    cue.Target.X,
		"target_y":    cue.Target.Y,
		"target_id":   cue.TargetID,
	})

	return err
}

// setSemanticState exposes lifecycle identifiers and timing for visual systems, not state magnitude
// or gameplay effects; authority remains responsible for applying the state.
func setSemanticState(
	world *akara.World,
	entity akara.Entity,
	event playeradapter.SemanticEvent,
) error {
	store, found := akara.GetDynamicStore(world, "d2legacy.state.event")
	if !found || event.State == nil {
		return fmt.Errorf("remote presentation: state event component is unavailable")
	}

	cue := event.State
	_, err := store.Set(entity, map[string]any{
		"kind":         cue.Kind,
		"tick":         int64(event.Tick),
		"target":       entity,
		"state_id":     cue.StateID,
		"source_id":    cue.SourceID,
		"expires_tick": int64(cue.ExpiresTick),
		"reason":       cue.Reason,
	})

	return err
}

// setSemanticEffect projects only the overlay and sound identifiers required to render an effect.
// It deliberately omits the gameplay cause and outcome.
func setSemanticEffect(
	world *akara.World,
	entity akara.Entity,
	event playeradapter.SemanticEvent,
) error {
	store, found := akara.GetDynamicStore(world, "d2legacy.presentation.effect_cue")
	if !found || event.Effect == nil {
		return fmt.Errorf("remote presentation: effect cue component is unavailable")
	}

	cue := event.Effect
	_, err := store.Set(entity, map[string]any{
		"kind":       cue.Kind,
		"tick":       int64(event.Tick),
		"target":     entity,
		"overlay_id": cue.OverlayID,
		"sound":      cue.Sound,
	})

	return err
}

// setSemanticMonsterDeath preserves the public death cue while zeroing private reward and loot
// fields required by the registered schema. Presentation cannot infer unrevealed outcomes.
func setSemanticMonsterDeath(
	world *akara.World,
	entity akara.Entity,
	event playeradapter.SemanticEvent,
) error {
	store, found := akara.GetDynamicStore(world, "d2legacy.monster.death_event")
	if !found || event.MonsterDeath == nil {
		return fmt.Errorf("remote presentation: monster death component is unavailable")
	}

	cue := event.MonsterDeath
	_, err := store.Set(entity, map[string]any{
		"kind":                      cue.Kind,
		"tick":                      int64(event.Tick),
		"monster_id":                cue.MonsterID,
		"killer_id":                 "",
		"credited_id":               "",
		"xp":                        int64(0),
		"loot_seed":                 "",
		"treasure_class":            "",
		"drops":                     "",
		"game_player_count":         int64(0),
		"effective_player_count":    int64(0),
		"nearby_party_member_count": int64(0),
		"monster_player_count":      int64(0),
		"no_drop_player_count":      int64(0),
	})

	return err
}

// reconcileMissiles treats each correction as the complete reliable projectile set. Duplicate IDs
// are rejected because accepting them would make update order determine the rendered projectile.
func (client *clientWorld) reconcileMissiles(
	app *application,
	projected []playeradapter.WorldMissile,
) error {
	if client == nil || app == nil || app.clientSimulation == nil {
		return nil
	}

	if client.missileEntities == nil {
		client.missileEntities = make(map[string]akara.Entity)
	}

	world := app.clientSimulation.World()
	seen := make(map[string]struct{}, len(projected))

	for _, missile := range projected {
		if _, duplicate := seen[missile.ID]; duplicate {
			return fmt.Errorf("remote presentation: duplicate missile %q", missile.ID)
		}

		seen[missile.ID] = struct{}{}

		if err := client.upsertMissile(world, missile); err != nil {
			return err
		}
	}

	client.removeStaleMissiles(world, seen)

	return nil
}

// upsertMissile retains entity identity across corrections for smooth rendering but replaces its
// allowlisted presentation components from authority on every update.
func (client *clientWorld) upsertMissile(
	world *akara.World,
	missile playeradapter.WorldMissile,
) error {
	entity, found := client.missileEntities[missile.ID]
	if !found {
		var err error

		entity, err = world.CreateEntity()
		if err != nil {
			return fmt.Errorf("remote presentation: create missile %q: %w", missile.ID, err)
		}

		client.missileEntities[missile.ID] = entity
	}

	for name, values := range missileComponents(missile) {
		store, available := akara.GetDynamicStore(world, name)
		if !available {
			return fmt.Errorf("remote presentation: missile component %q is unavailable", name)
		}

		if _, err := store.Set(entity, values); err != nil {
			return fmt.Errorf("remote presentation: set missile %s: %w", name, err)
		}
	}

	return nil
}

// missileComponents maps visual transform and animation data only. Damage, collision, ownership,
// and hit resolution remain absent so client systems cannot turn the mirror into authority.
func missileComponents(missile playeradapter.WorldMissile) map[string]map[string]any {
	return map[string]map[string]any{
		"d2legacy.world.position": {
			"x": missile.Position.X,
			"y": missile.Position.Y,
		},
		"d2legacy.world.location": {
			"act":      missile.Act,
			"level_id": missile.LevelID,
		},
		"d2legacy.presentation.missile": {
			"missile_id":        missile.MissileID,
			"dcc":               missile.DCC,
			"palette":           missile.Palette,
			"velocity_x":        missile.Velocity.X,
			"velocity_y":        missile.Velocity.Y,
			"logical_direction": missile.LogicalDirection,
			"directions":        missile.Directions,
			"frames_per_second": missile.FramesPerSecond,
			"loop":              missile.Loop,
			"transparency_mode": missile.TransparencyMode,
			"offset_x":          missile.OffsetX,
			"offset_y":          missile.OffsetY,
			"offset_z":          missile.OffsetZ,
		},
	}
}

// removeStaleMissiles removes any retained entity absent from the complete correction. Waiting for a
// client-side lifetime would disagree with authority after hits, despawns, or recovery.
func (client *clientWorld) removeStaleMissiles(
	world *akara.World,
	seen map[string]struct{},
) {
	for id, entity := range client.missileEntities {
		if _, retained := seen[id]; retained {
			continue
		}

		world.DestroyEntity(entity)
		delete(client.missileEntities, id)
	}
}

// reconcilePersistentStates binds public visual states to currently retained target mirrors. States
// whose targets are hidden or absent are not materialized, avoiding dangling ECS references.
func (client *clientWorld) reconcilePersistentStates(
	app *application,
	projected []playeradapter.WorldState,
) error {
	if client == nil || app == nil || app.clientSimulation == nil {
		return nil
	}

	world := app.clientSimulation.World()

	targets, states, err := persistentStateStores(world)
	if err != nil {
		return err
	}

	if client.stateEntities == nil {
		client.stateEntities = make(map[presentationStateKey]akara.Entity)
	}

	seen := make(map[presentationStateKey]struct{}, len(projected))

	for _, projectedState := range projected {
		key := presentationStateKey{
			targetID: projectedState.TargetID,
			stateID:  projectedState.StateID,
		}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf(
				"remote presentation: duplicate persistent state target=%q state=%q",
				key.targetID,
				key.stateID,
			)
		}

		seen[key] = struct{}{}

		target, available := targets[projectedState.TargetID]
		if !available {
			continue
		}

		if err := client.upsertPersistentState(world, states, key, target, projectedState); err != nil {
			return err
		}
	}

	client.removeStalePersistentStates(world, targets, seen)

	return nil
}

// persistentStateStores creates one target namespace from public selectable IDs and authenticated
// player IDs. The player prefix prevents accidental collisions between the two identity domains.
func persistentStateStores(
	world *akara.World,
) (map[string]akara.Entity, *akara.DynamicStore, error) {
	selectables, found := akara.GetDynamicStore(world, "d2legacy.world.selectable")
	if !found {
		return nil, nil, fmt.Errorf("remote presentation: selectable component is unavailable")
	}

	states, found := akara.GetDynamicStore(world, "d2legacy.presentation.state")
	if !found {
		return nil, nil, fmt.Errorf("remote presentation: persistent state component is unavailable")
	}

	targets := make(map[string]akara.Entity, selectables.Len())
	indexSelectableTargets(selectables, targets)

	if identities, available := akara.GetDynamicStore(world, "d2legacy.player.identity"); available {
		indexPlayerTargets(identities, targets)
	}

	return targets, states, nil
}

// indexSelectableTargets indexes only non-empty public IDs, because an empty identifier is not a
// stable attachment point for a persistent visual state.
func indexSelectableTargets(
	selectables *akara.DynamicStore,
	targets map[string]akara.Entity,
) {
	for _, entity := range selectables.Entities() {
		selectable, found := selectables.Get(entity)
		if !found {
			continue
		}

		id, _ := selectable.Get("id")
		if value, valid := id.(string); valid && value != "" {
			targets[value] = entity
		}
	}
}

// indexPlayerTargets prefixes player IDs to preserve their domain when combined with selectable
// targets in one lookup table.
func indexPlayerTargets(
	identities *akara.DynamicStore,
	targets map[string]akara.Entity,
) {
	for _, entity := range identities.Entities() {
		identity, found := identities.Get(entity)
		if !found {
			continue
		}

		player, _ := identity.Get("player")
		if value, valid := player.(string); valid && value != "" {
			targets["player:"+value] = entity
		}
	}
}

// upsertPersistentState retains one entity per target/state pair and writes only when its binding or
// visual period changes. Stable entities avoid restarting aura presentation every correction.
func (client *clientWorld) upsertPersistentState(
	world *akara.World,
	states *akara.DynamicStore,
	key presentationStateKey,
	target akara.Entity,
	projected playeradapter.WorldState,
) error {
	entity, retained := client.stateEntities[key]
	if !retained {
		var err error

		entity, err = world.CreateEntity()
		if err != nil {
			return fmt.Errorf(
				"remote presentation: create persistent state target=%q state=%q: %w",
				key.targetID,
				key.stateID,
				err,
			)
		}

		client.stateEntities[key] = entity
	}

	if !persistentStateChanged(states, entity, target, projected) {
		return nil
	}

	_, err := states.Set(entity, map[string]any{
		"target":       target,
		"state_id":     projected.StateID,
		"period_ticks": projected.PeriodTicks,
	})
	if err != nil {
		return fmt.Errorf(
			"remote presentation: set persistent state target=%q state=%q: %w",
			key.targetID,
			key.stateID,
			err,
		)
	}

	return nil
}

// persistentStateChanged compares all projected fields, including the resolved target entity; a
// mirror replacement must rebind even when public target and state IDs stayed the same.
func persistentStateChanged(
	states *akara.DynamicStore,
	entity akara.Entity,
	target akara.Entity,
	projected playeradapter.WorldState,
) bool {
	component, present := states.Get(entity)
	if !present {
		return true
	}

	currentTarget, _ := component.Get("target")
	currentState, _ := component.Get("state_id")
	currentPeriod, _ := component.Get("period_ticks")

	return currentTarget != target ||
		currentState != projected.StateID ||
		currentPeriod != projected.PeriodTicks
}

// removeStalePersistentStates requires both state projection and target mirror to remain present.
// Destroying either side prevents presentation systems from retaining invalid entity references.
func (client *clientWorld) removeStalePersistentStates(
	world *akara.World,
	targets map[string]akara.Entity,
	seen map[presentationStateKey]struct{},
) {
	for key, entity := range client.stateEntities {
		_, retained := seen[key]
		_, targetRetained := targets[key.targetID]

		if retained && targetRetained {
			continue
		}

		world.DestroyEntity(entity)
		delete(client.stateEntities, key)
	}
}
