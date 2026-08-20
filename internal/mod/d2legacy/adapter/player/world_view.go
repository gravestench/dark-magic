package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const WorldViewVersion uint32 = 5

const (
	WorldViewRadius         = 80.0
	MaxWorldViewEntities    = 256
	MaxWorldViewMissiles    = 512
	MaxWorldViewStates      = 512
	maxWorldIDBytes         = 128
	maxWorldKindBytes       = 32
	maxWorldLabelBytes      = 256
	maxWorldAssetBytes      = 512
	maxWorldComponentsBytes = 4096
	maxWorldStateBytes      = 128
)

var ErrWorldView = errors.New("player world view: invalid public projection")

type WorldView struct {
	Version   uint32         `json:"version"`
	Tick      uint64         `json:"tick"`
	Origin    HUDPosition    `json:"origin"`
	Entities  []WorldEntity  `json:"entities"`
	Missiles  []WorldMissile `json:"missiles"`
	States    []WorldState   `json:"states"`
	Truncated bool           `json:"truncated"`
}

// WorldState is a persistent presentation-safe relationship. The target and
// state vocabulary are sufficient to draw the effect; source identity, stats,
// range, party/filter decisions, and arbitration remain server-only.
type WorldState struct {
	TargetID    string `json:"target_id"`
	StateID     string `json:"state_id"`
	PeriodTicks int64  `json:"period_ticks"`
	distance2   float64
}

type worldStateKey struct {
	targetID string
	stateID  string
}

// WorldMissile is the bounded visual subset of an authoritative projectile or
// aftermath entity. It deliberately omits damage, collision, ownership, hit
// locks, and lifetime policy: a connected client can render these values but
// cannot reconstruct or advance gameplay from them.
type WorldMissile struct {
	ID               string      `json:"id"`
	Kind             string      `json:"kind"`
	MissileID        string      `json:"missile_id"`
	Position         HUDPosition `json:"position"`
	Velocity         HUDPosition `json:"velocity"`
	Act              int64       `json:"act"`
	LevelID          int64       `json:"level_id"`
	DCC              string      `json:"dcc"`
	Palette          string      `json:"palette,omitempty"`
	LogicalDirection int64       `json:"logical_direction"`
	Directions       int64       `json:"directions"`
	FramesPerSecond  int64       `json:"frames_per_second"`
	Loop             bool        `json:"loop"`
	TransparencyMode int64       `json:"transparency_mode"`
	OffsetX          float64     `json:"offset_x"`
	OffsetY          float64     `json:"offset_y"`
	OffsetZ          float64     `json:"offset_z"`
	distance2        float64
}

type WorldEntity struct {
	ID                     string      `json:"id"`
	Kind                   string      `json:"kind"`
	Label                  string      `json:"label,omitempty"`
	Owner                  string      `json:"owner,omitempty"`
	Position               HUDPosition `json:"position"`
	Radius                 float64     `json:"radius"`
	Priority               int64       `json:"priority"`
	Health                 *int64      `json:"health,omitempty"`
	MaxHealth              *int64      `json:"max_health,omitempty"`
	Class                  string      `json:"class,omitempty"`
	SpawnID                string      `json:"spawn_id,omitempty"`
	DefinitionID           string      `json:"definition_id,omitempty"`
	Token                  string      `json:"token,omitempty"`
	Mode                   string      `json:"mode,omitempty"`
	WeaponClass            string      `json:"weapon_class,omitempty"`
	Components             string      `json:"components,omitempty"`
	DeathSound             string      `json:"death_sound,omitempty"`
	OverlayHeight          int64       `json:"overlay_height,omitempty"`
	Direction              int64       `json:"direction,omitempty"`
	AnimationStartTick     uint64      `json:"animation_start_tick,omitempty"`
	VelocityPercent        int64       `json:"velocitypercent,omitempty"`
	ItemFasterMoveVelocity int64       `json:"item_fastermovevelocity,omitempty"`
	Act                    int64       `json:"act,omitempty"`
	LevelID                int64       `json:"level_id,omitempty"`
	distance2              float64
}

