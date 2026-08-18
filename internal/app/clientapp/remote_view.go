package clientapp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// installRemoteView copies the authenticated, allowlisted client projection
// into the dedicated disposable client ECS. This ECS is never advanced while
// connected and is never sent back to the server; authority remains the remote
// session while established Lua presentation reads familiar component shapes.
func (app *application) installRemoteView(session *clientsession.Session, snap bool) error {
	if session == nil {
		return nil
	}
	presentation := session.PresentationSnapshot()
	if presentation == nil {
		return nil
	}
	return app.installRemoteProjection(presentation.HUD, presentation.Private, presentation.Party, snap)
}

func (app *application) installRemoteProjection(hud playeradapter.HUD, private playeradapter.PrivateView, party playeradapter.PartyView, snap bool) error {
	if app.clientSimulation == nil {
		return nil
	}
	world := app.clientSimulation.World()
	controls, ok := akara.GetDynamicStore(world, "d2legacy.world.player_control")
	if !ok {
		return fmt.Errorf("remote presentation: player control store is unavailable")
	}
	var hero akara.Entity
	var stale []akara.Entity
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner != hud.Player.PlayerID {
			stale = append(stale, entity)
			continue
		}
		if hero == 0 {
			// Preserve the entity handle already cached by Lua gameplay.world.
			hero = entity
			continue
		}
		// Connected presentation has exactly one local owner.
		stale = append(stale, entity)
	}
	for _, entity := range stale {
		world.DestroyEntity(entity)
	}
	if hero == 0 {
		var err error
		hero, err = world.CreateEntity()
		if err != nil {
			return fmt.Errorf("remote presentation: create authenticated hero: %w", err)
		}
	}
	if control, found := controls.Get(hero); found {
		if err := control.Set("player", hud.Player.PlayerID); err != nil {
			return err
		}
	}
	if app.movementSource != nil {
		app.movementSource.SetPlayer(hud.Player.PlayerID)
	}
	updates := map[string]map[string]any{
		"d2legacy.player.identity":              {"character_id": hud.Player.CharacterID, "player": hud.Player.PlayerID, "name": hud.Player.Name, "class": hud.Player.Class},
		"d2legacy.player.vitals":                {"health": hud.Vitals.Health, "max_health": hud.Vitals.MaxHealth, "mana": hud.Vitals.Mana, "max_mana": hud.Vitals.MaxMana, "mana_raw": hud.Vitals.Mana * 256, "max_mana_raw": hud.Vitals.MaxMana * 256, "stamina": hud.Vitals.Stamina, "max_stamina": hud.Vitals.MaxStamina, "stamina_raw": hud.Vitals.StaminaRaw, "max_stamina_raw": hud.Vitals.MaxStaminaRaw},
		"d2legacy.player.progress":              {"level": hud.Progress.Level, "experience": hud.Progress.Experience, "unspent_skill_points": hud.Progress.UnspentSkillPoints},
		"d2legacy.player.combat_stats":          {"attack_rating": hud.Combat.AttackRating, "defense": hud.Combat.Defense},
		"d2legacy.player.animation":             {"direction": hud.Animation.Direction, "mode": hud.Animation.Mode, "start_tick": int64(hud.Animation.StartTick)},
		"d2legacy.presentation.animation_clock": {"seconds": float64(0)},
		"d2legacy.world.location":               {"act": hud.Location.Act, "level_id": hud.Location.LevelID},
		"d2legacy.player.appearance":            {"cof": "", "token": classToken(hud.Player.Class), "palette": "data/global/Palette/units/pal.dat", "weapon_class": "HTH"},
		"d2legacy.world.velocity":               {"x": hud.Movement.Velocity.X, "y": hud.Movement.Velocity.Y},
		"d2legacy.world.facing":                 {"direction": hud.Animation.Direction, "directions": int64(16)},
		"d2legacy.world.player_control":         {"player": hud.Player.PlayerID},
		"d2legacy.world.bounds":                 {"width": movementBound(hud.Movement.Bounds.X), "height": movementBound(hud.Movement.Bounds.Y)},
		"d2legacy.world.collider":               {"radius": movementRadius(hud.Movement.Radius)},
		"d2legacy.player.movement_mode":         {"running": hud.Movement.Running},
		"d2legacy.player.movement_stats":        {"run_drain": hud.Movement.RunDrain, "velocitypercent": hud.Movement.VelocityPercent, "item_fastermovevelocity": hud.Movement.ItemFasterMoveVelocity, "staminarecoverybonus": hud.Movement.StaminaRecoveryBonus, "item_staminadrainpct": hud.Movement.StaminaDrainPercent, "armor_run_drain": hud.Movement.ArmorRunDrain},
		"d2legacy.player.skill_assignment":      {"left": hud.Skills.Left, "right": hud.Skills.Right},
		"d2legacy.player.belt":                  beltFields(world, hud.Belt),
		"d2legacy.player.party_view":            partyViewFields(party),
		"d2legacy.world.selectable": {
			"id": "player:" + hud.Player.PlayerID, "kind": "player", "label": hud.Player.Name,
			"owner": hud.Player.PlayerID, "radius": float64(0.75), "priority": int64(10),
		},
	}
	for name, values := range updates {
		store, found := akara.GetDynamicStore(world, name)
		if !found {
			return fmt.Errorf("remote presentation: component %q is unavailable", name)
		}
		if _, err := store.Set(hero, values); err != nil {
			return fmt.Errorf("remote presentation: set %s: %w", name, err)
		}
	}
	privateKey, err := privateProjectionFingerprint(hud.Skills.Learned, private)
	if err != nil {
		return err
	}
	if privateKey != app.privateProjectionKey {
		if err := installPrivateProjection(world, hero, hud.Player.PlayerID, hud.Skills.Learned, private); err != nil {
			return err
		}
		app.privateProjectionKey = privateKey
	}
	if err := moveMirrorToward(world, hero, hud.Position, correctionAlpha(snap)); err != nil {
		return err
	}
	// The offline simulation is deliberately frozen while connected, including
	// its Lua camera-follow system. Keep the presentation camera attached to this
	// client's authenticated hero mirror on every rendered frame.
	if follows, found := akara.GetDynamicStore(world, "d2legacy.world.camera_follow"); found {
		for _, camera := range follows.Entities() {
			if err := moveMirrorToward(world, camera, currentPosition(world, hero), 1); err != nil {
				return err
			}
		}
	}
	app.activateWorld(int(hud.Location.LevelID))
	app.logNetworkRoster(hud)
	return nil
}

