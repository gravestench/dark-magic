package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/presentation/scene"
)

const (
	windowWidth           = 800
	windowHeight          = 600
	targetFramesPerSecond = 60
	heroSpeed             = 220.0
	defaultWorldWidth     = 1600
	defaultWorldHeight    = 1200
)

// keyDownFunc isolates deterministic movement arithmetic from Raylib without changing its key polling sequence.
type keyDownFunc func(key int32) bool

// startInteractiveScene owns Raylib resources in creation order so deferred cleanup runs texture-before-window.
func startInteractiveScene(config demoConfig, savePath string, mapPNG []byte) {
	rl.InitWindow(windowWidth, windowHeight, "Dark Magic — Interactive Scene")

	defer rl.CloseWindow()

	rl.SetTargetFPS(targetFramesPerSecond)

	state, mapTexture := createSceneResources(config.seed, mapPNG)
	if len(mapPNG) > 0 {
		// Preview presence, rather than texture validity, controls cleanup just as it did in the original runtime.
		defer rl.UnloadTexture(mapTexture)
	}

	runSceneLoop(state, mapTexture, config, savePath)
}

// createSceneResources releases CPU image memory immediately while returning the GPU texture to the window owner.
func createSceneResources(seed uint64, mapPNG []byte) (*scene.State, rl.Texture2D) {
	state := scene.New(seed, defaultWorldWidth, defaultWorldHeight)
	if len(mapPNG) == 0 {
		return state, rl.Texture2D{}
	}

	mapImage := rl.LoadImageFromMemory(".png", mapPNG, int32(len(mapPNG)))
	mapTexture := rl.LoadTextureFromImage(mapImage)
	rl.UnloadImage(mapImage)

	state = scene.New(seed, float64(mapTexture.Width), float64(mapTexture.Height))

	return state, mapTexture
}

// runSceneLoop preserves movement, save, load, and drawing order within every frame so loaded state renders
// immediately.
func runSceneLoop(state *scene.State, mapTexture rl.Texture2D, config demoConfig, savePath string) {
	for !rl.WindowShouldClose() {
		moveHeroFromKeyboard(state)
		state = applyPersistenceShortcuts(state, savePath)
		drawSceneFrame(state, mapTexture, config)
	}
}

// moveHeroFromKeyboard samples frame time before key state, preserving the distance applied to every direction in a
// frame.
func moveHeroFromKeyboard(state *scene.State) {
	distance := heroSpeed * float64(rl.GetFrameTime())
	dx, dy := heroMovement(distance, rl.IsKeyDown)

	state.MoveHero(dx, dy)
}

// heroMovement queries opposing keys in the original order so simultaneous input cancels deterministically on each
// axis.
func heroMovement(distance float64, isKeyDown keyDownFunc) (float64, float64) {
	dx, dy := 0.0, 0.0
	if isKeyDown(rl.KeyA) || isKeyDown(rl.KeyLeft) {
		dx -= distance
	}

	if isKeyDown(rl.KeyD) || isKeyDown(rl.KeyRight) {
		dx += distance
	}

	if isKeyDown(rl.KeyW) || isKeyDown(rl.KeyUp) {
		dy -= distance
	}

	if isKeyDown(rl.KeyS) || isKeyDown(rl.KeyDown) {
		dy += distance
	}

	return dx, dy
}

// applyPersistenceShortcuts saves before loading and ignores interactive I/O failures, preserving uninterrupted play.
func applyPersistenceShortcuts(state *scene.State, savePath string) *scene.State {
	if rl.IsKeyPressed(rl.KeyF5) {
		_ = saveScene(state, savePath)
	}

	if !rl.IsKeyPressed(rl.KeyF9) {
		return state
	}

	loaded, err := loadScene(savePath)
	if err != nil {
		return state
	}

	return loaded
}