// ProjectWorldView exposes only nearby entities carrying the mod's explicit
// public selectable contract. Raw ECS identity and non-allowlisted components
// never enter the result.
func ProjectWorldView(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	projection, err := newWorldProjection(playerID, checkpoint)
	if err != nil {
		return nil, err
	}

	if err := projection.projectSelectableEntities(); err != nil {
		return nil, err
	}

	if err := projection.projectCorpses(); err != nil {
		return nil, err
	}

	if err := projection.finish(); err != nil {
		return nil, err
	}

	return json.Marshal(projection.view)
}

type worldProjection struct {
	snapshot           gameecs.Snapshot
	playerID           string
	playerEntity       uint64
	origin             HUDPosition
	originAct          int64
	originLevel        int64
	selectables        gameecs.ComponentSnapshot
	inactive           gameecs.ComponentSnapshot
	positions          gameecs.ComponentSnapshot
	locations          gameecs.ComponentSnapshot
	monsters           gameecs.ComponentSnapshot
	monsterIdentities  gameecs.ComponentSnapshot
	monsterAppearances gameecs.ComponentSnapshot
	monsterDeaths      gameecs.ComponentSnapshot
	monsterAI          gameecs.ComponentSnapshot
	velocities         gameecs.ComponentSnapshot
	players            gameecs.ComponentSnapshot
	appearances        gameecs.ComponentSnapshot
	animations         gameecs.ComponentSnapshot
	movementStats      gameecs.ComponentSnapshot
	facings            gameecs.ComponentSnapshot
	view               WorldView
	publicIDs          map[uint64]string
	seen               map[string]struct{}
}

// newWorldProjection authenticates the owner and captures every component used
// by the public allowlist. Required spatial state fails before any partial view.
func newWorldProjection(playerID string, checkpoint simulation.Checkpoint) (*worldProjection, error) {
	if checkpoint.Snapshot == nil || strings.TrimSpace(playerID) == "" {
		return nil, ErrWorldView
	}

	snapshot := *checkpoint.Snapshot

	identities, found := findComponent(snapshot, "d2legacy.player.identity")
	if !found {
		return nil, ErrWorldView
	}

	playerEntity, _, found := findString(identities, "player", playerID)
	if !found {
		return nil, ErrWorldView
	}

	positions, found := findComponent(snapshot, "d2legacy.world.position")
	if !found {
		return nil, ErrWorldView
	}

	originFields, found := findInstance(positions, playerEntity)
	if !found {
		return nil, ErrWorldView
	}

	locations, found := findComponent(snapshot, "d2legacy.world.location")
	if !found {
		return nil, ErrWorldView
	}

	originLocation, found := findInstance(locations, playerEntity)
	if !found {
		return nil, ErrWorldView
	}

	projection := &worldProjection{
		snapshot:     snapshot,
		playerID:     playerID,
		playerEntity: playerEntity,
		origin:       HUDPosition{X: floatField(originFields, "x"), Y: floatField(originFields, "y")},
		originAct:    intField(originLocation, "act"),
		originLevel:  intField(originLocation, "level_id"),
		positions:    positions,
		locations:    locations,
		publicIDs:    map[uint64]string{playerEntity: "player:" + playerID},
		seen:         make(map[string]struct{}),
	}
	projection.loadOptionalComponents()
	projection.view = WorldView{
		Version:  WorldViewVersion,
		Tick:     checkpoint.Tick,
		Origin:   projection.origin,
		Entities: []WorldEntity{},
		Missiles: []WorldMissile{},
		States:   []WorldState{},
	}
	projection.indexPublicIDs()

	return projection, nil
}

