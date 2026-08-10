package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const AssignSkillsCommand = "player.assign_skills"

type AssignSkillsPayload struct {
	Left  *int64 `json:"left,omitempty"`
	Right *int64 `json:"right,omitempty"`
}

func RegisterSkillAssignments(session *Session) error {
	return session.Register(AssignSkillsCommand, CommandHandler{
		Validate: func(command simulation.Command) error {
			_, err := decodeAssignSkills(command.Payload)
			return err
		},
		Apply: func(engine *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeAssignSkills(command.Payload)
			if err != nil {
				return err
			}
			controls, found := akara.GetDynamicStore(engine.World(), "dm.world.player_control")
			if !found {
				return nil
			}
			assignments, found := akara.GetDynamicStore(engine.World(), "dm.player.skill_assignment")
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
	})
}

func learnedSkillAllows(engine *gameecs.Engine, owner akara.Entity, skillID int64, side string) bool {
	learned, found := akara.GetDynamicStore(engine.World(), "dm.player.learned_skill")
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
	if len(requests) == 0 {
		return nil
	}
	payload := AssignSkillsPayload{}
	if value, found := requests["left"]; found {
		payload.Left = &value
	}
	if value, found := requests["right"]; found {
		payload.Right = &value
	}
	encoded, _ := json.Marshal(payload)
	return []simulation.Command{{Tick: tick, Player: source.player, Authority: simulation.AuthorityPlayer, Sequence: source.controller.nextSequence(), Kind: AssignSkillsCommand, Payload: encoded}}
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
