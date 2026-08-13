package player

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const ClientViewVersion uint32 = 1

// ClientView is the complete initial/correction projection envelope. Owner
// private and nearby public schemas remain independently versioned within it.
type ClientView struct {
	Version uint32    `json:"version"`
	Tick    uint64    `json:"tick"`
	HUD     HUD       `json:"hud"`
	World   WorldView `json:"world"`
}

func ProjectClientView(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	hudPayload, err := ProjectHUD(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	worldPayload, err := ProjectWorldView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}
	var hud HUD
	var world WorldView
	if err := json.Unmarshal(hudPayload, &hud); err != nil {
		return nil, fmt.Errorf("client view: HUD: %w", err)
	}
	if err := json.Unmarshal(worldPayload, &world); err != nil {
		return nil, fmt.Errorf("client view: world: %w", err)
	}
	return json.Marshal(ClientView{Version: ClientViewVersion, Tick: checkpoint.Tick, HUD: hud, World: world})
}