// loadOptionalComponents centralizes the public projection's schema inventory.
// Missing optional components remain empty snapshots and naturally add no data.
func (projection *worldProjection) loadOptionalComponents() {
	projection.selectables, _ = findComponent(projection.snapshot, "d2legacy.world.selectable")
	projection.inactive, _ = findComponent(projection.snapshot, "d2legacy.world.inactive")
	projection.monsters, _ = findComponent(projection.snapshot, "d2legacy.monster.stats")
	projection.monsterIdentities, _ = findComponent(projection.snapshot, "d2legacy.monster.identity")
	projection.monsterAppearances, _ = findComponent(projection.snapshot, "d2legacy.monster.appearance")
	projection.monsterDeaths, _ = findComponent(projection.snapshot, "d2legacy.monster.death")
	projection.monsterAI, _ = findComponent(projection.snapshot, "d2legacy.monster.ai")
	projection.velocities, _ = findComponent(projection.snapshot, "d2legacy.world.velocity")
	projection.players, _ = findComponent(projection.snapshot, "d2legacy.player.identity")
	projection.appearances, _ = findComponent(projection.snapshot, "d2legacy.player.appearance")
	projection.animations, _ = findComponent(projection.snapshot, "d2legacy.player.animation")
	projection.movementStats, _ = findComponent(projection.snapshot, "d2legacy.player.movement_stats")
	projection.facings, _ = findComponent(projection.snapshot, "d2legacy.world.facing")
}

// indexPublicIDs records the only ECS entities eligible to become state-effect
// targets. Stable public IDs keep raw ECS handles out of the client protocol.
func (projection *worldProjection) indexPublicIDs() {
	for _, instance := range projection.selectables.Instances {
		fields, found := findInstance(projection.selectables, instance.Entity)
		if !found {
			continue
		}

		if id := stringField(fields, "id"); id != "" {
			projection.publicIDs[instance.Entity] = id
		}
	}
}

// projectSelectableEntities filters authoritative selectable entities by owner
// location and radius, then decorates only explicitly public presentation data.
func (projection *worldProjection) projectSelectableEntities() error {
	for _, instance := range projection.selectables.Instances {
		if instance.Entity == projection.playerEntity {
			continue
		}

		if _, dormant := findInstance(projection.inactive, instance.Entity); dormant {
			continue
		}

		entity, visible := projection.selectableEntity(instance.Entity)
		if !visible {
			continue
		}

		if _, duplicate := projection.seen[entity.ID]; duplicate {
			return fmt.Errorf("%w: duplicate public ID %q", ErrWorldView, entity.ID)
		}
		// Duplicate IDs remain invalid even when the later entity is outside the
		// radius; identity uniqueness is a world contract, not a view detail.
		projection.seen[entity.ID] = struct{}{}
		if entity.distance2 > WorldViewRadius*WorldViewRadius {
			continue
		}

		projection.decorateEntity(&entity, instance.Entity)

		if err := validateWorldEntity(entity); err != nil {
			return err
		}

		projection.view.Entities = append(projection.view.Entities, entity)
	}

	return nil
}

// selectableEntity builds the public base fields and applies the owner's level
// boundary before more expensive presentation joins.
func (projection *worldProjection) selectableEntity(source uint64) (WorldEntity, bool) {
	public, publicFound := findInstance(projection.selectables, source)
	position, positioned := findInstance(projection.positions, source)

	location, located := findInstance(projection.locations, source)
	if !publicFound || !positioned || !located {
		return WorldEntity{}, false
	}

	entity := WorldEntity{
		ID:       stringField(public, "id"),
		Kind:     stringField(public, "kind"),
		Label:    stringField(public, "label"),
		Owner:    stringField(public, "owner"),
		Position: HUDPosition{X: floatField(position, "x"), Y: floatField(position, "y")},
		Radius:   floatField(public, "radius"),
		Priority: intField(public, "priority"),
		Act:      intField(location, "act"),
		LevelID:  intField(location, "level_id"),
	}
	if entity.Act != projection.originAct || entity.LevelID != projection.originLevel {
		return WorldEntity{}, false
	}

	dx := entity.Position.X - projection.origin.X
	dy := entity.Position.Y - projection.origin.Y
	entity.distance2 = dx*dx + dy*dy

	return entity, true
}

