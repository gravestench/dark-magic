package clientapp

import (
	"fmt"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const playerPalettePath = "data/global/Palette/units/pal.dat"

// installRemoteView reads one immutable session snapshot and forwards only its authenticated HUD,
// private owner view, and party projection into the disposable ECS.
func (app *application) installRemoteView(
	session *clientsession.Session,
	snap bool,
) error {
	if session == nil {
		return nil
	}

	presentation := session.PresentationSnapshot()
	if presentation == nil {
		return nil
	}

	return app.installRemoteProjection(
		presentation.HUD,
		presentation.Private,
		presentation.Party,
		snap,
	)
}

// installRemoteProjection refreshes owner components and private graphs, while position changes obey
// the caller's snap policy. It also follows authoritative level changes in presentation state.
func (app *application) installRemoteProjection(
	hud playeradapter.HUD,
	private playeradapter.PrivateView,
	party playeradapter.PartyView,
	snap bool,
) error {
	if app.clientSimulation == nil {
		return nil
	}

	world := app.clientSimulation.World()

	hero, err := authenticatedHero(world, hud.Player.PlayerID)
	if err != nil {
		return err
	}

	if app.movementSource != nil {
		app.movementSource.SetPlayer(hud.Player.PlayerID)
	}

	if err := setRemoteComponents(world, hero, remoteHeroComponents(world, hud, party)); err != nil {
		return err
	}

	if err := app.refreshPrivateProjection(world, hero, hud, private); err != nil {
		return err
	}

	if err := moveMirrorToward(world, hero, hud.Position, correctionAlpha(snap)); err != nil {
		return err
	}

	// The offline camera system is frozen while connected, so attach it here.
	if err := followRemoteHero(world, hero); err != nil {
		return err
	}

	app.activateWorld(int(hud.Location.LevelID))
	app.logNetworkRoster(hud)

	return nil
}

// authenticatedHero preserves exactly one controlled entity for the admitted player and destroys
// every stale or duplicate control entity. Retention protects Lua code that caches the entity handle.
func authenticatedHero(world *akara.World, playerID string) (akara.Entity, error) {
	controls, found := akara.GetDynamicStore(world, "d2legacy.world.player_control")
	if !found {
		return 0, fmt.Errorf("remote presentation: player control store is unavailable")
	}

	var hero akara.Entity

	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")

		if owner == playerID && hero == 0 {
			// Lua gameplay.world may already cache this entity handle.
			hero = entity

			continue
		}

		world.DestroyEntity(entity)
	}

	if hero == 0 {
		var err error

		hero, err = world.CreateEntity()
		if err != nil {
			return 0, fmt.Errorf("remote presentation: create authenticated hero: %w", err)
		}
	}

	if control, available := controls.Get(hero); available {
		if err := control.Set("player", playerID); err != nil {
			return 0, err
		}
	}

	return hero, nil
}

// remoteHeroComponents is the owner projection allowlist. It provides UI and rendering facts but no
// simulation systems capable of overriding the server's canonical player.
func remoteHeroComponents(
	world *akara.World,
	hud playeradapter.HUD,
	party playeradapter.PartyView,
) map[string]map[string]any {
	return map[string]map[string]any{
		"d2legacy.player.identity": {
			"character_id": hud.Player.CharacterID,
			"player":       hud.Player.PlayerID,
			"name":         hud.Player.Name,
			"class":        hud.Player.Class,
		},
		"d2legacy.player.vitals":           remoteVitalsFields(hud.Vitals),
		"d2legacy.player.progress":         remoteProgressFields(hud.Progress),
		"d2legacy.player.combat_stats":     remoteCombatFields(hud.Combat),
		"d2legacy.player.animation":        remoteAnimationFields(hud.Animation),
		"d2legacy.player.movement_stats":   remoteMovementFields(hud.Movement),
		"d2legacy.player.skill_assignment": remoteSkillAssignmentFields(hud.Skills),
		"d2legacy.player.belt":             beltFields(world, hud.Belt),
		"d2legacy.player.party_view":       partyViewFields(party),
		"d2legacy.player.appearance": {
			"cof":          "",
			"token":        classToken(hud.Player.Class),
			"palette":      playerPalettePath,
			"weapon_class": "HTH",
		},
		"d2legacy.player.movement_mode": {
			"running": hud.Movement.Running,
		},
		"d2legacy.presentation.animation_clock": {
			"seconds": float64(0),
		},
		"d2legacy.world.location": {
			"act":      hud.Location.Act,
			"level_id": hud.Location.LevelID,
		},
		"d2legacy.world.velocity": {
			"x": hud.Movement.Velocity.X,
			"y": hud.Movement.Velocity.Y,
		},
		"d2legacy.world.facing": {
			"direction":  hud.Animation.Direction,
			"directions": int64(16),
		},
		"d2legacy.world.player_control": {
			"player": hud.Player.PlayerID,
		},
		"d2legacy.world.bounds": {
			"width":  movementBound(hud.Movement.Bounds.X),
			"height": movementBound(hud.Movement.Bounds.Y),
		},
		"d2legacy.world.collider": {
			"radius": movementRadius(hud.Movement.Radius),
		},
		"d2legacy.world.selectable": {
			"id":       "player:" + hud.Player.PlayerID,
			"kind":     "player",
			"label":    hud.Player.Name,
			"owner":    hud.Player.PlayerID,
			"radius":   float64(0.75),
			"priority": int64(10),
		},
	}
}

