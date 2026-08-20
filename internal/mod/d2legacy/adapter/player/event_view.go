package player

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	EventViewVersion      uint32 = 4
	EventViewHistoryTicks        = 64
	MaxEventViewEvents           = 256
	maxEventKindBytes            = 32
	maxEventSourceBytes          = 128
	maxEventAct                  = 5
	maxEventLevelID              = 4096
	maxEventDirection            = 15
	maxEventSkillID              = 1<<31 - 1
)

var ErrEventView = errors.New("player event view: invalid semantic projection")

// EventView is a bounded reliable tail of presentation-safe semantic facts.
// FromTick lets a client detect that a correction gap exceeded the tail rather
// than silently dropping spell or state presentation.
type EventView struct {
	Version   uint32          `json:"version"`
	Tick      uint64          `json:"tick"`
	FromTick  uint64          `json:"from_tick"`
	Events    []SemanticEvent `json:"events"`
	Truncated bool            `json:"truncated"`
}

type SemanticEvent struct {
	ID            uint64                   `json:"id"`
	Type          string                   `json:"type"`
	Tick          uint64                   `json:"tick"`
	Position      HUDPosition              `json:"position"`
	Act           int64                    `json:"act"`
	LevelID       int64                    `json:"level_id"`
	Direction     int64                    `json:"direction"`
	OverlayHeight int64                    `json:"overlay_height"`
	Cast          *SemanticCastCue         `json:"cast,omitempty"`
	State         *SemanticStateCue        `json:"state,omitempty"`
	Effect        *SemanticEffectCue       `json:"effect,omitempty"`
	MonsterDeath  *SemanticMonsterDeathCue `json:"monster_death,omitempty"`
}

type SemanticCastCue struct {
	Kind       string      `json:"kind"`
	EffectTick uint64      `json:"effect_tick"`
	Player     string      `json:"player"`
	SkillID    int64       `json:"skill_id"`
	Target     HUDPosition `json:"target"`
	TargetID   string      `json:"target_id,omitempty"`
}

type SemanticStateCue struct {
	Kind        string `json:"kind"`
	StateID     string `json:"state_id"`
	SourceID    string `json:"source_id,omitempty"`
	ExpiresTick uint64 `json:"expires_tick"`
	Reason      string `json:"reason,omitempty"`
}

type SemanticEffectCue struct {
	Kind      string `json:"kind"`
	OverlayID string `json:"overlay_id,omitempty"`
	Sound     string `json:"sound,omitempty"`
}

// SemanticMonsterDeathCue is the presentation-safe subset of the durable
// death event. XP, loot, player-count policy, and attribution never cross the
// client boundary.
type SemanticMonsterDeathCue struct {
	Kind      string `json:"kind"`
	MonsterID string `json:"monster_id"`
}

