package clientapp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// remoteRosterEntry captures enough stable identity for useful diagnostics while keeping the
// authority projection and ECS components private to the presentation layer.
type remoteRosterEntry struct {
	Entity uint64
	Owner  string
	Class  string
	Token  string
	X      float64
	Y      float64
}

// syncRemoteMirrors reconciles structural membership without moving retained entities. Snapshot
// arrival may change identity immediately, but transforms remain owned by frame interpolation.
func (app *application) syncRemoteMirrors(
	projected []playeradapter.WorldEntity,
	location playeradapter.HUDLocation,
) error {
	if app.clientSimulation == nil {
		return nil
	}

	app.ensureRemoteMirrorMaps()
	world := app.clientSimulation.World()
	seen := make(map[string]bool, len(projected))

	for _, remote := range projected {
		seen[remote.ID] = true

		if err := app.upsertRemoteMirror(world, remote, location); err != nil {
			return err
		}
	}

	app.removeStaleRemoteMirrors(world, seen)

	return nil
}

// ensureRemoteMirrorMaps initializes entity ownership and structural fingerprints together. The two
// maps form one cache: one preserves ECS identity and the other suppresses needless rebuilds.
func (app *application) ensureRemoteMirrorMaps() {
	if app.remoteMirrors == nil {
		app.remoteMirrors = make(map[string]akara.Entity)
	}

	if app.remoteMirrorKeys == nil {
		app.remoteMirrorKeys = make(map[string]string)
	}
}

// upsertRemoteMirror retains ECS identity for a public ID and rebuilds components only when its
// position-free fingerprint changes. Transform updates therefore cannot restart animation graphs.
func (app *application) upsertRemoteMirror(
	world *akara.World,
	remote playeradapter.WorldEntity,
	location playeradapter.HUDLocation,
) error {
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

	if exists && app.remoteMirrorKeys[remote.ID] == key {
		return nil
	}

	if err := installWorldMirror(world, entity, remote, location, false); err != nil {
		return err
	}

	app.remoteMirrorKeys[remote.ID] = key

	return nil
}

// removeStaleRemoteMirrors treats the projected roster as complete and removes both entity and
// fingerprint ownership, preventing stale units from surviving after despawn or level changes.
func (app *application) removeStaleRemoteMirrors(
	world *akara.World,
	seen map[string]bool,
) {
	for id, entity := range app.remoteMirrors {
		if seen[id] {
			continue
		}

		world.DestroyEntity(entity)
		delete(app.remoteMirrors, id)
		delete(app.remoteMirrorKeys, id)
	}
}

// applySampledWorldPositions is the sole writer of peer transforms after creation. This separation
// ensures packet arrival cannot bypass interpolation and produce visible correction jumps.
func (app *application) applySampledWorldPositions(
	projected []playeradapter.WorldEntity,
) error {
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

// applyLocalPredictedPosition selects the entity through authenticated player_control ownership, not
// roster order or character label. Local prediction can therefore never move a peer mirror.
func (app *application) applyLocalPredictedPosition(
	playerID string,
	predicted playeradapter.HUDPosition,
) error {
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

		if err := moveRemoteCameras(world, predicted); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("remote presentation: authenticated player %q is unavailable", playerID)
}

// moveRemoteCameras advances frozen camera-follow entities with the predicted owner because the
// offline camera systems are not authoritative and do not run in the connected replica.
func moveRemoteCameras(world *akara.World, position playeradapter.HUDPosition) error {
	follows, found := akara.GetDynamicStore(world, "d2legacy.world.camera_follow")
	if !found {
		return nil
	}

	for _, camera := range follows.Entities() {
		if err := moveMirrorToward(world, camera, position, 1); err != nil {
			return err
		}
	}

	return nil
}

// applyAnimationTimeline advances the owner's animation on the low-latency prediction clock and
// peers on the delayed interpolation clock. Using one clock would make either input or peers jitter.
func (app *application) applyAnimationTimeline(
	localPlayer string,
	timeline networkclock.Timeline,
	step time.Duration,
) error {
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
		if err := setRemoteAnimationClock(
			entity,
			localPlayer,
			timeline,
			step,
			animations,
			clocks,
			identities,
		); err != nil {
			return err
		}
	}

	return nil
}

