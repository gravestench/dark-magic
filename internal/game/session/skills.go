package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

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
