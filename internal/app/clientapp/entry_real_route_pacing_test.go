package clientapp

import (
	movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	"math"
	"testing"
)

// TestOwnedPointMovementMaintainsConfiguredRoutePace connects configured class data, ordinary
// pointer intent, Lua motion, and collision integration at the authority tick.
func TestOwnedPointMovementMaintainsConfiguredRoutePace(t *testing.T) {
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

	rates, found := fixture.app.movementCatalog.Rates("Amazon")
	if !found {
		t.Fatal("missing configured Amazon rates")
	}

	rates = movement.EffectiveRates(rates, movement.Modifiers{})

	for _, trial := range []struct {
		name    string
		running bool
		step    float64
	}{
		{name: "walk", step: rates.Walk / 25},
		{name: "run", running: true, step: rates.Run / 25},
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

			consecutiveFullSteps, longestFullRun := 0, 0

			for tick := 0; tick < 250 && control.HasMoveTarget(); tick++ {
				fixture.advanceOffline(t, 1)

				nextX, nextY := dynamicPosition(positions, player)
				step := math.Hypot(nextX-x, nextY-y)
				t.Logf("tick=%d position=(%.9f,%.9f) displacement=%.9f", tick, nextX, nextY, step)

				if step > trial.step+1e-8 {
					t.Fatalf("authority displacement = %.9f, exceeds configured step %.9f", step, trial.step)
				}

				if math.Abs(step-trial.step) < 1e-8 {
					consecutiveFullSteps++
					longestFullRun = max(longestFullRun, consecutiveFullSteps)
				} else {
					consecutiveFullSteps = 0
				}

				x, y = nextX, nextY
			}

			if control.HasMoveTarget() || math.Hypot(goalX-x, goalY-y) > 0.3 {
				t.Fatalf("movement did not arrive: position=(%f,%f), target=(%f,%f)", x, y, goalX, goalY)
			}

			if longestFullRun < 6 {
				t.Fatalf("consecutive full-rate steps = %d, want at least 6", longestFullRun)
			}
		})
	}
}