// remoteVitalsFields maps display totals and fixed-point stamina fields expected by presentation
// schemas. Derived mana raw values exist for compatibility, not as writable authority.
func remoteVitalsFields(vitals playeradapter.HUDVitals) map[string]any {
	return map[string]any{
		"health":          vitals.Health,
		"max_health":      vitals.MaxHealth,
		"mana":            vitals.Mana,
		"max_mana":        vitals.MaxMana,
		"mana_raw":        vitals.Mana * 256,
		"max_mana_raw":    vitals.MaxMana * 256,
		"stamina":         vitals.Stamina,
		"max_stamina":     vitals.MaxStamina,
		"stamina_raw":     vitals.StaminaRaw,
		"max_stamina_raw": vitals.MaxStaminaRaw,
	}
}

// remoteProgressFields exposes only progression values shown by connected UI; award calculation
// remains absent from the presentation replica.
func remoteProgressFields(progress playeradapter.HUDProgress) map[string]any {
	return map[string]any{
		"level":                progress.Level,
		"experience":           progress.Experience,
		"unspent_skill_points": progress.UnspentSkillPoints,
	}
}

// remoteCombatFields exposes summary ratings for UI without projecting the modifiers or rolls used
// by authority to calculate combat outcomes.
func remoteCombatFields(combat playeradapter.HUDCombat) map[string]any {
	return map[string]any{
		"attack_rating": combat.AttackRating,
		"defense":       combat.Defense,
	}
}

// remoteAnimationFields carries the authority-selected mode, direction, and start tick. The client
// may advance visual time but cannot select a different canonical action.
func remoteAnimationFields(animation playeradapter.HUDAnimation) map[string]any {
	return map[string]any{
		"direction":  animation.Direction,
		"mode":       animation.Mode,
		"start_tick": int64(animation.StartTick),
	}
}

// remoteMovementFields projects only modifiers required to reproduce local prediction. Collision and
// accepted position still come from authority corrections.
func remoteMovementFields(movement playeradapter.HUDMovement) map[string]any {
	return map[string]any{
		"run_drain":               movement.RunDrain,
		"velocitypercent":         movement.VelocityPercent,
		"item_fastermovevelocity": movement.ItemFasterMoveVelocity,
		"staminarecoverybonus":    movement.StaminaRecoveryBonus,
		"item_staminadrainpct":    movement.StaminaDrainPercent,
		"armor_run_drain":         movement.ArmorRunDrain,
	}
}

// remoteSkillAssignmentFields exposes current input-slot selection without granting either skill.
func remoteSkillAssignmentFields(skills playeradapter.HUDSkills) map[string]any {
	return map[string]any{
		"left":  skills.Left,
		"right": skills.Right,
	}
}

// setRemoteComponents requires every allowlisted schema to be registered and stops on the first
// failed write. Missing presentation structure is surfaced instead of silently dropping HUD facts.
func setRemoteComponents(
	world *akara.World,
	entity akara.Entity,
	updates map[string]map[string]any,
) error {
	for name, values := range updates {
		store, found := akara.GetDynamicStore(world, name)
		if !found {
			return fmt.Errorf("remote presentation: component %q is unavailable", name)
		}

		if _, err := store.Set(entity, values); err != nil {
			return fmt.Errorf("remote presentation: set %s: %w", name, err)
		}
	}

	return nil
}

