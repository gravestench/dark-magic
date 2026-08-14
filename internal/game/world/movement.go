package world

import "math"

// IntegrateVelocity applies one bounded, axis-separated collision step. It is
// shared by authoritative Lua through the world-map capability and by client
// rollback/replay prediction.
func IntegrateVelocity(collision *Map, position, velocity, bounds Point, radius, elapsed float64) Point {
	result := position
	result.X = clampMovement(position.X+velocity.X*elapsed, radius, bounds.X-radius)
	if footprintWalkable(collision, result.X, result.Y, radius) {
		position.X = result.X
	}
	result = position
	result.Y = clampMovement(position.Y+velocity.Y*elapsed, radius, bounds.Y-radius)
	if footprintWalkable(collision, result.X, result.Y, radius) {
		position.Y = result.Y
	}
	return position
}

func footprintWalkable(collision *Map, x, y, radius float64) bool {
	if collision == nil {
		return true
	}
	reach := int(math.Ceil(radius))
	for offsetY := -reach; offsetY <= reach; offsetY++ {
		for offsetX := -reach; offsetX <= reach; offsetX++ {
			if math.Hypot(float64(offsetX), float64(offsetY)) > radius+0.5 {
				continue
			}
			flags, found := collision.FlagsAtPosition(x+float64(offsetX), y+float64(offsetY))
			if found && flags.Blocked() {
				return false
			}
		}
	}
	return true
}

func clampMovement(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(maximum, value))
}