// ProjectEventView derives a bounded, nearby semantic history from one canonical
// checkpoint. It never exposes authoritative damage, loot, or AI decisions.
func ProjectEventView(playerID string, checkpoint simulation.Checkpoint) (EventView, error) {
	if checkpoint.Snapshot == nil || strings.TrimSpace(playerID) == "" {
		return EventView{}, ErrEventView
	}

	snapshot := *checkpoint.Snapshot

	identities, found := findComponent(snapshot, "d2legacy.player.identity")
	if !found {
		return EventView{}, ErrEventView
	}

	playerEntity, _, found := findString(identities, "player", playerID)
	if !found {
		return EventView{}, ErrEventView
	}

	positions, positioned := findComponent(snapshot, "d2legacy.world.position")
	locations, located := findComponent(snapshot, "d2legacy.world.location")
	facings, _ := findComponent(snapshot, "d2legacy.world.facing")
	monsterAppearances, _ := findComponent(snapshot, "d2legacy.monster.appearance")
	monsterIdentities, _ := findComponent(snapshot, "d2legacy.monster.identity")

	if !positioned || !located {
		return EventView{}, ErrEventView
	}

	positionByEntity, ok := indexEventComponent(positions)
	if !ok {
		return EventView{}, ErrEventView
	}

	locationByEntity, ok := indexEventComponent(locations)
	if !ok {
		return EventView{}, ErrEventView
	}

	facingByEntity, ok := indexEventComponent(facings)
	if !ok {
		return EventView{}, ErrEventView
	}

	monsterAppearanceByEntity, ok := indexEventComponent(monsterAppearances)
	if !ok {
		return EventView{}, ErrEventView
	}

	origin, found := positionByEntity[playerEntity]
	if !found {
		return EventView{}, ErrEventView
	}

	originLocation, found := locationByEntity[playerEntity]
	if !found {
		return EventView{}, ErrEventView
	}

	originX, originY := floatField(origin, "x"), floatField(origin, "y")
	originAct, originLevel := intField(originLocation, "act"), intField(originLocation, "level_id")
	fromTick := eventViewFromTick(checkpoint.Tick)
	view := EventView{
		Version:  EventViewVersion,
		Tick:     checkpoint.Tick,
		FromTick: fromTick,
		Events:   []SemanticEvent{},
	}
	// All semantic component kinds use the same proximity, time-window, and
	// presentation decoration policy, preventing subtle schema-specific leaks.
	retain := func(event SemanticEvent) {
		retainSemanticEvent(&view, event)
	}
	// appendEvent translates the three generic semantic component shapes through
	// one allowlist. Monster deaths use a separate identity join below.
	appendEvent := func(componentName, eventType, actorField string) error {
		component, exists := findComponent(snapshot, componentName)
		if !exists {
			return nil
		}

		for _, instance := range component.Instances {
			fields, ok := eventInstanceFields(component, instance)
			if !ok {
				return ErrEventView
			}

			tick, ok := nonnegativeTick(fields, "tick")
			if !ok || tick > checkpoint.Tick {
				return ErrEventView
			}

			if tick < fromTick {
				continue
			}

			actor := entityField(fields, actorField)
			position, hasPosition := positionByEntity[actor]

			location, hasLocation := locationByEntity[actor]
			if actor == 0 || !hasPosition || !hasLocation {
				continue
			}

			x, y := floatField(position, "x"), floatField(position, "y")

			act, levelID := intField(location, "act"), intField(location, "level_id")
			if act != originAct || levelID != originLevel || math.Hypot(x-originX, y-originY) > WorldViewRadius {
				continue
			}

			event := SemanticEvent{
				ID: instance.Entity, Type: eventType, Tick: tick,
				Position: HUDPosition{X: x, Y: y}, Act: act, LevelID: levelID,
			}
			if facing, exists := facingByEntity[actor]; exists {
				event.Direction = intField(facing, "direction")
			}

			if appearance, exists := monsterAppearanceByEntity[actor]; exists {
				event.OverlayHeight = intField(appearance, "overlay_height")
			} else if _, exists := eventInstanceByEntity(identities, actor); exists {
				event.OverlayHeight = 2
			}

			switch eventType {
			case "cast":
				effectTick, valid := nonnegativeTick(fields, "effect_tick")
				if !valid {
					return ErrEventView
				}

				event.Cast = &SemanticCastCue{
					Kind:       stringField(fields, "kind"),
					EffectTick: effectTick,
					Player:     stringField(fields, "player"),
					SkillID:    intField(fields, "skill_id"),
					Target: HUDPosition{
						X: floatField(fields, "target_x"),
						Y: floatField(fields, "target_y"),
					},
					TargetID: stringField(fields, "target_id"),
				}
			case "state":
				expiresTick, valid := nonnegativeTick(fields, "expires_tick")
				if !valid {
					return ErrEventView
				}

				event.State = &SemanticStateCue{
					Kind:        stringField(fields, "kind"),
					StateID:     stringField(fields, "state_id"),
					SourceID:    stringField(fields, "source_id"),
					ExpiresTick: expiresTick,
					Reason:      stringField(fields, "reason"),
				}
			case "effect":
				event.Effect = &SemanticEffectCue{
					Kind:      stringField(fields, "kind"),
					OverlayID: stringField(fields, "overlay_id"),
					Sound:     stringField(fields, "sound"),
				}
			}

			retain(event)
		}

		return nil
	}
	if err := appendEvent("d2legacy.skill.cast_cue", "cast", "caster"); err != nil {
		return EventView{}, fmt.Errorf("%w: cast cue", err)
	}

	if err := appendEvent("d2legacy.state.event", "state", "target"); err != nil {
		return EventView{}, fmt.Errorf("%w: state cue", err)
	}

	if err := appendEvent("d2legacy.presentation.effect_cue", "effect", "target"); err != nil {
		return EventView{}, fmt.Errorf("%w: presentation effect cue", err)
	}

	monsterBySpawnID := make(map[string]uint64, len(monsterIdentities.Instances))
	for _, instance := range monsterIdentities.Instances {
		fields, ok := eventInstanceFields(monsterIdentities, instance)
		if !ok {
			return EventView{}, ErrEventView
		}

		spawnID := stringField(fields, "spawn_id")
		if spawnID == "" {
			continue
		}

		if _, duplicate := monsterBySpawnID[spawnID]; duplicate {
			return EventView{}, ErrEventView
		}

		monsterBySpawnID[spawnID] = instance.Entity
	}

	if deaths, exists := findComponent(snapshot, "d2legacy.monster.death_event"); exists {
		for _, instance := range deaths.Instances {
			fields, ok := eventInstanceFields(deaths, instance)
			if !ok {
				return EventView{}, ErrEventView
			}

			if stringField(fields, "kind") != "monster_death_presented" {
				continue
			}

			tick, valid := nonnegativeTick(fields, "tick")
			if !valid || tick > checkpoint.Tick {
				return EventView{}, ErrEventView
			}

			if tick < fromTick {
				continue
			}

			monsterID := stringField(fields, "monster_id")
			actor := monsterBySpawnID[monsterID]
			position, hasPosition := positionByEntity[actor]

			location, hasLocation := locationByEntity[actor]
			if actor == 0 || !hasPosition || !hasLocation {
				continue
			}

			x, y := floatField(position, "x"), floatField(position, "y")

			act, levelID := intField(location, "act"), intField(location, "level_id")
			if act != originAct || levelID != originLevel || math.Hypot(x-originX, y-originY) > WorldViewRadius {
				continue
			}

			event := SemanticEvent{
				ID: instance.Entity, Type: "monster_death", Tick: tick,
				Position: HUDPosition{X: x, Y: y}, Act: act, LevelID: levelID,
				MonsterDeath: &SemanticMonsterDeathCue{Kind: "monster_death_presented", MonsterID: monsterID},
			}
			if facing, found := facingByEntity[actor]; found {
				event.Direction = intField(facing, "direction")
			}

			if appearance, found := monsterAppearanceByEntity[actor]; found {
				event.OverlayHeight = intField(appearance, "overlay_height")
			}

			retain(event)
		}
	}

	sortSemanticEvents(view.Events)

	if err := validateEventView(view, checkpoint.Tick); err != nil {
		return EventView{}, err
	}

	return view, nil
}

