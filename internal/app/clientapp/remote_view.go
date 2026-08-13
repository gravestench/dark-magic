package clientapp

import (
	"fmt"
	"log/slog"
	"math"
	"sort"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
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
	var stale []akara.Entity
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner != "local-player" && owner != hud.Player.PlayerID {
			continue
		}
		if hero == 0 {
			// Preserve the entity handle already cached by Lua gameplay.world.
			// The authenticated HUD below replaces all character-specific fields.
			hero = entity
			continue
		}
		// Character selection may happen after the offline fixture admitted an
		// earlier roster entry. Connected presentation has exactly one local
		// owner; destroy stale mirrors so Lua cannot bind whichever appears first.
		stale = append(stale, entity)
	}
	for _, entity := range stale {
		world.DestroyEntity(entity)
	}
	if control, found := controls.Get(hero); found {
		if err := control.Set("player", hud.Player.PlayerID); err != nil {
			return err
		}
	}
	if app.movementSource != nil {
		app.movementSource.SetPlayer(hud.Player.PlayerID)
	}
	if hero == 0 {
		return fmt.Errorf("remote presentation: local hero mirror is unavailable")
	}
	updates := map[string]map[string]any{
		"d2legacy.player.identity":     {"character_id": hud.Player.CharacterID, "player": hud.Player.PlayerID, "name": hud.Player.Name, "class": hud.Player.Class},
		"d2legacy.player.vitals":       {"health": hud.Vitals.Health, "max_health": hud.Vitals.MaxHealth, "mana": hud.Vitals.Mana, "max_mana": hud.Vitals.MaxMana, "mana_raw": hud.Vitals.Mana * 256, "max_mana_raw": hud.Vitals.MaxMana * 256},
		"d2legacy.player.progress":     {"level": hud.Progress.Level, "experience": hud.Progress.Experience, "unspent_skill_points": hud.Progress.UnspentSkillPoints},
		"d2legacy.player.combat_stats": {"attack_rating": hud.Combat.AttackRating, "defense": hud.Combat.Defense},
		"d2legacy.player.animation":    {"direction": hud.Animation.Direction, "mode": hud.Animation.Mode},
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
	if err := moveMirrorToward(world, hero, hud.Position, 0.35); err != nil {
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
		if err := installPlayerMirror(world, entity, remote, hud.Location); err != nil {
			return err
		}
	}
	for id, entity := range app.remoteMirrors {
		if !seen[id] {
			world.DestroyEntity(entity)
			delete(app.remoteMirrors, id)
		}
	}
	app.logNetworkRoster(hud)
	return nil
}

func (app *application) logNetworkRoster(hud playeradapter.HUD) {
	world := app.entitySimulation.World()
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
	key := fmt.Sprintf("%s:%s:%v", hud.Player.PlayerID, hud.Player.Class, entries)
	if key == app.networkRosterLogKey {
		return
	}
	app.networkRosterLogKey = key
	slog.Debug("connected player presentation roster", "authenticated_player", hud.Player.PlayerID,
		"authenticated_class", hud.Player.Class, "entities", entries)
}

func installPlayerMirror(world *akara.World, entity akara.Entity, player playeradapter.WorldEntity, location playeradapter.HUDLocation) error {
	values := map[string]map[string]any{
		"d2legacy.player.identity":   {"character_id": player.ID, "player": player.Owner, "name": player.Label, "class": player.Class},
		"d2legacy.player.appearance": {"cof": "", "token": player.Token, "palette": "data/global/Palette/units/pal.dat", "weapon_class": "HTH"},
		"d2legacy.player.animation":  {"direction": player.Direction, "mode": player.Mode},
		"d2legacy.world.facing":      {"direction": player.Direction, "directions": int64(16)},
		"d2legacy.world.location":    {"act": location.Act, "level_id": location.LevelID},
	}
	if err := moveMirrorToward(world, entity, player.Position, 0.35); err != nil {
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
