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
	EventViewVersion      uint32 = 2
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
	ID            uint64            `json:"id"`
	Type          string            `json:"type"`
	Tick          uint64            `json:"tick"`
	Position      HUDPosition       `json:"position"`
	Act           int64             `json:"act"`
	LevelID       int64             `json:"level_id"`
	Direction     int64             `json:"direction"`
	OverlayHeight int64             `json:"overlay_height"`
	Cast          *SemanticCastCue  `json:"cast,omitempty"`
	State         *SemanticStateCue `json:"state,omitempty"`
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
	view := EventView{Version: EventViewVersion, Tick: checkpoint.Tick, FromTick: fromTick, Events: []SemanticEvent{}}
	retain := func(event SemanticEvent) {
		view.Events = append(view.Events, event)
		if len(view.Events) <= MaxEventViewEvents {
			return
		}
		sortSemanticEvents(view.Events)
		view.Events = view.Events[1:]
		view.Truncated = true
	}
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
			event := SemanticEvent{ID: instance.Entity, Type: eventType, Tick: tick, Position: HUDPosition{X: x, Y: y}, Act: act, LevelID: levelID}
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
					Kind: stringField(fields, "kind"), EffectTick: effectTick, Player: stringField(fields, "player"),
					SkillID: intField(fields, "skill_id"), Target: HUDPosition{X: floatField(fields, "target_x"), Y: floatField(fields, "target_y")},
					TargetID: stringField(fields, "target_id"),
				}
			case "state":
				expiresTick, valid := nonnegativeTick(fields, "expires_tick")
				if !valid {
					return ErrEventView
				}
				event.State = &SemanticStateCue{
					Kind: stringField(fields, "kind"), StateID: stringField(fields, "state_id"), SourceID: stringField(fields, "source_id"),
					ExpiresTick: expiresTick, Reason: stringField(fields, "reason"),
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
	sortSemanticEvents(view.Events)
	if err := validateEventView(view, checkpoint.Tick); err != nil {
		return EventView{}, err
	}
	return view, nil
}

func sortSemanticEvents(events []SemanticEvent) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].Tick != events[j].Tick {
			return events[i].Tick < events[j].Tick
		}
		return events[i].ID < events[j].ID
	})
}

func eventInstanceFields(component gameecs.ComponentSnapshot, instance gameecs.InstanceSnapshot) (map[string]gameecs.ValueSnapshot, bool) {
	if len(instance.Values) != len(component.Fields) {
		return nil, false
	}
	fields := make(map[string]gameecs.ValueSnapshot, len(component.Fields))
	for index, field := range component.Fields {
		fields[field.Name] = instance.Values[index]
	}
	return fields, true
}

func indexEventComponent(component gameecs.ComponentSnapshot) (map[uint64]map[string]gameecs.ValueSnapshot, bool) {
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

func validateEventView(view EventView, tick uint64) error {
	if view.Version != EventViewVersion || view.Tick != tick || view.Tick > math.MaxInt64 || view.FromTick != eventViewFromTick(view.Tick) ||
		len(view.Events) > MaxEventViewEvents || view.Truncated && len(view.Events) != MaxEventViewEvents {
		return ErrClientView
	}
	seen := make(map[uint64]struct{}, len(view.Events))
	var previous SemanticEvent
	for index, event := range view.Events {
		if event.ID == 0 || event.Tick < view.FromTick || event.Tick > view.Tick || event.Tick > math.MaxInt64 || !finiteView(event.Position.X, event.Position.Y) ||
			!boundedRequired(event.Type, maxEventKindBytes) || event.Act < 1 || event.Act > maxEventAct ||
			event.LevelID < 0 || event.LevelID > maxEventLevelID || event.Direction < 0 || event.Direction > maxEventDirection ||
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
			if event.Cast == nil || event.State != nil || !validCastCue(*event.Cast) {
				return ErrClientView
			}
		case "state":
			if event.State == nil || event.Cast != nil || !validStateCue(*event.State) {
				return ErrClientView
			}
		default:
			return ErrClientView
		}
	}
	return nil
}

func eventInstanceByEntity(component gameecs.ComponentSnapshot, entity uint64) (map[string]gameecs.ValueSnapshot, bool) {
	for _, instance := range component.Instances {
		if instance.Entity != entity {
			continue
		}
		fields, ok := eventInstanceFields(component, instance)
		return fields, ok
	}
	return nil, false
}

func validCastCue(cue SemanticCastCue) bool {
	return (cue.Kind == "cast_started" || cue.Kind == "cast_effect") && cue.EffectTick <= math.MaxInt64 && cue.SkillID >= 0 && cue.SkillID <= maxEventSkillID &&
		boundedRequired(cue.Player, maxViewIdentityBytes) && bounded(cue.TargetID, maxViewIdentityBytes) && finiteView(cue.Target.X, cue.Target.Y)
}

func eventViewFromTick(tick uint64) uint64 {
	if tick < EventViewHistoryTicks {
		return 0
	}
	return tick - EventViewHistoryTicks + 1
}

func validStateCue(cue SemanticStateCue) bool {
	return boundedRequired(cue.Kind, maxEventKindBytes) && boundedRequired(cue.StateID, maxEventSourceBytes) &&
		bounded(cue.SourceID, maxEventSourceBytes) && bounded(cue.Reason, maxEventKindBytes) && cue.ExpiresTick <= math.MaxInt64
}

func nonnegativeTick(fields map[string]gameecs.ValueSnapshot, name string) (uint64, bool) {
	value := intField(fields, name)
	return uint64(max(value, 0)), value >= 0
}
