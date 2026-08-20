package movement

import (
	"math"

	"github.com/gravestench/akara"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// movementPathFinder isolates collision-aware route search from movement command generation.
type movementPathFinder interface {
	FindPath(gameworld.PathRequest) ([]gameworld.Point, error)
}

// SetNavigation binds routes to one map instance and invalidates coordinates when that instance changes.
func (source *MovementSource) SetNavigation(world *gameworld.Map) {
	source.mu.Lock()
	defer source.mu.Unlock()

	// Reliable corrections may reinstall the exact map pointer. That is not a world transition, so retaining the
	// accepted route prevents every correction from becoming an implicit stop command.
	if current, ok := source.navigation.(*gameworld.Map); ok && current == world {
		return
	}

	source.navigation = world
	source.resetAcceptedPath()
	// Pointer coordinates belong to the previous map. Retaining them could replan a town click in wilderness space.
	source.control.clearMoveTarget()
}

// resetAcceptedPath discards cached route geometry without changing the controller's pending pointer intent.
// Callers hold the source mutex so a command tick cannot observe half-reset route state.
func (source *MovementSource) resetAcceptedPath() {
	source.path = nil
	source.pathTarget = nil
}

// pathWaypoint converts the current pointer target into the next collision-safe waypoint.
func (source *MovementSource) pathWaypoint(target *MoveTarget) *MoveTarget {
	current, radius, found := source.playerPathOrigin()
	if !found {
		return target
	}

	if source.routeTargetChanged(target) {
		source.replaceAcceptedRoute(current, radius, target)
	} else if routeEndpointMoved(source.pathTarget, target) {
		source.updateRouteEndpoint(target)
	}

	source.consumeReachedWaypoints(current)

	if len(source.path) <= 1 {
		source.control.clearMoveTarget()
		source.resetAcceptedPath()

		return stationaryTarget(current)
	}

	return &MoveTarget{X: source.path[1].X, Y: source.path[1].Y}
}

// playerPathOrigin locates the controlled entity and includes its collision radius in every path request.
func (source *MovementSource) playerPathOrigin() (gameworld.Point, float64, bool) {
	positions, positionsPresent := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.position")

	controls, controlsPresent := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.player_control")
	if !positionsPresent || !controlsPresent {
		return gameworld.Point{}, 0, false
	}

	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)

		player, _ := control.Get("player")
		if player != source.player {
			continue
		}

		position, present := positions.Get(entity)
		if !present {
			continue
		}

		x, _ := position.Get("x")
		y, _ := position.Get("y")
		current := gameworld.Point{X: x.(float64), Y: y.(float64)}

		return current, source.entityCollisionRadius(entity), true
	}

	return gameworld.Point{}, 0, false
}

// entityCollisionRadius returns zero when the presentation replica has no collider for the controlled entity.
func (source *MovementSource) entityCollisionRadius(entity akara.Entity) float64 {
	colliders, present := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.collider")
	if !present {
		return 0
	}

	collider, exists := colliders.Get(entity)
	if !exists {
		return 0
	}

	value, _ := collider.Get("radius")

	return value.(float64)
}

// routeTargetChanged requests a new path only when the click enters a new collision cell or changes stop policy.
func (source *MovementSource) routeTargetChanged(target *MoveTarget) bool {
	if source.pathTarget == nil || source.pathTarget.StopRadius != target.StopRadius {
		return true
	}

	return gameworld.CollisionCell(source.pathTarget.X) != gameworld.CollisionCell(target.X) ||
		gameworld.CollisionCell(source.pathTarget.Y) != gameworld.CollisionCell(target.Y)
}

// replaceAcceptedRoute installs a valid replacement or preserves the previous route after a blocked click.
func (source *MovementSource) replaceAcceptedRoute(
	current gameworld.Point,
	radius float64,
	target *MoveTarget,
) {
	path, err := source.navigation.FindPath(gameworld.PathRequest{
		Start:      current,
		Goal:       gameworld.Point{X: target.X, Y: target.Y},
		Radius:     radius,
		StopRadius: target.StopRadius,
	})
	if err != nil {
		source.retainRouteOrStop()
		return
	}

	source.path = path
	copyTarget := *target
	source.pathTarget = &copyTarget
}

// retainRouteOrStop keeps accepted movement alive; only an initial blocked click resolves to a stop.
func (source *MovementSource) retainRouteOrStop() {
	if source.pathTarget != nil && len(source.path) > 1 {
		oldTarget := *source.pathTarget
		source.control.restoreMoveTarget(&oldTarget)

		return
	}

	source.control.clearMoveTarget()
	source.resetAcceptedPath()
}

// routeEndpointMoved recognizes sub-cell pointer drift that does not justify another collision search.
func routeEndpointMoved(previous, target *MoveTarget) bool {
	return previous.X != target.X || previous.Y != target.Y
}

// updateRouteEndpoint follows sub-cell camera drift while preserving the accepted collision-cell route.
func (source *MovementSource) updateRouteEndpoint(target *MoveTarget) {
	copyTarget := *target

	source.pathTarget = &copyTarget
	if len(source.path) > 0 {
		source.path[len(source.path)-1] = gameworld.Point{X: target.X, Y: target.Y}
	}
}

// consumeReachedWaypoints advances only past waypoints inside the established arrival tolerance.
func (source *MovementSource) consumeReachedWaypoints(current gameworld.Point) {
	for len(source.path) > 1 && distance(current, source.path[1]) <= 0.3 {
		source.path = source.path[1:]
	}
}

// distance measures world-space separation without exposing coordinate arithmetic in route orchestration.
func distance(first, second gameworld.Point) float64 {
	return math.Hypot(first.X-second.X, first.Y-second.Y)
}

// stationaryTarget emits the current position so a rejected or completed route cannot create residual velocity.
func stationaryTarget(current gameworld.Point) *MoveTarget {
	return &MoveTarget{X: current.X, Y: current.Y}
}
