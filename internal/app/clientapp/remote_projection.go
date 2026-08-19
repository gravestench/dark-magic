package clientapp

import (
	"fmt"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const playerPalettePath = "data/global/Palette/units/pal.dat"

// installRemoteView copies the latest authenticated presentation snapshot.
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

// installRemoteProjection updates the disposable ECS used by connected presentation.
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

// authenticatedHero preserves one local-owner entity and removes stale duplicates.
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

// remoteHeroComponents translates the authenticated HUD into allowlisted ECS fields.
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

// remoteVitalsFields maps HUD vitals without exposing authority-only state.
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

// remoteProgressFields maps level progression shown by the connected HUD.
func remoteProgressFields(progress playeradapter.HUDProgress) map[string]any {
	return map[string]any{
		"level":                progress.Level,
		"experience":           progress.Experience,
		"unspent_skill_points": progress.UnspentSkillPoints,
	}
}

// remoteCombatFields maps the allowlisted attack and defense totals.
func remoteCombatFields(combat playeradapter.HUDCombat) map[string]any {
	return map[string]any{
		"attack_rating": combat.AttackRating,
		"defense":       combat.Defense,
	}
}

// remoteAnimationFields maps the authoritative animation state.
func remoteAnimationFields(animation playeradapter.HUDAnimation) map[string]any {
	return map[string]any{
		"direction":  animation.Direction,
		"mode":       animation.Mode,
		"start_tick": int64(animation.StartTick),
	}
}

// remoteMovementFields maps presentation-safe movement modifiers.
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

// remoteSkillAssignmentFields maps the active left and right skill slots.
func remoteSkillAssignmentFields(skills playeradapter.HUDSkills) map[string]any {
	return map[string]any{
		"left":  skills.Left,
		"right": skills.Right,
	}
}

// setRemoteComponents applies one allowlisted component batch to an entity.
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

// refreshPrivateProjection rebuilds owner-private entities only when content changes.
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

// followRemoteHero keeps frozen offline camera entities attached to the hero.
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

// partyViewFields fills the fixed Lua-facing party roster shape.
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

// setPartyRosterSlot initializes one fixed party slot and optional member.
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

// beltFields includes only belt slots supported by the registered schema.
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

// movementBound supplies the legacy presentation fallback for absent bounds.
func movementBound(value float64) float64 {
	if value <= 0 {
		return 1 << 20
	}

	return value
}

// movementRadius supplies the legacy presentation fallback for absent radius.
func movementRadius(value float64) float64 {
	if value <= 0 {
		return 1
	}

	return value
}

// classToken maps save classes to their composite animation tokens.
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