// setRemoteAnimationClock derives elapsed seconds from the entity's authority-provided start tick
// and the timeline appropriate to its owner. Negative elapsed time is clamped during reordering.
func setRemoteAnimationClock(
	entity akara.Entity,
	localPlayer string,
	timeline networkclock.Timeline,
	step time.Duration,
	animations *akara.DynamicStore,
	clocks *akara.DynamicStore,
	identities *akara.DynamicStore,
) error {
	animation, animationFound := animations.Get(entity)

	identity, identityFound := identities.Get(entity)
	if !animationFound || !identityFound {
		return nil
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
	_, err := clocks.Set(entity, map[string]any{"seconds": seconds})

	return err
}

// remotePresentationFingerprint removes position before hashing so movement cannot trigger
// structural ECS writes. Location remains included because crossing worlds changes presentation.
func remotePresentationFingerprint(
	entity playeradapter.WorldEntity,
	location playeradapter.HUDLocation,
) (string, error) {
	entity = cloneWorldEntity(entity)
	entity.Position = playeradapter.HUDPosition{}

	payload, err := json.Marshal(struct {
		Entity   playeradapter.WorldEntity `json:"entity"`
		Location playeradapter.HUDLocation `json:"location"`
	}{
		Entity:   entity,
		Location: location,
	})
	if err != nil {
		return "", fmt.Errorf("remote presentation: fingerprint world entity: %w", err)
	}

	return string(payload), nil
}

// installWorldMirror dispatches the allowlisted public structure by kind. Corpses retain selection
// identity for applicable skills but explicitly lose living collision.
func installWorldMirror(
	world *akara.World,
	entity akara.Entity,
	value playeradapter.WorldEntity,
	location playeradapter.HUDLocation,
	snap bool,
) error {
	if value.Kind == "player" {
		return installPlayerMirror(world, entity, value, location, snap)
	}

	values := remoteMonsterComponents(value, location)
	if value.Kind == "corpse" {
		// Corpses remain selectable for skills but lose living collision.
		if colliders, found := akara.GetDynamicStore(world, "d2legacy.world.collider"); found {
			colliders.Remove(entity)
		}
	} else {
		values["d2legacy.world.collider"] = map[string]any{"radius": value.Radius}
	}

	if err := moveMirrorToward(world, entity, value.Position, correctionAlpha(snap)); err != nil {
		return err
	}

	return setRemoteComponents(world, entity, values)
}

// remoteMonsterComponents fills schema-required gameplay fields with neutral values while mapping
// only public appearance, health, location, and selection data from authority.
func remoteMonsterComponents(
	value playeradapter.WorldEntity,
	location playeradapter.HUDLocation,
) map[string]map[string]any {
	return map[string]map[string]any{
		"d2legacy.monster.identity": {
			"spawn_id":       value.SpawnID,
			"definition_id":  value.DefinitionID,
			"base_id":        value.DefinitionID,
			"graphics_id":    value.Token,
			"seed":           "",
			"treasure_class": "",
		},
		"d2legacy.monster.appearance": {
			"token":          value.Token,
			"mode":           value.Mode,
			"weapon_class":   monsterWeaponClass(value.WeaponClass),
			"name_key":       value.Label,
			"components":     value.Components,
			"death_sound":    value.DeathSound,
			"overlay_height": value.OverlayHeight,
		},
		"d2legacy.monster.stats": {
			"level":         int64(1),
			"health":        pointed(value.Health),
			"max_health":    pointed(value.MaxHealth),
			"defense":       int64(0),
			"attack_rating": int64(0),
			"physical_min":  int64(0),
			"physical_max":  int64(0),
			"experience":    int64(0),
		},
		"d2legacy.world.velocity": {
			"x": float64(0),
			"y": float64(0),
		},
		"d2legacy.world.facing": {
			"direction":  value.Direction,
			"directions": int64(16),
		},
		"d2legacy.world.location": {
			"act":      worldLocation(value.Act, location.Act),
			"level_id": worldLocation(value.LevelID, location.LevelID),
		},
		"d2legacy.world.selectable": {
			"id":       value.ID,
			"kind":     value.Kind,
			"label":    value.Label,
			"owner":    value.Owner,
			"radius":   value.Radius,
			"priority": value.Priority,
		},
	}
}

// installPlayerMirror installs only public peer identity, appearance, animation, movement modifiers,
// and location. Private owner HUD and inventory data use a separate projection path.
func installPlayerMirror(
	world *akara.World,
	entity akara.Entity,
	player playeradapter.WorldEntity,
	location playeradapter.HUDLocation,
	snap bool,
) error {
	values := map[string]map[string]any{
		"d2legacy.player.identity": {
			"character_id": player.ID,
			"player":       player.Owner,
			"name":         player.Label,
			"class":        player.Class,
		},
		"d2legacy.player.appearance": {
			"cof":          "",
			"token":        player.Token,
			"palette":      playerPalettePath,
			"weapon_class": "HTH",
		},
		"d2legacy.player.animation": {
			"direction":  player.Direction,
			"mode":       player.Mode,
			"start_tick": int64(player.AnimationStartTick),
		},
		"d2legacy.player.movement_stats": {
			"velocitypercent":         player.VelocityPercent,
			"item_fastermovevelocity": player.ItemFasterMoveVelocity,
		},
		"d2legacy.presentation.animation_clock": {
			"seconds": float64(0),
		},
		"d2legacy.world.facing": {
			"direction":  player.Direction,
			"directions": int64(16),
		},
		"d2legacy.world.location": {
			"act":      worldLocation(player.Act, location.Act),
			"level_id": worldLocation(player.LevelID, location.LevelID),
		},
	}

	if err := moveMirrorToward(world, entity, player.Position, correctionAlpha(snap)); err != nil {
		return err
	}

	return setRemoteComponents(world, entity, values)
}

// pointed converts an omitted optional public value to the registered schema's neutral integer.
func pointed(value *int64) int64 {
	if value == nil {
		return 0
	}

	return *value
}

// worldLocation uses the entity's explicit location when available and otherwise inherits the
// snapshot HUD location, supporting older projections without fabricating a different world.
func worldLocation(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}

	return value
}

