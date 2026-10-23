package clientapp

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestPredictionRetainsAndSettlesPendingTargets(t *testing.T) {
	catalog, err := movement.LoadCatalog(predictionMovementRecords{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name       string
		step       time.Duration
		moment     networkclock.Moment
		staminaRaw int64
		target     float64
		want       float64
	}{
		{"retained", 40 * time.Millisecond, networkclock.Moment{Tick: 14, Fraction: .5}, 84 * 256, 11.4, 11.4},
		{"larger_step", 80 * time.Millisecond, networkclock.Moment{Tick: 13}, 84 * 256, 10.3, 10.3},
		{"fraction", 40 * time.Millisecond, networkclock.Moment{Tick: 10, Fraction: .5}, 84 * 256, 10.3, 10.15},
		{"exhaustion", 40 * time.Millisecond, networkclock.Moment{Tick: 13}, 1, 10.25, 10.25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(movement.MovePayload{Running: true, Target: &movement.MoveTarget{X: tc.target, Y: 20}})
			if err != nil {
				t.Fatal(err)
			}

			hud := playeradapter.HUD{
				Tick: 10, Position: playeradapter.HUDPosition{X: 10, Y: 20},
				Player:   playeradapter.HUDIdentity{Class: "Amazon"},
				Vitals:   playeradapter.HUDVitals{StaminaRaw: tc.staminaRaw, MaxStaminaRaw: 84 * 256},
				Location: playeradapter.HUDLocation{LevelID: 2},
				Movement: playeradapter.HUDMovement{
					Bounds: playeradapter.HUDPosition{X: 100, Y: 100}, Radius: 1, RunDrain: 20,
				},
			}

			got := predictPosition(hud, []gameserver.CommandIntent{
				{TargetTick: 11, Sequence: 1, Kind: movement.MoveCommand, Payload: payload},
			}, tc.moment, nil, tc.step, catalog)
			if math.Abs(got.X-tc.want) > 1e-12 || got.Y != 20 {
				t.Fatalf("predicted position = %+v, want x=%v y=20", got, tc.want)
			}
		})
	}
}

func TestPredictionTownStaminaUsesRecoveredLevels(t *testing.T) {
	catalog, err := movement.LoadCatalog(predictionMovementRecords{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		levelID int64
		wantRaw int64
	}{
		{levelID: 40, wantRaw: 2602},
		{levelID: 45, wantRaw: 2520},
	} {
		hud := playeradapter.HUD{
			Tick:     10,
			Player:   playeradapter.HUDIdentity{Class: "Amazon"},
			Position: playeradapter.HUDPosition{X: 10, Y: 20},
			Vitals:   playeradapter.HUDVitals{StaminaRaw: 2560, MaxStaminaRaw: 84 * 256},
			Location: playeradapter.HUDLocation{LevelID: tc.levelID},
			Movement: playeradapter.HUDMovement{
				Running: true, Velocity: playeradapter.HUDPosition{X: 1},
				Bounds: playeradapter.HUDPosition{X: 100, Y: 100}, Radius: 1, RunDrain: 20,
			},
		}
		prediction := newPositionPrediction(hud, nil, nil, 40*time.Millisecond, catalog)
		prediction.advanceTick(11)

		if prediction.staminaRaw != tc.wantRaw {
			t.Errorf("level %d predicted stamina = %d raw, want %d", tc.levelID, prediction.staminaRaw, tc.wantRaw)
		}
	}
}
