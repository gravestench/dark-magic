package movement

import (
	"math"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func TestResolveUsesProductionMovementRules(t *testing.T) {
	rates := ClassRates{Walk: 6, Run: 9}
	diagonal := Resolve(gameworld.Point{}, MovePayload{X: 1, Y: 1, Running: true}, rates)
	expected := rates.Run / math.Sqrt2
	if math.Abs(diagonal.Velocity.X-expected) > 1e-12 || math.Abs(diagonal.Velocity.Y-expected) > 1e-12 || !diagonal.Moving {
		t.Fatalf("diagonal movement = %+v", diagonal)
	}
	arrived := Resolve(gameworld.Point{X: 10, Y: 20}, MovePayload{Target: &MoveTarget{X: 10.1, Y: 20}}, rates)
	if arrived.Moving || arrived.Velocity != (gameworld.Point{}) {
		t.Fatalf("arrival movement = %+v", arrived)
	}
	target := Resolve(gameworld.Point{X: 10, Y: 20}, MovePayload{Target: &MoveTarget{X: 13, Y: 24}}, rates)
	if math.Abs(target.Velocity.X-3.6) > 1e-12 || math.Abs(target.Velocity.Y-4.8) > 1e-12 {
		t.Fatalf("target velocity = %+v, want {3.6 4.8}", target.Velocity)
	}
}
