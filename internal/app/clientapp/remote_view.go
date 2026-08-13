package clientapp

import (
	"fmt"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
)

// installRemoteView copies the authenticated, allowlisted client projection
// into the existing disposable presentation ECS. This ECS is never advanced
// while connected and is never sent back to the server; authority remains the
// remote session while established Lua presentation can keep reading its
// familiar component contracts.
func (app *application) installRemoteView(session *clientsession.Session) error {
	if app.entitySimulation == nil || session == nil {
		return nil
	}
	hud, projected := session.View()
	world := app.entitySimulation.World()
	controls, ok := akara.GetDynamicStore(world, "d2legacy.world.player_control")
	if !ok {
		return fmt.Errorf("remote presentation: player control store is unavailable")
	}
	var hero akara.Entity
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner == "local-player" {
			hero = entity
			break
		}
	}
	if hero == 0 {
		return fmt.Errorf("remote presentation: local hero mirror is unavailable")
	}
	updates := map[string]map[string]any{
		"d2legacy.player.identity":     {"character_id": hud.Player.CharacterID, "player": "local-player", "name": hud.Player.Name, "class": hud.Player.Class},
		"d2legacy.player.vitals":       {"health": hud.Vitals.Health, "max_health": hud.Vitals.MaxHealth, "mana": hud.Vitals.Mana, "max_mana": hud.Vitals.MaxMana, "mana_raw": hud.Vitals.Mana * 256, "max_mana_raw": hud.Vitals.MaxMana * 256},
		"d2legacy.player.progress":     {"level": hud.Progress.Level, "experience": hud.Progress.Experience, "unspent_skill_points": hud.Progress.UnspentSkillPoints},
		"d2legacy.player.combat_stats": {"attack_rating": hud.Combat.AttackRating, "defense": hud.Combat.Defense},
		"d2legacy.world.position":      {"x": hud.Position.X, "y": hud.Position.Y},
		"d2legacy.world.location":      {"act": hud.Location.Act, "level_id": hud.Location.LevelID},
		"d2legacy.player.appearance":   {"cof": "", "token": classToken(hud.Player.Class), "palette": "data/global/Palette/units/pal.dat", "weapon_class": "HTH"},
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
	app.activateWorld(int(hud.Location.LevelID))
	if app.remoteMirrors == nil {
		app.remoteMirrors = map[string]akara.Entity{}
	}
	seen := map[string]bool{}
	for _, remote := range projected.Entities {
		if remote.Kind != "player" {
			continue
		}
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
		values := map[string]map[string]any{
			"d2legacy.player.identity":   {"character_id": remote.ID, "player": remote.Owner, "name": remote.Label, "class": remote.Class},
			"d2legacy.player.appearance": {"cof": "", "token": remote.Token, "palette": "data/global/Palette/units/pal.dat", "weapon_class": "HTH"},
			"d2legacy.player.animation":  {"direction": remote.Direction, "mode": remote.Mode},
			"d2legacy.world.facing":      {"direction": remote.Direction, "directions": int64(16)},
			"d2legacy.world.position":    {"x": remote.Position.X, "y": remote.Position.Y},
			"d2legacy.world.location":    {"act": hud.Location.Act, "level_id": hud.Location.LevelID},
		}
		for name, fields := range values {
			store, ok := akara.GetDynamicStore(world, name)
			if ok {
				if _, err := store.Set(entity, fields); err != nil {
					return err
				}
			}
		}
	}
	for id, entity := range app.remoteMirrors {
		if !seen[id] {
			world.DestroyEntity(entity)
			delete(app.remoteMirrors, id)
		}
	}
	return nil
}

func classToken(class string) string {
	return map[string]string{"Amazon": "AM", "Sorceress": "SO", "Necromancer": "NE", "Paladin": "PA", "Barbarian": "BA", "Druid": "DZ", "Assassin": "AI"}[class]
}
