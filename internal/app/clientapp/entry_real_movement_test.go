package clientapp

import (
	"math"
	"testing"
)

// TestOwnedPointMovementUsesRetailPace connects pinned class data, ordinary
// pointer intent, Lua motion, and collision integration at the authority tick.
func TestOwnedPointMovementUsesRetailPace(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	createSelectedAcceptanceCharacter(t, fixture)
	fixture.advanceOffline(t, 10)

	controls := requireRealStore(t, fixture, "d2legacy.world.player_control", "missing admitted player")
	positions := requireRealStore(t, fixture, "d2legacy.world.position", "missing authoritative positions")
	if controls.Len() != 1 {
		t.Fatalf("controlled players = %d, want 1", controls.Len())
	}
	player := controls.Entities()[0]
	control := fixture.app.playerControl

	for _, trial := range []struct {
		name    string
		running bool
		step    float64
	}{
		{name: "walk", step: 24576.0 / 65536},
		{name: "run", running: true, step: 36864.0 / 65536},
	} {
		t.Run(trial.name, func(t *testing.T) {
			x, y := dynamicPosition(positions, player)
			goalX, goalY := acceptanceMovementTarget(fixture.app.gameWorlds[1], x, y)
			if math.Hypot(goalX-x, goalY-y) < 6 {
				t.Fatal("generated town has no reachable movement target")
			}
			control.SetRunning(trial.running)
			if err := control.SetMoveTarget(goalX, goalY); err != nil {
				t.Fatal(err)
			}

			fullSteps := 0
			for tick := 0; tick < 250 && control.HasMoveTarget(); tick++ {
				fixture.advanceOffline(t, 1)
				nextX, nextY := dynamicPosition(positions, player)
				step := math.Hypot(nextX-x, nextY-y)
				t.Logf("tick=%d position=(%.9f,%.9f) displacement=%.9f", tick, nextX, nextY, step)
				if step > trial.step+1e-8 {
					t.Fatalf("authority displacement = %.9f, exceeds retail step %.9f", step, trial.step)
				}
				if math.Abs(step-trial.step) < 1e-8 {
					fullSteps++
				}
				x, y = nextX, nextY
			}
			if control.HasMoveTarget() || math.Hypot(goalX-x, goalY-y) > 0.3 {
				t.Fatalf("movement did not arrive: position=(%f,%f), target=(%f,%f)", x, y, goalX, goalY)
			}
			if fullSteps < 2 {
				t.Fatalf("full retail-sized authority steps = %d, want at least 2", fullSteps)
			}
		})
	}
}