// decorateEntity adds public monster or player presentation fields. Gameplay
// policy such as AI goals, damage, and collision never enters the result.
func (projection *worldProjection) decorateEntity(entity *WorldEntity, source uint64) {
	if stats, found := findInstance(projection.monsters, source); found {
		health := intField(stats, "health")
		maximum := intField(stats, "max_health")
		entity.Health, entity.MaxHealth = &health, &maximum
	}

	projection.decorateMonster(entity, source)

	if appearance, found := findInstance(projection.appearances, source); found {
		entity.Token = stringField(appearance, "token")
		entity.Mode = stringField(appearance, "mode")
	}

	if facing, found := findInstance(projection.facings, source); found {
		entity.Direction = intField(facing, "direction")
	}

	if identity, found := findInstance(projection.players, source); found {
		entity.Class = stringField(identity, "class")
		projection.decoratePlayerMotion(entity, source)
	}
}

// decorateMonster reduces private monster authority to its visual identity and
// derives an animation mode using the established presentation precedence.
func (projection *worldProjection) decorateMonster(entity *WorldEntity, source uint64) {
	if identity, found := findInstance(projection.monsterIdentities, source); found {
		entity.SpawnID = stringField(identity, "spawn_id")
		entity.DefinitionID = stringField(identity, "definition_id")
	}

	appearance, found := findInstance(projection.monsterAppearances, source)
	if !found {
		return
	}

	if nameKey := stringField(appearance, "name_key"); nameKey != "" {
		entity.Label = nameKey
	}

	entity.Token = stringField(appearance, "token")
	entity.Mode = stringField(appearance, "mode")
	entity.WeaponClass = stringField(appearance, "weapon_class")
	entity.Components = stringField(appearance, "components")
	entity.DeathSound = stringField(appearance, "death_sound")
	entity.OverlayHeight = intField(appearance, "overlay_height")
	brain, _ := findInstance(projection.monsterAI, source)
	velocity, _ := findInstance(projection.velocities, source)
	entity.Mode = monsterPresentationMode(
		entity.Mode,
		stringField(brain, "state"),
		floatField(velocity, "x"),
		floatField(velocity, "y"),
	)
}

// decoratePlayerMotion publishes animation timing and speed inputs needed for
// deterministic rendering while leaving movement authority on the server.
func (projection *worldProjection) decoratePlayerMotion(entity *WorldEntity, source uint64) {
	if animation, found := findInstance(projection.animations, source); found {
		entity.Mode = stringField(animation, "mode")
		entity.AnimationStartTick = uint64(max(0, intField(animation, "start_tick")))
	}

	if movement, found := findInstance(projection.movementStats, source); found {
		entity.VelocityPercent = intField(movement, "velocitypercent")
		entity.ItemFasterMoveVelocity = intField(movement, "item_fastermovevelocity")
	}
}

// projectCorpses restores non-selectable dead monsters to the visual stream.
// Loot, XP, corpse usability, and attribution remain server-only.
func (projection *worldProjection) projectCorpses() error {
	for _, instance := range projection.monsterDeaths.Instances {
		if _, dormant := findInstance(projection.inactive, instance.Entity); dormant {
			continue
		}

		entity, found := projection.corpseEntity(instance.Entity)
		if !found {
			continue
		}

		if _, alreadyProjected := projection.seen[entity.ID]; alreadyProjected {
			continue
		}

		projection.decorateEntity(&entity, instance.Entity)

		if err := validateWorldEntity(entity); err != nil {
			return err
		}

		projection.seen[entity.ID] = struct{}{}
		if entity.distance2 <= WorldViewRadius*WorldViewRadius {
			projection.view.Entities = append(projection.view.Entities, entity)
		}
	}

	return nil
}

