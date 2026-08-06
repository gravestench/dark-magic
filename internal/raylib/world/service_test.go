package gameScene

import (
	"encoding/json"
	"image"
	"image/color"
	"math"
	"os"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	input "github.com/gravestench/dark-magic/internal/raylib/input"
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

func TestSplitMapImagePreservesDimensionsAndPixels(t *testing.T) {
	source := image.NewRGBA(image.Rect(10, 20, 15, 24))
	source.Set(14, 23, color.RGBA{R: 12, G: 34, B: 56, A: 255})
	chunks := splitMapImage(source, 3)
	if len(chunks) != 4 {
		t.Fatalf("chunk count = %d, want 4", len(chunks))
	}
	last := chunks[3]
	if last.bounds != image.Rect(3, 3, 5, 4) {
		t.Fatalf("last bounds = %v", last.bounds)
	}
	if got := color.RGBAModel.Convert(last.image.At(1, 0)).(color.RGBA); got != (color.RGBA{R: 12, G: 34, B: 56, A: 255}) {
		t.Fatalf("last pixel = %v", got)
	}
}

func TestFloatBoundsIntersection(t *testing.T) {
	viewport := floatBounds{minX: 100, minY: 100, maxX: 200, maxY: 200}
	if !viewport.intersects(image.Rect(150, 150, 250, 250)) {
		t.Fatal("overlapping chunk was culled")
	}
	if viewport.intersects(image.Rect(200, 100, 300, 200)) {
		t.Fatal("edge-adjacent chunk should be culled")
	}
}