// reconcileSemanticEvents mirrors only newly observed reliable semantic facts
// into the disposable client ECS. Joining and reconnecting establish a
// baseline without replaying the authority's durable history.
func (client *clientWorld) reconcileSemanticEvents(app *application, view playeradapter.EventView, epoch uint64) error {
	if client == nil || app == nil || app.clientSimulation == nil || view.Version == 0 {
		return nil
	}
	world := app.clientSimulation.World()
	destroyPrevious := func() {
		for _, entity := range client.semanticEventEntities {
			world.DestroyEntity(entity)
		}
		client.semanticEventEntities = nil
	}
	if epoch != client.lastEventEpoch {
		destroyPrevious()
		client.lastEventEpoch = epoch
		client.lastEventViewTick = view.Tick
		client.eventCursorTick, client.eventCursorID = latestSemanticCursor(view.Events)
		return nil
	}
	if view.Tick <= client.lastEventViewTick {
		return nil
	}
	if view.FromTick > client.lastEventViewTick+1 {
		return fmt.Errorf("remote presentation: semantic event window gap after tick %d (starts at %d)", client.lastEventViewTick, view.FromTick)
	}
	if view.Truncated {
		return fmt.Errorf("remote presentation: semantic event window truncated at tick %d", view.Tick)
	}
	destroyPrevious()
	nextCursorTick, nextCursorID := client.eventCursorTick, client.eventCursorID
	for _, event := range view.Events {
		if event.Tick < nextCursorTick || event.Tick == nextCursorTick && event.ID <= nextCursorID {
			continue
		}
		entity, err := installSemanticEvent(world, event)
		if err != nil {
			destroyPrevious()
			return err
		}
		client.semanticEventEntities = append(client.semanticEventEntities, entity)
		nextCursorTick, nextCursorID = event.Tick, event.ID
	}
	client.eventCursorTick, client.eventCursorID = nextCursorTick, nextCursorID
	client.lastEventViewTick = view.Tick
	return nil
}

