package clientapp

import (
	"fmt"

	"github.com/gravestench/akara"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// reconcileSemanticEvents mirrors only reliable events newer than the cursor.
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

// baselineSemanticEvents advances the cursor without replaying durable history.
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

// validateSemanticEventWindow rejects gaps and server-truncated corrections.
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

// destroySemanticEvents removes transient cues from the previous correction.
func (client *clientWorld) destroySemanticEvents(world *akara.World) {
	for _, entity := range client.semanticEventEntities {
		world.DestroyEntity(entity)
	}

	client.semanticEventEntities = nil
}

// appendSemanticEvents installs events strictly after the existing cursor.
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

// semanticEventSeen reports whether an event is at or before the cursor.
func semanticEventSeen(
	event playeradapter.SemanticEvent,
	cursorTick uint64,
	cursorID uint64,
) bool {
	return event.Tick < cursorTick || event.Tick == cursorTick && event.ID <= cursorID
}

// latestSemanticCursor returns the final ordered event's cursor.
func latestSemanticCursor(events []playeradapter.SemanticEvent) (uint64, uint64) {
	if len(events) == 0 {
		return 0, 0
	}

	last := events[len(events)-1]

	return last.Tick, last.ID
}

// installSemanticEvent creates one presentation-only event entity.
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

// setSemanticAnchor installs the common position, location, facing, and height.
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

// setSemanticPayload dispatches one allowlisted semantic event type.
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

// setSemanticCast installs a presentation-safe skill cast cue.
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

// setSemanticState installs a presentation-only state lifecycle cue.
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

// setSemanticEffect installs an overlay and sound cue.
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

// setSemanticMonsterDeath installs identity without private loot or XP facts.
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

// reconcileMissiles mirrors the reliable visual subset of live projectiles.
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

// upsertMissile creates or updates one disposable projectile entity.
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

// missileComponents maps projectile presentation without damage or collision.
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

// removeStaleMissiles destroys projectiles absent from the reliable view.
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

// reconcilePersistentStates binds visual aura state to retained unit mirrors.
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

// persistentStateStores indexes selectable and player targets for visual states.
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

// indexSelectableTargets maps public selectable IDs to mirror entities.
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

// indexPlayerTargets maps player IDs to authenticated or peer hero entities.
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

// upsertPersistentState creates or updates one target-state binding.
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

// persistentStateChanged compares the retained state component with projection.
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

// removeStalePersistentStates removes absent states or missing targets.
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
