package movement

import (
	"math"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const (
	WalkSpeed       = 10.0
	RunSpeed        = 15.0
	ArrivalDistance = 0.2
	diagonalScale   = 0.7071067811865476
)

type ResolvedMovement struct {
	Velocity gameworld.Point
	Running  bool
	Moving   bool
}

// Resolve applies the production d2legacy input policy without touching ECS.
// Both authoritative Lua and client prediction call this implementation.
func Resolve(position gameworld.Point, payload MovePayload) ResolvedMovement {
	x, y := float64(payload.X), float64(payload.Y)
	if payload.Target != nil {
		x, y = payload.Target.X-position.X, payload.Target.Y-position.Y
		distance := math.Hypot(x, y)
		if distance <= ArrivalDistance {
			x, y = 0, 0
		} else {
			x, y = x/distance, y/distance
		}
	} else if x != 0 && y != 0 {
		x, y = x*diagonalScale, y*diagonalScale
	}
	speed := WalkSpeed
	if payload.Running {
		speed = RunSpeed
	}
	return ResolvedMovement{
		Velocity: gameworld.Point{X: x * speed, Y: y * speed},
		Running:  payload.Running,
		Moving:   x != 0 || y != 0,
	}
}