// reconcileMissiles mirrors the reliable, allowlisted visual subset of live
// authority into presentation-only ECS entities. The ordinary Lua world
// renderer can therefore draw connected projectiles and impacts without a
// network-specific render path, while the client never receives components
// capable of damage, collision, contact locking, or lifetime advancement.
func (client *clientWorld) reconcileMissiles(app *application, projected []playeradapter.WorldMissile) error {
	if client == nil || app == nil || app.clientSimulation == nil {
		return nil
	}
	world := app.clientSimulation.World()
	if client.missileEntities == nil {
		client.missileEntities = make(map[string]akara.Entity)
	}
	seen := make(map[string]struct{}, len(projected))
	for _, missile := range projected {
		if _, duplicate := seen[missile.ID]; duplicate {
			return fmt.Errorf("remote presentation: duplicate missile %q", missile.ID)
		}
		seen[missile.ID] = struct{}{}
		entity, found := client.missileEntities[missile.ID]
		if !found {
			var err error
			entity, err = world.CreateEntity()
			if err != nil {
				return fmt.Errorf("remote presentation: create missile %q: %w", missile.ID, err)
			}
			client.missileEntities[missile.ID] = entity
		}
		updates := map[string]map[string]any{
			"d2legacy.world.position": {"x": missile.Position.X, "y": missile.Position.Y},
			"d2legacy.world.location": {"act": missile.Act, "level_id": missile.LevelID},
			"d2legacy.presentation.missile": {
				"missile_id": missile.MissileID, "dcc": missile.DCC, "palette": missile.Palette,
				"velocity_x": missile.Velocity.X, "velocity_y": missile.Velocity.Y,
				"logical_direction": missile.LogicalDirection, "directions": missile.Directions,
				"frames_per_second": missile.FramesPerSecond, "loop": missile.Loop,
				"transparency_mode": missile.TransparencyMode,
				"offset_x":          missile.OffsetX, "offset_y": missile.OffsetY, "offset_z": missile.OffsetZ,
			},
		}
		for name, values := range updates {
			store, available := akara.GetDynamicStore(world, name)
			if !available {
				return fmt.Errorf("remote presentation: missile component %q is unavailable", name)
			}
			if _, err := store.Set(entity, values); err != nil {
				return fmt.Errorf("remote presentation: set missile %s: %w", name, err)
			}
		}
	}
	for id, entity := range client.missileEntities {
		if _, retained := seen[id]; retained {
			continue
		}
		world.DestroyEntity(entity)
		delete(client.missileEntities, id)
	}
	return nil
}

// reconcilePersistentStates binds the reliable semantic aura projection to
// existing disposable unit mirrors. The reconstructed component is visual
// only: no stat value, radius, source, party/filter decision, or authority
// lifecycle enters the connected client ECS.
func (client *clientWorld) reconcilePersistentStates(app *application, projected []playeradapter.WorldState) error {
	if client == nil || app == nil || app.clientSimulation == nil {
		return nil
	}
	world := app.clientSimulation.World()
	selectables, found := akara.GetDynamicStore(world, "d2legacy.world.selectable")
	if !found {
		return fmt.Errorf("remote presentation: selectable component is unavailable")
	}
	states, found := akara.GetDynamicStore(world, "d2legacy.presentation.state")
	if !found {
		return fmt.Errorf("remote presentation: persistent state component is unavailable")
	}
	targets := make(map[string]akara.Entity, selectables.Len())
	for _, entity := range selectables.Entities() {
		selectable, ok := selectables.Get(entity)
		if !ok {
			continue
		}
		id, _ := selectable.Get("id")
		if value, ok := id.(string); ok && value != "" {
			targets[value] = entity
		}
	}
	if identities, available := akara.GetDynamicStore(world, "d2legacy.player.identity"); available {
		for _, entity := range identities.Entities() {
			identity, ok := identities.Get(entity)
			if !ok {
				continue
			}
			player, _ := identity.Get("player")
			if value, ok := player.(string); ok && value != "" {
				targets["player:"+value] = entity
			}
		}
	}
	if client.stateEntities == nil {
		client.stateEntities = make(map[presentationStateKey]akara.Entity)
	}
	seen := make(map[presentationStateKey]struct{}, len(projected))
	for _, projectedState := range projected {
		key := presentationStateKey{targetID: projectedState.TargetID, stateID: projectedState.StateID}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("remote presentation: duplicate persistent state target=%q state=%q", key.targetID, key.stateID)
		}
		seen[key] = struct{}{}
		target, available := targets[projectedState.TargetID]
		if !available {
			continue
		}
		entity, retained := client.stateEntities[key]
		if !retained {
			var err error
			entity, err = world.CreateEntity()
			if err != nil {
				return fmt.Errorf("remote presentation: create persistent state target=%q state=%q: %w", key.targetID, key.stateID, err)
			}
			client.stateEntities[key] = entity
		}
		component, present := states.Get(entity)
		changed := !present
		if present {
			currentTarget, _ := component.Get("target")
			currentState, _ := component.Get("state_id")
			currentPeriod, _ := component.Get("period_ticks")
			changed = currentTarget != target || currentState != projectedState.StateID || currentPeriod != projectedState.PeriodTicks
		}
		if changed {
			if _, err := states.Set(entity, map[string]any{
				"target": target, "state_id": projectedState.StateID, "period_ticks": projectedState.PeriodTicks,
			}); err != nil {
				return fmt.Errorf("remote presentation: set persistent state target=%q state=%q: %w", key.targetID, key.stateID, err)
			}
		}
	}
	for key, entity := range client.stateEntities {
		_, retained := seen[key]
		_, targetRetained := targets[key.targetID]
		if !retained || !targetRetained {
			world.DestroyEntity(entity)
			delete(client.stateEntities, key)
		}
	}
	return nil
}

