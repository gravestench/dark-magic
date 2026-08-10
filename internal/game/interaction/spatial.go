package interaction

import (
	"fmt"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

// validateRange resolves the player-controlled entity from authoritative ECS
// components. Client payloads contain only a target ID; neither endpoint nor a
// claimed distance is accepted from presentation.
func (authority *Authority) validateRange(engine *gameecs.Engine, owner, targetID string) error {
	if engine == nil {
		return fmt.Errorf("interaction: world is required")
	}
	owner, targetID = strings.TrimSpace(owner), strings.ToLower(strings.TrimSpace(targetID))
	authority.mu.RLock()
	target, found := authority.targets[targetID]
	authority.mu.RUnlock()
	if !found {
		return fmt.Errorf("interaction: unknown target %q", targetID)
	}
	controls, found := akara.GetDynamicStore(engine.World(), "dm.world.player_control")
	if !found {
		return fmt.Errorf("interaction: player control state is unavailable")
	}
	positions, found := akara.GetDynamicStore(engine.World(), "dm.world.position")
	if !found {
		return fmt.Errorf("interaction: world position state is unavailable")
	}
	for _, entity := range controls.Entities() {
		control, present := controls.Get(entity)
		if !present {
			continue
		}
		player, err := control.Get("player")
		if err != nil {
			return err
		}
		if player != owner {
			continue
		}
		position, present := positions.Get(entity)
		if !present {
			return fmt.Errorf("interaction: owner %q has no world position", owner)
		}
		x, err := position.Get("x")
		if err != nil {
			return err
		}
		y, err := position.Get("y")
		if err != nil {
			return err
		}
		worldX, xOK := x.(float64)
		worldY, yOK := y.(float64)
		if !xOK || !yOK {
			return fmt.Errorf("interaction: owner %q has invalid world position", owner)
		}
		dx, dy := worldX-target.X, worldY-target.Y
		if dx*dx+dy*dy > target.Radius*target.Radius {
			return fmt.Errorf("interaction: target %q is out of range", target.ID)
		}
		return nil
	}
	return fmt.Errorf("interaction: owner %q has no controlled world entity", owner)
}

func (authority *Authority) openSpatial(engine *gameecs.Engine, owner, targetID string) error {
	if err := authority.validateRange(engine, owner, targetID); err != nil {
		return err
	}
	return authority.open(owner, targetID)
}