// retainSemanticEvent keeps the newest bounded tail independent of component
// iteration order. Truncation is explicit so clients can detect lost history.
func retainSemanticEvent(view *EventView, event SemanticEvent) {
	view.Events = append(view.Events, event)
	if len(view.Events) <= MaxEventViewEvents {
		return
	}

	sortSemanticEvents(view.Events)
	view.Events = view.Events[1:]
	view.Truncated = true
}

// sortSemanticEvents provides deterministic replay order, using stable event
// identity to break ties within a simulation tick.
func sortSemanticEvents(events []SemanticEvent) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].Tick != events[j].Tick {
			return events[i].Tick < events[j].Tick
		}

		return events[i].ID < events[j].ID
	})
}

// eventInstanceFields maps positional values onto a component schema. A shape
// mismatch invalidates the projection rather than shifting semantic meanings.
func eventInstanceFields(
	component gameecs.ComponentSnapshot,
	instance gameecs.InstanceSnapshot,
) (map[string]gameecs.ValueSnapshot, bool) {
	if len(instance.Values) != len(component.Fields) {
		return nil, false
	}

	fields := make(map[string]gameecs.ValueSnapshot, len(component.Fields))
	for index, field := range component.Fields {
		fields[field.Name] = instance.Values[index]
	}

	return fields, true
}

// indexEventComponent builds a unique entity lookup for repeated event joins;
// duplicate ECS rows fail closed because their selection would be ambiguous.
func indexEventComponent(
	component gameecs.ComponentSnapshot,
) (map[uint64]map[string]gameecs.ValueSnapshot, bool) {
	result := make(map[uint64]map[string]gameecs.ValueSnapshot, len(component.Instances))
	for _, instance := range component.Instances {
		fields, ok := eventInstanceFields(component, instance)
		if !ok {
			return nil, false
		}

		if _, duplicate := result[instance.Entity]; duplicate {
			return nil, false
		}

		result[instance.Entity] = fields
	}

	return result, true
}