// corpseEntity reconstructs the public identity and location that death removes
// from selectable state, preserving the existing corpse rendering contract.
func (projection *worldProjection) corpseEntity(source uint64) (WorldEntity, bool) {
	identity, identified := findInstance(projection.monsterIdentities, source)
	position, positioned := findInstance(projection.positions, source)

	location, located := findInstance(projection.locations, source)
	if !identified || !positioned || !located {
		return WorldEntity{}, false
	}

	entity := WorldEntity{
		ID:       "monster:" + stringField(identity, "spawn_id"),
		Kind:     "corpse",
		Label:    stringField(identity, "definition_id"),
		Position: HUDPosition{X: floatField(position, "x"), Y: floatField(position, "y")},
		Act:      intField(location, "act"),
		LevelID:  intField(location, "level_id"),
	}
	if entity.Act != projection.originAct || entity.LevelID != projection.originLevel {
		return WorldEntity{}, false
	}

	dx := entity.Position.X - projection.origin.X
	dy := entity.Position.Y - projection.origin.Y
	entity.distance2 = dx*dx + dy*dy

	return entity, true
}

// finish deterministically orders and bounds every public collection before
// computing states from only the entities that survived visibility truncation.
func (projection *worldProjection) finish() error {
	sort.Slice(projection.view.Entities, func(i, j int) bool {
		if projection.view.Entities[i].distance2 != projection.view.Entities[j].distance2 {
			return projection.view.Entities[i].distance2 < projection.view.Entities[j].distance2
		}

		return projection.view.Entities[i].ID < projection.view.Entities[j].ID
	})

	if len(projection.view.Entities) > MaxWorldViewEntities {
		projection.view.Entities = projection.view.Entities[:MaxWorldViewEntities]
		projection.view.Truncated = true
	}

	missiles, valid := projectWorldMissiles(
		projection.snapshot,
		projection.origin,
		projection.originAct,
		projection.originLevel,
	)
	if !valid {
		return ErrWorldView
	}

	projection.view.Missiles = missiles
	if len(projection.view.Missiles) > MaxWorldViewMissiles {
		projection.view.Missiles = projection.view.Missiles[:MaxWorldViewMissiles]
		projection.view.Truncated = true
	}

	states, valid := projectWorldStates(
		projection.snapshot,
		projection.visibleTargets(),
		projection.origin,
		projection.originAct,
		projection.originLevel,
	)
	if !valid {
		return ErrWorldView
	}

	projection.view.States = states
	if len(projection.view.States) > MaxWorldViewStates {
		projection.view.States = projection.view.States[:MaxWorldViewStates]
		projection.view.Truncated = true
	}

	return nil
}

// visibleTargets maps only retained public IDs back to ECS entities, preventing
// state effects for truncated or distant entities from leaking into the view.
func (projection *worldProjection) visibleTargets() map[uint64]string {
	visibleIDs := map[string]struct{}{"player:" + projection.playerID: {}}
	for _, entity := range projection.view.Entities {
		visibleIDs[entity.ID] = struct{}{}
	}

	targets := make(map[uint64]string, len(projection.publicIDs))
	for entity, id := range projection.publicIDs {
		if _, visible := visibleIDs[id]; visible {
			targets[entity] = id
		}
	}

	return targets
}