func latestSemanticCursor(events []playeradapter.SemanticEvent) (uint64, uint64) {
	if len(events) == 0 {
		return 0, 0
	}
	last := events[len(events)-1]
	return last.Tick, last.ID
}

func installSemanticEvent(world *akara.World, event playeradapter.SemanticEvent) (akara.Entity, error) {
	entity, err := world.CreateEntity()
	if err != nil {
		return 0, fmt.Errorf("remote presentation: create semantic event: %w", err)
	}
	fail := func(err error) (akara.Entity, error) {
		world.DestroyEntity(entity)
		return 0, err
	}
	for name, values := range map[string]map[string]any{
		"d2legacy.world.position":              {"x": event.Position.X, "y": event.Position.Y},
		"d2legacy.world.location":              {"act": event.Act, "level_id": event.LevelID},
		"d2legacy.world.facing":                {"direction": event.Direction, "directions": int64(16)},
		"d2legacy.presentation.overlay_anchor": {"height": event.OverlayHeight},
	} {
		store, found := akara.GetDynamicStore(world, name)
		if !found {
			return fail(fmt.Errorf("remote presentation: semantic component %q is unavailable", name))
		}
		if _, err := store.Set(entity, values); err != nil {
			return fail(fmt.Errorf("remote presentation: set semantic %s: %w", name, err))
		}
	}
	switch event.Type {
	case "cast":
		store, found := akara.GetDynamicStore(world, "d2legacy.skill.cast_cue")
		if !found || event.Cast == nil {
			return fail(fmt.Errorf("remote presentation: cast cue component is unavailable"))
		}
		cue := event.Cast
		_, err = store.Set(entity, map[string]any{
			"kind": cue.Kind, "tick": int64(event.Tick), "effect_tick": int64(cue.EffectTick), "caster": entity,
			"player": cue.Player, "skill_id": cue.SkillID, "target_x": cue.Target.X, "target_y": cue.Target.Y, "target_id": cue.TargetID,
		})
	case "state":
		store, found := akara.GetDynamicStore(world, "d2legacy.state.event")
		if !found || event.State == nil {
			return fail(fmt.Errorf("remote presentation: state event component is unavailable"))
		}
		cue := event.State
		_, err = store.Set(entity, map[string]any{
			"kind": cue.Kind, "tick": int64(event.Tick), "target": entity, "state_id": cue.StateID,
			"source_id": cue.SourceID, "expires_tick": int64(cue.ExpiresTick), "reason": cue.Reason,
		})
	case "monster_death":
		store, found := akara.GetDynamicStore(world, "d2legacy.monster.death_event")
		if !found || event.MonsterDeath == nil {
			return fail(fmt.Errorf("remote presentation: monster death component is unavailable"))
		}
		cue := event.MonsterDeath
		_, err = store.Set(entity, map[string]any{
			"kind": cue.Kind, "tick": int64(event.Tick), "monster_id": cue.MonsterID,
			"killer_id": "", "credited_id": "", "xp": int64(0), "loot_seed": "",
			"treasure_class": "", "drops": "", "game_player_count": int64(0),
			"effective_player_count": int64(0), "nearby_party_member_count": int64(0),
			"monster_player_count": int64(0), "no_drop_player_count": int64(0),
		})
	default:
		return fail(fmt.Errorf("remote presentation: unsupported semantic event %q", event.Type))
	}
	if err != nil {
		return fail(fmt.Errorf("remote presentation: set %s semantic event: %w", event.Type, err))
	}
	return entity, nil
}

func partyViewFields(view playeradapter.PartyView) map[string]any {
	values := map[string]any{
		"schema_version": int64(view.Version),
		"revision":       int64(view.Revision),
		"party_id":       view.PartyID,
		"roster_count":   int64(len(view.Roster)),
	}
	for slot := 1; slot <= playeradapter.MaxPartyViewRoster; slot++ {
		suffix := fmt.Sprintf("_%d", slot)
		values["player"+suffix] = ""
		values["name"+suffix] = ""
		values["class"+suffix] = ""
		values["level"+suffix] = int64(0)
		values["relationship"+suffix] = ""
		if slot <= len(view.Roster) {
			entry := view.Roster[slot-1]
			values["player"+suffix] = entry.PlayerID
			values["name"+suffix] = entry.Name
			values["class"+suffix] = entry.Class
			values["level"+suffix] = entry.Level
			values["relationship"+suffix] = entry.Relationship
		}
	}
	return values
}

func movementBound(value float64) float64 {
	if value <= 0 {
		return 1 << 20
	}
	return value
}

func movementRadius(value float64) float64 {
	if value <= 0 {
		return 1
	}
	return value
}

func monsterWeaponClass(value string) string {
	if value == "" {
		return "HTH"
	}
	return value
}

