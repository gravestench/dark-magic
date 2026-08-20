package player

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// ProjectClientView assembles all independently versioned owner projections
// from one checkpoint. Using a single tick prevents mixed-time correction state.
func ProjectClientView(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	hud, err := projectHUDValue(playerID, checkpoint)
	if err != nil {
		return nil, err
	}

	world, err := projectWorldValue(playerID, checkpoint)
	if err != nil {
		return nil, err
	}

	private, err := ProjectPrivateView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}

	party, err := ProjectPartyView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}

	events, err := ProjectEventView(playerID, checkpoint)
	if err != nil {
		return nil, err
	}

	view := ClientView{
		Version: ClientViewVersion,
		Tick:    checkpoint.Tick,
		HUD:     hud,
		World:   world,
		Private: private,
		Party:   party,
		Events:  events,
	}

	return json.Marshal(view)
}

// projectHUDValue decodes the canonical HUD projector's wire result rather than
// duplicating its allowlist, keeping the combined projection on one contract.
func projectHUDValue(playerID string, checkpoint simulation.Checkpoint) (HUD, error) {
	payload, err := ProjectHUD(playerID, checkpoint)
	if err != nil {
		return HUD{}, err
	}

	var hud HUD
	if err := json.Unmarshal(payload, &hud); err != nil {
		return HUD{}, fmt.Errorf("client view: HUD: %w", err)
	}

	return hud, nil
}

// projectWorldValue decodes the public projector's wire result so the combined
// envelope cannot drift from the standalone world-view schema.
func projectWorldValue(playerID string, checkpoint simulation.Checkpoint) (WorldView, error) {
	payload, err := ProjectWorldView(playerID, checkpoint)
	if err != nil {
		return WorldView{}, err
	}

	var world WorldView
	if err := json.Unmarshal(payload, &world); err != nil {
		return WorldView{}, fmt.Errorf("client view: world: %w", err)
	}

	return world, nil
}
