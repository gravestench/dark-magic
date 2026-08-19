package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/presentation/scene"
)

const gridSize = 80

// drawSceneFrame keeps Raylib's begin/end scopes balanced while presenting world content before the fixed HUD.
func drawSceneFrame(state *scene.State, mapTexture rl.Texture2D, config demoConfig) {
	camera := sceneCamera(state)

	rl.BeginDrawing()
	rl.ClearBackground(rl.Color{R: 18, G: 18, B: 22, A: 255})
	rl.BeginMode2D(camera)
	drawWorld(state, mapTexture)
	rl.EndMode2D()
	drawHUD(state, config)
	rl.EndDrawing()
}

// sceneCamera converts authoritative scene coordinates only at the Raylib boundary, avoiding precision loss in state.
func sceneCamera(state *scene.State) rl.Camera2D {
	return rl.Camera2D{
		Target: rl.Vector2{X: float32(state.Camera.X), Y: float32(state.Camera.Y)},
		Offset: rl.Vector2{X: windowWidth / 2, Y: windowHeight / 2},
		Zoom:   1,
	}
}

// drawWorld uses the decoded map when available and otherwise retains the generated grid fallback around the hero.
func drawWorld(state *scene.State, mapTexture rl.Texture2D) {
	if mapTexture.ID != 0 {
		rl.DrawTexture(mapTexture, 0, 0, rl.White)
	} else {
		drawGrid(state)
	}

	rl.DrawCircle(int32(state.Hero.X), int32(state.Hero.Y), 16, rl.Gold)
	rl.DrawCircleLines(int32(state.Hero.X), int32(state.Hero.Y), 16, rl.Maroon)
}

// drawHUD presents controls and current state after leaving camera space so the overlay remains screen-fixed.
func drawHUD(state *scene.State, config demoConfig) {
	rl.DrawRectangle(0, 0, windowWidth, 54, rl.Fade(rl.Black, 0.75))
	rl.DrawText("Move: WASD / arrows    Save: F5    Load: F9    Exit: Esc", 18, 10, 18, rl.RayWhite)

	mapLabel := config.mapLabel()
	status := fmt.Sprintf(
		"Seed %d  Position %.0f, %.0f  Map %s",
		state.Seed,
		state.Hero.X,
		state.Hero.Y,
		mapLabel,
	)
	rl.DrawText(status, 18, 32, 14, rl.LightGray)
}

// drawGrid renders both axes before the world boundary, keeping the border visible over intersecting grid lines.
func drawGrid(state *scene.State) {
	for x := 0; x <= int(state.World.Width); x += gridSize {
		rl.DrawLine(int32(x), 0, int32(x), int32(state.World.Height), rl.DarkGray)
	}

	for y := 0; y <= int(state.World.Height); y += gridSize {
		rl.DrawLine(0, int32(y), int32(state.World.Width), int32(y), rl.DarkGray)
	}

	rl.DrawRectangleLines(0, 0, int32(state.World.Width), int32(state.World.Height), rl.Red)
}