func (app *application) syncRemoteMirrors(projected []playeradapter.WorldEntity, location playeradapter.HUDLocation) error {
	if app.clientSimulation == nil {
		return nil
	}
	world := app.clientSimulation.World()
	if app.remoteMirrors == nil {
		app.remoteMirrors = map[string]akara.Entity{}
	}
	if app.remoteMirrorKeys == nil {
		app.remoteMirrorKeys = map[string]string{}
	}
	seen := map[string]bool{}
	for _, remote := range projected {
		seen[remote.ID] = true
		entity, exists := app.remoteMirrors[remote.ID]
		if !exists {
			var err error
			entity, err = world.CreateEntity()
			if err != nil {
				return err
			}
			app.remoteMirrors[remote.ID] = entity
		}
		key, err := remotePresentationFingerprint(remote, location)
		if err != nil {
			return err
		}
		if !exists || app.remoteMirrorKeys[remote.ID] != key {
			if err := installWorldMirror(world, entity, remote, location, false); err != nil {
				return err
			}
			app.remoteMirrorKeys[remote.ID] = key
		}
	}
	for id, entity := range app.remoteMirrors {
		if !seen[id] {
			world.DestroyEntity(entity)
			delete(app.remoteMirrors, id)
			delete(app.remoteMirrorKeys, id)
		}
	}
	return nil
}

// applySampledWorldPositions is the sole writer of public-world presentation
// transforms. Structural snapshot installation deliberately leaves existing
// transforms alone so packet arrival cannot produce a visible discontinuity.
func (app *application) applySampledWorldPositions(projected []playeradapter.WorldEntity) error {
	if app.clientSimulation == nil {
		return nil
	}
	world := app.clientSimulation.World()
	for _, remote := range projected {
		entity, found := app.remoteMirrors[remote.ID]
		if !found {
			continue
		}
		if err := moveMirrorToward(world, entity, remote.Position, 1); err != nil {
			return err
		}
	}
	return nil
}

