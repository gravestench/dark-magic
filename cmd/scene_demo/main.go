package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/pkg/assetinspect"
	"github.com/gravestench/dark-magic/pkg/scene"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
)

const (
	windowWidth  = 800
	windowHeight = 600
	heroSpeed    = 220.0
	gridSize     = 80
)

func main() {
	seed := flag.Uint64("seed", 1, "deterministic scene seed")
	savePath := flag.String("save", "dark-magic-scene.json", "scene save path")
	sourcePath := flag.String("source", "", "optional directory or MPQ containing a DS1 map")
	mapPath := flag.String("map", "", "optional DS1 path inside source")
	dt1Paths := flag.String("dt1", "", "comma-separated DT1 paths used to texture the DS1 map")
	palettePath := flag.String("palette", "", "optional PL2 palette for DT1 tiles")
	flag.Parse()

	var mapPNG []byte
	if *sourcePath != "" || *mapPath != "" {
		if *sourcePath == "" || *mapPath == "" {
			fatal("both -source and -map are required")
		}
		filesystem, err := fileLoader.NewSource(*sourcePath).Filesystem()
		if err != nil {
			fatal(err.Error())
		}
		if *dt1Paths == "" {
			mapPNG, err = assetinspect.DS1Preview(filesystem, *mapPath)
		} else {
			mapPNG, err = assetinspect.TexturedDS1Preview(filesystem, *mapPath, strings.Split(*dt1Paths, ","), *palettePath)
		}
		if err != nil {
			fatal(err.Error())
		}
	}

	rl.InitWindow(windowWidth, windowHeight, "Dark Magic — Interactive Scene")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	state := scene.New(*seed, 1600, 1200)
	var mapTexture rl.Texture2D
	if len(mapPNG) > 0 {
		mapImage := rl.LoadImageFromMemory(".png", mapPNG, int32(len(mapPNG)))
		mapTexture = rl.LoadTextureFromImage(mapImage)
		rl.UnloadImage(mapImage)
		defer rl.UnloadTexture(mapTexture)
		state = scene.New(*seed, float64(mapTexture.Width), float64(mapTexture.Height))
	}

	for !rl.WindowShouldClose() {
		distance := heroSpeed * float64(rl.GetFrameTime())
		dx, dy := 0.0, 0.0
		if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
			dx -= distance
		}
		if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
			dx += distance
		}
		if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
			dy -= distance
		}
		if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
			dy += distance
		}
		state.MoveHero(dx, dy)

		if rl.IsKeyPressed(rl.KeyF5) {
			_ = save(state, *savePath)
		}
		if rl.IsKeyPressed(rl.KeyF9) {
			if loaded, err := load(*savePath); err == nil {
				state = loaded
			}
		}

		camera := rl.Camera2D{
			Target: rl.Vector2{X: float32(state.Camera.X), Y: float32(state.Camera.Y)},
			Offset: rl.Vector2{X: windowWidth / 2, Y: windowHeight / 2},
			Zoom:   1,
		}
		rl.BeginDrawing()
		rl.ClearBackground(rl.Color{R: 18, G: 18, B: 22, A: 255})
		rl.BeginMode2D(camera)
		if mapTexture.ID != 0 {
			rl.DrawTexture(mapTexture, 0, 0, rl.White)
		} else {
			drawGrid(state)
		}
		rl.DrawCircle(int32(state.Hero.X), int32(state.Hero.Y), 16, rl.Gold)
		rl.DrawCircleLines(int32(state.Hero.X), int32(state.Hero.Y), 16, rl.Maroon)
		rl.EndMode2D()
		rl.DrawRectangle(0, 0, windowWidth, 54, rl.Fade(rl.Black, 0.75))
		rl.DrawText("Move: WASD / arrows    Save: F5    Load: F9    Exit: Esc", 18, 10, 18, rl.RayWhite)
		mapLabel := "generated grid"
		if *mapPath != "" {
			mapLabel = *mapPath
		}
		rl.DrawText(fmt.Sprintf("Seed %d  Position %.0f, %.0f  Map %s", state.Seed, state.Hero.X, state.Hero.Y, mapLabel), 18, 32, 14, rl.LightGray)
		rl.EndDrawing()
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func drawGrid(state *scene.State) {
	for x := 0; x <= int(state.World.Width); x += gridSize {
		rl.DrawLine(int32(x), 0, int32(x), int32(state.World.Height), rl.DarkGray)
	}
	for y := 0; y <= int(state.World.Height); y += gridSize {
		rl.DrawLine(0, int32(y), int32(state.World.Width), int32(y), rl.DarkGray)
	}
	rl.DrawRectangleLines(0, 0, int32(state.World.Width), int32(state.World.Height), rl.Red)
}

func save(state *scene.State, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return state.Save(file)
}

func load(path string) (*scene.State, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return scene.Load(file)
}