// refreshPrivateProjection fingerprints content without transport ticks and rebuilds only on semantic
// change. This preserves entity handles and avoids expensive graph churn every correction.
func (app *application) refreshPrivateProjection(
	world *akara.World,
	hero akara.Entity,
	hud playeradapter.HUD,
	private playeradapter.PrivateView,
) error {
	key, err := privateProjectionFingerprint(hud.Skills.Learned, private)
	if err != nil {
		return err
	}

	if key == app.privateProjectionKey {
		return nil
	}

	if err := installPrivateProjection(
		world,
		hero,
		hud.Player.PlayerID,
		hud.Skills.Learned,
		private,
	); err != nil {
		return err
	}

	app.privateProjectionKey = key

	return nil
}

// followRemoteHero aligns camera entities immediately because the offline follow system is frozen in
// connected play; subsequent predicted motion moves them with the authenticated owner.
func followRemoteHero(world *akara.World, hero akara.Entity) error {
	follows, found := akara.GetDynamicStore(world, "d2legacy.world.camera_follow")
	if !found {
		return nil
	}

	position := currentPosition(world, hero)

	for _, camera := range follows.Entities() {
		if err := moveMirrorToward(world, camera, position, 1); err != nil {
			return err
		}
	}

	return nil
}

// partyViewFields converts a variable authority roster into the fixed registered Lua schema. Every
// slot is initialized so members removed in a later revision cannot leave stale fields.
func partyViewFields(view playeradapter.PartyView) map[string]any {
	values := map[string]any{
		"schema_version": int64(view.Version),
		"revision":       int64(view.Revision),
		"party_id":       view.PartyID,
		"roster_count":   int64(len(view.Roster)),
	}

	for slot := 1; slot <= playeradapter.MaxPartyViewRoster; slot++ {
		setPartyRosterSlot(values, slot, view.Roster)
	}

	return values
}

// setPartyRosterSlot writes neutral values before an optional member, making replacement complete even
// when the new roster is shorter than the previous one.
func setPartyRosterSlot(
	values map[string]any,
	slot int,
	roster []playeradapter.PartyRosterEntry,
) {
	suffix := fmt.Sprintf("_%d", slot)
	values["player"+suffix] = ""
	values["name"+suffix] = ""
	values["class"+suffix] = ""
	values["level"+suffix] = int64(0)
	values["relationship"+suffix] = ""

	if slot > len(roster) {
		return
	}

	entry := roster[slot-1]
	values["player"+suffix] = entry.PlayerID
	values["name"+suffix] = entry.Name
	values["class"+suffix] = entry.Class
	values["level"+suffix] = entry.Level
	values["relationship"+suffix] = entry.Relationship
}

// beltFields intersects the projected belt with fields registered by the active package set. This
// allows older schemas to render safely without discarding supported slots.
func beltFields(world *akara.World, belt playeradapter.HUDBelt) map[string]any {
	fields := map[string]any{"capacity": belt.Capacity}

	store, found := akara.GetDynamicStore(world, "d2legacy.player.belt")
	if !found {
		return fields
	}

	available := make(map[string]bool, len(store.Schema().Fields))
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

// movementBound supplies a deliberately broad positive fallback for legacy projections that omit
// bounds, preventing invalid zero-sized movement while trusted collision still constrains play.
func movementBound(value float64) float64 {
	if value <= 0 {
		return 1 << 20
	}

	return value
}

// movementRadius supplies a conservative positive collision radius when an older HUD omits it.
func movementRadius(value float64) float64 {
	if value <= 0 {
		return 1
	}

	return value
}

// classToken maps canonical class names to the legacy two-character composite tokens consumed by
// presentation assets; unknown classes remain empty and fail visibly rather than guessing.
func classToken(class string) string {
	return map[string]string{
		"Amazon":      "AM",
		"Sorceress":   "SO",
		"Necromancer": "NE",
		"Paladin":     "PA",
		"Barbarian":   "BA",
		"Druid":       "DZ",
		"Assassin":    "AI",
	}[class]
}