// projectWorldStates joins persistent aura relationships to visible public IDs.
// Returning false rejects malformed or duplicate state instead of hiding drift.
func projectWorldStates(
	snapshot gameecs.Snapshot,
	publicIDs map[uint64]string,
	origin HUDPosition,
	originAct, originLevel int64,
) ([]WorldState, bool) {
	effects, exists := findComponent(snapshot, "d2legacy.skill.aura_effect")
	if !exists {
		return []WorldState{}, true
	}

	positions, positioned := findComponent(snapshot, "d2legacy.world.position")
	locations, located := findComponent(snapshot, "d2legacy.world.location")
	inactive, _ := findComponent(snapshot, "d2legacy.world.inactive")

	if !positioned || !located {
		return nil, false
	}

	result := make([]WorldState, 0, len(effects.Instances))

	seen := make(map[worldStateKey]struct{}, len(effects.Instances))
	for _, instance := range effects.Instances {
		fields, found := findInstance(effects, instance.Entity)
		if !found {
			return nil, false
		}

		target := entityField(fields, "target")
		targetID, public := publicIDs[target]
		position, hasPosition := findInstance(positions, target)

		location, hasLocation := findInstance(locations, target)
		if !public || !hasPosition || !hasLocation {
			continue
		}

		if _, dormant := findInstance(inactive, target); dormant {
			continue
		}

		if intField(location, "act") != originAct || intField(location, "level_id") != originLevel {
			continue
		}

		dx, dy := floatField(position, "x")-origin.X, floatField(position, "y")-origin.Y

		state := WorldState{
			TargetID: targetID, StateID: stringField(fields, "state_id"),
			PeriodTicks: intField(fields, "refresh_delay"), distance2: dx*dx + dy*dy,
		}
		if state.distance2 > WorldViewRadius*WorldViewRadius {
			continue
		}

		if err := validateWorldState(state); err != nil {
			return nil, false
		}

		key := worldStateKey{targetID: state.TargetID, stateID: state.StateID}
		if _, duplicate := seen[key]; duplicate {
			return nil, false
		}

		seen[key] = struct{}{}

		result = append(result, state)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].distance2 != result[j].distance2 {
			return result[i].distance2 < result[j].distance2
		}

		if result[i].TargetID != result[j].TargetID {
			return result[i].TargetID < result[j].TargetID
		}

		return result[i].StateID < result[j].StateID
	})

	return result, true
}

// validateWorldState bounds persistent visual effects before they reach client
// allocation and scheduling code.
func validateWorldState(state WorldState) error {
	if !boundedRequired(state.TargetID, maxWorldIDBytes) ||
		!boundedRequired(state.StateID, maxWorldStateBytes) ||
		state.PeriodTicks < 1 || state.PeriodTicks > math.MaxInt32 {
		return ErrWorldView
	}

	return nil
}

// monsterPresentationMode collapses server-only AI and velocity facts into
// the two-byte animation vocabulary already carried by the public transform
// stream. Its death -> attack -> movement -> authored precedence matches the
// offline monster snapshot adapter, so the client learns what to draw without
// learning why the monster chose it.
func monsterPresentationMode(authoredMode, aiState string, velocityX, velocityY float64) string {
	if authoredMode == "DT" {
		return authoredMode
	}

	if aiState == "attack" {
		return "A1"
	}

	if velocityX != 0 || velocityY != 0 {
		return "WL"
	}

	return authoredMode
}

// validateWorldEntity enforces the public schema's string, numeric, and corpse
// invariants so untrusted projections cannot create invalid presentation state.
func validateWorldEntity(entity WorldEntity) error {
	if entity.ID == "" || entity.Kind == "" ||
		len(entity.ID) > maxWorldIDBytes || len(entity.Kind) > maxWorldKindBytes ||
		len(entity.Label) > maxWorldLabelBytes || len(entity.Owner) > maxWorldIDBytes ||
		len(entity.SpawnID) > maxWorldIDBytes || len(entity.DefinitionID) > maxWorldIDBytes ||
		len(entity.Token) > maxWorldAssetBytes || len(entity.Mode) > maxWorldKindBytes ||
		len(entity.WeaponClass) > maxWorldKindBytes ||
		len(entity.Components) > maxWorldComponentsBytes ||
		len(entity.DeathSound) > maxWorldAssetBytes ||
		entity.OverlayHeight < 0 || entity.OverlayHeight > 4 ||
		math.IsNaN(entity.Position.X) || math.IsNaN(entity.Position.Y) ||
		math.IsInf(entity.Position.X, 0) || math.IsInf(entity.Position.Y, 0) ||
		entity.Radius < 0 || math.IsNaN(entity.Radius) || math.IsInf(entity.Radius, 0) {
		return ErrWorldView
	}

	if entity.Kind == "corpse" && (entity.SpawnID == "" || entity.Token == "" || entity.Mode != "DT") {
		return ErrWorldView
	}

	return nil
}