func (app *application) applyLocalPredictedPosition(playerID string, predicted playeradapter.HUDPosition) error {
	if app.clientSimulation == nil {
		return nil
	}
	world := app.clientSimulation.World()
	controls, found := akara.GetDynamicStore(world, "d2legacy.world.player_control")
	if !found {
		return fmt.Errorf("remote presentation: player control store is unavailable")
	}
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner != playerID {
			continue
		}
		if err := moveMirrorToward(world, entity, predicted, 1); err != nil {
			return err
		}
		if follows, found := akara.GetDynamicStore(world, "d2legacy.world.camera_follow"); found {
			for _, camera := range follows.Entities() {
				if err := moveMirrorToward(world, camera, predicted, 1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return fmt.Errorf("remote presentation: authenticated player %q is unavailable", playerID)
}

func (app *application) applyAnimationTimeline(localPlayer string, timeline networkclock.Timeline, step time.Duration) error {
	if app.clientSimulation == nil || step <= 0 || !timeline.Ready {
		return nil
	}
	world := app.clientSimulation.World()
	animations, animationOK := akara.GetDynamicStore(world, "d2legacy.player.animation")
	clocks, clockOK := akara.GetDynamicStore(world, "d2legacy.presentation.animation_clock")
	identities, identityOK := akara.GetDynamicStore(world, "d2legacy.player.identity")
	if !animationOK || !clockOK || !identityOK {
		return fmt.Errorf("remote presentation: animation timeline stores are unavailable")
	}
	for _, entity := range animations.Entities() {
		animation, animationFound := animations.Get(entity)
		identity, identityFound := identities.Get(entity)
		if !animationFound || !identityFound {
			continue
		}
		owner, _ := identity.Get("player")
		startValue, _ := animation.Get("start_tick")
		start := float64(startValue.(int64))
		moment := timeline.Interpolation
		if owner == localPlayer {
			moment = timeline.Prediction
		}
		current := float64(moment.Tick) + moment.Fraction
		seconds := max(0, current-start) * step.Seconds()
		if _, err := clocks.Set(entity, map[string]any{"seconds": seconds}); err != nil {
			return err
		}
	}
	return nil
}

func remotePresentationFingerprint(entity playeradapter.WorldEntity, location playeradapter.HUDLocation) (string, error) {
	entity = cloneWorldEntity(entity)
	entity.Position = playeradapter.HUDPosition{}
	payload, err := json.Marshal(struct {
		Entity   playeradapter.WorldEntity `json:"entity"`
		Location playeradapter.HUDLocation `json:"location"`
	}{Entity: entity, Location: location})
	if err != nil {
		return "", fmt.Errorf("remote presentation: fingerprint world entity: %w", err)
	}
	return string(payload), nil
}

func installWorldMirror(world *akara.World, entity akara.Entity, value playeradapter.WorldEntity, location playeradapter.HUDLocation, snap bool) error {
	if value.Kind == "player" {
		return installPlayerMirror(world, entity, value, location, snap)
	}
	values := map[string]map[string]any{
		"d2legacy.monster.identity": {
			"spawn_id": value.SpawnID, "definition_id": value.DefinitionID, "base_id": value.DefinitionID,
			"graphics_id": value.Token, "seed": "", "treasure_class": "",
		},
		"d2legacy.monster.appearance": {
			"token": value.Token, "mode": value.Mode, "weapon_class": monsterWeaponClass(value.WeaponClass),
			"name_key": value.Label, "components": value.Components, "death_sound": value.DeathSound,
			"overlay_height": value.OverlayHeight,
		},
		"d2legacy.monster.stats": {
			"level": int64(1), "health": pointed(value.Health), "max_health": pointed(value.MaxHealth),
			"defense": int64(0), "attack_rating": int64(0), "physical_min": int64(0),
			"physical_max": int64(0), "experience": int64(0),
		},
		"d2legacy.world.velocity": {"x": float64(0), "y": float64(0)},
		"d2legacy.world.facing":   {"direction": value.Direction, "directions": int64(16)},
		"d2legacy.world.location": {"act": worldLocation(value.Act, location.Act), "level_id": worldLocation(value.LevelID, location.LevelID)},
	}
	if value.Kind != "corpse" {
		values["d2legacy.world.collider"] = map[string]any{"radius": value.Radius}
		values["d2legacy.world.selectable"] = map[string]any{
			"id": value.ID, "kind": value.Kind, "label": value.Label, "owner": value.Owner,
			"radius": value.Radius, "priority": value.Priority,
		}
	} else {
		// A corpse keeps the living monster's presentation entity. Removing the
		// selectable and collider components changes the ECS queries it can enter,
		// so the mirror remains renderable without becoming a stale interaction or
		// locomotion obstacle in the disposable client simulation.
		if selectable, found := akara.GetDynamicStore(world, "d2legacy.world.selectable"); found {
			selectable.Remove(entity)
		}
		if colliders, found := akara.GetDynamicStore(world, "d2legacy.world.collider"); found {
			colliders.Remove(entity)
		}
	}
	if err := moveMirrorToward(world, entity, value.Position, correctionAlpha(snap)); err != nil {
		return err
	}
	for name, fields := range values {
		store, ok := akara.GetDynamicStore(world, name)
		if !ok {
			return fmt.Errorf("remote presentation: component %q is unavailable", name)
		}
		if _, err := store.Set(entity, fields); err != nil {
			return err
		}
	}
	return nil
}

func pointed(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func worldLocation(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}

func (app *application) logNetworkRoster(hud playeradapter.HUD) {
	world := app.clientSimulation.World()
	identities, identityOK := akara.GetDynamicStore(world, "d2legacy.player.identity")
	appearances, appearanceOK := akara.GetDynamicStore(world, "d2legacy.player.appearance")
	positions, positionOK := akara.GetDynamicStore(world, "d2legacy.world.position")
	if !identityOK || !appearanceOK || !positionOK {
		return
	}
	type rosterEntry struct {
		Entity uint64
		Owner  string
		Class  string
		Token  string
		X      float64
		Y      float64
	}
	entries := make([]rosterEntry, 0, identities.Len())
	for _, entity := range identities.Entities() {
		identity, identityFound := identities.Get(entity)
		appearance, appearanceFound := appearances.Get(entity)
		position, positionFound := positions.Get(entity)
		if !identityFound || !appearanceFound || !positionFound {
			continue
		}
		owner, _ := identity.Get("player")
		class, _ := identity.Get("class")
		token, _ := appearance.Get("token")
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		entries = append(entries, rosterEntry{Entity: uint64(entity), Owner: owner.(string), Class: class.(string), Token: token.(string), X: x.(float64), Y: y.(float64)})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Entity < entries[j].Entity })
	// Positions change every correction and belong in trace-level correction
	// diagnostics. A debug roster event is useful only when membership or
	// presentation identity changes; logging it at 10 Hz can itself disturb
	// frame pacing when logs are redirected to disk.
	key := hud.Player.PlayerID + ":" + hud.Player.Class
	for _, entry := range entries {
		key += fmt.Sprintf(":%s:%s:%s", entry.Owner, entry.Class, entry.Token)
	}
	if key == app.networkRosterLogKey {
		return
	}
	app.networkRosterLogKey = key
	slog.Debug("connected player presentation roster", "authenticated_player", hud.Player.PlayerID,
		"authenticated_class", hud.Player.Class, "entities", entries)
}

func installPlayerMirror(world *akara.World, entity akara.Entity, player playeradapter.WorldEntity, location playeradapter.HUDLocation, snap bool) error {
	values := map[string]map[string]any{
		"d2legacy.player.identity":              {"character_id": player.ID, "player": player.Owner, "name": player.Label, "class": player.Class},
		"d2legacy.player.appearance":            {"cof": "", "token": player.Token, "palette": "data/global/Palette/units/pal.dat", "weapon_class": "HTH"},
		"d2legacy.player.animation":             {"direction": player.Direction, "mode": player.Mode, "start_tick": int64(player.AnimationStartTick)},
		"d2legacy.player.movement_stats":        {"velocitypercent": player.VelocityPercent, "item_fastermovevelocity": player.ItemFasterMoveVelocity},
		"d2legacy.presentation.animation_clock": {"seconds": float64(0)},
		"d2legacy.world.facing":                 {"direction": player.Direction, "directions": int64(16)},
		"d2legacy.world.location":               {"act": worldLocation(player.Act, location.Act), "level_id": worldLocation(player.LevelID, location.LevelID)},
	}
	if err := moveMirrorToward(world, entity, player.Position, correctionAlpha(snap)); err != nil {
		return err
	}
	for name, fields := range values {
		store, ok := akara.GetDynamicStore(world, name)
		if !ok {
			return fmt.Errorf("remote presentation: component %q is unavailable", name)
		}
		if _, err := store.Set(entity, fields); err != nil {
			return err
		}
	}
	return nil
}

func beltFields(world *akara.World, belt playeradapter.HUDBelt) map[string]any {
	fields := map[string]any{"capacity": belt.Capacity}
	store, found := akara.GetDynamicStore(world, "d2legacy.player.belt")
	if !found {
		return fields
	}
	available := map[string]bool{}
	for _, field := range store.Schema().Fields {
		available[field.Name] = true
	}
	for slot := 1; slot <= 16; slot++ {
		name := fmt.Sprintf("slot_%d", slot)
		if !available[name] {
			continue
		}
		value := ""
		if slot <= len(belt.Slots) {
			value = belt.Slots[slot-1]
		}
		fields[name] = value
	}
	return fields
}

// installPrivateProjection mirrors only authenticated owner-private state.
// Removing and rebuilding this small graph on a correction keeps entity
// references coherent and avoids retaining an item or interaction after the
// server has removed it.
func installPrivateProjection(world *akara.World, hero akara.Entity, owner string, learned []playeradapter.HUDLearnedSkill, private playeradapter.PrivateView) error {
	for _, component := range []string{
		"d2legacy.player.learned_skill", "d2legacy.items.layout", "d2legacy.interaction.context",
		"d2legacy.interaction.target", "d2legacy.interaction.null_target",
	} {
		store, found := akara.GetDynamicStore(world, component)
		if !found {
			return fmt.Errorf("remote presentation: component %q is unavailable", component)
		}
		for _, entity := range store.Entities() {
			world.DestroyEntity(entity)
		}
	}
	for _, skill := range learned {
		entity, err := world.CreateEntity()
		if err != nil {
			return err
		}
		store, _ := akara.GetDynamicStore(world, "d2legacy.player.learned_skill")
		if _, err := store.Set(entity, map[string]any{
			"owner": hero, "skill_id": skill.SkillID, "level": skill.Level, "list_row": skill.ListRow,
			"left_allowed": skill.LeftAllowed, "right_allowed": skill.RightAllowed,
		}); err != nil {
			return err
		}
	}
	layout, err := world.CreateEntity()
	if err != nil {
		return err
	}
	layoutStore, _ := akara.GetDynamicStore(world, "d2legacy.items.layout")
	l := private.Items.Layout
	if _, err := layoutStore.Set(layout, map[string]any{
		"owner": owner, "inventory_width": l.InventoryWidth, "inventory_height": l.InventoryHeight,
		"stash_width": l.StashWidth, "stash_height": l.StashHeight, "cube_width": l.CubeWidth, "cube_height": l.CubeHeight,
		"belt_capacity": l.BeltCapacity, "active_weapon_set": l.ActiveWeaponSet, "vendor_width": l.VendorWidth,
		"vendor_height": l.VendorHeight, "carried_gold": l.CarriedGold, "stashed_gold": l.StashedGold,
	}); err != nil {
		return err
	}
	identityStore, identityOK := akara.GetDynamicStore(world, "d2legacy.item.identity")
	placementStore, placementOK := akara.GetDynamicStore(world, "d2legacy.item.placement")
	presentationStore, presentationOK := akara.GetDynamicStore(world, "d2legacy.item.presentation")
	if !identityOK || !placementOK || !presentationOK {
		return fmt.Errorf("remote presentation: item stores are unavailable")
	}
	for _, entity := range identityStore.Entities() {
		world.DestroyEntity(entity)
	}
	for _, item := range private.Items.Items {
		entity, createErr := world.CreateEntity()
		if createErr != nil {
			return createErr
		}
		if _, err := identityStore.Set(entity, map[string]any{
			"owner": layout, "id": item.ID, "code": item.Code, "width": item.Width, "height": item.Height,
			"body_slots": item.BodySlots, "belt_eligible": item.BeltEligible, "base_cost": item.BaseCost,
			"applied_services": item.AppliedServices,
		}); err != nil {
			return err
		}
		if _, err := placementStore.Set(entity, map[string]any{
			"container": item.Container, "x": item.X, "y": item.Y, "slot": item.Slot, "belt_slot": item.BeltSlot,
			"weapon_set": item.WeaponSet, "page": item.Page,
		}); err != nil {
			return err
		}
		if _, err := presentationStore.Set(entity, map[string]any{
			"inventory_dc6": item.InventoryDC6, "world_dc6": item.WorldDC6, "world_animated": item.WorldAnimated,
			"composite": item.Composite, "weapon_class": item.WeaponClass,
		}); err != nil {
			return err
		}
	}
	nullTarget, err := world.CreateEntity()
	if err != nil {
		return err
	}
	if store, found := akara.GetDynamicStore(world, "d2legacy.interaction.null_target"); found {
		if _, err := store.Set(nullTarget, map[string]any{}); err != nil {
			return err
		}
	}
	target := nullTarget
	if private.Interaction.Active && private.Interaction.Target != nil {
		target, err = world.CreateEntity()
		if err != nil {
			return err
		}
		value := private.Interaction.Target
		store, found := akara.GetDynamicStore(world, "d2legacy.interaction.target")
		if !found {
			return fmt.Errorf("remote presentation: interaction target store is unavailable")
		}
		if _, err := store.Set(target, map[string]any{
			"id": value.ID, "npc": value.NPC, "vendor": value.Vendor, "categories": value.Categories,
			"services": value.Services, "x": value.X, "y": value.Y, "radius": value.Radius,
		}); err != nil {
			return err
		}
	}
	contextEntity, err := world.CreateEntity()
	if err != nil {
		return err
	}
	contextStore, _ := akara.GetDynamicStore(world, "d2legacy.interaction.context")
	_, err = contextStore.Set(contextEntity, map[string]any{"owner": owner, "target": target})
	return err
}

func correctionAlpha(snap bool) float64 {
	if snap {
		return 1
	}
	// Existing mirrors are advanced by the frame interpolator. Applying even a
	// partial correction here creates a discontinuity at every snapshot edge.
	return 0
}

func privateProjectionFingerprint(learned []playeradapter.HUDLearnedSkill, private playeradapter.PrivateView) (string, error) {
	// Tick advances with every correction even when the owner-private graph is
	// unchanged. It is transport metadata, not presentation content.
	private.Tick = 0
	payload, err := json.Marshal(struct {
		Learned []playeradapter.HUDLearnedSkill `json:"learned"`
		Private playeradapter.PrivateView       `json:"private"`
	}{Learned: learned, Private: private})
	if err != nil {
		return "", fmt.Errorf("remote presentation: fingerprint private projection: %w", err)
	}
	return string(payload), nil
}

func currentPosition(world *akara.World, entity akara.Entity) playeradapter.HUDPosition {
	positions, found := akara.GetDynamicStore(world, "d2legacy.world.position")
	if !found {
		return playeradapter.HUDPosition{}
	}
	position, found := positions.Get(entity)
	if !found {
		return playeradapter.HUDPosition{}
	}
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	return playeradapter.HUDPosition{X: x.(float64), Y: y.(float64)}
}

func moveMirrorToward(world *akara.World, entity akara.Entity, target playeradapter.HUDPosition, alpha float64) error {
	positions, found := akara.GetDynamicStore(world, "d2legacy.world.position")
	if !found {
		return fmt.Errorf("remote presentation: position store is unavailable")
	}
	position, exists := positions.Get(entity)
	if !exists {
		_, err := positions.Set(entity, map[string]any{"x": target.X, "y": target.Y})
		return err
	}
	if alpha <= 0 {
		return nil
	}
	xValue, _ := position.Get("x")
	yValue, _ := position.Get("y")
	x, y := xValue.(float64), yValue.(float64)
	if math.Hypot(target.X-x, target.Y-y) > 4 {
		x, y = target.X, target.Y
	} else {
		x += (target.X - x) * alpha
		y += (target.Y - y) * alpha
	}
	if err := position.Set("x", x); err != nil {
		return err
	}
	return position.Set("y", y)
}

func classToken(class string) string {
	return map[string]string{"Amazon": "AM", "Sorceress": "SO", "Necromancer": "NE", "Paladin": "PA", "Barbarian": "BA", "Druid": "DZ", "Assassin": "AI"}[class]
}
