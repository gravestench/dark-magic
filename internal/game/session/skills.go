package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/gravestench/akara"
	gameaction "github.com/gravestench/dark-magic/internal/game/action"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	AssignSkillsCommand = "player.assign_skills"
	UseSkillCommand     = "player.use_skill"
)

type AssignSkillsPayload struct {
	Left  *int64 `json:"left,omitempty"`
	Right *int64 `json:"right,omitempty"`
}

type UseSkillPayload struct {
	Side     string  `json:"side"`
	TargetX  float64 `json:"target_x"`
	TargetY  float64 `json:"target_y"`
	TargetID string  `json:"target_id,omitempty"`
}

func RegisterSkillAssignments(session *Session) error {
	if err := session.Register(AssignSkillsCommand, CommandHandler{
		Validate: func(command simulation.Command) error {
			_, err := decodeAssignSkills(command.Payload)
			return err
		},
		Apply: func(engine *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeAssignSkills(command.Payload)
			if err != nil {
				return err
			}
			controls, found := akara.GetDynamicStore(engine.World(), "d2.world.player_control")
			if !found {
				return nil
			}
			assignments, found := akara.GetDynamicStore(engine.World(), "d2.player.skill_assignment")
			if !found {
				return nil
			}
			for _, entity := range controls.Entities() {
				control, _ := controls.Get(entity)
				player, _ := control.Get("player")
				if player != command.Player {
					continue
				}
				assignment, present := assignments.Get(entity)
				if !present {
					continue
				}
				if payload.Left != nil {
					if !learnedSkillAllows(engine, entity, *payload.Left, "left") {
						return fmt.Errorf("skill %d is not learned or cannot be assigned to left", *payload.Left)
					}
					if err := assignment.Set("left", *payload.Left); err != nil {
						return err
					}
				}
				if payload.Right != nil {
					if !learnedSkillAllows(engine, entity, *payload.Right, "right") {
						return fmt.Errorf("skill %d is not learned or cannot be assigned to right", *payload.Right)
					}
					if err := assignment.Set("right", *payload.Right); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}); err != nil {
		return err
	}
	return session.Register(UseSkillCommand, CommandHandler{
		Validate: func(command simulation.Command) error { _, err := decodeUseSkill(command.Payload); return err },
		Apply: func(engine *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeUseSkill(command.Payload)
			if err != nil {
				return err
			}
			controls, found := akara.GetDynamicStore(engine.World(), "d2.world.player_control")
			if !found {
				return nil
			}
			assignments, found := akara.GetDynamicStore(engine.World(), "d2.player.skill_assignment")
			if !found {
				return nil
			}
			intents, found := akara.GetDynamicStore(engine.World(), "d2.player.skill_intent")
			if !found {
				return nil
			}
			for _, entity := range controls.Entities() {
				control, _ := controls.Get(entity)
				player, _ := control.Get("player")
				if player != command.Player {
					continue
				}
				assignment, present := assignments.Get(entity)
				if !present {
					continue
				}
				assigned, err := assignment.Get(payload.Side)
				if err != nil {
					return err
				}
				skillID := assigned.(int64)
				if !learnedSkillAllows(engine, entity, skillID, payload.Side) {
					if skillID == 0 {
						return nil
					}
					return fmt.Errorf("assigned %s skill %d is not learned or allowed", payload.Side, skillID)
				}
				if !gameaction.MatchesExclusive(engine.World(), entity, skillID, payload.TargetID) {
					gameaction.CancelExclusive(engine.World(), entity)
				}
				intent, present := intents.Get(entity)
				if !present {
					continue
				}
				for field, value := range map[string]any{"side": payload.Side, "skill_id": skillID, "target_x": payload.TargetX, "target_y": payload.TargetY, "target_id": payload.TargetID} {
					if err := intent.Set(field, value); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}

func learnedSkillAllows(engine *gameecs.Engine, owner akara.Entity, skillID int64, side string) bool {
	learned, found := akara.GetDynamicStore(engine.World(), "d2.player.learned_skill")
	if !found {
		return false
	}
	for _, entity := range learned.Entities() {
		component, _ := learned.Get(entity)
		candidateOwner, _ := component.Get("owner")
		candidateID, _ := component.Get("skill_id")
		allowed, _ := component.Get(side + "_allowed")
		if candidateOwner == owner && candidateID == skillID && allowed == true {
			return true
		}
	}
	return false
}

type SkillSource struct {
	controller *MovementController
	player     string
}

func NewSkillSource(controller *MovementController, player string) (*SkillSource, error) {
	player = strings.TrimSpace(player)
	if controller == nil || player == "" {
		return nil, fmt.Errorf("game session: skill source requires controller and player")
	}
	return &SkillSource{controller: controller, player: player}, nil
}

func (source *SkillSource) Commands(tick uint64) []simulation.Command {
	requests := source.controller.drainSkills()
	commands := make([]simulation.Command, 0, 1)
	if len(requests) > 0 {
		payload := AssignSkillsPayload{}
		if value, found := requests["left"]; found {
			payload.Left = &value
		}
		if value, found := requests["right"]; found {
			payload.Right = &value
		}
		encoded, _ := json.Marshal(payload)
		commands = append(commands, simulation.Command{Tick: tick, Player: source.player, Authority: simulation.AuthorityPlayer, Sequence: source.controller.nextSequence(), Kind: AssignSkillsCommand, Payload: encoded})
	}
	for _, payload := range source.controller.drainSkillUses() {
		encoded, _ := json.Marshal(payload)
		commands = append(commands, simulation.Command{Tick: tick, Player: source.player, Authority: simulation.AuthorityPlayer, Sequence: source.controller.nextSequence(), Kind: UseSkillCommand, Payload: encoded})
	}
	if len(commands) == 0 {
		return nil
	}
	return commands
}

func decodeUseSkill(encoded []byte) (UseSkillPayload, error) {
	var payload UseSkillPayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return payload, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return payload, fmt.Errorf("skill use payload has trailing data")
	}
	payload.Side, payload.TargetID = strings.ToLower(strings.TrimSpace(payload.Side)), strings.TrimSpace(payload.TargetID)
	if payload.Side != "left" && payload.Side != "right" {
		return payload, fmt.Errorf("skill use side must be left or right")
	}
	if math.IsNaN(payload.TargetX) || math.IsNaN(payload.TargetY) || math.IsInf(payload.TargetX, 0) || math.IsInf(payload.TargetY, 0) {
		return payload, fmt.Errorf("skill target must be finite")
	}
	return payload, nil
}

func decodeAssignSkills(encoded []byte) (AssignSkillsPayload, error) {
	var payload AssignSkillsPayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return AssignSkillsPayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return AssignSkillsPayload{}, fmt.Errorf("skill assignment payload has trailing data")
	}
	if payload.Left == nil && payload.Right == nil {
		return AssignSkillsPayload{}, fmt.Errorf("skill assignment is empty")
	}
	if payload.Left != nil && *payload.Left < 0 || payload.Right != nil && *payload.Right < 0 {
		return AssignSkillsPayload{}, fmt.Errorf("skill IDs must be non-negative")
	}
	return payload, nil
}