// validateEventView enforces the bounded ordered tagged-union protocol before a
// projection is accepted by client presentation.
func validateEventView(view EventView, tick uint64) error {
	if view.Version != EventViewVersion || view.Tick != tick || view.Tick > math.MaxInt64 ||
		view.FromTick != eventViewFromTick(view.Tick) ||
		len(view.Events) > MaxEventViewEvents || view.Truncated && len(view.Events) != MaxEventViewEvents {
		return ErrClientView
	}

	seen := make(map[uint64]struct{}, len(view.Events))

	var previous SemanticEvent

	for index, event := range view.Events {
		if event.ID == 0 || event.Tick < view.FromTick || event.Tick > view.Tick ||
			event.Tick > math.MaxInt64 || !finiteView(event.Position.X, event.Position.Y) ||
			!boundedRequired(event.Type, maxEventKindBytes) || event.Act < 1 || event.Act > maxEventAct ||
			event.LevelID < 0 || event.LevelID > maxEventLevelID ||
			event.Direction < 0 || event.Direction > maxEventDirection ||
			event.OverlayHeight < 0 || event.OverlayHeight > 4 {
			return ErrClientView
		}

		if _, duplicate := seen[event.ID]; duplicate {
			return ErrClientView
		}

		seen[event.ID] = struct{}{}
		if index > 0 && (event.Tick < previous.Tick || event.Tick == previous.Tick && event.ID <= previous.ID) {
			return ErrClientView
		}

		previous = event
		switch event.Type {
		case "cast":
			if event.Cast == nil || event.State != nil || event.Effect != nil ||
				event.MonsterDeath != nil || !validCastCue(*event.Cast) {
				return ErrClientView
			}
		case "state":
			if event.State == nil || event.Cast != nil || event.Effect != nil ||
				event.MonsterDeath != nil || !validStateCue(*event.State) {
				return ErrClientView
			}
		case "effect":
			if event.Effect == nil || event.Cast != nil || event.State != nil ||
				event.MonsterDeath != nil || !validEffectCue(*event.Effect) {
				return ErrClientView
			}
		case "monster_death":
			if event.MonsterDeath == nil || event.Cast != nil || event.State != nil ||
				event.Effect != nil || !validMonsterDeathCue(*event.MonsterDeath) {
				return ErrClientView
			}
		default:
			return ErrClientView
		}
	}

	return nil
}

// eventInstanceByEntity performs a schema-checked lookup where building a full
// index would not reduce the small number of probes.
func eventInstanceByEntity(
	component gameecs.ComponentSnapshot,
	entity uint64,
) (map[string]gameecs.ValueSnapshot, bool) {
	for _, instance := range component.Instances {
		if instance.Entity != entity {
			continue
		}

		fields, ok := eventInstanceFields(component, instance)

		return fields, ok
	}

	return nil, false
}

// validCastCue accepts only known cast phases and bounded targeting values,
// leaving skill effects and authority decisions out of presentation.
func validCastCue(cue SemanticCastCue) bool {
	return (cue.Kind == "cast_started" || cue.Kind == "cast_effect") &&
		cue.EffectTick <= math.MaxInt64 && cue.SkillID >= 0 && cue.SkillID <= maxEventSkillID &&
		boundedRequired(cue.Player, maxViewIdentityBytes) &&
		bounded(cue.TargetID, maxViewIdentityBytes) && finiteView(cue.Target.X, cue.Target.Y)
}

// eventViewFromTick computes the inclusive history boundary without unsigned
// underflow during the first simulation ticks.
func eventViewFromTick(tick uint64) uint64 {
	if tick < EventViewHistoryTicks {
		return 0
	}

	return tick - EventViewHistoryTicks + 1
}

// validStateCue bounds the state vocabulary and expiry used by presentation;
// source identity remains optional because environmental effects have none.
func validStateCue(cue SemanticStateCue) bool {
	return boundedRequired(cue.Kind, maxEventKindBytes) &&
		boundedRequired(cue.StateID, maxEventSourceBytes) &&
		bounded(cue.SourceID, maxEventSourceBytes) &&
		bounded(cue.Reason, maxEventKindBytes) && cue.ExpiresTick <= math.MaxInt64
}

// validEffectCue requires at least one visual or audio asset, avoiding empty
// effect events that consume bounded history without presenting anything.
func validEffectCue(cue SemanticEffectCue) bool {
	return boundedRequired(cue.Kind, maxEventKindBytes) && bounded(cue.OverlayID, maxEventSourceBytes) &&
		bounded(cue.Sound, maxEventSourceBytes) && (cue.OverlayID != "" || cue.Sound != "")
}

// validMonsterDeathCue verifies the sole death presentation kind and its stable
// monster identity while excluding reward and attribution fields.
func validMonsterDeathCue(cue SemanticMonsterDeathCue) bool {
	return cue.Kind == "monster_death_presented" && boundedRequired(cue.MonsterID, maxViewIdentityBytes)
}

// nonnegativeTick converts signed ECS storage only after checking its domain,
// preventing negative values from wrapping into large future ticks.
func nonnegativeTick(fields map[string]gameecs.ValueSnapshot, name string) (uint64, bool) {
	value := intField(fields, name)
	return uint64(max(value, 0)), value >= 0
}
