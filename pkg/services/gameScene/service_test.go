package gameScene

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/pkg/services/input"
)

func TestMovementVector(t *testing.T) {
	x, y := movementVector(map[int32]input.InputState{
		rl.KeyW: input.StateDown,
		rl.KeyD: input.StatePressed,
	})
	if math.Abs(x-0.7071067811865476) > 0.0001 || math.Abs(y+0.7071067811865476) > 0.0001 {
		t.Fatalf("movement = (%f,%f)", x, y)
	}
}

func TestLoadConfiguredRealMap(t *testing.T) {
	source := os.Getenv("DARK_MAGIC_TEST_MPQ")
	if source == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ to run the real-asset scene test")
	}
	service := &Service{}
	if err := json.Unmarshal(service.DefaultConfigData(), &service.Config); err != nil {
		t.Fatal(err)
	}
	service.Config.Source = source
	img, err := service.loadMapImage()
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() < 1 || img.Bounds().Dy() < 1 {
		t.Fatalf("invalid map bounds %v", img.Bounds())
	}
}

func TestOpposingMovementCancels(t *testing.T) {
	x, y := movementVector(map[int32]input.InputState{
		rl.KeyA: input.StateDown,
		rl.KeyD: input.StateDown,
	})
	if x != 0 || y != 0 {
		t.Fatalf("movement = (%f,%f), want zero", x, y)
	}
}

func TestHUDImageIsOpaqueWhereBackgroundIsDrawn(t *testing.T) {
	img := hudImage("Dark Magic", 120, 30)
	_, _, _, alpha := img.At(0, 0).RGBA()
	if alpha == 0 {
		t.Fatal("HUD background is transparent")
	}
}

func TestHUDRefreshIsBoundedButInitialRenderIsImmediate(t *testing.T) {
	now := time.Unix(100, 0)
	if !hudRefreshDue(time.Time{}, now, true) {
		t.Fatal("initial HUD render should be immediate")
	}
	last := now
	if hudRefreshDue(last, now.Add(hudRefreshInterval-time.Millisecond), false) {
		t.Fatal("HUD refreshed before interval elapsed")
	}
	if !hudRefreshDue(last, now.Add(hudRefreshInterval), false) {
		t.Fatal("HUD did not refresh when interval elapsed")
	}
}