// monsterWeaponClass supplies the composite-animation system's unarmed token when authority omits a
// weapon class.
func monsterWeaponClass(value string) string {
	if value == "" {
		return "HTH"
	}

	return value
}

// logNetworkRoster keys logs on stable presentation identity instead of position, preventing normal
// movement from flooding diagnostics while still surfacing roster and appearance changes.
func (app *application) logNetworkRoster(hud playeradapter.HUD) {
	entries, found := app.networkRoster()
	if !found {
		return
	}

	key := hud.Player.PlayerID + ":" + hud.Player.Class
	for _, entry := range entries {
		key += fmt.Sprintf(":%s:%s:%s", entry.Owner, entry.Class, entry.Token)
	}

	if key == app.networkRosterLogKey {
		return
	}

	app.networkRosterLogKey = key

	slog.Debug(
		"connected player presentation roster",
		"authenticated_player", hud.Player.PlayerID,
		"authenticated_class", hud.Player.Class,
		"entities", entries,
	)
}

// networkRoster joins identity, appearance, and position only for complete player mirrors, then
// sorts by entity ID so debug output and change keys remain deterministic.
func (app *application) networkRoster() ([]remoteRosterEntry, bool) {
	world := app.clientSimulation.World()
	identities, identityOK := akara.GetDynamicStore(world, "d2legacy.player.identity")
	appearances, appearanceOK := akara.GetDynamicStore(world, "d2legacy.player.appearance")
	positions, positionOK := akara.GetDynamicStore(world, "d2legacy.world.position")

	if !identityOK || !appearanceOK || !positionOK {
		return nil, false
	}

	entries := make([]remoteRosterEntry, 0, identities.Len())

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

		entries = append(entries, remoteRosterEntry{
			Entity: uint64(entity),
			Owner:  owner.(string),
			Class:  class.(string),
			Token:  token.(string),
			X:      x.(float64),
			Y:      y.(float64),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Entity < entries[j].Entity
	})

	return entries, true
}

// correctionAlpha permits snapping only for initial construction. Retained mirrors receive zero
// alpha because frame interpolation exclusively owns their transforms.
func correctionAlpha(snap bool) float64 {
	if snap {
		return 1
	}

	// Frame interpolation is the sole transform writer between initial snapshots.
	return 0
}

// currentPosition returns the schema's neutral position when a mirror has not yet been initialized,
// allowing moveMirrorToward to own creation behavior.
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

// moveMirrorToward creates missing transforms, ignores zero-alpha correction writes, and snaps errors
// over four subtiles rather than visibly easing across invalid geometry.
func moveMirrorToward(
	world *akara.World,
	entity akara.Entity,
	target playeradapter.HUDPosition,
	alpha float64,
) error {
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
	x := xValue.(float64)
	y := yValue.(float64)

	if math.Hypot(target.X-x, target.Y-y) > 4 {
		x = target.X
		y = target.Y
	} else {
		x += (target.X - x) * alpha
		y += (target.Y - y) * alpha
	}

	if err := position.Set("x", x); err != nil {
		return err
	}

	return position.Set("y", y)
}
